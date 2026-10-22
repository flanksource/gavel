package ui

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/flanksource/gavel/git/gitstate"
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

func projectNamesOf(ps []Project) []string {
	names := make([]string, len(ps))
	for i, p := range ps {
		names[i] = p.Name
	}
	return names
}

var _ = Describe("hidden projects store", func() {
	BeforeEach(useTempProjectsCatalog)

	It("persists the hidden flag, keeps it across a form update, and filters it from the default load", func() {
		Expect(SaveProjects([]Project{{Name: "alpha", Dir: "/src/alpha"}, {Name: "beta", Dir: "/src/beta"}})).To(Succeed())

		_, err := SetProjectHidden("beta", true)
		Expect(err).NotTo(HaveOccurred())
		Expect(UpdateProject("beta", Project{Dir: "/src/beta-moved"})).To(Succeed())

		visible, err := LoadProjects(LoadProjectsOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(projectNamesOf(visible)).To(Equal([]string{"alpha"}))
		all, err := LoadProjects(LoadProjectsOptions{IncludeHidden: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(all).To(Equal([]Project{
			{Name: "alpha", Dir: "/src/alpha", Repos: []string{}},
			{Name: "beta", Dir: "/src/beta-moved", Repos: []string{}, Hidden: true},
		}))
	})

	It("names the catalog when hiding an unknown project", func() {
		_, err := SetProjectHidden("ghost", true)

		Expect(err).To(MatchError(ErrProjectNotFound))
		Expect(err).To(MatchError(ContainSubstring(projectsPath)))
	})
})

var _ = Describe("POST /api/projects/{name}/hidden", func() {
	var f *projectGitFixture
	var tracked trackedGitServer

	BeforeEach(func() {
		f = newProjectGitFixture()
		useTempProjectsCatalog()
		Expect(SaveProjects([]Project{{Name: "acme", Dir: f.repo}, {Name: "beta", Dir: f.worktree}})).To(Succeed())
		tracked = newTrackedGitServer()
	})

	It("hides the project, untracks its directory, and leaves it out of the git summary but not the project list", func() {
		Expect(serveProjectGit(tracked.server, http.MethodGet, "/api/projects/git-summary", nil).Code).To(Equal(http.StatusOK))
		_, wasTracked := tracked.tracker.Tracked(f.repo)
		Expect(wasTracked).To(BeTrue())

		hidden := serveProjectGit(tracked.server, http.MethodPost, "/api/projects/acme/hidden", map[string]bool{"hidden": true})

		Expect(hidden.Code).To(Equal(http.StatusOK), hidden.Body.String())
		info := decodeProjectGit[projectInfo](hidden)
		Expect([]any{info.Name, info.Hidden}).To(Equal([]any{"acme", true}))
		_, stillTracked := tracked.tracker.Tracked(f.repo)
		Expect(stillTracked).To(BeFalse())
		summaries := decodeProjectGit[[]gitstate.Summary](serveProjectGit(tracked.server, http.MethodGet, "/api/projects/git-summary", nil))
		Expect(summaries).To(HaveLen(1))
		Expect(summaries[0].Name).To(Equal("beta"))
		listed := decodeProjectGit[[]map[string]any](serveProjectGit(tracked.server, http.MethodGet, "/api/projects", nil))
		Expect([]any{listed[0]["name"], listed[0]["hidden"], listed[1]["name"], listed[1]["hidden"]}).
			To(Equal([]any{"acme", true, "beta", nil}))
		Expect(serveProjectGit(tracked.server, http.MethodGet, "/api/projects/acme/git", nil).Code).To(Equal(http.StatusNotFound))
	})

	It("answers 404 for an unknown project", func() {
		recorder := serveProjectGit(tracked.server, http.MethodPost, "/api/projects/ghost/hidden", map[string]bool{"hidden": true})

		Expect(recorder.Code).To(Equal(http.StatusNotFound), recorder.Body.String())
	})
})
