package ui

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/projects"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type launchedEditor struct {
	name string
	args []string
}

var _ = Describe("host editors", func() {
	var (
		launched []launchedEditor
		dir      string
	)

	BeforeEach(func() {
		appsDir := GinkgoT().TempDir()
		Expect(os.Mkdir(filepath.Join(appsDir, "Zed.app"), 0o755)).To(Succeed())
		originalLookPath, originalApps, originalGOOS, originalLaunch := editorLookPath, editorAppDirs, editorGOOS, launchEditor
		editorLookPath = func(name string) (string, error) {
			if name == "code" {
				return "/usr/local/bin/code", nil
			}
			return "", errors.New("not found")
		}
		editorAppDirs = func() []string { return []string{appsDir} }
		editorGOOS = "darwin"
		launched = nil
		launchEditor = func(name string, args []string) error {
			launched = append(launched, launchedEditor{name: name, args: args})
			return nil
		}
		useProjectsFile(filepath.Join(GinkgoT().TempDir(), "projects.json"))
		dir = GinkgoT().TempDir()
		writeDiffFile(dir, "src/main.go", "package main\n")
		Expect(seedProjects([]projects.Project{{Name: "gavel", Dir: dir}})).To(Succeed())
		DeferCleanup(func() {
			editorLookPath, editorAppDirs, editorGOOS, launchEditor = originalLookPath, originalApps, originalGOOS, originalLaunch
		})
	})

	It("lists the editors found on PATH or as application bundles", func() {
		recorder := httptest.NewRecorder()
		(&Server{}).handleEditors(recorder, httptest.NewRequest(http.MethodGet, "/api/editors", nil))

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Body.String()).To(MatchJSON(`[{"id":"vscode","label":"VS Code"},{"id":"zed","label":"Zed"}]`))
	})

	DescribeTable("opens a project file in the chosen editor",
		func(body string, expected launchedEditor) {
			response := postOpenInEditor(body)

			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
			Expect(response.Body.String()).To(MatchJSON(`{"ok":true}`))
			resolved, err := filepath.EvalSymlinks(filepath.Join(dir, "src", "main.go"))
			Expect(err).NotTo(HaveOccurred())
			for i, arg := range expected.args {
				expected.args[i] = os.Expand(arg, func(string) string { return resolved })
			}
			Expect(launched).To(Equal([]launchedEditor{expected}))
		},
		Entry("CLI editor at a line", `{"path":"src/main.go","editor":"vscode","line":7}`,
			launchedEditor{name: "/usr/local/bin/code", args: []string{"-g", "${file}:7"}}),
		Entry("CLI editor without a line", `{"path":"src/main.go","editor":"vscode"}`,
			launchedEditor{name: "/usr/local/bin/code", args: []string{"-g", "${file}"}}),
		Entry("application bundle", `{"path":"src/main.go","editor":"zed","line":3}`,
			launchedEditor{name: "open", args: []string{"-a", "Zed", "${file}"}}),
	)

	DescribeTable("rejects editors and paths it cannot open",
		func(body, message string) {
			response := postOpenInEditor(body)

			Expect(response.Code).To(Equal(http.StatusBadRequest), response.Body.String())
			Expect(response.Body.String()).To(ContainSubstring(message))
			Expect(launched).To(BeEmpty())
		},
		Entry("unknown editor", `{"path":"src/main.go","editor":"notepad"}`, "unknown editor"),
		Entry("undetected editor", `{"path":"src/main.go","editor":"cursor"}`, "is not installed"),
		Entry("path outside the project", `{"path":"../main.go","editor":"vscode"}`, "project-relative"),
		Entry("missing file", `{"path":"src/missing.go","editor":"vscode"}`, "src/missing.go"),
		Entry("negative line", `{"path":"src/main.go","editor":"vscode","line":-1}`, "line"),
	)
})

func postOpenInEditor(body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/projects/gavel/open", bytes.NewBufferString(body))
	request.SetPathValue("name", "gavel")
	(&Server{}).handleProjectOpen(recorder, request)
	return recorder
}
