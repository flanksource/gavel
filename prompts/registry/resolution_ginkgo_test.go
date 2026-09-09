package registry

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPromptResolution(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Registered prompt resolution")
}

var _ = Describe("Registered prompt resolution", func() {
	desc := prompts.Prompt{ID: "commit.message", ConfigPath: "commit.message", Default: "Summarize {{diff}}"}

	It("returns all effective fields from the same saved snapshot as dispatch", func() {
		cfg := verify.GavelConfig{Commit: verify.CommitConfig{Message: verify.PromptSpec{Spec: api.Spec{
			Model: api.Model{Name: "claude-sonnet-5"},
		}}}}
		saved := captainconfig.AIDefaults{Timeout: "17m", MaxTokens: 3000, NoCache: true,
			Providers: map[string]captainconfig.ProviderDefaults{"anthropic": {Mode: "agent", ReasoningEffort: "high"}}}
		item, err := ResolveOne(ResolveOptions{Trace: verify.GavelConfigTrace{Merged: cfg}, Saved: &saved,
			Data: map[string]any{"diff": "a change"}}, desc)
		Expect(err).NotTo(HaveOccurred())
		Expect(item.Effective.Prompt.User).To(Equal("Summarize a change"))
		Expect(item.Effective.Budget.Timeout).To(Equal("17m"))
		Expect(item.Effective.Budget.MaxTokens).To(Equal(3000))
		Expect(item.Effective.NoCache).To(BeTrue())
		Expect(item.Effective.Mode).To(Equal(api.ModeAgent))
		Expect(item.Provenance["/budget/timeout"].Source.Kind).To(Equal(api.FieldSourceSaved))
		resolved, err := cfg.Commit.Message.Resolve(verify.PromptResolveOptions{Base: cfg.AI, DefaultPrompt: desc.Default,
			Data: map[string]any{"diff": "a change"}, Saved: &saved, Name: desc.ConfigPath, RequireModel: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(item.Effective).To(Equal(resolved.Spec))
	})

	It("keeps model-free catalog composition independent of saved defaults", func() {
		item, err := ResolveOne(ResolveOptions{Preview: true}, desc)
		Expect(err).NotTo(HaveOccurred())
		Expect(item.Effective.Name).To(BeEmpty())
		Expect(item.Effective.Mode).To(BeEmpty())
		Expect(item.Effective.Prompt.User).To(Equal("Summarize"))
		Expect(item.Provenance).To(HaveKey("/prompt/user"))
	})

	It("fails a runtime preview without a configured model", func() {
		_, err := ResolveOne(ResolveOptions{Saved: &captainconfig.AIDefaults{}}, desc)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("model"))
	})

	It("resolves each fallback with its provider's saved mode", func() {
		cfg := verify.GavelConfig{Commit: verify.CommitConfig{Message: verify.PromptSpec{Spec: api.Spec{
			Model: api.Model{Name: "claude-sonnet-5", Fallbacks: api.ModelList{{Name: "gpt-5.6-sol"}}},
		}}}}
		saved := captainconfig.AIDefaults{Providers: map[string]captainconfig.ProviderDefaults{
			"anthropic": {Mode: "agent"}, "openai": {Mode: "cli"},
		}}
		item, err := ResolveOne(ResolveOptions{Trace: verify.GavelConfigTrace{Merged: cfg}, Saved: &saved}, desc)
		Expect(err).NotTo(HaveOccurred())
		Expect(item.Effective.Mode).To(Equal(api.ModeAgent))
		Expect(item.Effective.Fallbacks).To(HaveLen(1))
		Expect(item.Effective.Fallbacks[0].Mode).To(Equal(api.ModeCLI))
		Expect(item.Provenance["/fallbacks/0/mode"].Source.Kind).To(Equal(api.FieldSourceSaved))
	})

	It("attributes an equal value and explicit false to the last authored config file", func() {
		home := verify.GavelConfig{AI: api.Spec{Model: api.Model{NoCache: true}}, Commit: verify.CommitConfig{
			Message: verify.PromptSpec{Spec: api.Spec{Model: api.Model{Name: "claude-sonnet-5"}}}}}
		project := verify.GavelConfig{Commit: verify.CommitConfig{Message: verify.PromptSpec{Spec: (api.Spec{Model: api.Model{Name: "claude-sonnet-5"}}).WithExplicit("/noCache")}}}
		merged := verify.MergeGavelConfig(home, project)
		item, err := ResolveOne(ResolveOptions{Preview: true, Trace: verify.GavelConfigTrace{Merged: merged,
			Sources: []verify.GavelConfigSource{{Origin: "user home", Path: "/home/.gavel.yaml", Config: home},
				{Origin: "git root", Path: "/project/.gavel.yaml", Config: project}}}}, desc)
		Expect(err).NotTo(HaveOccurred())
		Expect(item.Effective.NoCache).To(BeFalse())
		Expect(item.Provenance["/model"].Source.Name).To(Equal("/project/.gavel.yaml"))
		Expect(item.Provenance["/noCache"].Source.Name).To(Equal("/project/.gavel.yaml"))
	})

	It("keeps replaced prompt files out of the fold while preserving earlier inline fields", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "selected.prompt")
		Expect(os.WriteFile(path, []byte("---\nmodel: claude-sonnet-5\n---\nSelected body"), 0o600)).To(Succeed())
		home := verify.GavelConfig{Commit: verify.CommitConfig{Message: verify.PromptSpec{
			File: filepath.Join(dir, "missing.prompt"), Spec: api.Spec{Model: api.Model{Effort: api.EffortHigh}},
		}}}
		project := verify.GavelConfig{Commit: verify.CommitConfig{Message: verify.PromptSpec{File: path}}}
		item, err := ResolveOne(ResolveOptions{Preview: true, Trace: verify.GavelConfigTrace{
			Merged: verify.MergeGavelConfig(home, project), Sources: []verify.GavelConfigSource{
				{Path: "/home/.gavel.yaml", Config: home}, {Path: "/project/.gavel.yaml", Config: project},
			},
		}}, desc)
		Expect(err).NotTo(HaveOccurred())
		Expect(item.Effective.Prompt.User).To(Equal("Selected body"))
		Expect(item.Provenance["/model"].Source.Name).To(Equal(path))
		Expect(item.Provenance["/effort"].Source.Name).To(Equal("/home/.gavel.yaml"))
	})

	It("renders a replacement draft without reading a broken stored prompt file", func() {
		cfg := verify.GavelConfig{Commit: verify.CommitConfig{Message: verify.PromptSpec{File: "/missing.prompt"}}}
		item, err := ResolveOne(ResolveOptions{Preview: true, Trace: verify.GavelConfigTrace{Merged: cfg},
			Draft: "---\nmodel: gpt-5.6-sol\n---\nDraft {{diff}}", Data: map[string]any{"diff": "body"}}, desc)
		Expect(err).NotTo(HaveOccurred())
		Expect(item.Effective.Name).To(Equal("gpt-5.6-sol"))
		Expect(item.Effective.Prompt.User).To(Equal("Draft body"))
		Expect(item.Provenance["/model"].Source.Name).To(Equal("draft commit.message"))
	})
})
