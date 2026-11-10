package editor

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/comments"
	ui "github.com/flanksource/gavel/pr/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Editor comment boundaries", func() {
	It("places threads under the file's project, and rejects foreign projects, unedited files and stale captures", func() {
		dir := GinkgoT().TempDir()
		useProject(dir)
		path := filepath.Join(dir, "conf", "settings.yaml")
		Expect(os.MkdirAll(filepath.Dir(path), 0700)).To(Succeed())
		Expect(os.WriteFile(path, []byte("name: example\n"), 0600)).To(Succeed())
		store, err := NewStore([]string{path})
		Expect(err).NotTo(HaveOccurred())
		doc, err := store.Read(store.Files()[0].ID)
		Expect(err).NotTo(HaveOccurred())
		place := mustTarget(doc)
		Expect([]string{place.Scope, place.URI}).To(Equal([]string{"gavel://project/demo", "gavel://project/demo/files/conf/settings.yaml"}))
		host := commentHost{store: store}
		ctx := context.Background()
		p, err := host.Principal(ctx)
		Expect(err).NotTo(HaveOccurred())
		raw, err := json.Marshal(CommentSelector{Kind: "document", Version: 1, View: "source"})
		Expect(err).NotTo(HaveOccurred())
		target := comments.Target{URI: place.URI, ScopeURI: place.Scope, Kind: "document", Revision: doc.Revision, Selector: raw}
		_, err = host.Resolve(ctx, p, target)
		Expect(err).NotTo(HaveOccurred())
		foreign := comments.Target{URI: "gavel://project/unknown/files/a.yaml", ScopeURI: "gavel://project/unknown"}
		Expect(host.Authorize(ctx, p, "read", foreign)).To(MatchError(comments.ErrForbidden))
		unedited := target
		unedited.URI = place.Scope + "/files/other.yaml"
		_, err = host.Resolve(ctx, p, unedited)
		Expect(err).To(MatchError(comments.ErrInvalid))
		target.Revision = "stale"
		_, err = host.Resolve(ctx, p, target)
		Expect(err).To(MatchError(comments.ErrConflict))
	})
	It("places a file opened through a symlinked directory under the project registered at its real path", func() {
		real := filepath.Join(GinkgoT().TempDir(), "envs", "base")
		Expect(os.MkdirAll(filepath.Join(real, "schemas"), 0700)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(real, "schemas", "pickers.json"), []byte("{}\n"), 0600)).To(Succeed())
		link := filepath.Join(GinkgoT().TempDir(), "envs")
		Expect(os.Symlink(filepath.Dir(real), link)).To(Succeed())
		useProject(real)
		store, err := NewStore([]string{filepath.Join(link, "base")})
		Expect(err).NotTo(HaveOccurred())
		doc, err := store.Read(store.Files()[0].ID)
		Expect(err).NotTo(HaveOccurred())

		place := mustTarget(doc)
		edited, err := store.editedFile(place.Scope, place.URI)

		Expect(err).NotTo(HaveOccurred())
		Expect([]string{doc.Path, place.URI, edited.ID}).To(Equal([]string{
			filepath.Join(link, "base", "schemas", "pickers.json"),
			"gavel://project/demo/files/schemas/pickers.json",
			doc.ID,
		}))
	})
	It("answers a file outside every project with a project to create for its directory", func() {
		useProject(GinkgoT().TempDir())
		outside := filepath.Join(GinkgoT().TempDir(), "notes.md")
		_, err := targetOf(context.Background(), outside)
		var missing NoProjectError
		Expect(errors.As(err, &missing)).To(BeTrue(), "got %v", err)
		Expect(missing.Suggestion.Dir).To(Equal(filepath.Dir(outside)))
		Expect(missing.Suggestion.Name).To(Equal(filepath.Base(filepath.Dir(outside))))
	})
	It("answers listing a file outside every project with 409 no_project, until the suggested project is created", func() {
		useProject(GinkgoT().TempDir())
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "notes.md")
		Expect(os.WriteFile(path, []byte("# notes\n"), 0600)).To(Succeed())
		store, err := NewStore([]string{path})
		Expect(err).NotTo(HaveOccurred())
		id := store.Files()[0].ID
		server := &Server{Store: store}
		handler := server.Handler()
		list := func() *httptest.ResponseRecorder {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest("GET", "/api/file/"+id+"/comments", nil))
			return response
		}
		missing := list()
		var body struct {
			Code       string     `json:"code"`
			Suggestion ui.Project `json:"suggestion"`
		}
		Expect(json.Unmarshal(missing.Body.Bytes(), &body)).To(Succeed())
		Expect([]any{missing.Code, body.Code, body.Suggestion.Dir}).To(Equal([]any{http.StatusConflict, "no_project", dir}))

		create := httptest.NewRequest("POST", "/api/projects", strings.NewReader(`{"name":"notes","dir":`+jsonString(dir)+`}`))
		create.Header.Set("Content-Type", "application/json")
		created := httptest.NewRecorder()
		handler.ServeHTTP(created, create)
		Expect(created.Code).To(Equal(http.StatusOK), created.Body.String())
		doc, err := store.Read(id)
		Expect(err).NotTo(HaveOccurred())
		Expect(mustTarget(doc).URI).To(Equal("gavel://project/notes/files/notes.md"))
	})
	It("validates the complete captured hunk against disk", func() {
		dir := GinkgoT().TempDir()
		useProject(dir)
		path := filepath.Join(dir, "settings.yaml")
		source := "before\nselected\nafter"
		Expect(os.WriteFile(path, []byte(source), 0600)).To(Succeed())
		store, err := NewStore([]string{path})
		Expect(err).NotTo(HaveOccurred())
		doc, err := store.Read(store.Files()[0].ID)
		Expect(err).NotTo(HaveOccurred())
		patch, err := comments.CapturePatch(source, 2, 2)
		Expect(err).NotTo(HaveOccurred())
		host := commentHost{store: store}
		ctx := context.Background()
		p, err := host.Principal(ctx)
		Expect(err).NotTo(HaveOccurred())
		place := mustTarget(doc)
		for _, valid := range []bool{true, false} {
			if !valid {
				patch.Before = []string{"forged"}
			}
			raw, err := json.Marshal(CommentSelector{Kind: "code-range", Version: 1, View: "source", Patch: &patch})
			Expect(err).NotTo(HaveOccurred())
			_, err = host.Resolve(ctx, p, comments.Target{URI: place.URI, ScopeURI: place.Scope, Kind: "code-range", Revision: doc.Revision, Selector: raw})
			if valid {
				Expect(err).NotTo(HaveOccurred())
			} else {
				Expect(err).To(MatchError(comments.ErrConflict))
			}
		}
	})
	Describe("selector views", func() {
		source := "one\ntwo\nthree\nfour\nselected\nnext\nlast"
		var host commentHost
		var p comments.Principal
		var doc Document
		resolve := func(selector CommentSelector) error {
			raw, err := json.Marshal(selector)
			Expect(err).NotTo(HaveOccurred())
			place := mustTarget(doc)
			_, err = host.Resolve(context.Background(), p, comments.Target{URI: place.URI, ScopeURI: place.Scope, Kind: selector.Kind, Revision: doc.Revision, Selector: raw})
			return err
		}
		BeforeEach(func() {
			dir := GinkgoT().TempDir()
			useProject(dir)
			path := filepath.Join(dir, "notes.md")
			Expect(os.WriteFile(path, []byte(source), 0600)).To(Succeed())
			store, err := NewStore([]string{path})
			Expect(err).NotTo(HaveOccurred())
			doc, err = store.Read(store.Files()[0].ID)
			Expect(err).NotTo(HaveOccurred())
			host = commentHost{store: store}
			p, err = host.Principal(context.Background())
			Expect(err).NotTo(HaveOccurred())
		})
		It("accepts a code-range captured from the preview view", func() {
			patch, err := comments.CapturePatch(source, 5, 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolve(CommentSelector{Kind: "code-range", Version: 1, View: "preview", Patch: &patch})).To(Succeed())
		})
		It("rejects code-range from rich and any unknown view as invalid", func() {
			patch, err := comments.CapturePatch(source, 5, 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(resolve(CommentSelector{Kind: "code-range", Version: 1, View: "rich", Patch: &patch})).To(MatchError(comments.ErrInvalid))
			Expect(resolve(CommentSelector{Kind: "document", Version: 1, View: "thumbnail"})).To(MatchError(comments.ErrInvalid))
		})
		It("accepts document and dom selectors from preview", func() {
			Expect(resolve(CommentSelector{Kind: "document", Version: 1, View: "preview"})).To(Succeed())
			Expect(resolve(CommentSelector{Kind: "dom", Version: 1, View: "preview", Anchor: "heading-1"})).To(Succeed())
		})
	})
	Describe("code-range re-anchoring", func() {
		source := "one\ntwo\nthree\nfour\nselected\nnext\nlast"
		patch, err := comments.CapturePatch(source, 5, 5)
		It("captures without error", func() { Expect(err).NotTo(HaveOccurred()) })
		DescribeTable("resolves the anchor against an edited source",
			func(edited string, expected comments.PatchLocation) {
				Expect(comments.ResolvePatch(edited, patch)).To(Equal(expected))
			},
			Entry("unchanged", source, comments.PatchLocation{State: "matched", StartLine: 5, EndLine: 5}),
			Entry("two lines inserted above the context", "new-a\nnew-b\n"+source, comments.PatchLocation{State: "matched", StartLine: 7, EndLine: 7}),
			Entry("selected line deleted", "one\ntwo\nthree\nfour\nnext\nlast", comments.PatchLocation{State: "outdated"}),
			Entry("context line edited", "one\nTWO\nthree\nfour\nselected\nnext\nlast", comments.PatchLocation{State: "outdated"}),
			Entry("hunk duplicated", source+"\n"+source, comments.PatchLocation{State: "matched", StartLine: 5, EndLine: 5}),
			Entry("hunk duplicated with prefix", "x\n"+source+"\n"+source, comments.PatchLocation{State: "ambiguous"}),
		)
	})
})
