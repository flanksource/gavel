package verify

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/flanksource/captain/pkg/api"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TODO runtime preset configuration", func() {
	DescribeTable("preserves global and step presets in the flat wire shape",
		func(format string) {
			original := map[string]any{"todos": map[string]any{
				"presets":        []any{"project-default", "project-review"},
				"runtimeProfile": "deprecated-profile",
				"lifecycle":      map[string]any{},
				"run":            map[string]any{"presets": []any{"implementer"}},
				"plan":           map[string]any{"presets": []any{"planner"}},
				"triage":         map[string]any{"presets": []any{"triager"}},
				"verify":         map[string]any{"presets": []any{"reviewer"}, "model": "cli:sonnet"},
				"steps": map[string]any{"handoff": map[string]any{
					"presets": []any{"publisher"}, "budget": map[string]any{"maxTurns": float64(3)},
				}},
			}}
			encode, decode := json.Marshal, json.Unmarshal
			if format == "yaml" {
				encode = yaml.Marshal
				decode = func(data []byte, into any) error { return yaml.Unmarshal(data, into) }
			}
			raw, err := encode(original)
			Expect(err).NotTo(HaveOccurred())
			var cfg GavelConfig
			Expect(decode(raw, &cfg)).To(Succeed())
			raw, err = encode(cfg)
			Expect(err).NotTo(HaveOccurred())
			var restored map[string]any
			Expect(decode(raw, &restored)).To(Succeed())
			Expect(restored["todos"]).To(Equal(original["todos"]))
		},
		Entry("JSON", "json"),
		Entry("YAML", "yaml"),
	)

	DescribeTable("rejects prompt files where the lifecycle owns the prompt",
		func(document, path string) {
			var cfg TodosConfig
			Expect(json.Unmarshal([]byte(document), &cfg)).To(Succeed())
			err := cfg.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(ContainSubstring(path)))
			Expect(err).To(MatchError(ContainSubstring("todos.lifecycle")))
		},
		Entry("verification", `{"verify":{"file":"review.prompt"}}`, "todos.verify.file"),
		Entry("custom steps", `{"steps":{"handoff":{"file":"handoff.prompt"}}}`, "todos.steps.handoff.file"),
	)

	It("keeps verifier defaults in its spec without selecting presets or a custom step", func() {
		defaults := DefaultGavelConfig().Todos
		Expect(defaults.Verify).To(Equal(PromptSpec{Spec: api.Spec{Model: api.Model{Mode: DefaultVerifyMode}}}))
		Expect(defaults.RuntimeProfile).To(BeEmpty())
		Expect(defaults.Presets).To(BeEmpty())
		Expect(defaults.Steps).To(BeEmpty())
	})

	It("inherits verification and custom step presets while applying partial spec overrides", func() {
		base := GavelConfig{Todos: TodosConfig{
			Presets: []string{"organization"},
			Verify: PromptSpec{Presets: []string{"reviewer"}, PresetsSet: true, Spec: api.Spec{
				Model: api.Model{Mode: DefaultVerifyMode}, Budget: api.Budget{MaxTurns: 8},
			}},
			Steps: map[string]PromptSpec{"handoff": {
				Presets: []string{"publisher"}, PresetsSet: true, Spec: api.Spec{Budget: api.Budget{MaxTurns: 5}},
			}},
		}}
		override := GavelConfig{Todos: TodosConfig{
			Presets: []string{"repository"},
			Verify:  PromptSpec{Spec: api.Spec{Model: api.Model{Name: "sonnet"}}},
			Steps: map[string]PromptSpec{"handoff": {
				Spec: api.Spec{Budget: api.Budget{MaxTokens: 200}},
			}},
		}}
		Expect(MergeGavelConfig(base, override).Todos).To(Equal(TodosConfig{
			Presets: []string{"repository"},
			Verify: PromptSpec{Presets: []string{"reviewer"}, PresetsSet: true, Spec: api.Spec{
				Model: api.Model{Name: "sonnet", Mode: DefaultVerifyMode}, Budget: api.Budget{MaxTurns: 8},
			}},
			Steps: map[string]PromptSpec{"handoff": {
				Presets: []string{"publisher"}, PresetsSet: true, Spec: api.Spec{Budget: api.Budget{MaxTurns: 5, MaxTokens: 200}},
			}},
		}))
	})

	It("preserves presets through saved configuration and rejects unsupported files on load", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, ".gavel.yaml")
		cfg := GavelConfig{Todos: TodosConfig{
			Presets: []string{"project-default"},
			Verify:  PromptSpec{Presets: []string{"reviewer"}, PresetsSet: true},
			Steps:   map[string]PromptSpec{"handoff": {Presets: []string{"publisher"}, PresetsSet: true}},
		}}
		Expect(SaveGavelConfig(dir, cfg)).To(Succeed())
		loaded, err := LoadSingleGavelConfig(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(loaded.Todos.Presets).To(Equal([]string{"project-default"}))
		Expect(loaded.Todos.Verify.Presets).To(Equal([]string{"reviewer"}))
		Expect(loaded.Todos.Steps["handoff"].Presets).To(Equal([]string{"publisher"}))
		Expect(os.WriteFile(path, []byte("todos:\n  verify:\n    file: review.prompt\n"), 0o600)).To(Succeed())
		_, err = LoadSingleGavelConfig(path)
		Expect(err).To(MatchError(ContainSubstring("todos.verify.file")))
		Expect(err).To(MatchError(ContainSubstring(path)))
	})

	It("documents global and step presets while marking profiles deprecated", func() {
		raw, err := ConfigJSONSchema()
		Expect(err).NotTo(HaveOccurred())
		var schema map[string]any
		Expect(json.Unmarshal([]byte(raw), &schema)).To(Succeed())
		todos := schema["properties"].(map[string]any)["todos"].(map[string]any)["properties"].(map[string]any)
		Expect(todos).To(HaveKey("presets"))
		Expect(todos).To(HaveKey("runtimeProfile"))
		Expect(todos["runtimeProfile"]).To(HaveKeyWithValue("deprecated", true))
		for _, name := range []string{"run", "plan", "triage", "verify"} {
			properties := todos[name].(map[string]any)["properties"].(map[string]any)
			Expect(properties).To(HaveKey("presets"), name)
			Expect(properties).To(HaveKey("runtimeProfile"), name)
			if name == "verify" {
				Expect(properties).NotTo(HaveKey("file"))
			}
		}
		custom := todos["steps"].(map[string]any)["additionalProperties"].(map[string]any)
		Expect(custom["properties"]).To(HaveKey("runtimeProfile"))
		Expect(custom["properties"]).NotTo(HaveKey("file"))
		commit := schema["properties"].(map[string]any)["commit"].(map[string]any)["properties"].(map[string]any)
		Expect(commit["message"].(map[string]any)["properties"]).NotTo(HaveKey("runtimeProfile"))
	})
})
