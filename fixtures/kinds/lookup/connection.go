package lookup

import (
	"context"
	"errors"
	"strings"

	"github.com/flanksource/clicky/schemaform"
	"github.com/flanksource/commons-db/models"

	"github.com/flanksource/gavel/fixtures/kinds/dbconn"
)

// ConnectionBinding answers a `format: connection` picker in a fixture that
// declares `connections:`: its named connections of Types, then the
// catalog's, as WorkDir sees it.
type ConnectionBinding struct {
	Types       []string
	Connections *dbconn.Connections
	WorkDir     string
	Options     dbconn.Options
}

func (b ConnectionBinding) validate() error {
	if b.Connections == nil {
		return errors.New("a connection binding needs the fixture's connections")
	}
	return nil
}

func (b ConnectionBinding) resolve(ctx context.Context, req Request) (Result, error) {
	named, err := b.Connections.NamedOfTypes(b.Types)
	if err != nil {
		return Result{}, err
	}
	result, err := ConnectionOptions(ctx, ConnectionQuery{WorkDir: b.WorkDir, Options: b.Options, Types: b.Types, Query: req.Query, Named: named})
	return Result{LookupResult: result}, err
}

// ConnectionQuery is a connection picker's request: the catalog WorkDir sees,
// the types it scopes to, the name search text, and the fixture's named
// connections listed ahead of the catalog's.
type ConnectionQuery struct {
	WorkDir string
	Options dbconn.Options
	Types   []string
	Query   string
	Named   []dbconn.NamedConnection
}

// ConnectionOptions lists the named connections whose name contains the
// search text, then the catalog's (a catalog entry a named one shadows is
// left out), each a dbconn.PickerOption, capped at MaxOptions.
func ConnectionOptions(ctx context.Context, q ConnectionQuery) (schemaform.LookupResult, error) {
	service, err := dbconn.Service(ctx, q.WorkDir, q.Options)
	if err != nil {
		return schemaform.LookupResult{}, err
	}
	catalog, err := dbconn.Lookup(service, dbconn.LookupQuery{Types: q.Types, Query: q.Query})
	if err != nil {
		return schemaform.LookupResult{}, err
	}
	needle := strings.ToLower(strings.TrimSpace(q.Query))
	all := listed(q.Named, catalog)
	var options []schemaform.Option
	seen := map[string]bool{}
	for _, named := range q.Named {
		if needle != "" && !strings.Contains(strings.ToLower(named.Name), needle) {
			continue
		}
		option := dbconn.PickerOption(named.Connection, all)
		option.Value, option.Label = named.Ref, named.Name+" ("+named.Connection.Type+")"
		options, seen[option.Value] = append(options, option), true
	}
	for _, connection := range catalog {
		if option := dbconn.PickerOption(connection, all); !seen[option.Value] {
			options = append(options, option)
		}
	}
	result := schemaform.LookupResult{Total: len(options)}
	if len(options) > MaxOptions {
		options, result.Truncated = options[:MaxOptions], true
	}
	result.Options = options
	return result, nil
}

// listed is every connection a picker lists, named first, for following a
// `connection://` delegation.
func listed(named []dbconn.NamedConnection, catalog []*models.Connection) []*models.Connection {
	all := make([]*models.Connection, 0, len(named)+len(catalog))
	for _, entry := range named {
		all = append(all, entry.Connection)
	}
	return append(all, catalog...)
}
