package lifecycle_test

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func stepNames(def lifecycle.Lifecycle) []string {
	var names []string
	for _, step := range def.Steps {
		names = append(names, step.Name)
	}
	return names
}

var _ = Describe("Load", func() {
	var workDir string
	BeforeEach(func() {
		workDir = GinkgoT().TempDir()
	})

	It("returns the default when nothing is configured", func() {
		def, err := lifecycle.LoadWith(verify.LifecycleConfig{}, workDir)
		Expect(err).NotTo(HaveOccurred())
		want, err := lifecycle.Default()
		Expect(err).NotTo(HaveOccurred())
		Expect(def).To(Equal(want))
	})

	It("replaces a default step of the same name and keeps the order", func() {
		override := verify.LifecycleConfig{Steps: []map[string]any{{
			"name": "run", "prompt": "file:prompts/custom-run.prompt",
			"outcomes": []map[string]any{{"status": "completed", "when": "run.state == 'succeeded'"}, {"status": "failed", "when": "true"}},
		}}}
		def, err := lifecycle.LoadWith(override, workDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(stepNames(def)).To(Equal([]string{"triage", "plan", "verify", "run"}))
		run, _ := def.Step("run")
		Expect(run.Prompt).To(Equal("file:prompts/custom-run.prompt"))
		Expect(run.When).To(BeEmpty(), "an override replaces the step wholesale")
		Expect(run.Outcomes).To(HaveLen(2))
		_, err = lifecycle.New(def)
		Expect(err).NotTo(HaveOccurred())
	})

	It("appends a new step and its subject declarations", func() {
		override := verify.LifecycleConfig{
			Subject: map[string]string{"owner": "string"},
			Steps: []map[string]any{{
				"name": "handoff", "prompt": "file:prompts/handoff.prompt",
				"when":     "subject.owner != '' && subject.status == 'verified'",
				"outcomes": []map[string]any{{"status": "completed", "when": "true"}},
			}},
		}
		def, err := lifecycle.LoadWith(override, workDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(stepNames(def)).To(Equal([]string{"triage", "plan", "verify", "run", "handoff"}))
		Expect(def.Subject).To(HaveKeyWithValue("owner", "string"))
		Expect(def.Subject).To(HaveKeyWithValue("status", "string"))
		_, err = lifecycle.New(def)
		Expect(err).NotTo(HaveOccurred())
	})

	It("loads an override from a file relative to the work dir", func() {
		Expect(os.MkdirAll(filepath.Join(workDir, ".gavel"), 0o755)).To(Succeed())
		doc := "name: acme\nsteps:\n  - name: plan\n    prompt: file:plan.prompt\n    envelope: plan\n    outcomes:\n      - { status: review, when: \"true\" }\n"
		Expect(os.WriteFile(filepath.Join(workDir, ".gavel", "lifecycle.yaml"), []byte(doc), 0o644)).To(Succeed())
		def, err := lifecycle.LoadWith(verify.LifecycleConfig{File: ".gavel/lifecycle.yaml"}, workDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(def.Name).To(Equal("acme"))
		plan, _ := def.Step("plan")
		Expect(plan.Prompt).To(Equal("file:plan.prompt"))
	})

	It("fails on an override file that does not exist", func() {
		_, err := lifecycle.LoadWith(verify.LifecycleConfig{File: "missing.yaml"}, workDir)
		Expect(err).To(MatchError(ContainSubstring("todos.lifecycle file")))
	})

	It("refuses a file override that also declares steps inline", func() {
		Expect(os.MkdirAll(filepath.Join(workDir, ".gavel"), 0o755)).To(Succeed())
		doc := "name: acme\nsteps:\n  - name: plan\n    prompt: todos.plan\n    envelope: plan\n    outcomes:\n      - { status: review, when: \"true\" }\n"
		Expect(os.WriteFile(filepath.Join(workDir, ".gavel", "lifecycle.yaml"), []byte(doc), 0o644)).To(Succeed())
		override := verify.LifecycleConfig{
			File:  ".gavel/lifecycle.yaml",
			Steps: []map[string]any{{"name": "run", "prompt": "todos.run", "outcomes": []map[string]any{{"status": "pending", "when": "true"}}}},
		}

		_, err := lifecycle.LoadWith(override, workDir)

		Expect(err).To(MatchError(ContainSubstring("mutually exclusive")))
		Expect(err).To(MatchError(ContainSubstring(".gavel/lifecycle.yaml")))
	})

	It("fails on an unknown key in an override", func() {
		override := verify.LifecycleConfig{Steps: []map[string]any{{
			"name": "run", "prompt": "todos.run", "outcome": []map[string]any{{"status": "pending", "when": "true"}},
		}}}
		_, err := lifecycle.LoadWith(override, workDir)
		Expect(err).To(MatchError(ContainSubstring("field outcome not found")))
	})

	It("fails on an override step with no outcomes", func() {
		override := verify.LifecycleConfig{Steps: []map[string]any{{"name": "run", "prompt": "todos.run"}}}
		_, err := lifecycle.LoadWith(override, workDir)
		Expect(err).To(MatchError(ContainSubstring("step run: at least one outcome is required")))
	})

	It("Parse rejects a document without a verify step", func() {
		doc := "name: bare\nsubject:\n  status: string\nsteps:\n  - name: run\n    prompt: todos.run\n    outcomes:\n      - { status: pending, when: \"true\" }\n"
		_, err := lifecycle.Parse([]byte(doc))
		Expect(err).To(MatchError(ContainSubstring(`has no "verify" step`)))
	})
})
