package procfile_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pf "github.com/flanksource/gavel/procfile"
)

// writeFakeShell writes an executable script that ignores its arguments and
// prints body verbatim, standing in for a login shell printing `env`.
func writeFakeShell(body string) string {
	path := filepath.Join(GinkgoT().TempDir(), "fakeshell")
	script := "#!/bin/sh\nprintf '%s' " + shellSingleQuote(body) + "\n"
	Expect(os.WriteFile(path, []byte(script), 0o755)).To(Succeed())
	return path
}

func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

var _ = Describe("homeForPath", func() {
	It("derives the home from a workspace path under /Users", func() {
		Expect(pf.HomeForPath("/Users/moshe/go/src/github.com/flanksource/clicky-ui")).
			To(Equal("/Users/moshe"))
	})

	It("derives the home from a workspace path under /home", func() {
		Expect(pf.HomeForPath("/home/dev/projects/app")).To(Equal("/home/dev"))
	})

	It("falls back to the running user's home outside any home parent", func() {
		home, err := os.UserHomeDir()
		Expect(err).NotTo(HaveOccurred())
		Expect(pf.HomeForPath("/opt/workspaces/app")).To(Equal(home))
	})
})

var _ = Describe("ParseEnvOutput", func() {
	It("parses KEY=value lines and ignores malformed and continuation lines", func() {
		env := pf.ParseEnvOutput(strings.NewReader(strings.Join([]string{
			"PATH=/usr/bin:/bin",
			"FOO=bar=baz",
			"",
			"not an env line",
			"EMPTY=",
		}, "\n")))
		Expect(env).To(Equal(map[string]string{
			"PATH":  "/usr/bin:/bin",
			"FOO":   "bar=baz",
			"EMPTY": "",
		}))
	})
})

var _ = Describe("LoadUserEnv", func() {
	It("captures the login-shell env and drops cwd-coupled vars", func() {
		restore := pf.SetShellForHome(func(string) string {
			return writeFakeShell("PATH=/injected/bin\nFOO=bar\nPWD=/somewhere\nSHLVL=2\n")
		})
		defer restore()

		env, err := pf.LoadUserEnv(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		Expect(env).To(HaveKeyWithValue("PATH", "/injected/bin"))
		Expect(env).To(HaveKeyWithValue("FOO", "bar"))
		Expect(env).NotTo(HaveKey("PWD"))
		Expect(env).NotTo(HaveKey("SHLVL"))
	})
})
