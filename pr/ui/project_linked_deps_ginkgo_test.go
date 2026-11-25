package ui

import (
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("project linked dependencies", func() {
	It("marks external links across working manifests while allowing internal workspace links and deletions", func() {
		dir := createDiffRepository()
		manifests := map[string]string{
			"go.mod":                "module example.com/app\n\ngo 1.26\n\nreplace example.com/lib => ../lib\n",
			"go.work":               "go 1.26\n\nuse (\n.\n../lib\n)\n",
			"web/package.json":      `{"dependencies":{"external":"link:../../ui","internal":"workspace:*"}}`,
			"pnpm-lock.yaml":        "lockfileVersion: '9.0'\nimporters:\n  web:\n    dependencies:\n      external:\n        version: link:../../ui\n",
			"internal/package.json": `{"dependencies":{"internal":"file:../web"}}`,
		}
		var files []projectFileStatus
		for path, content := range manifests {
			writeDiffFile(dir, path, content)
			files = append(files, projectFileStatus{Path: path, State: status.StateUntracked, WorkKind: status.KindUntracked})
		}
		files = append(files, projectFileStatus{Path: "deleted/package.json", State: status.StateUnstaged, WorkKind: status.KindDeleted})
		Expect(enrichProjectLinkedDeps(dir, files)).To(Succeed())
		for _, file := range files {
			if file.Path == "internal/package.json" || file.Path == "deleted/package.json" {
				Expect(file.LinkedDependencies).To(BeEmpty(), file.Path)
			} else {
				Expect(file.LinkedDependencies).To(HaveLen(1), file.Path)
			}
		}
	})

	It("checks the index for staged-only files and the working copy for partially staged files", func() {
		dir := createDiffRepository()
		writeDiffFile(dir, "package.json", `{"dependencies":{"ui":"file:../ui"}}`)
		runDiffGit(dir, "add", "package.json")
		writeDiffFile(dir, "package.json", `{"dependencies":{"ui":"^1.0.0"}}`)
		files := []projectFileStatus{{Path: "package.json", State: status.StateStaged, StagedKind: status.KindAdded}}
		Expect(enrichProjectLinkedDeps(dir, files)).To(Succeed())
		Expect(files[0].LinkedDependencies).To(HaveLen(1))
		files[0].State = status.StateBoth
		files[0].WorkKind = status.KindModified
		Expect(enrichProjectLinkedDeps(dir, files)).To(Succeed())
		Expect(files[0].LinkedDependencies).To(BeEmpty())
	})

	It("returns errors for unreadable or malformed manifests", func() {
		dir := createDiffRepository()
		files := []projectFileStatus{{Path: "package.json", State: status.StateUntracked}}
		Expect(enrichProjectLinkedDeps(dir, files)).NotTo(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":`), 0600)).To(Succeed())
		Expect(enrichProjectLinkedDeps(dir, files)).To(MatchError(ContainSubstring("parse package.json")))
	})
})
