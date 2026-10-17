// Checks a host's sandbox firewall from inside a job when GAVEL_SANDBOX_FIREWALL_TEST=1: on the
// gavel-sandbox network a job reaches the internet, but not the metadata address, private
// ranges, or services on the host itself.
package sandbox

import (
	"context"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTheSandboxFirewallConfinesJobs(t *testing.T) {
	if os.Getenv("GAVEL_SANDBOX_FIREWALL_TEST") != "1" {
		t.Skip("set GAVEL_SANDBOX_FIREWALL_TEST=1 on a host with the gavel-sandbox network and firewall (docs/sandbox.md)")
	}
	const network = "gavel-sandbox"
	gateway, err := exec.Command("docker", "network", "inspect", "-f", "{{(index .IPAM.Config 0).Gateway}}", network).Output()
	require.NoError(t, err, "the %s network exists", network)
	// A service on the host, as serve's database could be, listening where jobs could reach it.
	listener, err := net.Listen("tcp", strings.TrimSpace(string(gateway))+":0")
	require.NoError(t, err)
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}
	}()

	reach := func(address string) string {
		return `if timeout 5 bash -c "exec 3<>/dev/tcp/` + strings.Replace(address, ":", "/", 1) + `" 2>/dev/null; then echo "` + address + ` reachable"; else echo "` + address + ` blocked"; fi`
	}
	targets := map[string]string{
		"proxy.golang.org:443":   "reachable",
		"169.254.169.254:80":     "blocked",
		"10.0.0.1:80":            "blocked",
		"192.168.0.1:80":         "blocked",
		listener.Addr().String(): "blocked",
	}
	var script []string
	for target := range targets {
		script = append(script, reach(target))
	}
	executor := NewContainer("docker", ContainerOptions{Network: network})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	handle, err := executor.Submit(ctx, Job{RunID: uuid.New(), Image: "debian:bookworm-slim", Command: []string{"bash", "-c", strings.Join(script, "; ")}})
	require.NoError(t, err)
	logs, err := handle.Logs(ctx)
	require.NoError(t, err)
	output, err := io.ReadAll(logs)
	require.NoError(t, err)
	require.NoError(t, logs.Close())
	_, err = handle.Wait(ctx)
	require.NoError(t, err)
	for target, want := range targets {
		assert.Contains(t, string(output), target+" "+want, "from a job on %s", network)
	}
}
