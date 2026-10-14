package betterleaks

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	commonsContext "github.com/flanksource/commons/context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fakeBetterleaksScript writes an empty findings report to --report-path,
// standing in for the real scanner so the test only observes where gavel
// places its own artifacts.
const fakeBetterleaksScript = `#!/bin/sh
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--report-path" ]; then
    printf '[]' >"$2"
    exit 0
  fi
  shift
done
echo "missing --report-path" >&2
exit 1
`

func gitRun(dir string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-c", "user.name=test", "-c", "user.email=test@example.com"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %v: %s", args, out)
	return string(out)
}

var _ = Describe("Betterleaks run artifacts", func() {
	It("leaves a repository without .gitignore clean after a run that merges configs and git ignores", func() {
		home := GinkgoT().TempDir()
		GinkgoT().Setenv("HOME", home)
		GinkgoT().Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdg"))
		GinkgoT().Setenv("XDG_CACHE_HOME", filepath.Join(home, "cache"))
		Expect(os.WriteFile(filepath.Join(home, ".betterleaks.toml"), []byte("title = \"home\"\n"), 0o644)).To(Succeed())

		binDir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(binDir, "betterleaks"), []byte(fakeBetterleaksScript), 0o755)).To(Succeed())
		GinkgoT().Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

		repo := GinkgoT().TempDir()
		gitRun(repo, "init", "-q")
		Expect(os.WriteFile(filepath.Join(repo, ".betterleaks.toml"), []byte("title = \"repo\"\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repo, "main.go"), []byte("package main\n"), 0o644)).To(Succeed())
		gitRun(repo, "add", ".betterleaks.toml", "main.go")
		gitRun(repo, "commit", "-q", "-m", "seed")
		Expect(os.WriteFile(filepath.Join(repo, ".git", "info", "exclude"), []byte("debug.log\n"), 0o644)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(repo, "debug.log"), nil, 0o644)).To(Succeed())

		violations, err := NewBetterleaks(repo).Run(commonsContext.NewContext(context.Background()), nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(violations).To(BeEmpty())
		Expect(gitRun(repo, "status", "--porcelain", "--untracked-files=all")).To(BeEmpty())
	})
})
