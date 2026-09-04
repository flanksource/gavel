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
	"strings"
	"sync"

	"github.com/flanksource/clicky/shutdown"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/shell"
	flanksourceContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
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
type SetupSpec struct {
	shell.Setup `json:",inline" yaml:",inline"`
}

// UnmarshalYAML implements goccy's BytesUnmarshaler.
func (s *SetupSpec) UnmarshalYAML(data []byte) error {
	return ghodss.Unmarshal(data, &s.Setup)
}

// MarshalYAML implements goccy's BytesMarshaler.
func (s SetupSpec) MarshalYAML() ([]byte, error) {
	return ghodss.Marshal(s.Setup)
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
// nothing. It mirrors the AIStepRunner hook seam so this package never imports
// internal/database.
var SetupDBProvider func(context.Context) *gorm.DB

// prepareSetups prepares each markdown file's `setup:` block, once, keyed on the
// file that declared it.
//
// It runs before `build:` deliberately: a setup can relocate the run into a
// worktree, and a build that ran in the original repo would build tree A while
// the tests exercise tree B — and pass.
//
// A file whose tests were all filtered out contributes no node to walk and so
// gets no setup prepared, which is the behaviour you want: `--filter` should not
// clone a repo it has no tests for.
func (r *Runner) prepareSetups(ctx flanksourceContext.Context) error {
	for _, file := range r.tree.Children {
		spec, mdPath, sourceDir := fileSetup(file)
		if spec == nil {
			continue
		}
		prepared, err := prepareSetup(ctx, mdPath, sourceDir, spec)
		if err != nil {
			return err
		}
		if r.setups == nil {
			r.setups = make(map[string]*PreparedSetup)
		}
		r.setups[mdPath] = prepared
	}
	return nil
}

// cleanupSetups tears every prepared setup down. Failures are reported rather
// than returned — the fixture results are the run's answer, and a worktree that
// would not go away must not turn a passing run red.
func (r *Runner) cleanupSetups() {
	for file, prepared := range r.setups {
		if err := prepared.Cleanup(); err != nil {
			logger.Warnf("%s: setup cleanup failed: %v", file, err)
		}
	}
}

// envForNode returns the per-file runtime environment a node executes in.
func (r *Runner) envForNode(node *FixtureNode) fixtureEnv {
	if node == nil || node.Origin == nil {
		return fixtureEnv{}
	}
	return fixtureEnv{file: node.Origin.File, setup: r.setups[node.Origin.File]}
}

// fileSetup finds the `setup:` a file declared, along with the markdown path to
// key it on and the directory to anchor its relative paths against.
//
// Frontmatter is file-level, so it is reachable only through the tests that
// carry it — FixtureNode has no frontmatter of its own. The first test wins
// because every test in a file shares the same parsed frontmatter.
func fileSetup(file *FixtureNode) (spec *SetupSpec, mdPath, sourceDir string) {
	file.Walk(func(node *FixtureNode) {
		if spec != nil || node.Test == nil || node.Test.Setup == nil {
			return
		}
		spec = node.Test.Setup
		sourceDir = node.Test.SourceDir
		if node.Origin != nil {
			mdPath = node.Origin.File
		}
	})
	return spec, mdPath, sourceDir
}

// prepareSetup resolves one file's setup against the directory its markdown
// lives in and hands it to commons-db.
func prepareSetup(ctx flanksourceContext.Context, mdPath, sourceDir string, spec *SetupSpec) (*PreparedSetup, error) {
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

	result, err := shell.Prepare(dbCtx, &resolved)
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
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(mdPath)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(abs))
	return filepath.Join(cache, "gavel", "fixtures", hex.EncodeToString(sum[:8])), nil
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
