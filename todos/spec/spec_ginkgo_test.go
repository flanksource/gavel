package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/drivers"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTodosSpecResolution(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TODO spec resolution")
}

// workspace writes cfg as the only .gavel.yaml the seam can see. HOME is
// redirected because LoadGavelConfig merges ~/.gavel.yaml — without this the
// person running the suite is a hidden layer of every assertion below.
func workspace(cfg string) string {
	GinkgoT().Setenv("HOME", GinkgoT().TempDir())
	dir := GinkgoT().TempDir()
	if cfg != "" {
		Expect(os.WriteFile(filepath.Join(dir, ".gavel.yaml"), []byte(cfg), 0o644)).To(Succeed())
	}
	return dir
}

var _ = Describe("Resolve", func() {
	Describe("layering", func() {
		It("lets each layer override only the fields it names", func() {
			dir := workspace(`ai:
  budget:
    cost: 9
    maxTurns: 4
    maxTokens: 1000
todos:
  run:
    budget:
      maxTurns: 11
`)
			todo := &types.TODO{
				TODOFrontmatter: types.TODOFrontmatter{
					Title: "widen the budget",
					LLM:   &types.LLM{MaxCost: 21},
				},
			}

			resolved, err := Resolve(Input{
				WorkDir:  dir,
				Mode:     types.ModeRun,
				Todos:    []*types.TODO{todo},
				Override: api.Spec{Budget: api.Budget{MaxTokens: 4096}},
			})
			Expect(err).ToNot(HaveOccurred())

			// One field per layer, each surviving from a different depth.
			Expect(resolved.Spec.Budget.MaxTokens).To(Equal(4096), "the request wins")
			Expect(resolved.Spec.Budget.Cost).To(Equal(21.0), "the todo's llm: frontmatter wins over ai:")
			Expect(resolved.Spec.Budget.MaxTurns).To(Equal(11), "todos.run wins over ai:")

			Expect(resolved.Provenance).To(SatisfyAll(
				HaveKeyWithValue("budget.maxTokens", "request"),
				HaveKeyWithValue("budget.cost", "todo widen the budget"),
				HaveKeyWithValue("budget.maxTurns", ".gavel.yaml todos.run"),
			))
		})

		It("takes the model from the mode's .prompt frontmatter when nothing else names one", func() {
			resolved, err := Resolve(Input{WorkDir: workspace(""), Mode: types.ModeRun})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Spec.Name).ToNot(BeEmpty())
			Expect(resolved.Provenance["model"]).To(Equal("todos-run.prompt"))
		})

		It("applies todos.timeout to every mode and lets the request override it", func() {
			dir := workspace("todos:\n  timeout: 45m\n")
			for _, mode := range []types.RunMode{types.ModeRun, types.ModePlan, types.ModeVerify} {
				resolved, err := Resolve(Input{WorkDir: dir, Mode: mode})
				Expect(err).ToNot(HaveOccurred(), string(mode))
				Expect(resolved.Timeout).To(Equal(45*time.Minute), string(mode))
				Expect(resolved.Spec.Budget.Timeout).To(Equal("45m0s"), string(mode))
			}

			resolved, err := Resolve(Input{
				WorkDir:  dir,
				Mode:     types.ModeRun,
				Override: api.Spec{Budget: api.Budget{Timeout: "90s"}},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Timeout).To(Equal(90 * time.Second))
		})

		It("defaults the timeout when nothing configures one", func() {
			resolved, err := Resolve(Input{WorkDir: workspace(""), Mode: types.ModeRun})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Timeout).To(Equal(DefaultTimeout))
		})

		It("carries an inline prompt as the template, not as a body override", func() {
			dir := workspace(`todos:
  run:
    prompt:
      user: "Implement {{{body}}} carefully."
      system: Stay surgical.
`)
			resolved, err := Resolve(Input{WorkDir: dir, Mode: types.ModeRun})
			Expect(err).ToNot(HaveOccurred())

			// Carrying it on the spec as well would inject the raw, unrendered
			// source — `{{{body}}}` and all — alongside the rendered prompt.
			Expect(resolved.Template).To(Equal("Implement {{{body}}} carefully."))
			Expect(resolved.Spec.Prompt.User).To(BeEmpty())
			Expect(resolved.Spec.Prompt.System).To(Equal("Stay surgical."), "system is run config, not the templated body")
		})
	})

	Describe("mode invariants", func() {
		It("strips commits from every mode that produces no work to commit", func() {
			dir := workspace(`ai:
  workflow:
    commits:
      - on: run
        gates: full
`)
			for _, mode := range []types.RunMode{types.ModePlan, types.ModeVerify} {
				resolved, err := Resolve(Input{WorkDir: dir, Mode: mode})
				Expect(err).ToNot(HaveOccurred(), string(mode))
				if resolved.Spec.Workflow != nil {
					Expect(resolved.Spec.Workflow.Commits).To(BeEmpty(), string(mode))
				}
			}

			resolved, err := Resolve(Input{WorkDir: dir, Mode: types.ModeRun})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Spec.Workflow.Commits).To(HaveLen(1), "a run keeps what it was configured with")
		})

		It("defaults an unset mode to run", func() {
			resolved, err := Resolve(Input{WorkDir: workspace("")})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Mode).To(Equal(types.ModeRun))
		})
	})

	Describe("driver", func() {
		It("resolves request over config over the built-in default", func() {
			configured := workspace("todos:\n  driver: cli\n")

			resolved, err := Resolve(Input{WorkDir: configured, Mode: types.ModeRun})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Driver).To(Equal(drivers.Cli), ".gavel.yaml todos.driver")

			resolved, err = Resolve(Input{WorkDir: configured, Mode: types.ModeRun, Driver: "api"})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Driver).To(Equal(drivers.Api), "the request outranks the config")

			resolved, err = Resolve(Input{WorkDir: workspace(""), Mode: types.ModeRun})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Driver).To(Equal(drivers.Default))
		})

		// A captain backend is a (provider, mechanism) pair, so `ai.backend:
		// codex-agent` has already chosen the agent SDK. Pairing it with the
		// built-in cmux default would reject a coherent config for disagreeing
		// with a driver nobody asked for. The model sits at todos.run because
		// the mode's .prompt frontmatter outranks `ai:` — see the sibling spec.
		It("takes the mechanism a configured backend already names", func() {
			resolved, err := Resolve(Input{
				WorkDir: workspace("ai:\n  backend: codex-agent\ntodos:\n  run:\n    model: gpt-5.6-sol\n"),
				Mode:    types.ModeRun,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Driver).To(Equal(drivers.Cli))
			Expect(resolved.Spec.Backend).To(BeEquivalentTo("codex-agent"))
		})

		It("still lets an explicit driver outrank the backend's mechanism", func() {
			resolved, err := Resolve(Input{
				WorkDir: workspace("ai:\n  backend: codex-agent\ntodos:\n  run:\n    model: gpt-5.6-sol\n"),
				Mode:    types.ModeRun,
				Driver:  "api",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Driver).To(Equal(drivers.Api))
		})

		// The other half of the same pair going stale. `ai:` sits below the
		// mode's .prompt frontmatter, so `ai.backend: codex-agent` survives
		// under a frontmatter that pins `model: claude` — and the two halves
		// then name different providers. Left alone that pair is rejected
		// downstream, which would mean nobody with a codex `ai.backend` could
		// run a todo at all. The stale half is dropped so it is re-derived
		// from the model that actually won.
		It("drops a configured backend whose provider a higher layer overrode", func() {
			resolved, err := Resolve(Input{
				WorkDir: workspace("ai:\n  backend: codex-agent\n  model: gpt-5.6-sol\n"),
				Mode:    types.ModeRun,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Spec.Name).To(Equal("claude"), "todos-run.prompt frontmatter outranks ai.model")
			Expect(resolved.Spec.Backend).To(BeEmpty())
			Expect(resolved.Driver).To(Equal(drivers.Cli), "the mechanism half is still what ai.backend said")
		})
	})

	Describe("approvals", func() {
		It("defaults to what the entrypoint can service", func() {
			dir := workspace("")

			resolved, err := Resolve(Input{WorkDir: dir, Mode: types.ModeRun, CanApprove: true})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Approvals).To(BeTrue())

			resolved, err = Resolve(Input{WorkDir: dir, Mode: types.ModeRun, CanApprove: false})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Approvals).To(BeFalse())
		})

		It("refuses to enable approvals nothing can answer", func() {
			dir := workspace("todos:\n  approvals: true\n")

			_, err := Resolve(Input{WorkDir: dir, Mode: types.ModeRun, CanApprove: false})
			Expect(err).To(MatchError(ContainSubstring("cannot answer approval requests")))

			resolved, err := Resolve(Input{WorkDir: dir, Mode: types.ModeRun, CanApprove: true})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Approvals).To(BeTrue())
		})

		It("lets an explicit false turn approvals off where they are available", func() {
			resolved, err := Resolve(Input{
				WorkDir:    workspace("todos:\n  approvals: false\n"),
				Mode:       types.ModeRun,
				CanApprove: true,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(resolved.Approvals).To(BeFalse())
		})
	})

	Describe("failing loud", func() {
		It("rejects an effort the resolved model contradicts", func() {
			_, err := Resolve(Input{
				WorkDir:  workspace("todos:\n  run:\n    model: \"opus:low\"\n"),
				Mode:     types.ModeRun,
				Override: api.Spec{Model: api.Model{Effort: api.EffortHigh}},
			})
			Expect(err).To(MatchError(ContainSubstring("conflicts with model")))
		})

		It("rejects a malformed or non-positive timeout instead of falling back", func() {
			_, err := Resolve(Input{WorkDir: workspace("todos:\n  timeout: half-an-hour\n"), Mode: types.ModeRun})
			Expect(err).To(MatchError(ContainSubstring("invalid todos timeout")))

			_, err = Resolve(Input{WorkDir: workspace("todos:\n  timeout: 0s\n"), Mode: types.ModeRun})
			Expect(err).To(MatchError(ContainSubstring("must be greater than zero")))
		})

		// A tool posture that cannot be honoured must stop the run, not quietly
		// hand the agent whatever it inherited. Both spellings reach this: a
		// mistyped policy in .gavel.yaml is rejected by the config decode, and a
		// mistyped mode set programmatically by a caller is rejected here.
		It("rejects a mistyped tool policy in the config", func() {
			_, err := Resolve(Input{
				WorkDir: workspace("todos:\n  run:\n    permissions:\n      tools:\n        Bash: sometimes\n"),
				Mode:    types.ModeRun,
			})
			Expect(err).To(MatchError(ContainSubstring(`invalid tool policy "sometimes" for tool "Bash"`)))
		})

		It("rejects a mistyped tool mode from a caller-supplied override", func() {
			_, err := Resolve(Input{
				WorkDir: workspace(""),
				Mode:    types.ModeRun,
				Override: api.Spec{Permissions: api.Permissions{
					Tools: api.Tools{Modes: map[string]api.ToolMode{"Bash": "sometimes"}},
				}},
			})
			Expect(err).To(MatchError(ContainSubstring(`invalid tool mode "sometimes" for tool "Bash"`)))
		})

		It("names the offending model rather than running with a default", func() {
			_, err := Resolve(Input{WorkDir: workspace("todos:\n  run:\n    model: \"opus:sideways\"\n"), Mode: types.ModeRun})
			Expect(err).To(MatchError(ContainSubstring("invalid todos model")))
		})
	})

	// H3: the CLI and the dashboard used to be two independent hand-rolled
	// builders, so the same .gavel.yaml produced different runs depending on which
	// binary read it. Both now reduce their inputs to one api.Spec and fold it
	// through here, which is only worth anything if the fold itself is a pure
	// function of that Input.
	Describe("one answer per input", func() {
		It("produces byte-identical specs for equivalent inputs, in every mode", func() {
			dir := workspace(`ai:
  budget:
    cost: 6
todos:
  driver: cli
  timeout: 20m
  run:
    model: claude-sonnet-5
    permissions:
      mode: acceptEdits
  plan:
    budget:
      maxTurns: 5
`)
			// The same request expressed the way each entrypoint builds it: the CLI
			// assigns parsed flags onto a zero spec field by field, the dashboard
			// decodes the payload's spec object. Equivalent inputs, so equivalent runs.
			fromFlags := api.Spec{}
			fromFlags.Effort = api.EffortHigh
			fromFlags.Budget.Cost = 12.5

			fromPayload := api.Spec{
				Model:  api.Model{Effort: api.EffortHigh},
				Budget: api.Budget{Cost: 12.5},
			}

			for _, mode := range []types.RunMode{types.ModeRun, types.ModePlan, types.ModeVerify} {
				cli, err := Resolve(Input{WorkDir: dir, Mode: mode, Override: fromFlags, Driver: "cli"})
				Expect(err).ToNot(HaveOccurred(), string(mode))
				dashboard, err := Resolve(Input{WorkDir: dir, Mode: mode, Override: fromPayload, Driver: "cli"})
				Expect(err).ToNot(HaveOccurred(), string(mode))

				cliJSON, err := json.Marshal(cli.Spec)
				Expect(err).ToNot(HaveOccurred())
				dashboardJSON, err := json.Marshal(dashboard.Spec)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(cliJSON)).To(Equal(string(dashboardJSON)), string(mode))

				Expect(cli.Driver).To(Equal(dashboard.Driver), string(mode))
				Expect(cli.Timeout).To(Equal(dashboard.Timeout), string(mode))
				Expect(cli.Template).To(Equal(dashboard.Template), string(mode))
			}
		})

		It("does not let one resolution mutate the next", func() {
			dir := workspace("ai:\n  workflow:\n    verify:\n      commands: [go test ./...]\n")

			first, err := Resolve(Input{WorkDir: dir, Mode: types.ModeRun})
			Expect(err).ToNot(HaveOccurred())
			first.Spec.Workflow.Verify.Commands[0] = "rm -rf /"

			second, err := Resolve(Input{WorkDir: dir, Mode: types.ModeRun})
			Expect(err).ToNot(HaveOccurred())
			Expect(second.Spec.Workflow.Verify.Commands).To(Equal([]string{"go test ./..."}))
		})
	})
})
