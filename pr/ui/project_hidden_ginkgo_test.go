package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/filters"
	"github.com/flanksource/gavel/git/gitstate"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/tenant"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func useTempProjectsCatalog() {
	GinkgoHelper()
	originalPath, originalCounts := projectsPath, projectsTodoCounts
	projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
	projectsTodoCounts = func(_ context.Context, projects []Project) []todoCountsResult {
		return make([]todoCountsResult, len(projects))
	}
	DeferCleanup(func() {
		projectsPath, projectsTodoCounts = originalPath, originalCounts
	})
}

var _ = Describe("hidden-projects filter", func() {
	var f *projectGitFixture
	var tracked trackedGitServer

	BeforeEach(func() {
		f = newProjectGitFixture()
		useTempProjectsCatalog()
		Expect(SaveProjects([]Project{{Name: "acme", Dir: f.repo}, {Name: "beta", Dir: f.worktree}})).To(Succeed())
		tracked = newTrackedGitServer()
	})

	It("hides a project: untracks it, drops it from git surfaces, and flags it in /api/projects", func() {
		Expect(serveProjectGit(tracked.server, http.MethodGet, "/api/projects/git-summary", nil).Code).To(Equal(http.StatusOK))
		_, wasTracked := tracked.tracker.Tracked(f.repo)
		Expect(wasTracked).To(BeTrue())

		added := serveProjectGit(tracked.server, http.MethodPost, "/api/filters/hidden-projects/values", map[string]string{"value": f.repo})

		Expect(added.Code).To(Equal(http.StatusOK), added.Body.String())
		Expect(decodeProjectGit[filters.Filter](added)).To(Equal(filters.Filter{Name: filters.HiddenProjects, Kind: filters.KindWorkspace, Values: []string{f.repo}}))
		_, stillTracked := tracked.tracker.Tracked(f.repo)
		Expect(stillTracked).To(BeFalse())
		summaries := decodeProjectGit[[]gitstate.Summary](serveProjectGit(tracked.server, http.MethodGet, "/api/projects/git-summary", nil))
		Expect(summaries).To(HaveLen(1))
		Expect(summaries[0].Name).To(Equal("beta"))
		listed := decodeProjectGit[[]map[string]any](serveProjectGit(tracked.server, http.MethodGet, "/api/projects", nil))
		Expect([]any{listed[0]["name"], listed[0]["hidden"], listed[1]["name"], listed[1]["hidden"]}).
			To(Equal([]any{"acme", true, "beta", nil}))
		excluded := decodeProjectGit[[]map[string]any](serveProjectGit(tracked.server, http.MethodGet, "/api/projects?exclude=hidden-projects", nil))
		Expect(excluded).To(HaveLen(1))
		Expect(excluded[0]["name"]).To(Equal("beta"))
		Expect(serveProjectGit(tracked.server, http.MethodGet, "/api/projects/acme/git", nil).Code).To(Equal(http.StatusNotFound))
	})

	It("shows a hidden project again when its value is removed", func() {
		serveProjectGit(tracked.server, http.MethodPost, "/api/filters/hidden-projects/values", map[string]string{"value": f.repo})

		removed := serveProjectGit(tracked.server, http.MethodDelete, "/api/filters/hidden-projects/values", map[string]string{"value": f.repo})

		Expect(removed.Code).To(Equal(http.StatusOK), removed.Body.String())
		listed := decodeProjectGit[map[string][]filters.Filter](serveProjectGit(tracked.server, http.MethodGet, "/api/filters", nil))
		Expect(listed).To(Equal(map[string][]filters.Filter{"filters": {
			{Name: filters.HiddenGitHubOrgs, Kind: filters.KindGitHubOrg, Values: []string{}},
			{Name: filters.HiddenProjects, Kind: filters.KindWorkspace, Values: []string{}},
		}}))
		Expect(serveProjectGit(tracked.server, http.MethodGet, "/api/projects/acme/git", nil).Code).To(Equal(http.StatusOK))
	})

	It("answers 400 for an unknown filter, an invalid value, or an exclude of the wrong kind", func() {
		codes := []int{
			serveProjectGit(tracked.server, http.MethodGet, "/api/filters/favourite-projects", nil).Code,
			serveProjectGit(tracked.server, http.MethodPost, "/api/filters/hidden-projects/values", map[string]string{"value": "relative/dir"}).Code,
			serveProjectGit(tracked.server, http.MethodGet, "/api/projects?exclude=hidden-github-orgs", nil).Code,
			serveProjectGit(tracked.server, http.MethodGet, "/api/orgs?exclude=hidden-projects", nil).Code,
		}
		Expect(codes).To(Equal([]int{http.StatusBadRequest, http.StatusBadRequest, http.StatusBadRequest, http.StatusBadRequest}))
	})

	It("drops the orgs hidden-github-orgs holds only when /api/orgs excludes it", func() {
		tracked.server.orgs = []github.Org{{Login: "acme"}, {Login: "acme-legacy"}}
		tracked.server.orgsCachedAt = time.Now()
		serveProjectGit(tracked.server, http.MethodPost, "/api/filters/hidden-github-orgs/values", map[string]string{"value": "acme-legacy"})

		all := decodeProjectGit[[]github.Org](serveProjectGit(tracked.server, http.MethodGet, "/api/orgs", nil))
		shown := decodeProjectGit[[]github.Org](serveProjectGit(tracked.server, http.MethodGet, "/api/orgs?exclude=hidden-github-orgs", nil))

		Expect([][]github.Org{all, shown}).To(Equal([][]github.Org{{{Login: "acme"}, {Login: "acme-legacy"}}, {{Login: "acme"}}}))
	})

	It("keeps one user's hidden projects from another's", func() {
		ada, bob := uuid.New(), uuid.New()
		for id, clerk := range map[uuid.UUID]string{ada: "user_ada", bob: "user_bob"} {
			Expect(tracked.db.Exec(`INSERT INTO users (id, issuer, clerk_user_id) VALUES (?, 'https://clerk.example', ?)`, id, clerk).Error).To(Succeed())
		}
		as := func(user uuid.UUID, method, target string, body string) *httptest.ResponseRecorder {
			request := httptest.NewRequest(method, target, strings.NewReader(body))
			request = request.WithContext(tenant.WithScope(request.Context(), tenant.Scope{OrgID: tenant.LocalOrgID, UserID: &user}))
			recorder := httptest.NewRecorder()
			tracked.server.Handler().ServeHTTP(recorder, request)
			return recorder
		}
		Expect(as(ada, http.MethodPost, "/api/filters/hidden-projects/values", `{"value":"`+f.repo+`"}`).Code).To(Equal(http.StatusOK))

		seen := map[string]filters.Filter{}
		for name, user := range map[string]uuid.UUID{"ada": ada, "bob": bob} {
			seen[name] = decodeProjectGit[filters.Filter](as(user, http.MethodGet, "/api/filters/hidden-projects", ""))
		}
		Expect(seen).To(Equal(map[string]filters.Filter{
			"ada": {Name: filters.HiddenProjects, Kind: filters.KindWorkspace, Values: []string{f.repo}},
			"bob": {Name: filters.HiddenProjects, Kind: filters.KindWorkspace, Values: []string{}},
		}))
	})

	It("imports the hidden flags and ignored orgs of the config files once, then drops them from the files", func() {
		originalSettings := settingsPath
		settingsPath = filepath.Join(GinkgoT().TempDir(), "ui.settings.json")
		DeferCleanup(func() { settingsPath = originalSettings })
		writeFile := func(path, content string) { Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed()) }
		writeFile(projectsPath, `[{"name":"acme","dir":"`+f.repo+`","repos":[],"hidden":true},{"name":"beta","dir":"`+f.worktree+`","repos":[],"hidden":false}]`)
		writeFile(settingsPath, `{"repos":["acme/widget"],"ignoredOrgs":["acme-legacy","tenant-x"]}`)
		store := filters.NewStore(tracked.db)

		Expect(ImportFileFilters(context.Background(), store)).To(Succeed())
		Expect(ImportFileFilters(context.Background(), store)).To(Succeed())

		stored, err := store.List(context.Background(), tenant.Local())
		Expect(err).NotTo(HaveOccurred())
		Expect(stored).To(Equal([]filters.Filter{
			{Name: filters.HiddenGitHubOrgs, Kind: filters.KindGitHubOrg, Values: []string{"acme-legacy", "tenant-x"}},
			{Name: filters.HiddenProjects, Kind: filters.KindWorkspace, Values: []string{f.repo}},
		}))
		var projectsFile []map[string]any
		var settingsFile map[string]any
		Expect(readJSON(projectsPath, &projectsFile)).To(Succeed())
		Expect(readJSON(settingsPath, &settingsFile)).To(Succeed())
		Expect([]any{projectsFile, settingsFile}).To(Equal([]any{
			[]map[string]any{{"name": "acme", "dir": f.repo, "repos": []any{}}, {"name": "beta", "dir": f.worktree, "repos": []any{}}},
			map[string]any{"repos": []any{"acme/widget"}},
		}))
	})
})

func readJSON(path string, into any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, into)
}
