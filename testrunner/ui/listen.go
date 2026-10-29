package testui

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/flanksource/commons/logger"
)

// ListenOptions is where Listen serves a new Server and how it announces it:
// Addr names an interface only (the port is always picked, default 0.0.0.0);
// Title starts the banner line Writer gets (default os.Stdout); Open opens the
// page in a browser; DevProxy serves the page from a Vite dev server.
type ListenOptions struct {
	Addr     string
	Title    string
	Writer   io.Writer
	Open     bool
	DevProxy http.Handler
}

// Listen binds an ephemeral TCP listener and serves a new Server on it. It
// returns the listener so a caller can hand it off to a detached child
// process instead of tearing the connection down and rebinding, and the
// local URL it announced.
func Listen(opts ListenOptions) (*Server, net.Listener, string, error) {
	srv := NewServer()
	if opts.DevProxy != nil {
		srv.SetDevProxy(opts.DevProxy)
	}
	addr := opts.Addr
	if addr == "" {
		addr = "0.0.0.0"
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(addr, "0"))
	if err != nil {
		return nil, nil, "", fmt.Errorf("start %s on --addr %q (an interface, without a port): %w", opts.Title, addr, err)
	}
	url, networkURL := URLs(addr, listener.Addr().(*net.TCPAddr).Port)

	go http.Serve(listener, srv.Handler()) //nolint:errcheck

	time.Sleep(100 * time.Millisecond)
	writer := opts.Writer
	if writer == nil {
		writer = os.Stdout
	}
	if _, err := fmt.Fprintln(writer, Banner(opts.Title, url, networkURL)); err != nil {
		listener.Close() //nolint:errcheck
		return nil, nil, "", fmt.Errorf("announce %s: %w", opts.Title, err)
	}
	if opts.Open {
		OpenBrowser(url)
	}
	return srv, listener, url, nil
}

// OpenBrowser opens url in the desktop's browser, best effort.
func OpenBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	_ = exec.Command(cmd, args...).Start()
}

// URLs are the URLs a server bound to host:port answers on: local, for a
// browser on this machine, and network, the private IP a wildcard bind also
// answers on ("" for any other bind).
func URLs(host string, port int) (local, network string) {
	p := strconv.Itoa(port)
	local = "http://" + net.JoinHostPort(LocalHost(host), p)
	if host != "0.0.0.0" && host != "::" {
		return local, ""
	}
	ip := firstNonLoopbackIPv4()
	if ip == "" {
		logger.Warnf("--addr=%s but no non-loopback IPv4 interface found; announcing localhost only", host)
		return local, ""
	}
	return local, "http://" + net.JoinHostPort(ip, p)
}

// Banner formats the "<label> at <url>" line, appending the network URL when
// the server is reachable from other hosts.
func Banner(label, local, network string) string {
	if network == "" {
		return label + " at " + local
	}
	return fmt.Sprintf("%s at %s (network: %s)", label, local, network)
}

// LocalHost picks the hostname a browser or the native macOS menu-bar webview
// on this machine should load from. It must reach gavel over loopback:
// WKWebView blocks cleartext HTTP to non-loopback hosts under App Transport
// Security (which would render a blank popover), and a private IP is fragile
// across VPN/offline state. So every loopback/wildcard bind resolves to
// "localhost". An explicit interface (the user pinned --addr to one NIC) is
// kept as-is since that is the only address the server listens on.
func LocalHost(requested string) string {
	switch requested {
	case "", "0.0.0.0", "::", "localhost", "127.0.0.1", "::1":
		return "localhost"
	default:
		return requested
	}
}

func firstNonLoopbackIPv4() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Warnf("net.InterfaceAddrs failed: %v", err)
		return ""
	}
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() || ipnet.IP.IsLinkLocalUnicast() {
			continue
		}
		v4 := ipnet.IP.To4()
		if v4 == nil {
			continue
		}
		return v4.String()
	}
	return ""
}
