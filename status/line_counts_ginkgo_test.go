package status

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Status line counts", func() {
	DescribeTable("excludes tracked assets and binary source files",
		func(state FileState) {
			dir := GinkgoT().TempDir()
			git := func(args ...string) {
				cmd := exec.Command("git", args...)
				cmd.Dir = dir
				out, err := cmd.CombinedOutput()
				Expect(err).NotTo(HaveOccurred(), "%s", out)
			}
			git("init", "-q")
			git("config", "user.name", "Test User")
			git("config", "user.email", "test@example.com")
			git("config", "commit.gpgsign", "false")
			contents := map[string]string{"main.go": "package main\n", "image.svg": "<svg/>\n", "blob.go": "\x00binary\n"}
			for path, content := range contents {
				Expect(os.WriteFile(filepath.Join(dir, path), []byte(content), 0o600)).To(Succeed())
			}
			git("add", "main.go", "image.svg", "blob.go")
			git("commit", "-qm", "baseline")
			for path, content := range contents {
				Expect(os.WriteFile(filepath.Join(dir, path), []byte(content+"first\n"), 0o600)).To(Succeed())
			}
			if state != StateUnstaged {
				git("add", "main.go", "image.svg", "blob.go")
			}
			if state == StateBoth {
				for path, content := range contents {
					Expect(os.WriteFile(filepath.Join(dir, path), []byte(content+"first\nsecond\n"), 0o600)).To(Succeed())
				}
			}
			files := []FileStatus{{Path: "main.go", State: state}, {Path: "image.svg", State: state}, {Path: "blob.go", State: state}}
			Expect(enrichWithLineCounts(context.Background(), dir, nil, files)).To(Succeed())
			wantAdds := 1
			if state == StateBoth {
				wantAdds = 2
			}
			Expect(files).To(Equal([]FileStatus{
				{Path: "main.go", State: state, Adds: wantAdds}, {Path: "image.svg", State: state}, {Path: "blob.go", State: state},
			}))
		},
		Entry("staged", StateStaged), Entry("unstaged", StateUnstaged), Entry("both", StateBoth),
	)

	It("keeps untracked assets visible without counting their content as lines", func() {
		dir := GinkgoT().TempDir()
		cmd := exec.Command("git", "init", "-q", dir)
		Expect(cmd.Run()).To(Succeed())
		contents := map[string]string{
			"main.go": "package main\n", "README.md": "# Example\nDetails",
			"settings.yaml": "enabled: true\n", "notes.custom": "one\ntwo\n",
			"image.svg": "<svg>\n</svg>\n", "report.pdf": "%PDF-1.7\nbody\n",
			"image.PNG": "image\ndata\n", "blob": "\x00binary\nbytes\n",
			"disguised.go": "\x00binary\nbytes\n", "empty.go": "",
		}
		for path, content := range contents {
			Expect(os.WriteFile(filepath.Join(dir, path), []byte(content), 0o600)).To(Succeed())
		}
		result, err := GatherBase(dir, Options{NoRepomap: true, NoResults: true})
		Expect(err).NotTo(HaveOccurred())
		counts := map[string][2]int{}
		for _, file := range result.Files {
			counts[file.Path] = [2]int{file.Adds, file.Dels}
		}
		Expect(counts).To(Equal(map[string][2]int{
			"main.go": {1, 0}, "README.md": {2, 0}, "settings.yaml": {1, 0}, "notes.custom": {2, 0},
			"image.svg": {}, "report.pdf": {}, "image.PNG": {}, "blob": {}, "disguised.go": {}, "empty.go": {},
		}))
		Expect(result.Counts()).To(Equal(Counts{Untracked: len(contents), Adds: 6}))
	})
})
