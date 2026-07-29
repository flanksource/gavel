package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/prompts"
	"github.com/flanksource/gavel/verify"
)

// malformedPrompt is a .prompt whose frontmatter YAML does not parse (an
// unterminated flow sequence), the canonical broken-override case. In the
// prompt-spec model only a file override can hold such text — an inline override
// is a structured api.Spec, so its body can never be malformed frontmatter.
const malformedPrompt = "---\nmodel: [broken\n---\nbody\n"

// newProjectDir registers a single project (Ginkgo analogue of withProject,
// which needs a *testing.T) and returns its directory.
func newProjectDir(name, repo string) string {
	t := GinkgoT()
	dir := t.TempDir()
	orig := projectsPath
	projectsPath = filepath.Join(t.TempDir(), "projects.json")
	DeferCleanup(func() { projectsPath = orig })
	Expect(SaveProjects([]Project{{Name: name, Dir: dir, Repos: []string{repo}}})).To(Succeed())
	return dir
}

// seedInlineOverride seeds a valid inline structured override on commit.message.
func seedInlineOverride(dir string, spec api.Spec) {
	cfg := verify.GavelConfig{}
	cfg.Commit.Message = verify.PromptSpec{Spec: spec}
	Expect(verify.SaveGavelConfig(dir, cfg)).To(Succeed())
}

// seedFileOverride writes text to a .prompt file and points commit.message at it.
func seedFileOverride(dir, file, text string) {
	Expect(os.WriteFile(filepath.Join(dir, file), []byte(text), 0o644)).To(Succeed())
	cfg := verify.GavelConfig{}
	cfg.Commit.Message = verify.PromptSpec{File: file}
	Expect(verify.SaveGavelConfig(dir, cfg)).To(Succeed())
}

func promptDetailCall(method, scope, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	(&Server{}).handleSettingsPromptDetail(rec, promptReq(method, prompts.CommitMessage, scope, body))
	return rec
}

func decodePromptDetail(rec *httptest.ResponseRecorder) promptDetailResponse {
	var got promptDetailResponse
	Expect(json.Unmarshal(rec.Body.Bytes(), &got)).To(Succeed())
	return got
}

func mustJSON(req promptDetailRequest) string {
	b, err := json.Marshal(req)
	Expect(err).NotTo(HaveOccurred())
	return string(b)
}

var _ = Describe("settings prompt override repair and round-trip", func() {
	const scope = "project=gavel"
	var dir string

	BeforeEach(func() {
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		dir = newProjectDir("gavel", "flanksource/gavel")
	})

	It("round-trips a valid inline structured override", func() {
		seedInlineOverride(dir, api.Spec{
			Model:  api.Model{Name: "claude-test"},
			Prompt: api.Prompt{User: "Body {{diff}}."},
		})

		rec := promptDetailCall("GET", scope, "")
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())

		got := decodePromptDetail(rec)
		Expect(got.Source).To(Equal("inline"))
		Expect(got.ParseError).To(BeEmpty())
		Expect(got.Spec).NotTo(BeNil())
		Expect((*got.Spec)["model"]).To(Equal("claude-test"))
		Expect(got.Body).NotTo(BeNil())
		Expect(*got.Body).To(ContainSubstring("Body {{diff}}"))
	})

	It("returns the repair contract for a malformed file override and retains its path", func() {
		seedFileOverride(dir, "bad.prompt", malformedPrompt)

		rec := promptDetailCall("GET", scope, "")
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())

		got := decodePromptDetail(rec)
		Expect(got.Source).To(Equal("file"))
		Expect(got.Path).To(Equal("bad.prompt"))
		Expect(got.Raw).To(Equal(malformedPrompt))
		Expect(got.ParseError).NotTo(BeEmpty())
		Expect(got.Spec).To(BeNil(), "no fabricated spec for a malformed prompt")
		Expect(got.Body).To(BeNil(), "no fabricated body for a malformed prompt")
	})

	It("repairs raw source even when the base override is invalid, then GET returns a parsed detail", func() {
		seedFileOverride(dir, "bad.prompt", malformedPrompt)
		fixed := "---\nmodel: fixed-model\n---\nFixed body {{diff}}.\n"

		rec := promptDetailCall("PUT", scope, mustJSON(promptDetailRequest{Source: "inline", Raw: ptr(fixed)}))
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		got := decodePromptDetail(rec)
		Expect(got.Source).To(Equal("inline"))
		Expect(got.ParseError).To(BeEmpty())
		Expect(got.Spec).NotTo(BeNil())
		Expect((*got.Spec)["model"]).To(Equal("fixed-model"))

		rec = promptDetailCall("GET", scope, "")
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		got = decodePromptDetail(rec)
		Expect(got.ParseError).To(BeEmpty())
		Expect(got.Body).NotTo(BeNil())
		Expect(*got.Body).To(ContainSubstring("Fixed body"))
	})

	It("rejects a still-invalid inline raw repair with 400 and leaves .gavel.yaml unchanged", func() {
		seedInlineOverride(dir, api.Spec{Prompt: api.Prompt{User: "valid body"}})
		before, err := os.ReadFile(filepath.Join(dir, ".gavel.yaml"))
		Expect(err).NotTo(HaveOccurred())

		rec := promptDetailCall("PUT", scope, mustJSON(promptDetailRequest{Source: "inline", Raw: ptr("---\nstill: [broken\n---\n")}))
		Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())

		after, err := os.ReadFile(filepath.Join(dir, ".gavel.yaml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(before))
	})

	It("rejects a still-invalid file raw repair with 400 and leaves the prompt file unchanged", func() {
		seedFileOverride(dir, "bad.prompt", malformedPrompt)

		rec := promptDetailCall("PUT", scope, mustJSON(promptDetailRequest{Source: "file", Path: "bad.prompt", Raw: ptr("---\nstill: [broken\n---\n")}))
		Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())

		content, err := os.ReadFile(filepath.Join(dir, "bad.prompt"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(content)).To(Equal(malformedPrompt))
	})

	It("rejects ambiguous raw-plus-structured payloads", func() {
		rec := promptDetailCall("PUT", scope, mustJSON(promptDetailRequest{
			Source: "inline",
			Raw:    ptr("---\nmodel: x\n---\nb\n"),
			Spec:   ptr(map[string]any{"model": "x"}),
			Body:   ptr("b"),
		}))
		Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())
	})

	It("rejects an incomplete structured payload with spec but no body", func() {
		rec := promptDetailCall("PUT", scope, mustJSON(promptDetailRequest{Source: "inline", Spec: ptr(map[string]any{"model": "x"})}))
		Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())
	})

	It("rejects a default reset that carries content fields", func() {
		rec := promptDetailCall("PUT", scope, mustJSON(promptDetailRequest{Source: "default", Raw: ptr("x")}))
		Expect(rec.Code).To(Equal(http.StatusBadRequest), rec.Body.String())
	})

	It("resets an invalid override to default", func() {
		seedFileOverride(dir, "bad.prompt", malformedPrompt)

		rec := promptDetailCall("PUT", scope, mustJSON(promptDetailRequest{Source: "default"}))
		Expect(rec.Code).To(Equal(http.StatusOK), rec.Body.String())
		got := decodePromptDetail(rec)
		Expect(got.Source).To(Equal("default"))
		Expect(got.ParseError).To(BeEmpty())

		cfg, err := verify.LoadSingleGavelConfig(filepath.Join(dir, ".gavel.yaml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Commit.Message.IsEmpty()).To(BeTrue())
	})
})
