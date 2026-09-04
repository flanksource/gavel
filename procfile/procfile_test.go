package procfile_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pf "github.com/flanksource/gavel/procfile"
	"github.com/flanksource/gavel/verify"
)

var _ = Describe("Parse", func() {
	It("parses string entries, preserving order and ignoring comments", func() {
		src := strings.Join([]string{
			"# a comment",
			"",
			"web: bundle exec rails server",
			"worker: bundle exec rake jobs:work",
		}, "\n")

		entries, err := pf.Parse(strings.NewReader(src))
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(Equal([]pf.Entry{
			{Name: "web", Command: "bundle exec rails server"},
			{Name: "worker", Command: "bundle exec rake jobs:work"},
		}))
	})

	It("keeps colons that appear inside the command", func() {
		entries, err := pf.Parse(strings.NewReader("db: psql postgres://localhost:5432/app"))
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1))
		Expect(entries[0].Command).To(Equal("psql postgres://localhost:5432/app"))
	})

	It("parses the object form with all fields", func() {
		src := `worker:
  command: rake jobs:work
  default: false
  autoRestart: on-failure
  cpu: 100
  mem: 512Mi
  maxRestarts: 10
  env:
    WORKER_CONCURRENCY: "4"
`
		entries, err := pf.Parse(strings.NewReader(src))
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(HaveLen(1))
		e := entries[0]
		Expect(e.Name).To(Equal("worker"))
		Expect(e.Command).To(Equal("rake jobs:work"))
		Expect(e.Default).NotTo(BeNil())
		Expect(*e.Default).To(BeFalse())
		Expect(e.AutoRestart).To(Equal(verify.RestartOnFailure))
		Expect(e.CPU).To(Equal(100.0))
		Expect(e.Mem).To(Equal("512Mi"))
		Expect(e.MaxRestarts).NotTo(BeNil())
		Expect(*e.MaxRestarts).To(Equal(10))
		Expect(e.Env).To(Equal(map[string]string{"WORKER_CONCURRENCY": "4"}))
	})

	It("accepts autoRestart as a bool (true => on-failure)", func() {
		entries, err := pf.Parse(strings.NewReader("web:\n  command: serve\n  autoRestart: true\n"))
		Expect(err).NotTo(HaveOccurred())
		Expect(entries[0].AutoRestart).To(Equal(verify.RestartOnFailure))
	})

	It("accepts profiles as a scalar or a list", func() {
		one, err := pf.Parse(strings.NewReader("a:\n  command: x\n  profiles: dev\n"))
		Expect(err).NotTo(HaveOccurred())
		Expect(one[0].Profiles).To(Equal([]string{"dev"}))

		many, err := pf.Parse(strings.NewReader("a:\n  command: x\n  profiles: [dev, ci]\n"))
		Expect(err).NotTo(HaveOccurred())
		Expect(many[0].Profiles).To(Equal([]string{"dev", "ci"}))
	})

	It("preserves declaration order across mixed string and object forms", func() {
		src := "z: run-z\na:\n  command: run-a\nm: run-m\n"
		entries, err := pf.Parse(strings.NewReader(src))
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{entries[0].Name, entries[1].Name, entries[2].Name}).
			To(Equal([]string{"z", "a", "m"}))
	})

	It("fails when the top level is not a mapping", func() {
		_, err := pf.Parse(strings.NewReader("just a bare string"))
		Expect(err).To(MatchError(ContainSubstring("expected a mapping")))
	})

	It("fails on an invalid process name", func() {
		_, err := pf.Parse(strings.NewReader("web server: run"))
		Expect(err).To(MatchError(ContainSubstring("invalid process name")))
	})

	It("fails when a string command is empty", func() {
		_, err := pf.Parse(strings.NewReader("web:"))
		Expect(err).To(MatchError(ContainSubstring("has no command")))
	})

	It("fails when an object entry has no command", func() {
		_, err := pf.Parse(strings.NewReader("web:\n  cpu: 10\n"))
		Expect(err).To(MatchError(ContainSubstring("has no command")))
	})

	It("fails on a duplicate process name", func() {
		_, err := pf.Parse(strings.NewReader("web: a\nweb: b"))
		Expect(err).To(MatchError(ContainSubstring("duplicate process name")))
	})

	It("fails when there are no process definitions", func() {
		_, err := pf.Parse(strings.NewReader("# only comments\n\n"))
		Expect(err).To(MatchError(ContainSubstring("no process definitions")))
	})

	It("fails on malformed YAML", func() {
		_, err := pf.Parse(strings.NewReader("web: [unterminated"))
		Expect(err).To(MatchError(ContainSubstring("invalid Procfile")))
	})
})

var _ = Describe("Load", func() {
	It("errors when the Procfile is missing", func() {
		_, err := pf.Load(filepath.Join(GinkgoT().TempDir(), "Procfile"))
		Expect(err).To(MatchError(ContainSubstring("open Procfile")))
	})

	It("parses a file from disk", func() {
		dir := GinkgoT().TempDir()
		path := filepath.Join(dir, "Procfile")
		Expect(os.WriteFile(path, []byte("web: serve\n"), 0o644)).To(Succeed())
		entries, err := pf.Load(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(entries).To(Equal([]pf.Entry{{Name: "web", Command: "serve"}}))
	})
})

var _ = Describe("Find", func() {
	It("returns an absolute override verbatim", func() {
		Expect(pf.Find("/somewhere", "/etc/Procfile")).To(Equal("/etc/Procfile"))
	})

	It("resolves a relative override against dir", func() {
		Expect(pf.Find("/repo/app", "Procfile.dev")).To(Equal("/repo/app/Procfile.dev"))
	})

	It("discovers the nearest Procfile up to the git root", func() {
		root := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(root, ".git"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "Procfile"), []byte("web: serve\n"), 0o644)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "cmd", "app"), 0o755)).To(Succeed())

		Expect(pf.Find(filepath.Join(root, "cmd", "app"), "")).To(Equal(filepath.Join(root, "Procfile")))
	})

	It("returns empty when no Procfile exists", func() {
		root := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(root, ".git"), 0o755)).To(Succeed())
		Expect(pf.Find(root, "")).To(BeEmpty())
	})
})

var _ = Describe("Select", func() {
	entries := []pf.Entry{
		{Name: "web", Command: "serve"},
		{Name: "worker", Command: "work"},
		{Name: "clock", Command: "tick"},
	}

	It("returns all entries when no names are given", func() {
		got, err := pf.Select(entries, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(entries))
	})

	It("returns the named subset in Procfile order", func() {
		got, err := pf.Select(entries, []string{"clock", "web"})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]pf.Entry{
			{Name: "web", Command: "serve"},
			{Name: "clock", Command: "tick"},
		}))
	})

	It("errors on an unknown name", func() {
		_, err := pf.Select(entries, []string{"web", "ghost"})
		Expect(err).To(MatchError(ContainSubstring("ghost")))
	})
})
