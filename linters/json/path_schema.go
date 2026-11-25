package jsonlint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// BuiltinScheme names a schema gavel itself provides: gavel:<name>.
const BuiltinScheme = "gavel"

// BuiltinSchema is a schema gavel provides for its own file kinds, such as the
// frontmatter of a fixture Markdown file. Matches reports whether it applies
// to path (absolute) under cfg, the file's layered .gavel.yaml.
type BuiltinSchema struct {
	Name    string
	Matches func(path string, cfg verify.GavelConfig) (bool, error)
	Load    func() (map[string]any, error)
}

var (
	builtinsMu sync.Mutex
	builtins   []BuiltinSchema
	builtinRaw = map[string]func() ([]byte, error){}
)

// RegisterBuiltinSchema adds schema as gavel:<Name>, the schema of the files
// it matches that declare none and that no .gavel.yaml lint.schemas glob names.
// The schema is generated once, on first use.
func RegisterBuiltinSchema(schema BuiltinSchema) {
	builtinsMu.Lock()
	defer builtinsMu.Unlock()
	if slices.ContainsFunc(builtins, func(b BuiltinSchema) bool { return b.Name == schema.Name }) {
		panic(fmt.Sprintf("jsonlint: builtin schema %q registered twice", schema.Name))
	}
	builtins = append(builtins, schema)
	builtinRaw[schema.Name] = sync.OnceValues(func() ([]byte, error) {
		value, err := schema.Load()
		if err != nil {
			return nil, fmt.Errorf("generate builtin schema %s: %w", schema.Name, err)
		}
		data, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("encode builtin schema %s: %w", schema.Name, err)
		}
		return data, nil
	})
}

// loadBuiltin decodes a fresh copy of a builtin schema on every call, since
// EditorSchema rewrites the schemas it loads in place.
func loadBuiltin(location string) (any, error) {
	name := strings.TrimPrefix(location, BuiltinScheme+":")
	builtinsMu.Lock()
	load, ok := builtinRaw[name]
	builtinsMu.Unlock()
	if !ok {
		return nil, fmt.Errorf("unknown builtin schema %q", location)
	}
	data, err := load()
	if err != nil {
		return nil, err
	}
	return jsonschema.UnmarshalJSON(bytes.NewReader(data))
}

// PathSchema is the schema reference for path (absolute) when the file
// declares none: the lint.schemas glob of its .gavel.yaml layers that matches
// it, else the builtin schema of its gavel file kind; "" when neither applies.
func PathSchema(path string) (string, error) {
	cfg, err := verify.LoadGavelConfig(filepath.Dir(path))
	if err != nil {
		return "", err
	}
	var matched []string
	for glob := range cfg.Lint.Schemas {
		ok, err := doublestar.Match(filepath.ToSlash(glob), filepath.ToSlash(path))
		if err != nil {
			return "", fmt.Errorf("lint.schemas glob %q: %w", glob, err)
		}
		if ok {
			matched = append(matched, glob)
		}
	}
	slices.Sort(matched)
	if len(matched) > 1 {
		return "", fmt.Errorf("%s matches several lint.schemas globs: %s", path, strings.Join(matched, ", "))
	}
	if len(matched) == 1 {
		return cfg.Lint.Schemas[matched[0]], nil
	}
	builtinsMu.Lock()
	candidates := slices.Clone(builtins)
	builtinsMu.Unlock()
	for _, builtin := range candidates {
		ok, err := builtin.Matches(path, cfg)
		if err != nil {
			return "", fmt.Errorf("builtin schema %s: %w", builtin.Name, err)
		}
		if ok {
			return BuiltinScheme + ":" + builtin.Name, nil
		}
	}
	return "", nil
}

// MatchesFixtureFiles reports whether path is a fixture file under cfg's
// fixtures.files globs, which are relative to the file's git root.
func MatchesFixtureFiles(path string, cfg verify.GavelConfig, extensions ...string) (bool, error) {
	if !slices.Contains(extensions, strings.ToLower(filepath.Ext(path))) {
		return false, nil
	}
	root := utils.FindGitRoot(filepath.Dir(path))
	if root == "" {
		root = filepath.Dir(path)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false, err
	}
	for _, glob := range cfg.Fixtures.ResolvedFiles() {
		ok, err := doublestar.Match(filepath.ToSlash(glob), filepath.ToSlash(rel))
		if err != nil {
			return false, fmt.Errorf("fixtures.files glob %q: %w", glob, err)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}
