// Package lookup holds the server-side bindings behind a form's
// `x-clicky-lookup` fields that a fixture declares — an `x-query` field read
// by SQL, a `format: connection` picker that also lists the fixture's named
// connections — and answers them for the terminal form and the HTTP lookup
// endpoint alike. A form step registers a binding per field while it asks and
// releases it when the step ends; the unguessable token in the field's lookup
// URL is the only way to reach it.
package lookup

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/flanksource/clicky/entity"
	"github.com/flanksource/clicky/schemaform"
)

// URLPrefix is the path a binding's lookup URL starts with:
// URLPrefix + <token> + <json pointer of the field>.
const URLPrefix = "/api/v1/lookup/"

// MaxOptions caps a lookup's options, as clicky's lookup endpoints do.
const MaxOptions = entity.MaxLookupOptions

// Request is one lookup of a registered field: the token and JSON pointer of
// its URL, the search text, and the form state it is scoped by — Scope is the
// `x-clicky-lookup` scope params a browser sends (`connection`), Values the
// whole form a terminal holds. A linked field reads Scope ahead of Values.
type Request struct {
	Token   string
	Pointer string
	Query   string
	Scope   map[string]any
	Values  map[string]any
}

// Result is a lookup's options and the label its filter shows: empty, or
// "Choose <param> first" while the param a query is linked to is unset.
type Result struct {
	schemaform.LookupResult
	Label string
}

// Binding is what answers a registered field: a QueryBinding or a
// ConnectionBinding.
type Binding interface {
	validate() error
	resolve(ctx context.Context, req Request) (Result, error)
}

type entry struct {
	pointer string
	binding Binding
}

var registry = struct {
	sync.Mutex
	entries map[string]entry
}{entries: map[string]entry{}}

// Register holds binding for the field at pointer until Release, returning
// the token its lookup URL carries.
func Register(pointer string, binding Binding) (string, error) {
	if !strings.HasPrefix(pointer, "/") {
		return "", fmt.Errorf("lookup binding pointer %q must be a JSON pointer to the field", pointer)
	}
	if err := binding.validate(); err != nil {
		return "", fmt.Errorf("lookup binding %s: %w", pointer, err)
	}
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("lookup token: %w", err)
	}
	token := hex.EncodeToString(raw)
	registry.Lock()
	defer registry.Unlock()
	registry.entries[token] = entry{pointer: pointer, binding: binding}
	return token, nil
}

// Release drops the binding token names; its URL answers no more lookups.
func Release(token string) {
	registry.Lock()
	defer registry.Unlock()
	delete(registry.entries, token)
}

// Resolve answers req with the binding its token names, refusing a token that
// was never registered or was released, and a pointer that is not the field
// the token was registered for.
func Resolve(ctx context.Context, req Request) (Result, error) {
	registry.Lock()
	found, ok := registry.entries[req.Token]
	registry.Unlock()
	if !ok {
		return Result{}, fmt.Errorf("lookup token %q is unknown or its form step has ended: %w", req.Token, ErrNotFound)
	}
	if found.pointer != req.Pointer {
		return Result{}, fmt.Errorf("lookup token %q is for field %s, not %s: %w", req.Token, found.pointer, req.Pointer, ErrNotFound)
	}
	return found.binding.resolve(ctx, req)
}

// ErrNotFound is what Resolve wraps when no registered field answers a
// request's token and pointer.
var ErrNotFound = errors.New("no such lookup field")

// Response is result in clicky's lookup wire shape under filter, the key the
// field's `x-clicky-lookup` reads.
func Response(filter string, result Result) entity.LookupResponse {
	options := make(map[string]entity.LookupNode, len(result.Options))
	for _, option := range result.Options {
		options[option.Value] = Node(option)
	}
	return entity.LookupResponse{Filters: map[string]entity.LookupFilter{filter: {
		Label: result.Label, Options: options, Truncated: result.Truncated, Total: result.Total,
	}}}
}

// Node is option as a lookup node: its label and how pickers decorate it.
func Node(option schemaform.Option) entity.LookupNode {
	node := entity.LookupNode{Kind: "text", Text: option.Label, Plain: option.Label, Icon: option.Icon,
		Description: option.Description, Tone: option.Tone, Group: option.Group}
	for _, badge := range option.Badges {
		node.Badges = append(node.Badges, entity.LookupBadge{Label: badge.Label, Icon: badge.Icon, Tone: badge.Tone})
	}
	return node
}

// Pointer is the JSON pointer of the property at path.
func Pointer(path []string) string {
	escaper := strings.NewReplacer("~", "~0", "/", "~1")
	var b strings.Builder
	for _, segment := range path {
		b.WriteString("/" + escaper.Replace(segment))
	}
	return b.String()
}

// URL is the lookup URL of the field at pointer registered under token.
func URL(token, pointer string) string {
	segments := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return URLPrefix + token + "/" + strings.Join(segments, "/")
}

// ParseURL reads the token and pointer of a lookup URL; ok is false for a URL
// that is not a binding's (e.g. the global connection catalog).
func ParseURL(raw string) (token, pointer string, ok bool, err error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", "", false, fmt.Errorf("lookup url %q: %w", raw, err)
	}
	rest, ok := strings.CutPrefix(parsed.Path, URLPrefix)
	if !ok {
		return "", "", false, nil
	}
	token, field, found := strings.Cut(rest, "/")
	if !found || token == "" || field == "" {
		return "", "", true, fmt.Errorf("lookup url %q: want %s<token>/<field pointer>", raw, URLPrefix)
	}
	return token, "/" + field, true, nil
}
