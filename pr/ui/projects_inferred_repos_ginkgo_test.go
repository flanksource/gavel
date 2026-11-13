package ui

import (
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project repos inferred from git remotes", func() {
	const remoteRepo = "acme/widget"
	var originalPath, checkout string

	BeforeEach(func() {
		originalPath = projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		checkout = GinkgoT().TempDir()
		repo, err := git.PlainInit(checkout, false)
		Expect(err).NotTo(HaveOccurred())
		_, err = repo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{"https://github.com/" + remoteRepo + ".git"}})
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		projectsPath = originalPath
	})

	storedRepos := func() string {
		data, err := os.ReadFile(projectsPath)
		Expect(err).NotTo(HaveOccurred())
		return string(data)
	}

	It("fills an unregistered repo list from origin and marks it inferred", func() {
		Expect(SaveProjects([]Project{{Name: "widget", Dir: checkout}})).To(Succeed())

		project, err := GetProject("widget")

		Expect(err).NotTo(HaveOccurred())
		Expect(project).To(Equal(Project{Name: "widget", Dir: checkout, Repos: []string{remoteRepo}, ReposInferred: true}))
	})

	It("keeps registered repos instead of the remote", func() {
		Expect(SaveProjects([]Project{{Name: "widget", Dir: checkout, Repos: []string{"acme/registered"}}})).To(Succeed())

		project, err := GetProject("widget")

		Expect(err).NotTo(HaveOccurred())
		Expect(project.Repos).To(Equal([]string{"acme/registered"}))
		Expect(project.ReposInferred).To(BeFalse())
	})

	It("does not write inferred repos back when a project is updated unchanged", func() {
		Expect(SaveProjects([]Project{{Name: "widget", Dir: checkout}})).To(Succeed())

		Expect(UpdateProject("widget", Project{Dir: checkout, Repos: []string{remoteRepo}})).To(Succeed())

		Expect(storedRepos()).NotTo(ContainSubstring(remoteRepo))
		Expect(GetProject("widget")).To(HaveField("ReposInferred", BeTrue()))
	})

	It("registers repos that an update changes", func() {
		Expect(SaveProjects([]Project{{Name: "widget", Dir: checkout}})).To(Succeed())

		Expect(UpdateProject("widget", Project{Dir: checkout, Repos: []string{"acme/other"}})).To(Succeed())

		Expect(storedRepos()).To(ContainSubstring("acme/other"))
		Expect(GetProject("widget")).To(HaveField("ReposInferred", BeFalse()))
	})

	It("leaves a project outside any git checkout without repos", func() {
		Expect(SaveProjects([]Project{{Name: "notes", Dir: GinkgoT().TempDir()}})).To(Succeed())

		project, err := GetProject("notes")

		Expect(err).NotTo(HaveOccurred())
		Expect(project.Repos).To(BeEmpty())
		Expect(project.ReposInferred).To(BeFalse())
	})
})
