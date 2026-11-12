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
	DescribeTable("fails only on an execution failure or an error in a changed file", func(report string, linterExit, wantExit int) {
		root := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		GinkgoT().Setenv("GOWORK", "off")
		GinkgoT().Setenv("REACT_DOCTOR_TEST_REPORT", report)
		GinkgoT().Setenv("REACT_DOCTOR_TEST_EXIT", strconv.Itoa(linterExit))
		Expect(os.Mkdir(filepath.Join(root, "bin"), 0o755)).To(Succeed())
		script, err := os.ReadFile("../../linters/reactdoctor/testdata/react-doctor.sh")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.WriteFile(filepath.Join(root, "bin", "react-doctor"), script, 0o700)).To(Succeed())
		GinkgoT().Setenv("PATH", filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
		for file, content := range map[string]string{
			"go.mod":           "module example.com/changed\n\ngo 1.26\n",
			"package.json":     `{"name":"changed","dependencies":{"react":"18.0.0"}}` + "\n",
			"selected.go":      "package sample\n",
			"selected.tsx":     "export const selected = 1\n",
			"selected_test.go": "package sample\nimport \"testing\"\nfunc TestSelected(t *testing.T) {}\n",
			".gavel.yaml":      "lint:\n  linters:\n    golangci-lint:\n      enabled: false\n    betterleaks:\n      enabled: false\n",
		} {
			Expect(os.WriteFile(filepath.Join(root, file), []byte(content), 0o600)).To(Succeed())
		}
		previousExit := exitCode
		DeferCleanup(func() { exitCode = previousExit })
		exitCode = 0
		_, err = runTests(testrunner.RunOptions{WorkDir: root, Frameworks: []string{"go"}, ChangedFiles: []string{"selected.go", "selected.tsx"}, Lint: true, SkipHooks: true}, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(exitCode).To(Equal(wantExit))
	},
		Entry("execution failure without findings", ``, 2, 1),
		Entry("warning in a changed file", `{"diagnostics":[{"filePath":"selected.tsx","rule":"r","severity":"warning"}]}`, 0, 0),
		Entry("error only in an unchanged file", `{"diagnostics":[{"filePath":"other.tsx","rule":"r","severity":"error"}]}`, 0, 0),
		Entry("error in a changed file", `{"diagnostics":[{"filePath":"selected.tsx","rule":"r","severity":"error"}]}`, 0, 1),
	)
})
