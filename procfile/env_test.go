package procfile_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pf "github.com/flanksource/gavel/procfile"
)

var _ = Describe("LoadDotEnv", func() {
	write := func(body string) string {
		path := filepath.Join(GinkgoT().TempDir(), ".env")
		Expect(os.WriteFile(path, []byte(body), 0o644)).To(Succeed())
		return path
	}

	It("returns an empty map when the file is missing", func() {
		env, err := pf.LoadDotEnv(filepath.Join(GinkgoT().TempDir(), ".env"))
		Expect(err).NotTo(HaveOccurred())
		Expect(env).To(BeEmpty())
	})

	It("parses keys, comments, quotes and export prefixes", func() {
		path := write(strings.Join([]string{
			"# comment",
			"",
			"PLAIN=value",
			"export EXPORTED=exp",
			`QUOTED="with spaces"`,
			"SINGLE='single quoted'",
			"EMPTY=",
		}, "\n"))

		env, err := pf.LoadDotEnv(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(env).To(Equal(map[string]string{
			"PLAIN":    "value",
			"EXPORTED": "exp",
			"QUOTED":   "with spaces",
			"SINGLE":   "single quoted",
			"EMPTY":    "",
		}))
	})

	It("fails on a line without =", func() {
		_, err := pf.LoadDotEnv(write("NOEQUALS"))
		Expect(err).To(MatchError(ContainSubstring("expected KEY=value")))
	})
})

var _ = Describe("MergeEnv", func() {
	It("layers maps left-to-right with later layers winning", func() {
		merged := pf.MergeEnv(
			map[string]string{"A": "1", "B": "2"},
			map[string]string{"B": "20", "C": "3"},
		)
		Expect(merged).To(Equal(map[string]string{"A": "1", "B": "20", "C": "3"}))
	})

	It("ignores nil layers and never returns the parent environment", func() {
		merged := pf.MergeEnv(nil, map[string]string{"X": "y"})
		Expect(merged).To(Equal(map[string]string{"X": "y"}))
	})
})
