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

// goodInputRules is an INPUT chain that keeps jobs on bridge off the host's services but
// gavel serve's, with the host's own rules after gavel's.
func goodInputRules(bridge string) string {
	return strings.Join([]string{
		"-P INPUT ACCEPT",
		`-A INPUT -i ` + bridge + ` -m conntrack --ctstate RELATED,ESTABLISHED -m comment --comment gavel-sandbox -j ACCEPT`,
		`-A INPUT -d 10.0.0.5/32 -i ` + bridge + ` -p tcp -m tcp --dport 9092 -m comment --comment gavel-sandbox -j ACCEPT`,
		`-A INPUT -i ` + bridge + ` -m comment --comment gavel-sandbox -j REJECT --reject-with icmp-port-unreachable`,
		"-A INPUT -i lo -j ACCEPT",
		"-A INPUT -p tcp -m tcp --dport 22 -j ACCEPT",
	}, "\n")
}

// firewalledHost answers the host check's commands for a host whose firewall is set up.
func firewalledHost(bridge string) fakeCommands {
	return fakeCommands{
		dockerNetworkInspect:      "abcdef0123456789 \n",
		"iptables -S DOCKER-USER": goodRules(bridge),
		"iptables -S INPUT":       goodInputRules(bridge),
	}
}

func TestCheckHostAcceptsAFirewalledNetwork(t *testing.T) {
	commands := firewalledHost("br-abcdef012345")
	check := HostCheck{Binary: "docker", Network: "gavel-sandbox", ServeAddr: "10.0.0.5:9092", run: commands.run}
	require.NoError(t, check.Run(context.Background()))

	commands = firewalledHost("gavel0")
	commands[dockerNetworkInspect] = "abcdef0123456789 gavel0\n"
	check.run = commands.run
	require.NoError(t, check.Run(context.Background()), "a bridge with its own name")
}

func TestCheckHostRefusesHostServicesOpenToJobs(t *testing.T) {
	commands := firewalledHost("br-abcdef012345")
	commands["iptables -S INPUT"] = "-P INPUT ACCEPT\n-A INPUT -i lo -j ACCEPT"
	check := HostCheck{Binary: "docker", Network: "gavel-sandbox", ServeAddr: "10.0.0.5:9092", run: commands.run}
	err := check.Run(context.Background())
	require.Error(t, err)
	message := err.Error()
	assert.Contains(t, message, "the host firewall lets sandbox jobs on br-abcdef012345 reach this host's own services")
	assert.Contains(t, message, "iptables -I INPUT 1 -i br-abcdef012345 -m conntrack --ctstate RELATED,ESTABLISHED -j ACCEPT -m comment --comment gavel-sandbox")
	assert.Contains(t, message, "iptables -I INPUT 2 -i br-abcdef012345 -d 10.0.0.5 -p tcp --dport 9092 -j ACCEPT -m comment --comment gavel-sandbox")
	assert.Contains(t, message, "iptables -I INPUT 3 -i br-abcdef012345 -j REJECT -m comment --comment gavel-sandbox")

	for name, accept := range map[string]string{
		"every interface": "-A INPUT -p tcp -m tcp --dport 5432 -j ACCEPT",
		"the bridge":      "-A INPUT -i br-abcdef012345 -j ACCEPT",
		"a matching glob": "-A INPUT -i br+ -j ACCEPT",
	} {
		commands["iptables -S INPUT"] = "-P INPUT ACCEPT\n" + accept + "\n" + strings.TrimPrefix(goodInputRules("br-abcdef012345"), "-P INPUT ACCEPT\n")
		err := check.Run(context.Background())
		require.Error(t, err, name)
		assert.Contains(t, err.Error(), "accepts traffic from sandbox jobs before gavel's REJECT", name)
	}

	commands["iptables -S INPUT"] = "-P INPUT ACCEPT\n-A INPUT -i eth0 -p tcp -m tcp --dport 22 -j ACCEPT\n" + strings.TrimPrefix(goodInputRules("br-abcdef012345"), "-P INPUT ACCEPT\n")
	require.NoError(t, check.Run(context.Background()), "an ACCEPT for another interface doesn't reach jobs")
}

func TestCheckHostRefusesWithTheCommandsToRun(t *testing.T) {
	commands := firewalledHost("br-abcdef012345")
	commands["iptables -S DOCKER-USER"] = "-N DOCKER-USER\n-A DOCKER-USER -j RETURN"
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
	commands := firewalledHost("br-abcdef012345")
	commands["iptables -S DOCKER-USER"] = strings.Join(rules, "\n")
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
