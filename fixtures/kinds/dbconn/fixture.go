package dbconn

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/models"

	"github.com/flanksource/gavel/fixtures"
)

const refPrefix = "connection://"

// Connections is a fixture's declared `connections:` layered over the
// catalog: named entries answer their name (or connection://<name>) ahead of
// the host resolver and the gavel store, and each list entry is resolved once
// as the default for its connection type.
type Connections struct {
	context  dbcontext.Context
	set      fixtures.ConnectionSet
	defaults []declaredDefault
}

type declaredDefault struct {
	ref      string
	resolved *Resolved
}

// NamedConnection is a named `connections:` entry as a picker offers it: Ref
// is the connection://<name> value the named layer answers. Connection is
// unhydrated; an inline entry's carries the credentials it was declared with.
type NamedConnection struct {
	Name       string
	Ref        string
	Connection *models.Connection
}

// StepConnections is the step's fixture connections, resolved against the
// step's directory and the host's resolver (RunOptions.Connections).
func StepConnections(ctx context.Context, step *fixtures.StepContext, opts Options) (*Connections, error) {
	return OpenConnections(ctx, OpenOptions{WorkDir: step.WorkDir, Host: step.Opts.Connections, Set: step.Test.Connections, Options: opts})
}

// OpenOptions are where OpenConnections resolves: the fixture directory, the
// host resolver (nil for none) and the fixture's declaration (nil for none).
type OpenOptions struct {
	WorkDir string
	Host    dbcontext.ConnectionResolver
	Set     *fixtures.ConnectionSet
	Options
}

// OpenConnections layers the declaration over AttachCatalog and resolves its
// defaults, failing on an inline entry that does not decode, a default that
// does not resolve, or two defaults of one type — every offender named.
func OpenConnections(ctx context.Context, open OpenOptions) (*Connections, error) {
	dbctx, err := AttachCatalog(ctx, open.WorkDir, open.Host, open.Options)
	if err != nil {
		return nil, err
	}
	c := &Connections{context: dbctx}
	if open.Set == nil {
		return c, nil
	}
	c.set = *open.Set
	inline, err := decodeNamed(c.set)
	if err != nil {
		return nil, err
	}
	c.context = dbctx.WithConnectionResolver(namedLayer(c.set.Named, inline, dbctx.ConnectionResolver()))
	return c, c.resolveDefaults()
}

func decodeNamed(set fixtures.ConnectionSet) (map[string]*models.Connection, error) {
	inline := map[string]*models.Connection{}
	var errs []error
	for _, name := range set.NamedKeys() {
		value, ok := set.Named[name].(map[string]any)
		if !ok {
			continue
		}
		if _, named := value["name"]; !named {
			value = mergedName(value, name)
		}
		connection, err := inlineConnection(value)
		if err != nil {
			errs = append(errs, fmt.Errorf("connections.%s: %w", name, err))
			continue
		}
		inline[name] = connection
	}
	return inline, errors.Join(errs...)
}

func mergedName(value map[string]any, name string) map[string]any {
	named := map[string]any{"name": name}
	for k, v := range value {
		named[k] = v
	}
	return named
}

// namedLayer answers a named entry's name and connection://<name>: an inline
// entry with its decoded connection, a string entry by resolving its
// reference through next.
func namedLayer(named map[string]any, inline map[string]*models.Connection, next dbcontext.ConnectionResolver) dbcontext.ConnectionResolver {
	return func(ctx dbcontext.Context, reference string) (*models.Connection, error) {
		name := strings.TrimPrefix(reference, refPrefix)
		entry, ok := named[name]
		if !ok {
			return next(ctx, reference)
		}
		if connection, decoded := inline[name]; decoded {
			copied := *connection
			return &copied, nil
		}
		ref := entry.(string)
		connection, err := next(ctx, ref)
		if err == nil && connection == nil {
			err = fmt.Errorf("%q: %w", ref, ErrNotFound)
		}
		if err != nil {
			return nil, fmt.Errorf("connections.%s: %w", name, err)
		}
		return connection, nil
	}
}

func (c *Connections) resolveDefaults() error {
	var errs []error
	for i, ref := range c.set.Defaults {
		resolved, err := c.Resolve(ref)
		if err != nil {
			errs = append(errs, fmt.Errorf("connections[%d] %q: %w", i, ref, err))
			continue
		}
		if prior, ok := c.defaultEntry(resolved.Connection.Type); ok {
			errs = append(errs, fmt.Errorf("connections %q and %q are both %s connections: a type has one default", prior.ref, ref, resolved.Connection.Type))
			continue
		}
		c.defaults = append(c.defaults, declaredDefault{ref: ref, resolved: resolved})
	}
	return errors.Join(errs...)
}

func (c *Connections) defaultEntry(connectionType string) (declaredDefault, bool) {
	for _, d := range c.defaults {
		if d.resolved.Connection.Type == connectionType {
			return d, true
		}
	}
	return declaredDefault{}, false
}

// Resolve resolves ref — a string reference or an inline connection object —
// through the named layer, the host resolver and the gavel store.
func (c *Connections) Resolve(ref any) (*Resolved, error) {
	var connection *models.Connection
	var err error
	switch value := ref.(type) {
	case string:
		connection, err = resolveReference(c.context, value)
	case map[string]any:
		connection, err = inlineConnection(value)
	case nil:
		return nil, errors.New("no connection given: set a connection name, connection:// reference or inline connection")
	default:
		return nil, fmt.Errorf("connection must be a name, connection:// reference or object, got %T", ref)
	}
	if err != nil {
		return nil, err
	}
	hydrated, err := dbcontext.HydrateConnection(c.context, connection)
	if err != nil {
		return nil, fmt.Errorf("hydrate connection %q: %w", connection.Name, err)
	}
	return answering(c.context, hydrated), nil
}

// Context is the commons-db context resolving references through the named
// layer, the host resolver and the gavel store.
func (c *Connections) Context() dbcontext.Context { return c.context }

// DefaultFor is the declared default of connectionType.
func (c *Connections) DefaultFor(connectionType string) (*Resolved, bool) {
	d, ok := c.defaultEntry(connectionType)
	return d.resolved, ok
}

// SingleDefault is the one declared default, for a step that names no
// connection; none or several is an error listing what was declared.
func (c *Connections) SingleDefault() (*Resolved, error) {
	switch len(c.defaults) {
	case 1:
		return c.defaults[0].resolved, nil
	case 0:
		return nil, errors.New("no default connection: name the step's connection or declare one under the fixture's connections: [<ref>]")
	}
	declared := make([]string, len(c.defaults))
	for i, d := range c.defaults {
		declared[i] = fmt.Sprintf("%q (%s)", d.ref, d.resolved.Connection.Type)
	}
	return nil, fmt.Errorf("%d default connections declared (%s): name the step's connection", len(c.defaults), strings.Join(declared, ", "))
}

// Named resolves the named entry name.
func (c *Connections) Named(name string) (*Resolved, error) {
	if _, ok := c.set.Named[name]; !ok {
		return nil, fmt.Errorf("connection %q: %w among the fixture's connections (%s)", name, ErrNotFound, strings.Join(c.set.NamedKeys(), ", "))
	}
	return c.Resolve(name)
}

// NamedOfTypes is every named entry, sorted by name, whose connection type is
// one of types (every entry when types is empty).
func (c *Connections) NamedOfTypes(types []string) ([]NamedConnection, error) {
	var out []NamedConnection
	for _, name := range c.set.NamedKeys() {
		connection, err := c.context.ConnectionResolver()(c.context, name)
		if err != nil {
			return nil, err
		}
		if len(types) > 0 && !slices.Contains(types, connection.Type) {
			continue
		}
		out = append(out, NamedConnection{Name: name, Ref: refPrefix + name, Connection: connection})
	}
	return out, nil
}
