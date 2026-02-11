package git_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/git"
)

var _ = Describe("Argument Detection", func() {
	Describe("ParseArgs", func() {
		It("should separate commits, ranges, and paths correctly", func() {
			opts := git.HistoryOptions{
				Args: []string{
					"abc1234567890abcdef1234567890abcdef12345", // full SHA
					"abc1234",            // short SHA
					"main..feature",      // range
					"*.go",               // path
					"cmd/*.go",           // path with directory
					"main...develop",     // triple dot range
					"**/*kubernetes*.go", // glob path
				},
			}

			err := opts.ParseArgs()
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.CommitShas).To(ConsistOf(
				"abc1234567890abcdef1234567890abcdef12345",
				"abc1234",
			))

			Expect(opts.CommitRanges).To(ConsistOf(
				"main..feature",
				"main...develop",
			))

			Expect(opts.FilePaths).To(ConsistOf(
				"*.go",
				"cmd/*.go",
				"**/*kubernetes*.go",
			))
		})

		It("should handle branch names as commits", func() {
			opts := git.HistoryOptions{
				Args: []string{"main", "feature-branch", "v1.0.0"},
			}

			err := opts.ParseArgs()
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.CommitShas).To(ConsistOf("main", "feature-branch", "v1.0.0"))
		})

		It("should handle empty Args", func() {
			opts := git.HistoryOptions{
				Args: []string{},
			}

			err := opts.ParseArgs()
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.CommitShas).To(BeEmpty())
			Expect(opts.CommitRanges).To(BeEmpty())
			Expect(opts.FilePaths).To(BeEmpty())
		})

		It("should handle paths with wildcards", func() {
			opts := git.HistoryOptions{
				Args: []string{
					"**/*.yaml",
					"cmd/**/*.go",
					"*.md",
					"test?file.go",
				},
			}

			err := opts.ParseArgs()
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.FilePaths).To(HaveLen(4))
			Expect(opts.CommitShas).To(BeEmpty())
			Expect(opts.CommitRanges).To(BeEmpty())
		})

		It("should handle mixed content", func() {
			opts := git.HistoryOptions{
				Args: []string{
					"abc123",         // commit
					"main..branch",   // range
					"git/*.go",       // path
					"def456",         // commit
					"test/data/*.md", // path
				},
			}

			err := opts.ParseArgs()
			Expect(err).ToNot(HaveOccurred())

			Expect(opts.CommitShas).To(ConsistOf("abc123", "def456"))
			Expect(opts.CommitRanges).To(ConsistOf("main..branch"))
			Expect(opts.FilePaths).To(ConsistOf("git/*.go", "test/data/*.md"))
		})
	})

	Describe("detectArgType", func() {
		DescribeTable("should correctly detect argument types",
			func(arg string, expectedType string) {
				// We can't call detectArgType directly as it's not exported,
				// so we test via ParseArgs
				opts := git.HistoryOptions{Args: []string{arg}}
				err := opts.ParseArgs()
				Expect(err).ToNot(HaveOccurred())

				switch expectedType {
				case "commit":
					Expect(opts.CommitShas).To(ContainElement(arg))
				case "range":
					Expect(opts.CommitRanges).To(ContainElement(arg))
				case "path":
					Expect(opts.FilePaths).To(ContainElement(arg))
				}
			},
			// Commit SHAs
			Entry("full SHA (40 chars)", "abc1234567890abcdef1234567890abcdef12345", "commit"),
			Entry("short SHA (7 chars)", "abc1234", "commit"),
			Entry("short SHA (10 chars)", "abc1234567", "commit"),
			Entry("branch name", "main", "commit"),
			Entry("branch with dash", "feature-branch", "commit"),
			Entry("tag", "v1.0.0", "commit"),

			// Commit ranges
			Entry("two-dot range", "main..feature", "range"),
			Entry("three-dot range", "main...develop", "range"),
			Entry("SHA range", "abc123..def456", "range"),

			// File paths
			Entry("wildcard", "*.go", "path"),
			Entry("glob pattern", "**/*.yaml", "path"),
			Entry("path with directory", "cmd/test.go", "path"),
			Entry("path with wildcard", "cmd/*.go", "path"),
			Entry("question mark wildcard", "test?.go", "path"),
			Entry("nested path", "git/kubernetes/*.go", "path"),
		)
	})
})
