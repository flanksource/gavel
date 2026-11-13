package runners

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// writeTree creates each Cwd-relative file under root with the given content.
func writeTree(root string, files map[string]string) {
	for rel, content := range files {
		path := filepath.Join(root, rel)
		Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
		Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed())
	}
}

var _ = Describe("FileSelection", func() {
	DescribeTable("Pattern renders one RE2 expression",
		func(sel FileSelection, want string) {
			Expect(sel.Pattern()).To(Equal(want))
		},
		Entry("no entries", FileSelection{}, ""),
		Entry("exact is quoted and anchored", FileSelection{Exact: []string{"Test.A(b)"}}, `^Test\.A\(b\)$`),
		Entry("glob * is any run", FileSelection{Glob: []string{"Test*"}}, `^Test.*$`),
		Entry("glob ? is one character and metachars are quoted", FileSelection{Glob: []string{"a.b?"}}, `^a\.b.$`),
		Entry("a single regexp is kept as given", FileSelection{Regex: []string{"login|logout"}}, "login|logout"),
		Entry("several entries are each wrapped and ORed",
			FileSelection{Exact: []string{"A"}, Glob: []string{"B*"}, Regex: []string{"c+"}},
			`(?:^A$)|(?:^B.*$)|(?:c+)`),
	)

	It("rejects an invalid regexp", func() {
		_, err := FileSelection{Regex: []string{"("}}.Pattern()
		Expect(err).To(MatchError(ContainSubstring(`invalid test pattern "("`)))
	})

	DescribeTable("Matches",
		func(sel FileSelection, value string, want bool) {
			Expect(sel.Matches(value)).To(Equal(want))
		},
		Entry("exact value", FileSelection{Exact: []string{"src/a.ts"}}, "src/a.ts", true),
		Entry("exact is not a prefix", FileSelection{Exact: []string{"src/a"}}, "src/a.ts", false),
		Entry("doublestar glob crosses directories", FileSelection{Glob: []string{"src/**/*.test.ts"}}, "src/x/y/a.test.ts", true),
		Entry("single-star glob stays in one directory", FileSelection{Glob: []string{"src/*.ts"}}, "src/x/a.ts", false),
		Entry("unanchored regexp", FileSelection{Regex: []string{`\.spec\.`}}, "e2e/home.spec.ts", true),
	)

	It("reports an invalid glob from Matches", func() {
		_, err := FileSelection{Glob: []string{"["}}.Matches("a")
		Expect(err).To(MatchError(ContainSubstring(`invalid glob "["`)))
	})

	It("Merge unions each list without duplicates, keeping first-seen order", func() {
		merged := FileSelection{Exact: []string{"a", "b"}, Regex: []string{"x"}}.
			Merge(FileSelection{Exact: []string{"b", "c"}, Glob: []string{"g"}, Regex: []string{"x"}})
		Expect(merged).To(Equal(FileSelection{Exact: []string{"a", "b", "c"}, Glob: []string{"g"}, Regex: []string{"x"}}))
	})

	Describe("Expand", func() {
		var root string

		BeforeEach(func() {
			root = GinkgoT().TempDir()
			writeTree(root, map[string]string{
				"src/a.test.ts":                  "",
				"src/nested/b.test.ts":           "",
				"src/c.ts":                       "",
				"lib/e.test.ts":                  "",
				"node_modules/dep/d.test.ts":     "",
				"src/node_modules/dep/f.test.ts": "",
			})
		})

		DescribeTable("resolves to sorted root-relative files",
			func(sel, skip FileSelection, want []string) {
				Expect(sel.Expand(root, skip)).To(Equal(want))
			},
			Entry("doublestar glob", FileSelection{Glob: []string{"src/**/*.test.ts"}}, FileSelection{},
				[]string{"src/a.test.ts", "src/nested/b.test.ts"}),
			Entry("regexp skips node_modules", FileSelection{Regex: []string{`\.test\.ts$`}}, FileSelection{},
				[]string{"lib/e.test.ts", "src/a.test.ts", "src/nested/b.test.ts"}),
			Entry("exact with ./ prefix", FileSelection{Exact: []string{"./src/c.ts"}}, FileSelection{},
				[]string{"src/c.ts"}),
			Entry("skip selection drops matches", FileSelection{Glob: []string{"**/*.test.ts"}}, FileSelection{Glob: []string{"src/nested/**"}},
				[]string{"lib/e.test.ts", "src/a.test.ts"}),
		)

		It("fails on a missing exact file", func() {
			_, err := FileSelection{Exact: []string{"src/missing.ts"}}.Expand(root, FileSelection{})
			Expect(err).To(MatchError(ContainSubstring("test file src/missing.ts")))
		})

		It("fails when the selection resolves to nothing", func() {
			_, err := FileSelection{Glob: []string{"**/*.go"}}.Expand(root, FileSelection{})
			Expect(err).To(MatchError(ContainSubstring("match glob:**/*.go")))
		})

		It("fails when every match is skipped", func() {
			_, err := FileSelection{Exact: []string{"src/c.ts"}}.Expand(root, FileSelection{Exact: []string{"src/c.ts"}})
			Expect(err).To(MatchError(ContainSubstring("no test files under")))
		})
	})
})
