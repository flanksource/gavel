package jsonlint

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
)

type schemaTestServer struct {
	URL   string
	Close func()
}

// servePublicSchema exercises the production address policy without external
// traffic: only the public fixture address is routed to the local test server.
func servePublicSchema(handler http.Handler) schemaTestServer {
	server := httptest.NewServer(handler)
	transport := &http.Transport{DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		if address != "93.184.215.14:80" {
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
	return schemaTestServer{URL: "http://93.184.215.14", Close: server.Close}
}
