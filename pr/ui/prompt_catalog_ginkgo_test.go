package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

const (
	catalogProject       = "acme"
	catalogHomeModel     = "gpt-5"
	catalogCommitModel   = "claude-sonnet-4-6"
	catalogRunPromptFile = ".gavel/prompts/todos-run.prompt"
	catalogRunPrompt     = "---\nmodel: claude-opus-4-6\n---\nRun {{{body}}} now, then {{#if existingPlan}}{{existingPlan}}{{/if}}.\n"
)

// seedCatalogLayers builds a two-layer chain: ~/.gavel.yaml sets the base ai
// model and an inline commit.message model, the project sets a file override
// for todos.run.
func seedCatalogLayers(home, dir string) {
	global := verify.GavelConfig{}
	global.AI = api.Spec{Model: api.Model{Name: catalogHomeModel}}
	global.Commit.Message = verify.PromptSpec{Spec: api.Spec{Model: api.Model{Name: catalogCommitModel}}}
	Expect(verify.SaveGavelConfig(home, global)).To(Succeed())

	Expect(os.MkdirAll(filepath.Join(dir, filepath.Dir(catalogRunPromptFile)), 0o755)).To(Succeed())
	Expect(os.WriteFile(filepath.Join(dir, catalogRunPromptFile), []byte(catalogRunPrompt), 0o644)).To(Succeed())
	project := verify.GavelConfig{}
	project.Todos.Run = verify.PromptSpec{File: catalogRunPromptFile}
	Expect(verify.SaveGavelConfig(dir, project)).To(Succeed())
}

func catalogCall(query string) []promptCatalogEntry {
	rec := httptest.NewRecorder()
	(&Server{}).Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/settings/prompts/catalog?"+query, nil))
	Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
	var entries []promptCatalogEntry
	Expect(json.Unmarshal(rec.Body.Bytes(), &entries)).To(Succeed())
	return entries
}

func catalogEntry(entries []promptCatalogEntry, id string) promptCatalogEntry {
	for _, entry := range entries {
		if entry.ID == id {
			return entry
		}
	}
	Fail("catalog has no entry " + id)
	return promptCatalogEntry{}
}

func catalogLayer(entry promptCatalogEntry, origin string) promptCatalogLayer {
	for _, layer := range entry.Layers {
		if layer.Origin == origin {
			return layer
		}
	}
	Fail(entry.ID + " has no layer " + origin)
	return promptCatalogLayer{}
}

func renderCall(id, query, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/settings/prompts/"+id+"/render?"+query, strings.NewReader(body))
	(&Server{}).Handler().ServeHTTP(rec, req)
	return rec
}

var _ = Describe("settings prompt catalog", func() {
	var home, dir string
	scope := "project=" + catalogProject

	BeforeEach(func() {
		home = GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		dir = newProjectDir(catalogProject, "acme/"+catalogProject)
		seedCatalogLayers(home, dir)
	})

	It("resolves every registered prompt through the layer chain", func() {
		entries := catalogCall(scope)
		ids := make([]string, 0, len(entries))
		for _, entry := range entries {
			ids = append(ids, entry.ID)
		}
		for _, desc := range registeredPrompts() {
			Expect(ids).To(ContainElement(desc.ID))
		}

		commit := catalogEntry(entries, prompts.CommitMessage)
		Expect(commit.Source).To(Equal("inline"))
		Expect(commit.UsedBy).To(ContainElement("gavel commit"))
		Expect(commit.Effective.Model).To(Equal(catalogCommitModel))
		Expect(commit.Effective.ModelSource).To(Equal("operation"))
		Expect(commit.Provenance["model"]).To(Equal("user-home"))
		Expect(commit.Provenance["body"]).To(Equal("prompt default"))
		Expect(commit.Version).To(HaveLen(16))
		Expect(catalogLayer(commit, "user-home")).To(MatchFields(IgnoreExtras, Fields{
			"Editable": BeTrue(), "Scope": Equal("scope=global"), "Source": Equal("inline"), "Fields": ConsistOf("model"),
		}))
		Expect(catalogLayer(commit, "target-directory")).To(MatchFields(IgnoreExtras, Fields{
			"Editable": BeTrue(), "Scope": Equal("project=" + catalogProject), "Source": Equal("none"),
		}))

		run := catalogEntry(entries, prompts.TodosRun)
		Expect(run.Source).To(Equal("file"))
		Expect(run.Path).To(Equal(filepath.Join(dir, catalogRunPromptFile)))
		Expect(run.Raw).To(Equal(catalogRunPrompt))
		Expect(run.Variables).To(Equal([]string{"body", "existingPlan"}))
		Expect(run.Effective.Model).To(Equal("claude-opus-4-6"))
		Expect(run.Provenance["body"]).To(Equal("target-directory"))
		Expect(run.Provenance["model"]).To(Equal("target-directory"))
		Expect(catalogLayer(run, "target-directory").Fields).To(ConsistOf("file"))

		lint := catalogEntry(entries, prompts.LintFix)
		Expect(lint.Source).To(Equal("builtin"))
		Expect(lint.UsedBy).To(ContainElement("gavel lint --ai-fix"))
		Expect(lint.Effective.ModelSource).NotTo(BeEmpty())
		Expect(lint.Body).NotTo(BeEmpty())
	})

	It("marks a layer read-only when the scope cannot write it", func() {
		commit := catalogEntry(catalogCall("scope=global"), prompts.CommitMessage)
		Expect(catalogLayer(commit, "user-home").Editable).To(BeTrue())
		Expect(commit.Layers).To(HaveLen(1))
	})

	It("reports the effective override on a layer that does not set one", func() {
		rec := promptDetailCall(http.MethodGet, scope, "")
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		got := decodePromptDetail(rec)
		Expect(got.Source).To(Equal("default"))
		Expect(got.Version).To(Equal(promptSourceVersion(got.Raw)))
		Expect(got.Effective).NotTo(BeNil())
		Expect(*got.Effective).To(MatchFields(IgnoreExtras, Fields{
			"Source": Equal("inline"), "Origin": Equal("user-home"),
		}))
		Expect(got.Effective.Raw).To(ContainSubstring("model: " + catalogCommitModel))
	})

	It("refuses a save whose base is stale", func() {
		stale := "---\nmodel: something-else\n---\nold body\n"
		body := "new body"
		spec := map[string]any{}
		rec := promptDetailCall(http.MethodPut, "scope=global", mustJSON(promptDetailRequest{
			Source: "inline", Spec: &spec, Body: &body, BaseRaw: &stale,
		}))
		Expect(rec.Code).To(Equal(http.StatusConflict), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring("reload before saving"))
	})

	It("refuses an inline save that would drop dotprompt-only frontmatter", func() {
		raw := "---\nmodel: claude-sonnet-4-6\noutput:\n  schema:\n    type: object\n---\nbody\n"
		rec := promptDetailCall(http.MethodPut, scope, mustJSON(promptDetailRequest{Source: "inline", Raw: &raw}))
		Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())
		Expect(rec.Body.String()).To(SatisfyAll(ContainSubstring("output"), ContainSubstring("as a file")))

		rec = promptDetailCall(http.MethodPut, scope, mustJSON(promptDetailRequest{
			Source: "file", Path: ".gavel/prompts/commit-message.prompt", Raw: &raw,
		}))
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		Expect(decodePromptDetail(rec).Raw).To(Equal(raw))
	})

	It("renders the effective template and an unsaved draft with caller variables", func() {
		Expect(os.WriteFile(filepath.Join(home, ".captain.yaml"), []byte("ai:\n  timeout: 17m\n  providers:\n    anthropic:\n      mode: agent\n    openai:\n      mode: agent\n"), 0o600)).To(Succeed())
		rec := renderCall(prompts.TodosRun, scope, `{"variables":{"body":"fix the build"}}`)
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		var got promptRenderResponse
		Expect(json.Unmarshal(rec.Body.Bytes(), &got)).To(Succeed())
		Expect(got.User).To(Equal("Run fix the build now, then ."))
		Expect(got.Model).To(Equal("claude-opus-4-6"))
		Expect(got.Resolution.Spec.Budget.Timeout).To(Equal("17m"))
		Expect(got.Resolution.Provenance["/budget/timeout"].Source.Kind).To(Equal(api.FieldSourceSaved))

		rec = renderCall(prompts.TodosRun, scope, `{"variables":{"body":"x"},"raw":"---\nmodel: gpt-5\n---\nDraft {{{body}}}\n"}`)
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		Expect(json.Unmarshal(rec.Body.Bytes(), &got)).To(Succeed())
		Expect(got.User).To(Equal("Draft x"))
		Expect(got.Model).To(Equal("gpt-5"))

		Expect(renderCall("nope", scope, `{}`).Code).To(Equal(http.StatusNotFound))
	})

	It("keeps catalog access independent of malformed saved runtime defaults", func() {
		Expect(os.WriteFile(filepath.Join(home, ".captain.yaml"), []byte("ai: [broken"), 0o600)).To(Succeed())
		Expect(catalogEntry(catalogCall(scope), prompts.CommitMessage).ParseError).To(BeEmpty())
		rec := renderCall(prompts.CommitMessage, scope, `{}`)
		Expect(rec.Code).To(Equal(http.StatusInternalServerError), rec.Body.String())
		Expect(rec.Body.String()).To(ContainSubstring("saved AI defaults"))
	})
})

var _ = Describe("prompt catalog runtime projection", func() {
	It("preserves structural selectors without independently normalizing them", func() {
		runtime := catalogRuntime(api.Model{Name: "agent:opus", Mode: api.ModeAPI}, "operation")

		Expect(runtime.Error).To(BeEmpty())
		Expect(runtime.Mode).To(Equal("api"))
		Expect(runtime.Model).To(Equal("agent:opus"))
	})
})

var _ = Describe("template variables", func() {
	It("lists top-level variables once, skipping helpers, closers and paths", func() {
		body := "{{diff}} {{{details}}} {{#if linters}}{{#each commits}}{{this.hash}}{{/each}}{{/if}} {{role \"user\"}} {{diff}}"
		Expect(templateVariables(body)).To(Equal([]string{"commits", "details", "diff", "linters"}))
	})
})
