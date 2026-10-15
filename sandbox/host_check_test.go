// Tests the container host check against canned command output: the sandbox network, the
// firewall rules that keep jobs off private ranges, and rootless podman.
package sandbox

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeCommands answers each command line with canned output or an error.
type fakeCommands map[string]string

func (f fakeCommands) run(_ context.Context, name string, args ...string) (string, error) {
	line := strings.Join(append([]string{name}, args...), " ")
	out, ok := f[line]
	if !ok {
		return "", errors.New("exit status 1: no such command in the fake: " + line)
	}
	return out, nil
}

const dockerNetworkInspect = `docker network inspect gavel-sandbox --format {{.Id}} {{index .Options "com.docker.network.bridge.name"}}`

func goodRules(bridge string) string {
	lines := []string{
		"-N DOCKER-USER",
		`-A DOCKER-USER -d 10.0.0.5/32 -i ` + bridge + ` -p tcp -m tcp --dport 9092 -m comment --comment gavel-sandbox -j ACCEPT`,
	}
	for _, cidr := range privateRanges {
		lines = append(lines, `-A DOCKER-USER -d `+cidr+` -i `+bridge+` -m comment --comment gavel-sandbox -j REJECT --reject-with icmp-port-unreachable`)
	}
	return strings.Join(append(lines, "-A DOCKER-USER -j RETURN"), "\n")
}

func TestCheckHostAcceptsAFirewalledNetwork(t *testing.T) {
	commands := fakeCommands{
		dockerNetworkInspect:      "abcdef0123456789 \n",
		"iptables -S DOCKER-USER": goodRules("br-abcdef012345"),
	}
	check := HostCheck{Binary: "docker", Network: "gavel-sandbox", ServeAddr: "10.0.0.5:9092", run: commands.run}
	require.NoError(t, check.Run(context.Background()))

	commands[dockerNetworkInspect] = "abcdef0123456789 gavel0\n"
	commands["iptables -S DOCKER-USER"] = goodRules("gavel0")
	require.NoError(t, check.Run(context.Background()), "a bridge with its own name")
}

func TestCheckHostRefusesWithTheCommandsToRun(t *testing.T) {
	commands := fakeCommands{dockerNetworkInspect: "abcdef0123456789 \n", "iptables -S DOCKER-USER": "-N DOCKER-USER\n-A DOCKER-USER -j RETURN"}
	check := HostCheck{Binary: "docker", Network: "gavel-sandbox", ServeAddr: "10.0.0.5:9092", run: commands.run}
	err := check.Run(context.Background())
	require.Error(t, err)
	message := err.Error()
	assert.Contains(t, message, "the host firewall doesn't keep sandbox jobs on br-abcdef012345 off private networks")
	assert.Contains(t, message, "iptables -I DOCKER-USER -i br-abcdef012345 -d 10.0.0.5 -p tcp --dport 9092 -j ACCEPT -m comment --comment gavel-sandbox")
	assert.Contains(t, message, "iptables -A DOCKER-USER -i br-abcdef012345 -d 169.254.0.0/16 -j REJECT -m comment --comment gavel-sandbox")
}

func TestCheckHostRefusesAnAcceptAfterTheRejects(t *testing.T) {
	rules := strings.Split(goodRules("br-abcdef012345"), "\n")
	accept := rules[1]
	rules = append(append([]string{rules[0]}, rules[2:len(rules)-1]...), accept, rules[len(rules)-1])
	commands := fakeCommands{dockerNetworkInspect: "abcdef0123456789 \n", "iptables -S DOCKER-USER": strings.Join(rules, "\n")}
	check := HostCheck{Binary: "docker", Network: "gavel-sandbox", ServeAddr: "10.0.0.5:9092", run: commands.run}
	err := check.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "the ACCEPT for gavel serve must come before the REJECT rules")
}

func TestCheckHostRefusesAMissingNetwork(t *testing.T) {
	check := HostCheck{Binary: "docker", Network: "gavel-sandbox", ServeAddr: "10.0.0.5:9092", run: fakeCommands{}.run}
	err := check.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "docker network create gavel-sandbox")
}

func TestCheckHostRefusesRootlessPodman(t *testing.T) {
	commands := fakeCommands{"podman info --format {{.Host.Security.Rootless}}": "true\n"}
	check := HostCheck{Binary: "podman", Network: "gavel-sandbox", ServeAddr: "10.0.0.5:9092", run: commands.run}
	err := check.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rootless podman is not supported")
}

func TestCheckHostNeedsAnIPForGavelServe(t *testing.T) {
	check := HostCheck{Binary: "docker", Network: "gavel-sandbox", ServeAddr: ":9092", run: fakeCommands{}.run}
	err := check.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gavel serve's address :9092 needs the IP sandbox jobs reach it at")
}
