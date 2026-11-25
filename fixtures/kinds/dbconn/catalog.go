package dbconn

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	connections "github.com/flanksource/commons-db/cmd/query/connections"
	"github.com/flanksource/commons-db/connection/store"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/models"
	"github.com/flanksource/commons-db/query"
)

// Catalog is gavel's layered connection store as a fixture in workDir sees
// it: the shared and gavel layers, then workDir's `.connections`.
func Catalog(workDir string, opts Options) (*store.Store, error) {
	app := opts.App
	if app == "" {
		app = DefaultApp
	}
	layers, err := store.DefaultLayers(app)
	if err != nil {
		return nil, err
	}
	layers[len(layers)-1].Dir = filepath.Join(workDir, ".connections")
	return store.New(layers...)
}

// Service is the commons-db connection service over Catalog(workDir): its
// List scopes by type and redacts every credential, and Create saves to
// workDir's `.connections` unless the body names another writable layer.
func Service(ctx context.Context, workDir string, opts Options) (*connections.Service, error) {
	catalog, err := Catalog(workDir, opts)
	if err != nil {
		return nil, err
	}
	provider, err := connections.NewStoreProvider(connections.StoreProviderOptions{Store: catalog, WriteLayer: store.LayerCWD})
	if err != nil {
		return nil, err
	}
	dbctx := catalog.Attach(dbcontext.NewContext(ctx))
	return connections.New(connections.Options{
		Virtual:    provider,
		Context:    func() dbcontext.Context { return dbctx },
		DecodeBody: func(_ context.Context, body map[string]any) (map[string]any, error) { return body, nil },
		Profiles:   func(context.Context) ([]query.Profile, error) { return nil, nil },
		Secrets:    opts.Secrets,
	})
}

// LookupQuery is a `format: connection` picker's request: the connection
// types its lookup URL scopes to (`?types=a,b`) and the name search text.
type LookupQuery struct {
	Types []string
	Query string
}

// LookupTypes reads the `types` a picker's lookup URL scopes it to.
func LookupTypes(lookupURL string) ([]string, error) {
	parsed, err := url.Parse(lookupURL)
	if err != nil {
		return nil, fmt.Errorf("connection lookup url %q: %w", lookupURL, err)
	}
	return splitTypes(parsed.Query().Get("types")), nil
}

func splitTypes(raw string) []string {
	var types []string
	for _, t := range strings.Split(raw, ",") {
		if t = strings.TrimSpace(t); t != "" {
			types = append(types, t)
		}
	}
	return types
}

// Lookup lists the catalog's connections a picker offers: those of the
// requested types whose name contains the search text, case-insensitively,
// sorted by name and redacted.
func Lookup(service *connections.Service, q LookupQuery) ([]*models.Connection, error) {
	listed, err := service.List(connections.ListOptions{Types: strings.Join(q.Types, ",")})
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(strings.TrimSpace(q.Query))
	var matched []*models.Connection
	for _, connection := range listed {
		if needle == "" || strings.Contains(strings.ToLower(connection.Name), needle) {
			matched = append(matched, connection)
		}
	}
	return matched, nil
}

// PickerValue is what a picker commits for connection: its connection://
// reference (namespaced when it has a namespace), which Resolve, a setup's
// `connection:` and a query step read back alike.
func PickerValue(connection *models.Connection) string {
	if connection.Namespace != "" {
		return "connection://" + connection.Namespace + "/" + connection.Name
	}
	return "connection://" + connection.Name
}

// PickerLabel is a picker option's label: the name and connection type.
func PickerLabel(connection *models.Connection) string {
	name := connection.Name
	if connection.Namespace != "" {
		name = connection.Namespace + "/" + name
	}
	if connection.Type == "" {
		return name
	}
	return name + " (" + connection.Type + ")"
}
