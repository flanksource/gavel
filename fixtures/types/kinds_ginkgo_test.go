package types

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky/task"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
)

func parseYAMLDocument(body string) []*fixtures.FixtureNode {
	GinkgoHelper()
	path := filepath.Join(GinkgoT().TempDir(), "steps.fixture.yaml")
	Expect(os.WriteFile(path, []byte(body), 0o600)).To(Succeed())
	file, err := fixtures.ParseFixtureFile(path)
	Expect(err).ToNot(HaveOccurred())
	return []*fixtures.FixtureNode{file}
}

var _ = Describe("exec kind in a YAML fixture", func() {
	It("hands a later step's subprocess the variables an earlier step exported, and its outputs to CEL", func() {
		nodes := parseYAMLDocument(`
kind: Fixture
steps:
  - id: mint
    kind: exec
    script: 'printf "{\"token\":\"t-42\"}"'
    cel: json.token == 't-42'
    export: { TOKEN: "{{.outputs.json.token}}" }
  - id: use
    kind: exec
    script: 'echo "token=$TOKEN from={{.steps.mint.json.token}}"'
    cel: stdout.contains('token=t-42 from=t-42') && exitCode == 0
`)
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(results).To(HaveLen(2))
		for _, result := range results {
			Expect(result.Status).To(Equal(task.StatusPASS), result.Error+result.CELTrace)
			Expect(result.Type).To(Equal("exec"))
		}
		Expect(results[0].Metadata["exports"]).To(Equal(map[string]fixtures.ExportRecord{"TOKEN": {Value: "t-42", Step: "mint"}}))
	})

	It("waits by re-running the command until the until expression holds over its fresh outputs", func() {
		nodes := parseYAMLDocument(`
kind: Fixture
steps:
  - id: counter
    kind: exec
    script: 'n=$(cat {{.tmp}}/n 2>/dev/null || echo 0); n=$((n+1)); echo $n > {{.tmp}}/n; echo "count=$n"'
    wait: { until: "stdout.contains('count=3')", interval: 10ms, timeout: 10s }
    cel: stdout.contains('count=3')
`)
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error+results[0].CELTrace)
		wait := results[0].Metadata["wait"].(fixtures.WaitReport)
		Expect(wait.Outcome).To(Equal(fixtures.WaitSatisfied))
		Expect(wait.Outputs).To(Equal(results[0].Metadata["outputs"]))
	})

	It("fails an exec step on an unexpected exit code before its cel runs", func() {
		nodes := parseYAMLDocument("kind: Fixture\nsteps:\n  - { id: x, kind: exec, script: 'exit 3', cel: 'true' }\n")
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(results[0].Status).To(Equal(task.StatusFAIL))
		Expect(results[0].Error).To(ContainSubstring("expected exit code 0, got 3"))
	})

	It("answers a pty prompt from the step's input script and asserts over what the terminal showed", func() {
		nodes := parseYAMLDocument(`
kind: Fixture
steps:
  - id: ask
    kind: exec
    terminal: pty
    script: 'printf "pick one: "; read choice; echo "picked $choice"'
    input:
      - { expect: "pick one:", send: sqlite, keys: [enter] }
    cel: stdout.contains('picked sqlite')
`)
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error+results[0].CELTrace)
	})

	It("fails a step that scripts input without a pty", func() {
		nodes := parseYAMLDocument("kind: Fixture\nsteps:\n  - { id: x, kind: exec, script: 'read a', input: [{ expect: a, keys: [enter] }] }\n")
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(results[0].Status).ToNot(Equal(task.StatusPASS))
		Expect(results[0].Error).To(ContainSubstring("input needs terminal: pty"))
	})
})

var _ = Describe("exec kind in a markdown fixture", func() {
	It("fails a command killed by the fixture's timeout as timed out", func() {
		timeout := 200 * time.Millisecond
		test := fixtures.FixtureTest{Name: "sleeper", ExecFixtureBase: fixtures.ExecFixtureBase{Exec: "bash", Args: []string{"-c", "sleep 5"}},
			Expected: fixtures.Expectations{Timeout: &timeout}}

		result := fixtures.RunNode(context.Background(), test, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect([]any{result.Status, result.TimedOut}).To(Equal([]any{task.StatusFAIL, true}), result.Error)
		Expect(result.Error).To(HavePrefix("step sleeper timed out after 200ms: "))
	})
})

var _ = Describe("ai kind in a YAML fixture", func() {
	It("grades a criteria checklist declared in the step config", func() {
		var graded fixtures.FixtureTest
		DeferCleanup(fixtures.DefaultKinds.Replace(NewAIKind(func(test fixtures.FixtureTest, _ fixtures.RunOptions) fixtures.FixtureResult {
			graded = test
			return fixtures.FixtureResult{Name: test.Name, Status: task.StatusPASS,
				Metadata: map[string]interface{}{"checklist": []fixtures.ChecklistResult{{Item: "has a test", Passed: true}}}}
		})))
		nodes := parseYAMLDocument(`
kind: Fixture
steps:
  - id: review
    kind: ai
    model: api:gpt-5
    prompt: Focus on the greet package.
    criteria: [has a test]
    cel: passed && checklist[0].passed
`)
		results, _, err := fixtures.RunNodes(context.Background(), nodes, fixtures.RunOptions{WorkDir: GinkgoT().TempDir()})

		Expect(err).ToNot(HaveOccurred())
		Expect(results[0].Status).To(Equal(task.StatusPASS), results[0].Error+results[0].CELTrace)
		Expect(results[0].Type).To(Equal("ai"))
		Expect(graded.AI.Model).To(Equal("api:gpt-5"))
		Expect(graded.AIStep.Prompt).To(Equal("Focus on the greet package."))
		Expect(graded.AIStep.Criteria).To(Equal([]fixtures.ChecklistItem{{Text: "has a test"}}))
	})

	It("rejects an ai step with neither prompt nor criteria at parse time", func() {
		path := filepath.Join(GinkgoT().TempDir(), "bad.fixture.yaml")
		Expect(os.WriteFile(path, []byte("kind: Fixture\nsteps:\n  - { id: r, kind: ai, model: x }\n"), 0o600)).To(Succeed())
		_, err := fixtures.ParseFixtureFile(path)
		Expect(err).To(MatchError(ContainSubstring(`step "r": config does not match the ai schema`)))
	})
})
