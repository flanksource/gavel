package git_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/flanksource/gavel/git"
	. "github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/models/kubernetes"
	"github.com/flanksource/gavel/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetChanges", func() {
	Context("with empty patch", func() {
		It("should return nil", func() {
			got, err := ParsePatch("")
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(BeNil())
		})
	})

	Context("with single file modification", func() {
		It("should parse adds and dels correctly", func() {
			patch := `diff --git a/file.go b/file.go
index abc123..def456 100644
--- a/file.go
+++ b/file.go
@@ -10,7 +10,8 @@
 line
-removed line
+added line
+another added line`

			got, err := ParsePatch(patch)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].File).To(Equal("file.go"))
			Expect(got[0].Type).To(Equal(SourceChangeTypeModified))
			Expect(got[0].Adds).To(Equal(2))
			Expect(got[0].Dels).To(Equal(1))
			Expect(got[0].LinesChanged).To(Equal(LineRanges("11-12")))
		})
	})

	Context("with new file", func() {
		It("should set type to added", func() {
			patch := `diff --git a/new.go b/new.go
new file mode 100644
index 0000000..abc123
--- /dev/null
+++ b/new.go
@@ -0,0 +1,3 @@
+line 1
+line 2
+line 3`

			got, err := ParsePatch(patch)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].File).To(Equal("new.go"))
			Expect(got[0].Type).To(Equal(SourceChangeTypeAdded))
			Expect(got[0].Adds).To(Equal(3))
			Expect(got[0].Dels).To(Equal(0))
			Expect(got[0].LinesChanged).To(Equal(LineRanges("1-3")))
		})
	})

	Context("with deleted file", func() {
		It("should set type to deleted", func() {
			patch := `diff --git a/old.go b/old.go
deleted file mode 100644
index abc123..0000000
--- a/old.go
+++ /dev/null
@@ -1,3 +0,0 @@
-line 1
-line 2
-line 3`

			got, err := ParsePatch(patch)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].File).To(Equal("old.go"))
			Expect(got[0].Type).To(Equal(SourceChangeTypeDeleted))
			Expect(got[0].Adds).To(Equal(0))
			Expect(got[0].Dels).To(Equal(3))
			Expect(got[0].LinesChanged).To(BeEmpty())
		})
	})

	Context("with renamed file", func() {
		It("should set type to renamed", func() {
			patch := `diff --git a/old.go b/new.go
rename from old.go
rename to new.go`

			got, err := ParsePatch(patch)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].File).To(Equal("new.go"))
			Expect(got[0].Type).To(Equal(SourceChangeTypeRenamed))
			Expect(got[0].Adds).To(Equal(0))
			Expect(got[0].Dels).To(Equal(0))
			Expect(got[0].LinesChanged).To(BeEmpty())
		})
	})

	Context("with multiple files", func() {
		It("should parse all files correctly", func() {
			patch := `diff --git a/file1.go b/file1.go
index abc123..def456 100644
--- a/file1.go
+++ b/file1.go
@@ -1,2 +1,3 @@
 line
+added line
diff --git a/file2.go b/file2.go
index xyz789..uvw012 100644
--- a/file2.go
+++ b/file2.go
@@ -1,3 +1,2 @@
 line
-removed line`

			got, err := ParsePatch(patch)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(HaveLen(2))
			Expect(got[0].File).To(Equal("file1.go"))
			Expect(got[0].Type).To(Equal(SourceChangeTypeModified))
			Expect(got[0].Adds).To(Equal(1))
			Expect(got[0].Dels).To(Equal(0))
			Expect(got[0].LinesChanged).To(Equal(LineRanges("2")))
			Expect(got[1].File).To(Equal("file2.go"))
			Expect(got[1].Type).To(Equal(SourceChangeTypeModified))
			Expect(got[1].Adds).To(Equal(0))
			Expect(got[1].Dels).To(Equal(1))
			Expect(got[1].LinesChanged).To(BeEmpty())
		})
	})

	Context("with spaces in folder names", func() {
		It("should parse file paths correctly", func() {
			patch := `diff --git a/folder with spaces/file.go b/folder with spaces/file.go
index abc123..def456 100644
--- a/folder with spaces/file.go
+++ b/folder with spaces/file.go
@@ -1,2 +1,3 @@
 line
+added line
+another added line`

			got, err := ParsePatch(patch)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].File).To(Equal("folder with spaces/file.go"))
			Expect(got[0].Type).To(Equal(SourceChangeTypeModified))
			Expect(got[0].Adds).To(Equal(2))
			Expect(got[0].Dels).To(Equal(0))
			Expect(got[0].LinesChanged).To(Equal(LineRanges("2-3")))
		})

		It("should handle quoted paths with spaces", func() {
			patch := `diff --git "a/folder with spaces/file.go" "b/folder with spaces/file.go"
index abc123..def456 100644
--- "a/folder with spaces/file.go"
+++ "b/folder with spaces/file.go"
@@ -1,1 +1,2 @@
 line
+new line`

			got, err := ParsePatch(patch)
			Expect(err).ToNot(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].File).To(Equal("folder with spaces/file.go"))
			Expect(got[0].Type).To(Equal(SourceChangeTypeModified))
			Expect(got[0].Adds).To(Equal(1))
			Expect(got[0].Dels).To(Equal(0))
		})
	})
})

var _ = Describe("ParseGitLogOutput", func() {
	Context("with empty output", func() {
		It("should return nil", func() {
			commits, err := ParseGitLogOutput([]byte(""))
			Expect(err).ToNot(HaveOccurred())
			Expect(commits).To(BeNil())
		})
	})

	Context("with single commit", func() {
		It("should parse all fields correctly", func() {
			// Build sample git log output with proper delimiters
			const (
				RS  = "\x1e"
				US  = "\x1f"
				GS  = "\x1d"
				NUL = "\x00"
			)

			hash := "abc123def456789"
			authorName := "John Doe"
			authorEmail := "john@example.com"
			authorDate := "2024-01-15T10:30:00-05:00"
			committerName := "Jane Smith"
			committerEmail := "jane@example.com"
			committerDate := "2024-01-15T11:00:00-05:00"
			subject := "feat(api): add new endpoint"
			body := "This adds a new API endpoint\n\nfor user management"
			trailerKeys := "Signed-off-by"
			trailerValues := "John Doe <john@example.com>"

			output := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s",
				RS, hash, US, authorName, US, authorEmail, US, authorDate, US,
				committerName, US, committerEmail, US, committerDate, US,
				subject, US, body, US, trailerKeys, US, trailerValues, NUL)

			commits, err := ParseGitLogOutput([]byte(output))
			Expect(err).ToNot(HaveOccurred())
			Expect(commits).To(HaveLen(1))

			commit := commits[0]
			Expect(commit.Hash).To(Equal(hash))
			Expect(commit.Author.Name).To(Equal(authorName))
			Expect(commit.Author.Email).To(Equal(authorEmail))
			Expect(commit.Committer.Name).To(Equal(committerName))
			Expect(commit.Committer.Email).To(Equal(committerEmail))
			Expect(commit.Subject).To(Equal("add new endpoint"))
			Expect(commit.CommitType).To(Equal(CommitTypeFeat))
			Expect(commit.Scope).To(Equal(ScopeType("api")))
			Expect(commit.Trailers).To(HaveKey("Signed-off-by"))
		})
	})

	Context("with multiple trailers of same key", func() {
		It("should concatenate values with comma", func() {
			const (
				RS  = "\x1e"
				US  = "\x1f"
				GS  = "\x1d"
				NUL = "\x00"
			)

			output := fmt.Sprintf("%sabc%sjohn%sjohn@example.com%s2024-01-15T10:30:00Z%sjane%sjane@example.com%s2024-01-15T11:00:00Z%sfeat: test%sbody%sCo-authored-by%sCo-authored-by%sPerson A <a@example.com>%sPerson B <b@example.com>%s",
				RS, US, US, US, US, US, US, US, US, US, GS, US, GS, NUL)

			commits, err := ParseGitLogOutput([]byte(output))
			Expect(err).ToNot(HaveOccurred())
			Expect(commits).To(HaveLen(1))
			Expect(commits[0].Trailers["Co-authored-by"]).To(ContainSubstring(","))
			Expect(commits[0].Trailers["Co-authored-by"]).To(ContainSubstring("Person A"))
			Expect(commits[0].Trailers["Co-authored-by"]).To(ContainSubstring("Person B"))
		})
	})

	Context("with patch data", func() {
		It("should include patch in commit", func() {
			const (
				RS  = "\x1e"
				US  = "\x1f"
				NUL = "\x00"
			)

			patch := "diff --git a/file.go b/file.go\n+new line"
			output := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s",
				RS, "abc", US, "john", US, "john@example.com", US, "2024-01-15T10:30:00Z", US,
				"jane", US, "jane@example.com", US, "2024-01-15T11:00:00Z", US,
				"fix: bug", US, "body", US, "", US, "", NUL, patch)

			commits, err := ParseGitLogOutput([]byte(output))
			Expect(err).ToNot(HaveOccurred())
			Expect(commits).To(HaveLen(1))
			Expect(commits[0].Patch).To(Equal(patch))
		})
	})
})

var _ = Describe("GetCommitHistory", func() {
	Context("basic usage", func() {
		It("should retrieve commits from current repo", func() {
			filter := HistoryOptions{
				Path: ".",
			}

			commits, err := GetCommitHistory(filter)
			Expect(err).ToNot(HaveOccurred())
			Expect(commits).ToNot(BeNil())
		})
	})

	Context("with ShowPatch option", func() {
		It("should include patch data when enabled", func() {
			filter := HistoryOptions{
				Path:      ".",
				ShowPatch: true,
			}

			commits, err := GetCommitHistory(filter)
			Expect(err).ToNot(HaveOccurred())

			// At least some commits should have patches
			if len(commits) > 0 {
				hasPatches := false
				for _, commit := range commits {
					if commit.Patch != "" {
						hasPatches = true
						break
					}
				}
				// This may not always be true for merge commits, but should be true for most
				if len(commits) > 5 {
					Expect(hasPatches).To(BeTrue())
				}
			}
		})

		It("should not include patch data when disabled", func() {
			filter := HistoryOptions{
				Path:      ".",
				ShowPatch: false,
			}

			commits, err := GetCommitHistory(filter)
			Expect(err).ToNot(HaveOccurred())

			// All patches should be empty or contain only whitespace
			for _, commit := range commits {
				trimmed := strings.TrimSpace(commit.Patch)
				Expect(trimmed).To(BeEmpty(), "Commit %s has patch data when ShowPatch is false", commit.Hash[:8])
			}
		})
	})

	Context("with date filtering", func() {
		It("should filter by Since date", func() {
			since := time.Now().AddDate(0, 0, -7) // Last 7 days
			filter := HistoryOptions{
				Path:  ".",
				Since: since,
			}

			commits, err := GetCommitHistory(filter)
			Expect(err).ToNot(HaveOccurred())

			for _, commit := range commits {
				Expect(commit.Author.Date.After(since) || commit.Author.Date.Equal(since)).To(BeTrue())
			}
		})

		It("should filter by Until date", func() {
			until := time.Now().AddDate(0, 0, -30) // More than 30 days ago
			filter := HistoryOptions{
				Path:  ".",
				Until: until,
			}

			commits, err := GetCommitHistory(filter)
			Expect(err).ToNot(HaveOccurred())

			for _, commit := range commits {
				Expect(commit.Author.Date.Before(until) || commit.Author.Date.Equal(until)).To(BeTrue())
			}
		})
	})

	Context("with author filtering", func() {
		It("should filter by author name or email", func() {
			// First get all commits to find an author
			allCommits, err := GetCommitHistory(HistoryOptions{Path: "."})
			Expect(err).ToNot(HaveOccurred())

			if len(allCommits) == 0 {
				Skip("No commits found in repository")
			}

			targetAuthor := allCommits[0].Author.Name
			filter := HistoryOptions{
				Path:   ".",
				Author: []string{targetAuthor},
			}

			commits, err := GetCommitHistory(filter)
			Expect(err).ToNot(HaveOccurred())
			Expect(commits).ToNot(BeEmpty())

			for _, commit := range commits {
				Expect(commit.Author.Name).To(ContainSubstring(targetAuthor))
			}
		})

		It("should filter by multiple authors with OR logic", func() {
			// First get all commits to find authors
			allCommits, err := GetCommitHistory(HistoryOptions{Path: "."})
			Expect(err).ToNot(HaveOccurred())

			if len(allCommits) < 2 {
				Skip("Need at least 2 commits with different authors")
			}

			// Find two different authors
			author1 := allCommits[0].Author.Name
			author2 := ""
			for _, commit := range allCommits {
				if commit.Author.Name != author1 {
					author2 = commit.Author.Name
					break
				}
			}

			if author2 == "" {
				Skip("All commits have same author")
			}

			// Test with multiple authors
			filter := HistoryOptions{
				Path:   ".",
				Author: []string{author1, author2},
			}

			commits, err := GetCommitHistory(filter)
			Expect(err).ToNot(HaveOccurred())
			Expect(commits).ToNot(BeEmpty())

			// Verify all commits match at least one of the authors
			for _, commit := range commits {
				matchesAny := strings.Contains(commit.Author.Name, author1) || strings.Contains(commit.Author.Name, author2)
				Expect(matchesAny).To(BeTrue(), "Commit %s author '%s' should match one of: %s, %s",
					commit.Hash[:8], commit.Author.Name, author1, author2)
			}
		})
	})

	Context("with message filtering", func() {
		It("should filter by commit message", func() {
			// Note: Git --grep uses regex to search raw commit messages.
			// The Matches() filter then applies glob patterns to parsed Subject.
			// Use "Merge branch*" which:
			// - Git interprets as regex matching "Merge branch" (the * matches 0+ of 'h')
			// - collections.MatchItem interprets as glob matching subjects starting with "Merge branch"
			filter := HistoryOptions{
				Path:    ".",
				Message: "Merge branch*",
			}

			commits, err := GetCommitHistory(filter)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(commits)).To(BeNumerically(">", 0))

			for _, commit := range commits {
				Expect(commit.Subject).To(HavePrefix("Merge branch"))
			}
		})
	})

	Context("with invalid path", func() {
		It("should return an error", func() {
			filter := HistoryOptions{
				Path: "/nonexistent/path/to/nowhere",
			}

			_, err := GetCommitHistory(filter)
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("AnalyzeCommitHistory", func() {
	Context("with commits containing patches", func() {
		It("should extract changes from patches", func() {
			patch := `diff --git a/file.go b/file.go
index abc123..def456 100644
--- a/file.go
+++ b/file.go
@@ -1,2 +1,3 @@
 line
+added line`

			commits := []Commit{
				{
					Hash:    "abc123def456",
					Subject: "add new feature",
					Patch:   patch,
				},
			}

			// Create analyzer context
			ctx, err := NewAnalyzerContext(context.Background(), ".")
			Expect(err).ToNot(HaveOccurred())

			options := AnalyzeOptions{}

			analyses, err := AnalyzeCommitHistory(ctx, commits, options)
			Expect(err).ToNot(HaveOccurred())
			Expect(analyses).To(HaveLen(1))
			Expect(analyses[0].Hash).To(Equal("abc123def456"))
			Expect(analyses[0].Changes).To(HaveLen(1))
			Expect(analyses[0].Changes[0].File).To(Equal("file.go"))
			Expect(analyses[0].Changes[0].Adds).To(Equal(1))
		})
	})
})

var _ = Describe("LoadCommitAnalysesFromJSON", func() {
	Context("with valid JSON files", func() {
		It("should load and merge multiple files", func() {
			// Create temporary test files
			tempDir := GinkgoT().TempDir()

			// Sample commit analysis data
			analysis1 := CommitAnalyses{
				{
					Commit: Commit{
						Hash:       "abc123",
						Repository: "repo1",
						Subject:    "feat: add feature",
						Author:     Author{Name: "Alice", Email: "alice@example.com", Date: time.Now()},
					},
					Changes: Changes{
						{File: "file1.go", Type: SourceChangeTypeAdded, Adds: 10},
					},
				},
			}

			analysis2 := CommitAnalyses{
				{
					Commit: Commit{
						Hash:       "def456",
						Repository: "repo2",
						Subject:    "fix: bug fix",
						Author:     Author{Name: "Bob", Email: "bob@example.com", Date: time.Now()},
					},
					Changes: Changes{
						{File: "file2.go", Type: SourceChangeTypeModified, Adds: 5, Dels: 3},
					},
				},
			}

			// Write test files
			file1 := fmt.Sprintf("%s/repo1-analysis.json", tempDir)
			file2 := fmt.Sprintf("%s/repo2-analysis.json", tempDir)

			data1, err := json.Marshal(analysis1)
			Expect(err).ToNot(HaveOccurred())
			err = os.WriteFile(file1, data1, 0644)
			Expect(err).ToNot(HaveOccurred())

			data2, err := json.Marshal(analysis2)
			Expect(err).ToNot(HaveOccurred())
			err = os.WriteFile(file2, data2, 0644)
			Expect(err).ToNot(HaveOccurred())

			// Load files
			result, err := LoadCommitAnalysesFromJSON([]string{file1, file2})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(2))
			Expect(result[0].Hash).To(Equal("abc123"))
			Expect(result[0].Repository).To(Equal("repo1"))
			Expect(result[1].Hash).To(Equal("def456"))
			Expect(result[1].Repository).To(Equal("repo2"))
		})
	})

	Context("with missing repository field", func() {
		It("should return an error", func() {
			tempDir := GinkgoT().TempDir()

			// Create commit without repository field
			analysis := CommitAnalyses{
				{
					Commit: Commit{
						Hash:    "abc123",
						Subject: "feat: add feature",
					},
				},
			}

			file := fmt.Sprintf("%s/invalid-analysis.json", tempDir)
			data, err := json.Marshal(analysis)
			Expect(err).ToNot(HaveOccurred())
			err = os.WriteFile(file, data, 0644)
			Expect(err).ToNot(HaveOccurred())

			_, err = LoadCommitAnalysesFromJSON([]string{file})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing repository field"))
		})
	})

	Context("with non-existent file", func() {
		It("should return an error", func() {
			_, err := LoadCommitAnalysesFromJSON([]string{"/nonexistent/file.json"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to read file"))
		})
	})

	Context("with invalid JSON", func() {
		It("should return an error", func() {
			tempDir := GinkgoT().TempDir()
			file := fmt.Sprintf("%s/invalid.json", tempDir)
			err := os.WriteFile(file, []byte("not valid json"), 0644)
			Expect(err).ToNot(HaveOccurred())

			_, err = LoadCommitAnalysesFromJSON([]string{file})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to unmarshal"))
		})
	})
})

var _ = Describe("ApplyFilters", func() {
	Context("with author filter", func() {
		It("should filter commits by author", func() {
			commits := CommitAnalyses{
				{
					Commit: Commit{
						Hash:       "abc123",
						Repository: "repo1",
						Subject:    "feat: add feature",
						Author:     Author{Name: "Alice", Email: "alice@example.com", Date: time.Now()},
					},
				},
				{
					Commit: Commit{
						Hash:       "def456",
						Repository: "repo1",
						Subject:    "fix: bug fix",
						Author:     Author{Name: "Bob", Email: "bob@example.com", Date: time.Now()},
					},
				},
			}

			filter := HistoryOptions{
				Author: []string{"Alice"},
			}

			result := ApplyFilters(commits, filter)
			Expect(result).To(HaveLen(1))
			Expect(result[0].Author.Name).To(Equal("Alice"))
		})
	})

	Context("with date filter", func() {
		It("should filter commits by date range", func() {
			now := time.Now()
			yesterday := now.AddDate(0, 0, -1)
			lastWeek := now.AddDate(0, 0, -7)

			commits := CommitAnalyses{
				{
					Commit: Commit{
						Hash:       "abc123",
						Repository: "repo1",
						Subject:    "recent commit",
						Author:     Author{Date: now},
					},
				},
				{
					Commit: Commit{
						Hash:       "def456",
						Repository: "repo1",
						Subject:    "old commit",
						Author:     Author{Date: lastWeek},
					},
				},
			}

			filter := HistoryOptions{
				Since: yesterday,
			}

			result := ApplyFilters(commits, filter)
			Expect(result).To(HaveLen(1))
			Expect(result[0].Hash).To(Equal("abc123"))
		})
	})

	Context("with no filters", func() {
		It("should return all commits", func() {
			commits := CommitAnalyses{
				{
					Commit: Commit{Hash: "abc123", Repository: "repo1"},
				},
				{
					Commit: Commit{Hash: "def456", Repository: "repo1"},
				},
			}

			filter := HistoryOptions{}

			result := ApplyFilters(commits, filter)
			Expect(result).To(HaveLen(2))
		})
	})
})

// Helper function to run git commands in a directory
func runCommand(dir, cmd string) error {
	// Parse command respecting quoted strings
	parts := parseCommand(cmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	command := exec.Command(parts[0], parts[1:]...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command '%s' failed: %w\nOutput: %s", cmd, err, string(output))
	}
	return nil
}

// parseCommand parses a command string respecting quoted arguments
func parseCommand(cmd string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for _, r := range cmd {
		switch {
		case r == '"' || r == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = r
			} else if r == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current.WriteRune(r)
			}
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// Helper function to get commit hash
func getCommitHash(dir, ref string) (string, error) {
	command := exec.Command("git", "rev-parse", ref)
	command.Dir = dir
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash for %s: %w", ref, err)
	}
	return strings.TrimSpace(string(output)), nil
}

var _ = Describe("E2E: Git Analyzer and DiffMap for Deleted K8s Files", func() {
	var (
		tempDir     string
		repoPath    string
		ctx         *AnalyzerContext
		initialHash string
		deleteHash  string
	)

	BeforeEach(func() {
		var err error
		tempDir = GinkgoT().TempDir()
		repoPath = fmt.Sprintf("%s/test-repo", tempDir)

		// Create and initialize git repository
		err = os.MkdirAll(repoPath, 0755)
		Expect(err).ToNot(HaveOccurred())

		// Initialize git repo
		commands := []string{
			"git init",
			"git config user.name 'Test User'",
			"git config user.email 'test@example.com'",
		}
		for _, cmd := range commands {
			err = runCommand(repoPath, cmd)
			Expect(err).ToNot(HaveOccurred())
		}

		// Create Kubernetes YAML file with multiple resources
		k8sYAML := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: nginx
        image: nginx:1.21.0
        ports:
        - containerPort: 80
        env:
        - name: ENVIRONMENT
          value: production
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: production
data:
  database.url: postgres://db.example.com:5432/mydb
  cache.ttl: "3600"
  max.connections: "100"
---
apiVersion: v1
kind: Service
metadata:
  name: web-service
  namespace: production
spec:
  type: LoadBalancer
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
`
		k8sFile := fmt.Sprintf("%s/k8s-resources.yaml", repoPath)
		err = os.WriteFile(k8sFile, []byte(k8sYAML), 0644)
		Expect(err).ToNot(HaveOccurred())

		// Commit the file
		err = runCommand(repoPath, "git add k8s-resources.yaml")
		Expect(err).ToNot(HaveOccurred())
		err = runCommand(repoPath, `git commit -m "Add Kubernetes resources"`)
		Expect(err).ToNot(HaveOccurred())

		// Get the initial commit hash
		initialHash, err = getCommitHash(repoPath, "HEAD")
		Expect(err).ToNot(HaveOccurred())
		Expect(initialHash).ToNot(BeEmpty())

		// Delete the file
		err = os.Remove(k8sFile)
		Expect(err).ToNot(HaveOccurred())

		// Commit the deletion
		err = runCommand(repoPath, "git add k8s-resources.yaml")
		Expect(err).ToNot(HaveOccurred())
		err = runCommand(repoPath, `git commit -m "Remove Kubernetes resources"`)
		Expect(err).ToNot(HaveOccurred())

		// Get the deletion commit hash
		deleteHash, err = getCommitHash(repoPath, "HEAD")
		Expect(err).ToNot(HaveOccurred())
		Expect(deleteHash).ToNot(BeEmpty())
		Expect(deleteHash).ToNot(Equal(initialHash))

		// Create analyzer context
		ctx, err = NewAnalyzerContext(context.Background(), repoPath)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should analyze deleted Kubernetes resources correctly", func() {
		// Get the deletion commit
		commits, err := GetCommitHistory(HistoryOptions{
			Path:      repoPath,
			ShowPatch: true,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(commits).ToNot(BeEmpty())

		// Find the deletion commit
		var deletionCommit *Commit
		for i := range commits {
			if commits[i].Hash == deleteHash {
				deletionCommit = &commits[i]
				break
			}
		}
		Expect(deletionCommit).ToNot(BeNil(), "Should find deletion commit %s", deleteHash)
		Expect(deletionCommit.Patch).ToNot(BeEmpty(), "Deletion commit should have patch data")

		// Analyze the deletion commit
		analysis, err := AnalyzeCommit(ctx, *deletionCommit, AnalyzeOptions{})
		Expect(err).ToNot(HaveOccurred())

		// Verify commit analysis
		Expect(analysis.Hash).To(Equal(deleteHash))
		Expect(analysis.Changes).To(HaveLen(1), "Should have one file change")
		Expect(analysis.Changes[0].File).To(Equal("k8s-resources.yaml"))
		Expect(analysis.Changes[0].Type).To(Equal(SourceChangeTypeDeleted))

		// Verify Kubernetes changes were detected
		Expect(analysis.Changes[0].KubernetesChanges).To(HaveLen(3), "Should detect 3 deleted resources")

		// Verify each deleted resource
		resourcesFound := make(map[string]bool)
		for _, k8sChange := range analysis.Changes[0].KubernetesChanges {
			// All changes should be deletions
			Expect(k8sChange.ChangeType).To(Equal(kubernetes.SourceChangeTypeDeleted))
			Expect(k8sChange.Severity).To(Equal(kubernetes.ChangeSeverityHigh))

			// Track which resources we found
			key := fmt.Sprintf("%s/%s", k8sChange.Kind, k8sChange.Name)
			resourcesFound[key] = true

			// Verify namespace
			Expect(k8sChange.Namespace).To(Equal("production"))

			// Verify Before map is populated (the resource that was deleted)
			Expect(k8sChange.Before).ToNot(BeEmpty(), "Before map should contain deleted resource data")

			// Verify After map is empty (resource no longer exists)
			Expect(k8sChange.After).To(BeEmpty(), "After map should be empty for deleted resources")

			// Verify specific resource details
			switch k8sChange.Kind {
			case "Deployment":
				Expect(k8sChange.Name).To(Equal("web-app"))
				Expect(k8sChange.Before["spec"].(map[string]any)["replicas"]).To(BeNumerically("==", 3))
			case "ConfigMap":
				Expect(k8sChange.Name).To(Equal("app-config"))
				data := k8sChange.Before["data"].(map[string]any)
				Expect(data).To(HaveKey("database.url"))
				Expect(data["database.url"]).To(Equal("postgres://db.example.com:5432/mydb"))
			case "Service":
				Expect(k8sChange.Name).To(Equal("web-service"))
				Expect(k8sChange.Before["spec"].(map[string]any)["type"]).To(Equal("LoadBalancer"))
			default:
				Fail(fmt.Sprintf("Unexpected resource kind: %s", k8sChange.Kind))
			}

			// Verify line numbers are tracked
			Expect(k8sChange.StartLine).To(BeNumerically(">", 0))
			Expect(k8sChange.EndLine).To(BeNumerically(">=", k8sChange.StartLine))
		}

		// Verify we found all expected resources
		Expect(resourcesFound).To(HaveKey("Deployment/web-app"))
		Expect(resourcesFound).To(HaveKey("ConfigMap/app-config"))
		Expect(resourcesFound).To(HaveKey("Service/web-service"))
	})

	It("should format deleted resources with DiffMap correctly", func() {
		// Get and analyze the deletion commit
		commits, err := GetCommitHistory(HistoryOptions{
			Path:      repoPath,
			ShowPatch: true,
		})
		Expect(err).ToNot(HaveOccurred())

		var deletionCommit *Commit
		for i := range commits {
			if commits[i].Hash == deleteHash {
				deletionCommit = &commits[i]
				break
			}
		}
		Expect(deletionCommit).ToNot(BeNil())

		analysis, err := AnalyzeCommit(ctx, *deletionCommit, AnalyzeOptions{})
		Expect(err).ToNot(HaveOccurred())
		Expect(analysis.Changes[0].KubernetesChanges).To(HaveLen(3))

		// Test DiffMap formatting for each deleted resource
		for _, k8sChange := range analysis.Changes[0].KubernetesChanges {
			// Create DiffMap from before/after
			beforeMap := utils.DiffMap[any](k8sChange.Before)
			afterMap := utils.DiffMap[any](k8sChange.After)

			// Generate diff
			diff := beforeMap.Diff(afterMap)

			// For deletions, all fields from before should appear in diff as removed
			Expect(diff).ToNot(BeEmpty(), "Diff should show removed fields for %s/%s", k8sChange.Kind, k8sChange.Name)

			// Verify specific fields are in the diff
			switch k8sChange.Kind {
			case "Deployment":
				// Check that key deployment fields appear in diff
				hasReplicasField := false
				hasImageField := false
				for key := range diff {
					if strings.Contains(key, "replicas") {
						hasReplicasField = true
					}
					if strings.Contains(key, "image") {
						hasImageField = true
					}
				}
				Expect(hasReplicasField).To(BeTrue(), "Diff should contain replicas field")
				Expect(hasImageField).To(BeTrue(), "Diff should contain image field")

			case "ConfigMap":
				// Check that data fields appear in diff
				hasDataField := false
				for key := range diff {
					if strings.Contains(key, "data") {
						hasDataField = true
						break
					}
				}
				Expect(hasDataField).To(BeTrue(), "Diff should contain data fields")

			case "Service":
				// Check that service type appears in diff
				hasTypeField := false
				for key := range diff {
					if strings.Contains(key, "type") {
						hasTypeField = true
						break
					}
				}
				Expect(hasTypeField).To(BeTrue(), "Diff should contain type field")
			}

			// Test Pretty formatting
			prettyOutput := beforeMap.Pretty()
			Expect(prettyOutput).ToNot(BeNil())
			outputStr := prettyOutput.String()
			Expect(outputStr).ToNot(BeEmpty(), "Pretty output should not be empty for %s/%s", k8sChange.Kind, k8sChange.Name)
		}
	})
})

var _ = Describe("E2E: Helm Chart Analysis with Go Templates", func() {
	var (
		tempDir     string
		repoPath    string
		ctx         *AnalyzerContext
		initialHash string
		modifyHash  string
	)

	BeforeEach(func() {
		var err error
		tempDir = GinkgoT().TempDir()
		repoPath = fmt.Sprintf("%s/test-repo", tempDir)

		// Create and initialize git repository
		err = os.MkdirAll(repoPath, 0755)
		Expect(err).ToNot(HaveOccurred())

		// Initialize git repo
		commands := []string{
			"git init",
			"git config user.name 'Test User'",
			"git config user.email 'test@example.com'",
		}
		for _, cmd := range commands {
			err = runCommand(repoPath, cmd)
			Expect(err).ToNot(HaveOccurred())
		}

		// Create Helm chart directory structure
		err = os.MkdirAll(fmt.Sprintf("%s/charts/app/templates", repoPath), 0755)
		Expect(err).ToNot(HaveOccurred())

		// Create Helm template with Go template syntax
		helmTemplate := `{{- if eq .Values.puppeteer.enabled true}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: iiab-puppeteer-pdf
  labels: {{- include "puppeteer.metadata.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.puppeteer.replicas }}
  selector:
    matchLabels: {{- include "puppeteer.metadata.labels" . | nindent 6 }}
  template:
    metadata:
      labels: {{- include "puppeteer.metadata.labels" . | nindent 8 }}
      annotations:
        config/puppeteer: {{ tpl (print $.Template.BasePath "/puppeteer/secret.yaml") . | sha256sum }}
        {{- if not (empty (merge (default dict .Values.annotations) (default dict .Values.puppeteer.annotations))) }}
        {{- toYaml (merge (default dict .Values.annotations) (default dict .Values.puppeteer.annotations)) | nindent 8 }}
        {{- end }}
    spec:
      {{- if eq .Values.puppeteer.serviceAccount.enabled true }}
      serviceAccountName: {{ .Values.puppeteer.serviceAccount.name }}
      {{- end }}
      containers:
        - name: iiab-puppeteer-pdf
          image: {{ include "puppeteer.image" . }}
          imagePullPolicy: {{ .Values.puppeteer.image.pullPolicy }}
          command:
          - dotnet
          - IIAB.PuppeteerPdfService.Api.dll
          ports:
          - containerPort: 8080
            protocol: TCP
            name: http
          {{- if eq (include "puppeteer.metrics.enabled" .) "true" }}
          - containerPort: 9090
            protocol: TCP
            name: metrics
          {{- end }}
          env:
            - name: IIAB_ConfigLocation
              value: '/configs'
            {{- with .Values.aspnetcoreEnvironment }}
            - name: ASPNETCORE_ENVIRONMENT
              value: {{ . }}
            {{- end }}
          {{- with .Values.puppeteer.resources }}
          resources:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          readinessProbe:
            httpGet:
              path: /api/health/echo
              port: 8080
              scheme: HTTP
            initialDelaySeconds: 10
            timeoutSeconds: 30
            periodSeconds: 30
          livenessProbe:
            httpGet:
              path: /api/health/echo
              port: 8080
            initialDelaySeconds: 60
            periodSeconds: 20
            failureThreshold: 1
      terminationGracePeriodSeconds: 200
{{- end }}
`
		helmFile := fmt.Sprintf("%s/charts/app/templates/deployment.yaml", repoPath)
		err = os.WriteFile(helmFile, []byte(helmTemplate), 0644)
		Expect(err).ToNot(HaveOccurred())

		// Commit the Helm template
		err = runCommand(repoPath, "git add charts/")
		Expect(err).ToNot(HaveOccurred())
		err = runCommand(repoPath, `git commit -m "Add Helm chart with Go templates"`)
		Expect(err).ToNot(HaveOccurred())

		// Get the initial commit hash
		initialHash, err = getCommitHash(repoPath, "HEAD")
		Expect(err).ToNot(HaveOccurred())
		Expect(initialHash).ToNot(BeEmpty())

		// Modify some values in the template
		modifiedTemplate := `{{- if eq .Values.puppeteer.enabled true}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: iiab-puppeteer-pdf
  labels: {{- include "puppeteer.metadata.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.puppeteer.replicas }}
  selector:
    matchLabels: {{- include "puppeteer.metadata.labels" . | nindent 6 }}
  template:
    metadata:
      labels: {{- include "puppeteer.metadata.labels" . | nindent 8 }}
      annotations:
        config/puppeteer: {{ tpl (print $.Template.BasePath "/puppeteer/secret.yaml") . | sha256sum }}
        {{- if not (empty (merge (default dict .Values.annotations) (default dict .Values.puppeteer.annotations))) }}
        {{- toYaml (merge (default dict .Values.annotations) (default dict .Values.puppeteer.annotations)) | nindent 8 }}
        {{- end }}
    spec:
      {{- if eq .Values.puppeteer.serviceAccount.enabled true }}
      serviceAccountName: {{ .Values.puppeteer.serviceAccount.name }}
      {{- end }}
      containers:
        - name: iiab-puppeteer-pdf
          image: {{ include "puppeteer.image" . }}
          imagePullPolicy: {{ .Values.puppeteer.image.pullPolicy }}
          command:
          - dotnet
          - IIAB.PuppeteerPdfService.Api.dll
          ports:
          - containerPort: 8080
            protocol: TCP
            name: http
          {{- if eq (include "puppeteer.metrics.enabled" .) "true" }}
          - containerPort: 9090
            protocol: TCP
            name: metrics
          {{- end }}
          env:
            - name: IIAB_ConfigLocation
              value: '/configs'
            {{- with .Values.aspnetcoreEnvironment }}
            - name: ASPNETCORE_ENVIRONMENT
              value: {{ . }}
            {{- end }}
          {{- with .Values.puppeteer.resources }}
          resources:
            {{- toYaml . | nindent 12 }}
          {{- end }}
          readinessProbe:
            httpGet:
              path: /api/health/echo
              port: 8080
              scheme: HTTP
            initialDelaySeconds: 15
            timeoutSeconds: 45
            periodSeconds: 60
          livenessProbe:
            httpGet:
              path: /api/health/echo
              port: 8080
            initialDelaySeconds: 90
            periodSeconds: 30
            failureThreshold: 3
      terminationGracePeriodSeconds: 300
{{- end }}
`
		err = os.WriteFile(helmFile, []byte(modifiedTemplate), 0644)
		Expect(err).ToNot(HaveOccurred())

		// Commit the modification
		err = runCommand(repoPath, "git add charts/")
		Expect(err).ToNot(HaveOccurred())
		err = runCommand(repoPath, `git commit -m "Update Helm chart probe timings and grace period"`)
		Expect(err).ToNot(HaveOccurred())

		// Get the modification commit hash
		modifyHash, err = getCommitHash(repoPath, "HEAD")
		Expect(err).ToNot(HaveOccurred())
		Expect(modifyHash).ToNot(BeEmpty())
		Expect(modifyHash).ToNot(Equal(initialHash))

		// Create analyzer context
		ctx, err = NewAnalyzerContext(context.Background(), repoPath)
		Expect(err).ToNot(HaveOccurred())
	})

	PIt("should analyze Helm chart changes with Go template syntax", func() {
		// Get the modification commit
		commits, err := GetCommitHistory(HistoryOptions{
			Path:      repoPath,
			ShowPatch: true,
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(commits).ToNot(BeEmpty())

		// Find the modification commit
		var modifyCommit *Commit
		for i := range commits {
			if commits[i].Hash == modifyHash {
				modifyCommit = &commits[i]
				break
			}
		}
		Expect(modifyCommit).ToNot(BeNil(), "Should find modification commit %s", modifyHash)
		Expect(modifyCommit.Patch).ToNot(BeEmpty(), "Modification commit should have patch data")

		// Analyze the modification commit
		analysis, err := AnalyzeCommit(ctx, *modifyCommit, AnalyzeOptions{})
		Expect(err).ToNot(HaveOccurred())

		// Verify commit analysis
		Expect(analysis.Hash).To(Equal(modifyHash))
		Expect(analysis.Changes).To(HaveLen(1), "Should have one file change")
		Expect(analysis.Changes[0].File).To(Equal("charts/app/templates/deployment.yaml"))
		Expect(analysis.Changes[0].Type).To(Equal(SourceChangeTypeModified))

		// The analyzer should attempt to parse the Helm template
		// It may or may not detect Kubernetes changes depending on whether it can handle the Go template syntax
		// At minimum, it should not crash
		GinkgoWriter.Printf("Detected %d Kubernetes changes in Helm template\n", len(analysis.Changes[0].KubernetesChanges))

		// If Kubernetes changes were detected, verify they make sense
		if len(analysis.Changes[0].KubernetesChanges) > 0 {
			for _, k8sChange := range analysis.Changes[0].KubernetesChanges {
				// Verify basic structure
				Expect(k8sChange.Kind).ToNot(BeEmpty(), "Kind should be detected")
				GinkgoWriter.Printf("  - Found %s change: %s\n", k8sChange.ChangeType, k8sChange.Kind)

				// If it's a Deployment, check for expected fields
				if k8sChange.Kind == "Deployment" {
					Expect(k8sChange.ChangeType).To(Equal(kubernetes.SourceChangeTypeModified))
					GinkgoWriter.Printf("    Deployment change detected\n")
				}
			}
		}
	})

	PIt("should handle Helm templates gracefully even if unable to parse Go syntax", func() {
		// Get and analyze the commit
		commits, err := GetCommitHistory(HistoryOptions{
			Path:      repoPath,
			ShowPatch: true,
		})
		Expect(err).ToNot(HaveOccurred())

		var modifyCommit *Commit
		for i := range commits {
			if commits[i].Hash == modifyHash {
				modifyCommit = &commits[i]
				break
			}
		}
		Expect(modifyCommit).ToNot(BeNil())

		// The main requirement: should not crash or error
		analysis, err := AnalyzeCommit(ctx, *modifyCommit, AnalyzeOptions{})
		Expect(err).ToNot(HaveOccurred())
		Expect(analysis).ToNot(BeNil())

		// Should at least detect that a YAML file was modified
		Expect(analysis.Changes).To(HaveLen(1))
		Expect(analysis.Changes[0].File).To(HaveSuffix(".yaml"))
		Expect(analysis.Changes[0].Type).To(Equal(SourceChangeTypeModified))
		Expect(analysis.Changes[0].Adds).To(BeNumerically(">", 0), "Should track added lines")
		Expect(analysis.Changes[0].Dels).To(BeNumerically(">", 0), "Should track deleted lines")

		GinkgoWriter.Printf("Successfully analyzed Helm template: %d adds, %d dels\n",
			analysis.Changes[0].Adds, analysis.Changes[0].Dels)
	})
})
