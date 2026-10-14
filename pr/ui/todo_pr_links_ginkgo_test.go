package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	todoruntime "github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TODO PR links over HTTP", Ordered, func() {
	const (
		primaryURL = "https://github.com/acme/widgets/pull/12"
		linkedURL  = "https://github.com/Acme/Widgets/pull/30"
	)
	var (
		ctx       = context.Background()
		handler   http.Handler
		server    *Server
		workDir   string
		todo      *types.TODO
		duplicate *types.TODO
	)

	BeforeAll(func() {
		GinkgoT().Setenv(database.EnvDSN, dbtest.ForGinkgo(dbtest.Options{Name: "gavel_prui_pr_links"}).DSN())
		GinkgoT().Setenv(database.EnvDisable, "")
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		store, err := cache.Open()
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(store.Close()).To(Succeed()) })
		// Both handles are open; clear the DSN so no handler reaching the
		// process-wide database (and Captain's registry, which accepts one pool
		// per process) picks this spec's database up. See TestMain.
		GinkgoT().Setenv(database.EnvDSN, "")
		previousStore := prIndexStore
		prIndexStore = func() *cache.Store { return store }
		DeferCleanup(func() { prIndexStore = previousStore })

		workDir = GinkgoT().TempDir()
		provider, err := todoruntime.New(ctx, opened.Gorm(), todoruntime.WorkspaceOptions{
			Name: "pr-links", RootPath: workDir, Repositories: []string{"acme/widgets"},
		})
		Expect(err).NotTo(HaveOccurred())
		previous := openTodoProvider
		openTodoProvider = func(context.Context, string) (todos.Provider, error) { return provider, nil }
		DeferCleanup(func() { openTodoProvider = previous })

		todo, err = provider.Create(ctx, todos.CreateRequest{Title: "Fix the parser", PR: &types.PR{Number: 12, URL: primaryURL}})
		Expect(err).NotTo(HaveOccurred())
		duplicate, err = provider.Create(ctx, todos.CreateRequest{Title: "Parser panics"})
		Expect(err).NotTo(HaveOccurred())

		server = &Server{ghOpts: github.Options{WorkDir: workDir}, knownBots: map[string]struct{}{}}
		Expect(server.SetResults(github.PRSearchResults{{
			Number: 30, Repo: "Acme/Widgets", Title: "Speed up the parser", Author: "octocat", State: "OPEN",
			URL: linkedURL, UpdatedAt: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		}}, false)).To(Succeed())
		handler = server.Handler()
	})

	do := func(method, path, body string) *httptest.ResponseRecorder {
		GinkgoHelper()
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(recorder, request)
		return recorder
	}
	dirQuery := func() string { return "dir=" + url.QueryEscape(workDir) }

	It("searches the PR index the poller persisted", func() {
		recorder := do(http.MethodGet, "/api/prs/search?limit=20", "")
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var rows []map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &rows)).To(Succeed())
		Expect(rows).To(HaveLen(1))
		Expect(rows[0]).To(SatisfyAll(
			HaveKeyWithValue("provider", "github"), HaveKeyWithValue("repo", "Acme/Widgets"), HaveKeyWithValue("number", float64(30)),
			HaveKeyWithValue("title", "Speed up the parser"), HaveKeyWithValue("isDraft", false), HaveKey("updatedAt"),
		))
		Expect(do(http.MethodGet, "/api/prs/search?q=speed", "").Body.String()).To(ContainSubstring(`"number":30`))
		Expect(do(http.MethodGet, "/api/prs/search?q=unrelated", "").Body.String()).To(Equal("[]\n"))
		Expect(do(http.MethodGet, "/api/prs/search?state=draft", "").Code).To(Equal(http.StatusBadRequest))
	})

	It("links a PR, lists it with the indexed title, and reverse-looks-up the TODO", func() {
		recorder := do(http.MethodPost, "/api/todos/pr-links", `{"dir":`+strconvQuote(workDir)+`,"id":`+strconvQuote(todo.ID)+`,"url":`+strconvQuote(linkedURL)+`}`)
		Expect(recorder.Code).To(Equal(http.StatusCreated), recorder.Body.String())
		var created map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &created)).To(Succeed())
		Expect(created).To(SatisfyAll(
			HaveKeyWithValue("relation", "related"), HaveKeyWithValue("source", "manual"),
			HaveKeyWithValue("repo", "acme/widgets"), HaveKeyWithValue("title", "Speed up the parser"),
		))

		recorder = do(http.MethodGet, "/api/todos/pr-links?"+dirQuery()+"&id="+url.QueryEscape(todo.ID), "")
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var listed []map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &listed)).To(Succeed())
		Expect(listed).To(HaveLen(2))
		Expect(listed[0]).To(SatisfyAll(HaveKeyWithValue("url", primaryURL), HaveKeyWithValue("relation", "implements"), Not(HaveKey("title"))))
		Expect(listed[1]).To(SatisfyAll(HaveKeyWithValue("url", linkedURL), HaveKeyWithValue("state", "OPEN"), HaveKeyWithValue("author", "octocat")))

		recorder = do(http.MethodGet, "/api/prs/todos?"+dirQuery()+"&url="+url.QueryEscape("https://github.com/acme/widgets/pull/30"), "")
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var linked []todos.PRLinkedTODO
		Expect(json.Unmarshal(recorder.Body.Bytes(), &linked)).To(Succeed())
		Expect(linked).To(HaveLen(1))
		Expect(linked[0].ID).To(Equal(todo.ID))
		Expect(linked[0].Title).To(Equal("Fix the parser"))
		Expect(linked[0].Relation).To(Equal(types.PRRelationRelated))
	})

	It("refuses bad input with 4xx JSON errors", func() {
		for _, tc := range []struct {
			method, path, body string
			status             int
		}{
			{http.MethodPost, "/api/todos/pr-links", `{"dir":` + strconvQuote(workDir) + `,"id":` + strconvQuote(todo.ID) + `,"url":"https://github.com/acme/widgets/issues/3"}`, http.StatusBadRequest},
			{http.MethodPost, "/api/todos/pr-links", `{"dir":` + strconvQuote(workDir) + `,"id":` + strconvQuote(todo.ID) + `,"url":` + strconvQuote(linkedURL) + `}`, http.StatusConflict},
			{http.MethodPost, "/api/todos/pr-links", `{"dir":` + strconvQuote(workDir) + `,"id":"00000000","url":` + strconvQuote(linkedURL) + `}`, http.StatusNotFound},
			{http.MethodGet, "/api/todos/pr-links?" + dirQuery(), "", http.StatusBadRequest},
			{http.MethodGet, "/api/prs/todos?url=" + url.QueryEscape("https://example.com/x"), "", http.StatusBadRequest},
		} {
			recorder := do(tc.method, tc.path, tc.body)
			Expect(recorder.Code).To(Equal(tc.status), "%s %s: %s", tc.method, tc.path, recorder.Body.String())
			Expect(recorder.Body.String()).To(ContainSubstring(`"error"`))
		}
	})

	It("unlinks the primary PR with 204 and clears the TODO's pr", func() {
		recorder := do(http.MethodDelete, "/api/todos/pr-links?"+dirQuery()+"&id="+url.QueryEscape(todo.ID)+"&url="+url.QueryEscape(primaryURL), "")
		Expect(recorder.Code).To(Equal(http.StatusNoContent), recorder.Body.String())
		recorder = do(http.MethodGet, "/api/todos/pr-links?"+dirQuery()+"&id="+url.QueryEscape(todo.ID), "")
		Expect(recorder.Body.String()).NotTo(ContainSubstring(primaryURL))
		recorder = do(http.MethodDelete, "/api/todos/pr-links?"+dirQuery()+"&id="+url.QueryEscape(todo.ID)+"&url="+url.QueryEscape(primaryURL), "")
		Expect(recorder.Code).To(Equal(http.StatusNotFound), recorder.Body.String())
	})

	It("carries the duplicates relation through /api/todos/links", func() {
		recorder := do(http.MethodPost, "/api/todos/links?"+dirQuery(), `{"ref":`+strconvQuote(todo.ID)+`,"target":`+strconvQuote(duplicate.ID)+`,"relation":"duplicates"}`)
		Expect(recorder.Code).To(Equal(http.StatusCreated), recorder.Body.String())
		recorder = do(http.MethodGet, "/api/todos/links?"+dirQuery()+"&ref="+url.QueryEscape(duplicate.ID), "")
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var listed todoLinksResponse
		Expect(json.Unmarshal(recorder.Body.Bytes(), &listed)).To(Succeed())
		Expect(listed.Links).To(HaveLen(1))
		Expect(listed.Links[0].Relation).To(Equal(types.RelationDuplicates))
		Expect(listed.Links[0].TargetID).To(Equal(todo.ID))
		Expect(listed.Links[0].TargetTitle).To(Equal("Fix the parser"))
	})

	It("documents the new routes in the OpenAPI spec", func() {
		recorder := do(http.MethodGet, "/api/openapi.json", "")
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var document struct {
			Paths map[string]map[string]any `json:"paths"`
		}
		Expect(json.Unmarshal(recorder.Body.Bytes(), &document)).To(Succeed())
		Expect(document.Paths["/api/todos/pr-links"]).To(SatisfyAll(HaveKey("get"), HaveKey("post"), HaveKey("delete")))
		Expect(document.Paths["/api/prs/todos"]).To(HaveKey("get"))
		Expect(document.Paths["/api/prs/search"]).To(HaveKey("get"))
	})
})
