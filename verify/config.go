package verify

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/collections"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/todos/types"
	yamlv3 "gopkg.in/yaml.v3"
)

// DefaultVerifyMode is a CONSTRAINT, not a preference, which is why it is the
// one runtime value still pinned in code: the grader that marks a definition of
// done is told to inspect the change with its own tools, so it must run on an
// agentic mechanism. A grader that cannot read the diff still returns confident
// verdicts, so letting configuration select an API mode here would silently
// produce authoritative-looking nonsense. It pins the mechanism only — the model
// itself comes from configuration like every other operation's.
const DefaultVerifyMode = api.ModeAgent

// DefaultAIConfig is the built-in base spec. It deliberately names NO model.
//
// It used to seed "claude-haiku-4-5", which meant every repo looked configured
// whether or not anyone had chosen a model, and made the per-operation defaults
// downstream unreachable. The model now comes from .gavel.yaml, then
// ~/.captain.yaml, and a run that finds neither fails loudly telling the user to
// run `gavel configure` or `captain configure`.
func DefaultAIConfig() api.Spec {
	return api.Spec{}
}

// DefaultGavelConfig seeds the built-in defaults every loaded config layers on.
func DefaultGavelConfig() GavelConfig {
	return GavelConfig{
		AI:    DefaultAIConfig(),
		Todos: TodosConfig{Verify: api.Spec{Model: api.Model{Mode: DefaultVerifyMode}}},
	}
}

// RestartPolicy is the supervisor restart policy for a process. It accepts a
// bool (true→on-failure, false→no) or an enum string (no|on-failure|always)
// from both .gavel.yaml (JSON via ghodss/yaml) and the Procfile (yaml.v3). The
// empty value means "unset" so the resolver applies the default (no).
type RestartPolicy string

const (
	RestartNo        RestartPolicy = "no"
	RestartOnFailure RestartPolicy = "on-failure"
	RestartAlways    RestartPolicy = "always"
)

func (p RestartPolicy) String() string { return string(p) }

// restartPolicyFromString validates an enum string, mapping "" to unset.
func restartPolicyFromString(s string) (RestartPolicy, error) {
	switch RestartPolicy(s) {
	case "", RestartNo, RestartOnFailure, RestartAlways:
		return RestartPolicy(s), nil
	default:
		return "", fmt.Errorf("invalid restart policy %q (want no, on-failure, always, or a bool)", s)
	}
}

func (p *RestartPolicy) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*p = ""
		return nil
	}
	if bytes.Equal(data, []byte("true")) {
		*p = RestartOnFailure
		return nil
	}
	if bytes.Equal(data, []byte("false")) {
		*p = RestartNo
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	v, err := restartPolicyFromString(s)
	if err != nil {
		return err
	}
	*p = v
	return nil
}

func (p *RestartPolicy) UnmarshalYAML(node *yamlv3.Node) error {
	var b bool
	if err := node.Decode(&b); err == nil {
		if b {
			*p = RestartOnFailure
		} else {
			*p = RestartNo
		}
		return nil
	}
	var s string
	if err := node.Decode(&s); err != nil {
		return err
	}
	v, err := restartPolicyFromString(s)
	if err != nil {
		return err
	}
	*p = v
	return nil
}

// LintIgnoreRule suppresses lint violations matching every populated field.
// Rule and Source accept collections.MatchItem patterns: literal strings,
// `*` globs ("acme-*"), and `!`-prefixed negations. File uses doublestar
// globs so directory traversal patterns ("pkg/**/*.go") behave as expected.
// At least one of Rule/Source/File must be set; an empty rule never matches.
type LintIgnoreRule struct {
	Rule   string `yaml:"rule,omitempty" json:"rule,omitempty"`
	Source string `yaml:"source,omitempty" json:"source,omitempty"`
	File   string `yaml:"file,omitempty" json:"file,omitempty"`
}

func (r LintIgnoreRule) MatchesViolation(v models.Violation) bool {
	if r.Source != "" {
		matched, negated := collections.MatchItem(v.Source, r.Source)
		if negated || !matched {
			return false
		}
	}
	if r.Rule != "" {
		if v.Rule == nil {
			return false
		}
		matched, negated := collections.MatchItem(v.Rule.Method, r.Rule)
		if negated || !matched {
			return false
		}
	}
	if r.File != "" {
		matched, _ := doublestar.Match(r.File, v.File)
		if !matched {
			return false
		}
	}
	return r.Rule != "" || r.Source != "" || r.File != ""
}

type LintConfig struct {
	// AIFix is the AI spec used to repair lint violations. It is independent of
	// commit.message because source repair needs an editing-capable agent while
	// commit-message generation only needs to summarize a diff.
	Fix PromptSpec `yaml:"fix,omitempty" json:"fix,omitempty"`
	// Ignore accumulates across config layers: a suppression is a standing
	// statement about this repo, so adding one must not revoke the home config's.
	Ignore  []LintIgnoreRule            `yaml:"ignore,omitempty" json:"ignore,omitempty" merge:"append"`
	Linters map[string]LintLinterConfig `yaml:"linters,omitempty" json:"linters,omitempty"`
}

type LintLinterConfig struct {
	Enabled *bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

func (c LintConfig) IsLinterEnabled(name string, defaultEnabled bool) bool {
	if c.Linters == nil {
		return defaultEnabled
	}
	cfg, ok := c.Linters[name]
	if !ok || cfg.Enabled == nil {
		return defaultEnabled
	}
	return *cfg.Enabled
}

type CommitHook struct {
	Name  string   `yaml:"name" json:"name"`
	Run   string   `yaml:"run" json:"run"`
	Files []string `yaml:"files,omitempty" json:"files,omitempty"`
}

type CheckMode string

func (m CheckMode) String() string {
	return string(m)
}

func (m *CheckMode) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*m = ""
		return nil
	}
	if bytes.Equal(data, []byte("false")) {
		*m = "skip"
		return nil
	}
	if bytes.Equal(data, []byte("true")) {
		return fmt.Errorf("expected mode string or false, got true")
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*m = CheckMode(s)
	return nil
}

type CommitConfig struct {
	// Hooks accumulate across config layers, in execution order. GitIgnore and
	// Allow accumulate too and collapse repeats — naming the same path in two
	// configs says nothing more than naming it once.
	Hooks     []CommitHook     `yaml:"hooks,omitempty" json:"hooks,omitempty" merge:"append"`
	GitIgnore []string         `yaml:"gitignore,omitempty" json:"gitignore,omitempty" merge:"append,unique"`
	Allow     []string         `yaml:"allow,omitempty" json:"allow,omitempty" merge:"append,unique"`
	Precommit PrecommitConfig  `yaml:"precommit,omitempty" json:"precommit,omitempty"`
	Lint      CommitLintConfig `yaml:"lint,omitempty" json:"lint,omitempty"`
	Tidy      CommitTidyConfig `yaml:"tidy,omitempty" json:"tidy,omitempty"`
	// Message, Grouping, and Summary are the AI specs (model/prompt/budget/effort)
	// for commit-message generation, AI commit grouping (`gavel commit -G`), and
	// commit-group summaries (`gavel git analyze --summary`). Each overrides the
	// base ai: spec field-wise; empty uses the built-in default.
	Message  PromptSpec `yaml:"message,omitempty" json:"message,omitempty"`
	Grouping PromptSpec `yaml:"grouping,omitempty" json:"grouping,omitempty"`
	Summary  PromptSpec `yaml:"summary,omitempty" json:"summary,omitempty"`
	// MaxCommits caps the number of commits AI grouping may produce (0 = default).
	MaxCommits int `yaml:"maxCommits,omitempty" json:"maxCommits,omitempty"`
	// Types is the set of conventional-commit types AI message generation may
	// choose from; it becomes the generated message's `type:` enum, so the model
	// cannot invent one. Empty uses models.SelectableCommitTypes(). Unlike the
	// lists above this replaces rather than appends — naming types here narrows
	// the vocabulary, and appending would make narrowing impossible.
	Types []string `yaml:"types,omitempty" json:"types,omitempty"`
}

// CommitTidyConfig controls whether `gavel commit` runs `go mod tidy` in every
// Go module in the repo before committing and stages any go.mod / go.sum
// updates into the in-flight commit. Enabled is on by default (nil = on);
// set to false to disable. CLI flag --tidy overrides per-invocation.
type CommitTidyConfig struct {
	Enabled *bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

// CommitLintConfig controls whether `gavel commit` runs linters over the
// staged file set before creating the commit. Two independent gates:
//
//   - Enabled toggles every non-secrets linter. nil = off (default), true = on.
//   - Secrets toggles the betterleaks/secrets linter. nil = on (default —
//     secrets are the highest-value pre-commit check), true = on, false = off.
//
// CLI flags --lint and --lint-secrets override these per-invocation.
type CommitLintConfig struct {
	Enabled *bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Secrets *bool `yaml:"secrets,omitempty" json:"secrets,omitempty"`
}

// PrecommitConfig configures the combined pre-commit gate for commit.gitignore
// and linked dependency checks. Mode is "prompt" (default), "fail", "skip",
// or false (alias for skip).
type PrecommitConfig struct {
	Mode CheckMode `yaml:"mode,omitempty" json:"mode,omitempty"`
}

// HookStep is a single shell command rendered into the SSH post-receive hook.
// Used by top-level Pre/Post in GavelConfig.
type HookStep struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Run  string `yaml:"run" json:"run"`
}

// SSHConfig overrides the main command run by the SSH post-receive hook.
// When Cmd is empty, the hook falls back to `gavel test --lint`.
type SSHConfig struct {
	Cmd string `yaml:"cmd,omitempty" json:"cmd,omitempty"`
}

// DefaultFixturesGlob is the default glob pattern used to discover fixture files.
const DefaultFixturesGlob = "**/*.fixture.md"

type FixturesConfig struct {
	Enabled bool     `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Files   []string `yaml:"files,omitempty" json:"files,omitempty"`
}

// ResolvedFiles returns the configured globs, falling back to the default when none are set.
func (f FixturesConfig) ResolvedFiles() []string {
	if len(f.Files) > 0 {
		return f.Files
	}
	return []string{DefaultFixturesGlob}
}

type GavelConfig struct {
	// AI is the base spec (model/fallbacks/budget/effort defaults) that every AI
	// operation inherits and overrides field-wise. See DefaultAIConfig.
	AI       api.Spec       `yaml:"ai,omitempty" json:"ai,omitempty"`
	Lint     LintConfig     `yaml:"lint,omitempty" json:"lint,omitempty"`
	Commit   CommitConfig   `yaml:"commit,omitempty" json:"commit,omitempty"`
	Fixtures FixturesConfig `yaml:"fixtures,omitempty" json:"fixtures,omitempty"`
	SSH      SSHConfig      `yaml:"ssh,omitempty" json:"ssh,omitempty"`
	// Pre and Post accumulate across config layers, in execution order: a repo
	// config adds its steps to the home config's rather than replacing them.
	Pre      []HookStep     `yaml:"pre,omitempty" json:"pre,omitempty" merge:"append"`
	Post     []HookStep     `yaml:"post,omitempty" json:"post,omitempty" merge:"append"`
	Secrets  SecretsConfig  `yaml:"secrets,omitempty" json:"secrets,omitempty"`
	Procfile ProcfileConfig `yaml:"procfile,omitempty" json:"procfile,omitempty"`
	// Checks is the project default for the `yaml test`/`yaml lint` fences the
	// lifecycle host appends to a todo's definition of done; the run step's
	// verify loop re-runs them until they pass. A TODO's frontmatter `checks:`
	// overrides it.
	Checks types.AgentChecksConfig `yaml:"checks,omitempty" json:"checks,omitempty"`
	Todos  TodosConfig             `yaml:"todos,omitempty" json:"todos,omitempty"`
	Status StatusConfig            `yaml:"status,omitempty" json:"status,omitempty"`
	Test   TestConfig              `yaml:"test,omitempty" json:"test,omitempty"`
	PR     PRConfig                `yaml:"pr,omitempty" json:"pr,omitempty"`
}

// PRConfig configures `gavel pr`.
type PRConfig struct {
	// Content is the AI spec for generating the PR title, body, and branch name.
	Content PromptSpec `yaml:"content,omitempty" json:"content,omitempty"`
	// Fix is the AI spec driving `gavel pr status --ai-fix`. Its
	// workflow.verify.commands are the loop's definition of done (re-polling
	// `gavel pr status`) and workflow.commits the per-turn commit policy.
	// See prompts.PRFix.
	Fix PromptSpec `yaml:"fix,omitempty" json:"fix,omitempty"`
	// Base is the base branch for the PR (e.g. origin/main).
	Base string `yaml:"base,omitempty" json:"base,omitempty"`
	// Draft opens the PR as a draft.
	Draft bool `yaml:"draft,omitempty" json:"draft,omitempty"`
}

// StatusConfig configures `gavel status`.
type StatusConfig struct {
	// Summary is the AI spec for the per-file summary used by `gavel status --ai`.
	// Empty uses the built-in default. See prompts.StatusSummary.
	Summary PromptSpec `yaml:"summary,omitempty" json:"summary,omitempty"`
}

// TestConfig configures `gavel test`.
type TestConfig struct {
	// OutlineSummary is the AI spec for the per-test summary used by
	// `gavel test outline --ai-summary`. Empty uses the built-in default. See
	// prompts.TestOutlineSummary.
	OutlineSummary PromptSpec `yaml:"outlineSummary,omitempty" json:"outlineSummary,omitempty"`
	// Timeout, TestTimeout and LintTimeout are Go duration strings ("20m")
	// carrying the deadlines `gavel test` otherwise takes from its flag
	// defaults. They live in config because the deadline a repo needs is a
	// property of that repo's suites, not of the invocation: CI, a local run and
	// the SSH git-push backend all have to agree on it, and the backend runs
	// `gavel test` with no timeout flags of its own.
	//
	// An explicitly passed flag still wins, so a one-off run can shorten or
	// lengthen a deadline without editing the file.
	Timeout     string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	TestTimeout string `yaml:"testTimeout,omitempty" json:"testTimeout,omitempty"`
	LintTimeout string `yaml:"lintTimeout,omitempty" json:"lintTimeout,omitempty"`
}

// ProcfileConfig configures `gavel proc` — global defaults for the processes
// declared in the Procfile (per-process settings live in the Procfile itself).
// Every field is optional.
type ProcfileConfig struct {
	// Profile is the default active profile: a Procfile entry with `profiles`
	// auto-starts only when one of them is active. `gavel proc --profile` overrides.
	Profile string `yaml:"profile,omitempty" json:"profile,omitempty"`
	// AutoRestart is the default restart policy for every process (bool or enum).
	AutoRestart RestartPolicy `yaml:"autoRestart,omitempty" json:"autoRestart,omitempty"`
	// MaxRestarts caps automatic restarts per process (0 = unlimited).
	MaxRestarts int `yaml:"maxRestarts,omitempty" json:"maxRestarts,omitempty"`
	// Env is injected into every process on top of the parent environment and
	// any sibling .env file.
	Env map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	// Mem caps resident memory per process (e.g. "512Mi", "2g"). Empty disables
	// the limit. A process whose tree exceeds it is killed.
	Mem string `yaml:"mem,omitempty" json:"mem,omitempty"`
	// CPU caps sustained CPU per process as a percentage (100 = one full core).
	// 0 disables the limit. A process that stays above it is killed.
	CPU float64 `yaml:"cpu,omitempty" json:"cpu,omitempty"`
}

// SecretsConfig turns the betterleaks linter on/off and optionally points at
// extra betterleaks/gitleaks TOML configs beyond the ones gavel discovers
// from the home dir, git root, and cwd. Rule authoring lives in those TOML
// files, not here — gavel only orchestrates discovery + merge.
type SecretsConfig struct {
	// Disabled turns off the betterleaks linter even when the binary is on
	// PATH. Defaults to false (enabled).
	Disabled bool `yaml:"disabled,omitempty" json:"disabled,omitempty"`
	// Configs is an optional list of additional .betterleaks.toml /
	// .gitleaks.toml paths to merge in (relative paths resolve against the
	// .gavel.yaml's directory). Layers accumulate and repeats collapse.
	Configs []string `yaml:"configs,omitempty" json:"configs,omitempty" merge:"append,unique"`
}
