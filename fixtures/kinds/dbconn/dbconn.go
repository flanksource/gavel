// Package dbconn resolves the database connection a fixture step names: a
// catalog name or connection:// reference, answered by the host's resolver
// (RunOptions.Connections) first and then by gavel's layered connection store
// (the shared, gavel and fixture-directory `.connections` layers), or an inline
// commons-db connection spec whose `valueFrom` references are hydrated.
package dbconn

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/flanksource/commons-db/connection/crd"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flanksource/gavel/fixtures"
)

// ErrNotFound is wrapped when no resolver defines the named connection.
var ErrNotFound = errors.New("connection not found")

// DefaultApp is the connection store application layer used when Options
// names none.
const DefaultApp = "gavel"

// inlineName names an inline connection that declares no name of its own.
const inlineName = "inline"

// Options tunes resolution; the zero value resolves through the gavel layers.
type Options struct {
	// App is the connection store application layer (<config>/<app>/connections).
	App string
}

// Resolved is a hydrated connection and the commons-db context that answers
// Ref (and every other catalog reference) with it, so a query profile pointed
// at Ref, or at another connection://name, resolves the same way.
type Resolved struct {
	Connection *models.Connection
	// Ref is the connection:// reference Context answers with Connection.
	Ref     string
	Context dbcontext.Context
}

// Resolve resolves ref — a string reference or an inline connection object.
func Resolve(ctx context.Context, step *fixtures.StepContext, ref any, opts Options) (*Resolved, error) {
	dbctx, err := catalogContext(ctx, step, opts)
	if err != nil {
		return nil, err
	}
	var connection *models.Connection
	switch value := ref.(type) {
	case string:
		connection, err = resolveReference(dbctx, value)
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
	hydrated, err := dbcontext.HydrateConnection(dbctx, connection)
	if err != nil {
		return nil, fmt.Errorf("hydrate connection %q: %w", connection.Name, err)
	}
	return answering(dbctx, hydrated), nil
}

// A params-reading `setup:` resolves its connection:// references through the
// same catalog a step and a `format: connection` picker read.
func init() {
	fixtures.StepConnectionCatalog = func(ctx context.Context, workDir string, host dbcontext.ConnectionResolver) (dbcontext.Context, error) {
		return AttachCatalog(ctx, workDir, host, Options{})
	}
}

// catalogContext is a commons-db context resolving references through the
// host resolver, then the layered gavel store.
func catalogContext(ctx context.Context, step *fixtures.StepContext, opts Options) (dbcontext.Context, error) {
	return AttachCatalog(ctx, step.WorkDir, step.Opts.Connections, opts)
}

// AttachCatalog is a commons-db context resolving references through host
// (when set), then the layered gavel store as a fixture in workDir sees it.
func AttachCatalog(ctx context.Context, workDir string, host dbcontext.ConnectionResolver, opts Options) (dbcontext.Context, error) {
	catalog, err := Catalog(workDir, opts)
	if err != nil {
		return dbcontext.Context{}, err
	}
	dbctx := catalog.Attach(dbcontext.NewContext(ctx))
	if host == nil {
		return dbctx, nil
	}
	stored := dbctx.ConnectionResolver()
	return dbctx.WithConnectionResolver(func(ctx dbcontext.Context, reference string) (*models.Connection, error) {
		connection, err := host(ctx, reference)
		if err != nil || connection != nil {
			return connection, err
		}
		return stored(ctx, reference)
	}), nil
}

func resolveReference(dbctx dbcontext.Context, reference string) (*models.Connection, error) {
	if reference == "" {
		return nil, errors.New("connection reference is empty")
	}
	connection, err := dbctx.ConnectionResolver()(dbctx, reference)
	if err != nil {
		return nil, fmt.Errorf("resolve connection %q: %w", reference, err)
	}
	if connection == nil {
		return nil, fmt.Errorf("connection %q: %w (looked in the host catalog and the gavel connection layers)", reference, ErrNotFound)
	}
	copied := *connection
	return &copied, nil
}

// flatEnvVarKeys are the ConnectionSpec keys an inline connection may give as
// a plain string, read as that EnvVar's `value`.
var flatEnvVarKeys = []string{"url", "port", "username", "password", "certificate"}

// inlineConnection reads a commons-db ConnectionSpec (`type`/`url`/`username`/
// `password`, each a string or an EnvVar with `valueFrom`, or a typed variant
// such as `postgres:`), plus an optional `name` and `namespace`.
func inlineConnection(value map[string]any) (*models.Connection, error) {
	spec := make(map[string]any, len(value))
	for k, v := range value {
		spec[k] = v
	}
	for _, key := range flatEnvVarKeys {
		if text, ok := spec[key].(string); ok {
			spec[key] = map[string]any{"value": text}
		}
	}
	meta := metav1.ObjectMeta{Name: inlineName}
	for key, target := range map[string]*string{"name": &meta.Name, "namespace": &meta.Namespace} {
		if raw, ok := spec[key]; ok {
			text, isString := raw.(string)
			if !isString || text == "" {
				return nil, fmt.Errorf("inline connection %s must be a non-empty string, got %v", key, raw)
			}
			*target = text
			delete(spec, key)
		}
	}
	resource := crd.Connection{ObjectMeta: meta}
	if err := decodeStrict(spec, &resource.Spec); err != nil {
		return nil, fmt.Errorf("inline connection: %w", err)
	}
	connection, err := crd.ToModel(&resource)
	if err != nil {
		return nil, fmt.Errorf("inline connection: %w", err)
	}
	return &connection, nil
}

func decodeStrict(value any, target any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

// answering is dbctx with connection answering its own connection:// ref.
func answering(dbctx dbcontext.Context, connection *models.Connection) *Resolved {
	ref := "connection://" + connection.Name
	if connection.Namespace != "" {
		ref = "connection://" + connection.Namespace + "/" + connection.Name
	}
	next := dbctx.ConnectionResolver()
	answered := dbctx.WithConnectionResolver(func(ctx dbcontext.Context, reference string) (*models.Connection, error) {
		if reference == ref {
			copied := *connection
			return &copied, nil
		}
		return next(ctx, reference)
	})
	return &Resolved{Connection: connection, Ref: ref, Context: answered}
}
