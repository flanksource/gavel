package main

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/flanksource/gavel/testrunner"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Changed combined lint verdict", func() {
	DescribeTable("preserves the shared execution and severity verdict", func(report string, wantExit int) {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("GOWORK", "off")
		GinkgoT().Setenv("UIR_TEST_REPORT", report)
		GinkgoT().Setenv("UIR_TEST_EXIT", strconv.Itoa(wantExit))
		Expect(os.Mkdir(filepath.Join(root, "bin"), 0o755)).To(Succeed())
		script, err := os.ReadFile("../../linters/uir/testdata/uir.sh")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(root, "bin", "uir"), script, 0o700)).To(Succeed())
		GinkgoT().Setenv("PATH", filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
		for file, content := range map[string]string{
			"go.mod":           "module example.com/changed\n\ngo 1.26\n",
			"selected.go":      "package sample\n",
			"selected_test.go": "package sample\nimport \"testing\"\nfunc TestSelected(t *testing.T) {}\n",
			".uir.yaml":        "version: 1\n",
			".gavel.yaml":      "lint:\n  linters:\n    golangci-lint:\n      enabled: false\n    betterleaks:\n      enabled: false\n",
		} {
			Expect(os.WriteFile(filepath.Join(root, file), []byte(content), 0o600)).To(Succeed())
		}
		previousExit := exitCode
		DeferCleanup(func() { exitCode = previousExit })
		exitCode = 0
		_, err = runTests(testrunner.RunOptions{WorkDir: root, Frameworks: []string{"go"}, ChangedFiles: []string{"selected.go"}, Lint: true, SkipHooks: true}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(exitCode).To(Equal(wantExit))
	},
		Entry("execution failure without findings", `[{"linter":"uir","success":false,"error":"analysis failed","violations":[]}]`, 1),
		Entry("warning in a selected file", `[{"linter":"uir","success":true,"violations":[{"file":"selected.go","severity":"warning","message":"review this declaration"}]}]`, 0),
	)
})
