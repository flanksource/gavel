// Package queryapp serves the commons-db query app over the test UI's run
// trace stores, for its "Open in query" links. It is kept out of
// testrunner/ui so a host that imports the test UI without opening the app,
// such as oipa-cli, does not embed it: only gavel's own serve wiring imports
// this package.
package queryapp

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/flanksource/commons-db/cmd/query/www"
)

// app is built once: it serves embedded files, the same for every mount.
var app = sync.OnceValues(www.Handler)

// Host is a test UI server that serves the app over its stores.
type Host interface {
	SetQueryApp(app http.Handler)
}

// Serve has host serve the query app over its run trace stores.
func Serve(host Host) error {
	handler, err := app()
	if err != nil {
		return fmt.Errorf("serve the query app over the run trace stores: %w", err)
	}
	host.SetQueryApp(handler)
	return nil
}
