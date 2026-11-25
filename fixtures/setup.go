package fixtures

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/flanksource/clicky/shutdown"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/shell"
	flanksourceContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/sdk/status"
	ghodss "github.com/ghodss/yaml"
	"gorm.io/gorm"
)

// SetupSpec is the fixture-frontmatter carrier for commons-db's shell.Setup: the
// dotenv files, env vars, cloud/k8s connections and git checkout a markdown file
// declares before any of its tests run.
//
// It exists only to route decoding through ghodss/yaml (YAML → JSON →
// encoding/json). Fixture frontmatter is parsed by goccy, but the connection
// types underneath Setup are tagged for JSON only — connection.KubernetesConnection
// spells its field `json:"connection"` with no yaml tag, and the embedded
// credentials are `json:",inline"`. Handed to goccy those bind by lowercased field
// name and nest rather than inline, so `connections:` would silently mis-decode.
// goccy honours the BytesUnmarshaler/BytesMarshaler interfaces, which is the seam
// this uses to hand the raw bytes to a decoder that reads json tags.
//
// A setup that reads `{{.params.…}}` or holds a `{$cel: …}` value is also kept
// as authored (raw): it is rendered once the params form answered. Under
// `connections:`, a string names that connection kind's reference
// (`kubernetes: connection://acme`), and a null leaves it unconfigured.
type SetupSpec struct {
	shell.Setup `json:",inline" yaml:",inline"`
	raw         json.RawMessage
}

// UnmarshalYAML implements goccy's BytesUnmarshaler.
func (s *SetupSpec) UnmarshalYAML(data []byte) error {
	encoded, err := ghodss.YAMLToJSON(data)
	if err != nil {
		return fmt.Errorf("setup: %w", err)
	}
	return s.UnmarshalJSON(encoded)
}

// UnmarshalJSON strictly decodes the setup, keeping it as authored when its
// values come from the params form.
func (s *SetupSpec) UnmarshalJSON(data []byte) error {
	tree, err := decodeOrderedJSON(data)
	if err != nil {
		return fmt.Errorf("setup: %w", err)
	}
	tree = expandConnectionShorthand(tree)
	encoded, err := encodeOrderedJSON(tree)
	if err != nil {
		return fmt.Errorf("setup: %w", err)
	}
	shape := tree
	if isDeferredSetup(tree, encoded) {
		s.raw = encoded
		shape = withoutDeferredValues(tree)
	}
	if s.Setup, err = decodeSetup(shape); err != nil {
		return fmt.Errorf("setup: %w", err)
	}
	return nil
}

// MarshalJSON is the setup as authored when it is rendered from the params
// form, so a worker it is handed to renders the same setup.
func (s SetupSpec) MarshalJSON() ([]byte, error) {
	if s.raw != nil {
		return s.raw, nil
	}
	return json.Marshal(s.Setup)
}

// MarshalYAML implements goccy's BytesMarshaler.
func (s SetupSpec) MarshalYAML() ([]byte, error) {
	encoded, err := s.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return ghodss.JSONToYAML(encoded)
}

// declares reports whether the authored setup sets any of keys at its top level.
func (s *SetupSpec) declares(keys ...string) (bool, error) {
	if s.raw == nil {
		return false, nil
	}
	tree, err := decodeOrderedJSON(s.raw)
	if err != nil {
		return false, err
	}
	root, _ := tree.(orderedObject)
	for _, member := range root {
		if member.Value != nil && slices.Contains(keys, member.Key) {
			return true, nil
		}
	}
	return false, nil
}

// applyWorktreeDefaults fills in the worktree defaults a fixture didn't spell out
// and returns any downgrade warnings the caller should log. It is the single place
// the JSON-schema `default:` values, the docs and the runtime agree.
//
// Order matters: `base` defaults to HEAD first, because whether `uncommitted`
// defaults to clone is derived from it — uncommitted work is a diff against *your*
// HEAD, so replaying it onto a worktree branched elsewhere applies to the wrong
// context.
func applyWorktreeDefaults(setup *shell.Setup) []string {
	if setup == nil || setup.Checkout == nil {
		return nil
	}
	return setup.Checkout.Worktree.ApplyDefaults()
}

// PreparedSetup is what one markdown file's `setup:` block produced: the
// directory its tests run in, the environment they inherit, and how to undo it.
//
// One per file, not one per test. Every test in a file shares this tree and they
// run concurrently in the same task group, so a worktree isolates the file from
// the rest of the repo — it does not isolate the file's tests from each other.
type PreparedSetup struct {
	File  string            // markdown file that declared it
	Cwd   string            // where the prepared tree landed
	Env   map[string]string // prepared environment, split on the first '='
	Extra map[string]any    // commit, worktree, path, dirtyFiles

	cleanup func() error
	once    sync.Once
}

// Dir returns where a command belonging to this setup should run, falling back
// to fallback when no setup relocated the run. Nil-safe.
func (p *PreparedSetup) Dir(fallback string) string {
	if p == nil || p.Cwd == "" {
		return fallback
	}
	return p.Cwd
}

// Environ returns the environment for a command belonging to this setup, or nil
// to inherit unchanged. It layers the prepared variables over os.Environ()
// rather than replacing it: os/exec and clicky/exec both start from the process
// environment, so a setup is additive and overriding, never hermetic.
func (p *PreparedSetup) Environ() []string {
	if p == nil || len(p.Env) == 0 {
		return nil
	}
	env := os.Environ()
	for k, v := range p.Env {
		env = append(env, k+"="+v)
	}
	return env
}

// Cleanup tears the setup down exactly once, however it is reached — the
// runner's defer on the happy path, the shutdown hook when the process is
// interrupted, and whichever loses the race does nothing.
func (p *PreparedSetup) Cleanup() error {
	if p == nil {
		return nil
	}
	var err error
	p.once.Do(func() {
		if p.cleanup != nil {
			err = p.cleanup()
		}
	})
	return err
}

// SetupDBProvider supplies the database that `connection://…` references in a
// fixture's `setup:` resolve against. Left nil — the default, and what every
// test binary gets — those references fail loud instead of silently resolving to
// nothing. It is a hook so this package never imports internal/database.
var SetupDBProvider func(context.Context) *gorm.DB

// prepareSetup resolves one file's setup against the directory its markdown
// lives in and hands it to commons-db. A non-nil catalog resolves the setup's
// connection references (a params-reading setup, see StepConnectionCatalog);
// nil resolves them against the process database.
func prepareSetup(ctx flanksourceContext.Context, mdPath, sourceDir string, spec *SetupSpec, catalog *dbcontext.Context) (*PreparedSetup, error) {
	resolved, err := spec.Resolve(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("%s: setup: resolve: %w", mdPath, err)
	}

	// Never leave Cwd blank. commons-db's defaultCommandDir invents a scratch
	// <baseDir>/tmp/<uuid> when it is, so a setup that declares nothing but
	// `dotenv:` would silently relocate the file's tests into an empty directory.
	if resolved.Cwd == "" {
		resolved.Cwd = sourceDir
	}
	if spec.BaseDir == "" {
		base, err := setupBaseDir(mdPath)
		if err != nil {
			return nil, fmt.Errorf("%s: setup: baseDir: %w", mdPath, err)
		}
		resolved.BaseDir = base
	}
	for _, warning := range applyWorktreeDefaults(&resolved) {
		logger.Warnf("%s: setup: %s", mdPath, warning)
	}

	if catalog != nil {
		return finishSetup(mdPath, &resolved, *catalog)
	}
	dbCtx := dbcontext.NewContext(ctx)
	var db *gorm.DB
	if SetupDBProvider != nil {
		if db = SetupDBProvider(ctx); db != nil {
			dbCtx = dbCtx.WithDB(db, nil)
		}
	}
	if db == nil && setupReferencesConnection(&resolved) {
		return nil, fmt.Errorf("%s: setup: resolving a connection:// reference needs a database, but none is configured", mdPath)
	}

	if SetupConnectionResolver != nil {
		dbCtx = dbCtx.WithConnectionResolver(SetupConnectionResolver)
	}
	return finishSetup(mdPath, &resolved, dbCtx)
}

// finishSetup prepares a resolved setup in dbCtx and registers its teardown.
func finishSetup(mdPath string, resolved *shell.Setup, dbCtx dbcontext.Context) (*PreparedSetup, error) {
	result, err := shell.Prepare(dbCtx, resolved)
	if err != nil {
		return nil, fmt.Errorf("%s: setup: prepare: %w", mdPath, err)
	}

	prepared := &PreparedSetup{
		File:    mdPath,
		Cwd:     result.Cwd,
		Env:     splitEnv(result.Env),
		Extra:   result.Extra,
		cleanup: result.Cleanup,
	}
	// A worktree outlives a SIGINT unless something removes it, and the runner's
	// defer never runs on one. PriorityWorkers matches the git clone manager's
	// hook, so both tear down before the database closes.
	shutdown.AddHookWithPriority("fixture setup "+mdPath, shutdown.PriorityWorkers, func() {
		if err := prepared.Cleanup(); err != nil {
			logger.Warnf("%s: setup cleanup failed: %v", mdPath, err)
		}
	})
	return prepared, nil
}

// setupBaseDir is where a file's clones and worktrees land when it didn't say.
// commons-db would otherwise default to `<sourceDir>/.shell`, which writes into
// whatever repository the markdown happens to live in. Keying the cache
// directory on the markdown path keeps concurrent files off each other's
// worktrees while staying stable across runs of the same file.
func setupBaseDir(mdPath string) (string, error) {
	cache, err := status.CacheDir()
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(mdPath)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(abs))
	return filepath.Join(cache, "fixtures", hex.EncodeToString(sum[:8])), nil
}

// setupReferencesConnection reports whether the setup names a stored connection
// anywhere — under `connections:`, or as a checkout's credentials. Without this
// the failure surfaces from deep inside commons-db as a bare "db is not
// configured" with nothing to tie it back to a fixture.
//
// It recognises the `connection://` form only; a reference spelled as a bare
// UUID falls through to commons-db's own error.
func setupReferencesConnection(setup *shell.Setup) bool {
	raw, err := json.Marshal(setup)
	if err != nil {
		return false
	}
	return bytes.Contains(raw, []byte("connection://"))
}

// splitEnv turns commons-db's `KEY=value` slice into a map. Later entries win,
// matching how a process resolves duplicate keys in its environment.
func splitEnv(env []string) map[string]string {
	if len(env) == 0 {
		return nil
	}
	out := make(map[string]string, len(env))
	for _, kv := range env {
		if key, value, found := strings.Cut(kv, "="); found {
			out[key] = value
		}
	}
	return out
}

// fixtureEnv is the per-file runtime environment a fixture executes in: the
// file it came from, its prepared `setup:` and the step scope its steps share.
type fixtureEnv struct {
	file  string
	setup *PreparedSetup
	scope *StepScope
}

// envForNode returns the per-file runtime environment a node executes in.
func (r *Runner) envForNode(node *FixtureNode) fixtureEnv {
	file := nodeFile(node)
	return fixtureEnv{file: file, setup: r.setups.get(file)}
}
