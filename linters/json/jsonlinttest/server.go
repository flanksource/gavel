// Package jsonlinttest serves JSON schemas to specs under the json linter's
// production network policy, which refuses loopback and private hosts.
package jsonlinttest

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
)

// PublicHost is the public host specs fetch schemas from; the test
// transport routes it, and only it, to the local test server.
const PublicHost = "93.184.215.14"

// SchemaServer is a local schema server reachable through a public URL.
type SchemaServer struct {
	URL   string
	Close func()
}

// ServePublicSchema serves handler at http://PublicHost for the current
// spec, so remote-schema fetching is exercised under its public-host policy
// without external traffic. http.DefaultTransport is restored on cleanup.
func ServePublicSchema(handler http.Handler) SchemaServer {
	server := httptest.NewServer(handler)
	transport := &http.Transport{DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != PublicHost+":80" {
			return nil, fmt.Errorf("unexpected schema test destination %s", address)
		}
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}}
	previous := http.DefaultTransport
	http.DefaultTransport = transport
	DeferCleanup(func() {
		transport.CloseIdleConnections()
		http.DefaultTransport = previous
	})
	return SchemaServer{URL: "http://" + PublicHost, Close: server.Close}
}
