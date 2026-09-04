package todos

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/fixtures"

	// Registers the fixture engine's `yaml test` / `yaml lint` step runners
	// (fixtures.TestStepRunner / LintStepRunner) via its init.
	_ "github.com/flanksource/gavel/fixtures/types"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/repomap"
	"github.com/goccy/go-yaml"
)

// The run loop's definition of done is one gavel fixture document: the
// configured `checks` test/lint steps, the todo's `## Verification` section and
// its acceptance criteria, rendered as markdown and stamped onto
// api.Workflow.Verify.Fixture.
//
// Captain dispatches that declaration to whatever claimed the `fixture` verifier
// kind — gavel/fixtures/verifier in this process, `gavel fixtures run` out of
// it. Rendering the whole definition of done as one document rather than as a
// bespoke in-process verifier is what makes those two paths the same code: an
// external runner receives nothing but this markdown on stdin, so anything kept
// beside the document would not reach it.

// DefinitionOfDoneOptions describes the todo whose definition of done is
// rendered, and what grades its acceptance criteria. The implementer's own run
// spec has no say here: a run may be a cmux TUI resuming a long session, and
// none of that may grade its own work. Whether the document runs, and how many
// rounds, is the lifecycle step's — see its workflow.verify.
type DefinitionOfDoneOptions struct {
	WorkDir string
	Todos   []*types.TODO
	// Grader is the resolved verification spec — request > .gavel.yaml
	// todos.verify > ai: — carrying no session of its own.
	Grader api.Spec
}

// DefinitionOfDone is a run's executable definition of done.
type DefinitionOfDone struct {
	// Fixture is the markdown document, ready to stamp onto Workflow.Verify.
	// Empty when the todo declares no definition of done at all — such a run
	// ends `completed` rather than verified.
	Fixture string
	// Warnings names what was dropped while building the document: front-matter
	// keys the todo set that the generated document has no field for. The
	// document is still usable, which is why these are warnings — see
	// adoptVerificationFrontMatter.
	Warnings []string
}

// Declared reports whether there is anything to verify.
func (d DefinitionOfDone) Declared() bool { return strings.TrimSpace(d.Fixture) != "" }

// dodFrontMatter is the generated document's front matter: where and how its
// steps run, the retry predicate the verifier evaluates, and — only when the
// todo has acceptance criteria — the `ai:` block whose presence turns the
// document's task list into a graded checklist step.
//
// Its fields are exactly the front-matter keys a todo's `## Verification`
// section may set: each one is honoured by the fixture engine's node runner,
// which is what executes this document. A key with no field here is one the
// runner would silently ignore, and adoptVerificationFrontMatter refuses it.
type dodFrontMatter struct {
	CodeBlocks []string                      `yaml:"codeBlocks,omitempty"`
	CWD        string                        `yaml:"cwd,omitempty"`
	Exec       string                        `yaml:"exec,omitempty"`
	Args       []string                      `yaml:"args,omitempty"`
	Env        map[string]any                `yaml:"env,omitempty"`
	Terminal   string                        `yaml:"terminal,omitempty"`
	OS         string                        `yaml:"os,omitempty"`
	Arch       string                        `yaml:"arch,omitempty"`
	Skip       string                        `yaml:"skip,omitempty"`
	AI         *fixtures.FixtureAIConfig     `yaml:"ai,omitempty"`
	Verify     *fixtures.FixtureVerifyConfig `yaml:"verify,omitempty"`
}

// BuildDefinitionOfDone renders the run's definition of done. It returns an
// undeclared DefinitionOfDone only when the todo has no definition of done at
// all.
func BuildDefinitionOfDone(in DefinitionOfDoneOptions) (DefinitionOfDone, error) {
	gitRoot := checksWorkDirFor(in.WorkDir, in.Todos)
	project, err := verify.LoadGavelConfig(gitRoot)
	if err != nil {
		return DefinitionOfDone{}, fmt.Errorf("checks: load .gavel.yaml: %w", err)
	}
	cfg := types.ResolveAgentChecks(project.Checks, firstChecksConfig(in.Todos))

	var sections []string
	// The `checks` test/lint suite is opt-in: `.gavel.yaml checks.enabled` or
	// the todo's own `checks:` front matter turns it on.
	if cfg.IsEnabled() {
		if cfg.Test != nil {
			section, err := runnerStepSection("checks:test", fixtures.RunnerKindTest, cfg.Test, gitRoot)
			if err != nil {
				return DefinitionOfDone{}, err
			}
			sections = append(sections, section)
		}
		if cfg.Lint != nil {
			section, err := runnerStepSection("checks:lint", fixtures.RunnerKindLint, cfg.Lint, gitRoot)
			if err != nil {
				return DefinitionOfDone{}, err
			}
			sections = append(sections, section)
		}
	}

	// The todo's own `## Verification` fixture is the definition of done — it
	// gates the loop whenever present, independent of the `checks` toggle. Its
	// front matter is lifted into the generated document's, because a fixture
	// document has exactly one and the todo's is the one that knows how its own
	// steps want to run.
	front := dodFrontMatter{CWD: gitRoot}
	var criteria []string
	var warnings []string
	for _, todo := range in.Todos {
		if todo == nil {
			continue
		}
		body, dropped, err := adoptVerificationFrontMatter(&front, todo.VerificationMarkdown)
		if err != nil {
			return DefinitionOfDone{}, err
		}
		for _, warning := range dropped {
			// Named, because several todos can contribute to one document and a
			// warning nobody can trace back to a record is one nobody will fix.
			warnings = append(warnings, fmt.Sprintf("todo %s: %s", todo.DisplayID(), warning))
			logger.Warnf("definition of done: todo %s: %s", todo.DisplayID(), warning)
		}
		if body != "" {
			sections = append(sections, body)
		}
		for _, c := range todo.AcceptanceCriteria {
			criteria = append(criteria, c.Text)
		}
	}

	// Acceptance criteria become the document's task list. The `ai:` block is
	// what makes the parser read that list as a graded checklist step, so it is
	// declared only when there is something to grade — an `ai:` with no criteria
	// is a step that skips itself on every iteration.
	if len(criteria) > 0 {
		sections = append(sections, criteriaSection(criteria))
		front.AI = graderAIConfig(in.Grader)
	}
	// The warnings ride along even on the paths that produce no document: a todo
	// whose front matter was ignored should hear about it whether or not the rest
	// of it amounted to something executable.
	if len(sections) == 0 {
		return DefinitionOfDone{Warnings: warnings}, nil
	}
	front.Verify = &fixtures.FixtureVerifyConfig{Retry: cfg.Retry}

	document, err := renderDefinitionOfDone(front, sections)
	if err != nil {
		return DefinitionOfDone{}, err
	}
	// A `## Verification` section can be prose — notes on how a human would check
	// the change — which declares no executable step. Rendering it as a document
	// anyway would give the loop a definition of done that can never go green,
	// so a document that parses to no steps is no definition of done at all.
	steps, err := fixtures.ParseMarkdownDocument("definition-of-done", document, gitRoot)
	if err != nil {
		return DefinitionOfDone{}, fmt.Errorf("definition of done: %w", err)
	}
	if types.CountTests(steps.Children) == 0 {
		return DefinitionOfDone{Warnings: warnings}, nil
	}
	return DefinitionOfDone{Fixture: document, Warnings: warnings}, nil
}

func renderDefinitionOfDone(front dodFrontMatter, sections []string) (string, error) {
	header, err := yaml.Marshal(front)
	if err != nil {
		return "", fmt.Errorf("definition of done: marshal front matter: %w", err)
	}
	var b strings.Builder
	b.WriteString("---\n")
	b.Write(header)
	b.WriteString("---\n\n# Definition of done\n\n")
	b.WriteString(strings.Join(sections, "\n\n"))
	b.WriteString("\n")
	return b.String(), nil
}

// adoptVerificationFrontMatter splits a todo's verification markdown, folds its
// front matter into the generated document's, and returns the body plus the keys
// it had to drop.
//
// A key the generated document cannot honour is dropped with a warning rather
// than refused. `setup:`, `record:`, `build:`, `daemon:` and `files:` are
// prepared by the whole-file fixture runner and never by the node runner this
// document executes under, so a todo that set one has steps running in an
// environment it did not describe; `ai:` and `verify:` belong to the generated
// document itself (the criteria grader and the retry predicate), so a todo's own
// would take over the whole run's definition of done. Neither is what the author
// meant, and both are worth saying out loud.
//
// Refusing outright was worse. This runs inside lifecycle evaluation, so the
// error did not surface as "your verification front matter has a bad key" — it
// surfaced as "Failed to load todo <uuid>", and the todo could not be opened,
// listed or edited. The one action that fixes the front matter was the one
// action the error prevented. A warning keeps the record reachable and still
// says exactly what was ignored and why.
func adoptVerificationFrontMatter(front *dodFrontMatter, markdown string) (string, []string, error) {
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return "", nil, nil
	}
	parsed, body, err := fixtures.SplitFrontMatter(markdown)
	if err != nil {
		return "", nil, fmt.Errorf("definition of done: verification front matter: %w", err)
	}
	if parsed == nil {
		return strings.TrimSpace(body), nil, nil
	}
	var warnings []string
	if unhonoured := unhonouredVerificationKeys(parsed); len(unhonoured) > 0 {
		warnings = append(warnings, fmt.Sprintf(
			"ignoring %s in the todo's verification front matter: the run loop cannot honour %s; "+
				"it accepts codeBlocks, cwd, exec, args, env, terminal, os, arch and skip "+
				"(acceptance criteria and the retry predicate are the generated document's own, not an ai:/verify: block)",
			strings.Join(unhonoured, ", "), pluralKeys(len(unhonoured))))
	}
	if len(parsed.CodeBlocks) > 0 {
		front.CodeBlocks = parsed.CodeBlocks
	}
	if parsed.CWD != "" {
		front.CWD = parsed.CWD
	}
	front.Exec, front.Args, front.Env, front.Terminal = parsed.Exec, parsed.Args, parsed.Env, parsed.Terminal
	front.OS, front.Arch, front.Skip = parsed.OS, parsed.Arch, parsed.Skip
	return strings.TrimSpace(body), warnings, nil
}

func pluralKeys(n int) string {
	if n == 1 {
		return "it"
	}
	return "them"
}

// unhonouredVerificationKeys names the front-matter keys a todo's verification
// section set that the generated document has no field for, in document order.
func unhonouredVerificationKeys(parsed *fixtures.FrontMatter) []string {
	var keys []string
	for _, key := range []struct {
		name string
		set  bool
	}{
		{"setup", parsed.Setup != nil},
		{"record", parsed.Record != nil},
		{"build", parsed.Build != ""},
		{"daemon", parsed.Daemon != ""},
		{"files", parsed.Files != ""},
		{"timeout", parsed.Timeout != nil},
		{"ai", parsed.AI != nil},
		{"verify", parsed.Verify != nil},
	} {
		if key.set {
			keys = append(keys, key.name)
		}
	}
	parsed.CleanMetadata()
	extra := make([]string, 0, len(parsed.Metadata))
	for key := range parsed.Metadata {
		extra = append(extra, key)
	}
	sort.Strings(extra)
	return append(keys, extra...)
}

// criteriaHeading is the section the acceptance criteria are rendered under, and
// the section the AI step is told to read them from. Scoping matters here: the
// same document holds the todo's own fixture steps, whose `- cel:` expectation
// bullets would otherwise be adopted as the checklist's own expectation.
const criteriaHeading = "Acceptance Criteria"

// graderAIConfig carries the resolved verification model into the document. It
// is the only part of the grading runtime the document can hold, and therefore
// the only part an external fixture runner ever sees; a document with its own
// `ai:` front matter still outranks it inside the fixture engine.
func graderAIConfig(grader api.Spec) *fixtures.FixtureAIConfig {
	cfg := &fixtures.FixtureAIConfig{CriteriaSection: criteriaHeading}
	if name := strings.TrimSpace(grader.Name); name != "" {
		cfg.Model = name
	}
	return cfg
}

// criteriaSection renders the acceptance criteria as a task list under
// criteriaHeading — the shape fixtures.ExtractChecklist reads them back out of.
func criteriaSection(criteria []string) string {
	var b strings.Builder
	b.WriteString("## " + criteriaHeading + "\n")
	for _, item := range criteria {
		b.WriteString("\n- [ ] " + strings.TrimSpace(item))
	}
	return b.String()
}

// runnerStepSection renders one configured check as a `yaml test` / `yaml lint`
// fence — the same wire contract a hand-written fixture file uses. work-dir is
// stamped explicitly so the suite runs at the git root the config was resolved
// against, not wherever the agent's turn happened to leave the run.
func runnerStepSection(name, kind string, options any, workDir string) (string, error) {
	body, err := yaml.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("%s: marshal step options: %w", name, err)
	}
	config := strings.TrimRight(string(body), "\n")
	if config == "{}" {
		config = ""
	}
	if config != "" {
		config += "\n"
	}
	return fmt.Sprintf("## %s\n\n```yaml %s\n%swork-dir: %s\n```", name, kind, config, workDir), nil
}

// checksWorkDirFor resolves the directory checks run in: the git root of the
// group's working directory, mirroring how the commit step derives its dir.
func checksWorkDirFor(workDir string, todosInGroup []*types.TODO) string {
	dir := workDir
	for _, todo := range todosInGroup {
		if todo == nil || todo.CWD == "" {
			continue
		}
		if filepath.IsAbs(todo.CWD) {
			dir = filepath.Clean(todo.CWD)
		} else if workDir != "" {
			dir = filepath.Clean(filepath.Join(workDir, todo.CWD))
		} else {
			dir = filepath.Clean(todo.CWD)
		}
		break
	}
	if root := repomap.FindGitRoot(dir); root != "" {
		return root
	}
	return dir
}

// firstChecksConfig returns the first TODO frontmatter `checks` block in the
// group, or nil when none set one. Frontmatter overrides the project default.
func firstChecksConfig(todosInGroup []*types.TODO) *types.AgentChecksConfig {
	for _, todo := range todosInGroup {
		if todo != nil && todo.Checks != nil {
			return todo.Checks
		}
	}
	return nil
}
