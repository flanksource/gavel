package procfile_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pf "github.com/flanksource/gavel/procfile"
)

var _ = Describe("Parse", func() {
	It("parses name: command lines, skipping blanks and comments", func() {
		src := strings.Join([]string{
			"# a comment",
			"",
			"web: bundle exec rails server",
			"  worker:   bundle exec rake jobs:work  ",
			"# trailing comment",
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

	It("fails on a line without a colon", func() {
		_, err := pf.Parse(strings.NewReader("web bundle exec rails"))
		Expect(err).To(MatchError(ContainSubstring("expected \"name: command\"")))
	})

	It("fails on an empty process name", func() {
		_, err := pf.Parse(strings.NewReader(": something"))
		Expect(err).To(MatchError(ContainSubstring("empty process name")))
	})

	It("fails on an invalid process name", func() {
		_, err := pf.Parse(strings.NewReader("web server: run"))
		Expect(err).To(MatchError(ContainSubstring("invalid process name")))
	})

	It("fails when a command is empty", func() {
		_, err := pf.Parse(strings.NewReader("web:   "))
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
