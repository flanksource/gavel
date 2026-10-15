// Checks that a container host can run sandbox jobs safely: the bridge network exists and the
// host firewall keeps jobs on it off private networks except gavel serve itself.
package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

// privateRanges are the destinations sandbox jobs must not reach: RFC 1918, carrier-grade NAT,
// link-local (cloud metadata) and loopback.
var privateRanges = []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "100.64.0.0/10", "169.254.0.0/16", "127.0.0.0/8"}

// firewallComment marks the iptables rules gavel expects.
const firewallComment = "gavel-sandbox"

// HostCheck verifies a container host before gavel serve submits jobs to it.
type HostCheck struct {
	Binary  string // docker or podman
	Network string // the bridge jobs join
	// ServeAddr is the IP and port sandbox jobs reach gavel serve at, the one private
	// destination they may use.
	ServeAddr string

	run func(ctx context.Context, name string, args ...string) (string, error)
}

// Run returns an error, with the commands that fix it, when the host isn't ready. It never
// changes the host: the fixes need root.
func (c HostCheck) Run(ctx context.Context) error {
	run := c.run
	if run == nil {
		run = runCommand
	}
	host, port, err := net.SplitHostPort(c.ServeAddr)
	if err != nil || net.ParseIP(host) == nil || net.ParseIP(host).IsUnspecified() {
		return fmt.Errorf("gavel serve's address %s needs the IP sandbox jobs reach it at, such as 10.0.0.5:9092", c.ServeAddr)
	}
	if c.Binary == "podman" {
		out, err := run(ctx, "podman", "info", "--format", "{{.Host.Security.Rootless}}")
		if err != nil {
			return fmt.Errorf("check whether podman is rootless: %w", err)
		}
		if strings.TrimSpace(out) == "true" {
			return fmt.Errorf("rootless podman is not supported for sandbox jobs: its networking bypasses the host firewall rules that keep jobs off private networks; run podman as root")
		}
	}
	bridge, err := c.bridge(ctx, run)
	if err != nil {
		return err
	}
	rules, err := run(ctx, "iptables", "-S", "DOCKER-USER")
	if err != nil {
		return fmt.Errorf("read the DOCKER-USER firewall chain (gavel serve needs to read iptables to check the sandbox network): %w", err)
	}
	return checkFirewall(rules, bridge, host, port)
}

func (c HostCheck) bridge(ctx context.Context, run func(context.Context, string, ...string) (string, error)) (string, error) {
	missing := func(err error) error {
		return fmt.Errorf("the sandbox network %s is missing (%v); create it with:\n  %s network create %s", c.Network, err, c.Binary, c.Network)
	}
	if c.Binary == "podman" {
		out, err := run(ctx, "podman", "network", "inspect", c.Network, "--format", "{{.NetworkInterface}}")
		if err != nil {
			return "", missing(err)
		}
		return strings.TrimSpace(out), nil
	}
	out, err := run(ctx, c.Binary, "network", "inspect", c.Network, "--format", `{{.Id}} {{index .Options "com.docker.network.bridge.name"}}`)
	if err != nil {
		return "", missing(err)
	}
	fields := strings.Fields(out)
	if len(fields) == 0 {
		return "", missing(fmt.Errorf("%s printed no network id", c.Binary))
	}
	if len(fields) > 1 {
		return fields[1], nil
	}
	if len(fields[0]) < 12 {
		return "", missing(fmt.Errorf("%s printed the network id %q", c.Binary, fields[0]))
	}
	return "br-" + fields[0][:12], nil
}

// checkFirewall finds, in `iptables -S DOCKER-USER` output, an ACCEPT for gavel serve followed by
// a REJECT for every private range, all from bridge.
func checkFirewall(rules, bridge, host, port string) error {
	acceptAt, firstReject, missing := -1, -1, false
	lines := strings.Split(rules, "\n")
	ours := func(line string, parts ...string) bool {
		fields := " " + line + " "
		for _, part := range append(parts, "-A DOCKER-USER", "-i "+bridge, "--comment "+firewallComment) {
			if !strings.Contains(fields, " "+part+" ") {
				return false
			}
		}
		return true
	}
	for i, line := range lines {
		if acceptAt < 0 && ours(line, "-d "+host+"/32", "--dport "+port, "-j ACCEPT") {
			acceptAt = i
		}
	}
	for _, cidr := range privateRanges {
		found := -1
		for i, line := range lines {
			if ours(line, "-d "+cidr, "-j REJECT") {
				found = i
				break
			}
		}
		if found < 0 {
			missing = true
			continue
		}
		if firstReject < 0 || found < firstReject {
			firstReject = found
		}
	}
	if acceptAt < 0 || missing {
		return fmt.Errorf("the host firewall doesn't keep sandbox jobs on %s off private networks; as root, run:\n%s", bridge, firewallCommands(bridge, host, port))
	}
	if acceptAt > firstReject {
		return fmt.Errorf("the ACCEPT for gavel serve must come before the REJECT rules in DOCKER-USER, or jobs can't report results; as root, delete the gavel-sandbox rules and run:\n%s", firewallCommands(bridge, host, port))
	}
	return nil
}

// firewallCommands are the iptables commands that set up bridge: the ACCEPT for gavel serve
// inserted first, then a REJECT per private range.
func firewallCommands(bridge, host, port string) string {
	commands := []string{fmt.Sprintf("  iptables -I DOCKER-USER -i %s -d %s -p tcp --dport %s -j ACCEPT -m comment --comment %s", bridge, host, port, firewallComment)}
	for _, cidr := range privateRanges {
		commands = append(commands, fmt.Sprintf("  iptables -A DOCKER-USER -i %s -d %s -j REJECT -m comment --comment %s", bridge, cidr, firewallComment))
	}
	return strings.Join(commands, "\n")
}

func runCommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if message := strings.TrimSpace(stderr.String()); message != "" {
			return "", fmt.Errorf("%w: %s", err, message)
		}
		return "", err
	}
	return stdout.String(), nil
}
