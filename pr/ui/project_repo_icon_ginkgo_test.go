package ui

import (
	"github.com/flanksource/gavel/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project repo icons", func() {
	const widgetsHome = "https://widgets.example.com"

	It("picks the first project repo whose homepage a tracked PR reports, matching repos case-insensitively", func() {
		projects := []projectInfo{
			{Name: "widgets", Repos: []string{"acme/widgets-fork", "acme/widgets"}},
			{Name: "tools", Repos: []string{"acme/tools"}},
			{Name: "stale", Repos: []string{"acme/stale"}, RepoIcon: &projectRepoIcon{Repo: "acme/stale", HomepageURL: "https://stale.example.com"}},
		}
		withRepoIcons(projects, github.PRSearchResults{
			{Number: 1, Repo: "Acme/Widgets", RepoHomepageURL: widgetsHome},
			{Number: 2, Repo: "acme/tools"},
		})

		Expect([]*projectRepoIcon{projects[0].RepoIcon, projects[1].RepoIcon, projects[2].RepoIcon}).To(Equal([]*projectRepoIcon{
			{Repo: "acme/widgets", HomepageURL: widgetsHome},
			nil,
			nil,
		}))
	})
})
