package prwatch

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// stickyComment builds the PR comment gavel posts per matrix shard: a sticky id
// plus the artifact link FindGavelArtifacts scans for.
func stickyComment(id int64, stickyID string, artifactID int64) github.PRComment {
	return github.PRComment{
		ID: id,
		Body: fmt.Sprintf("<!-- sticky-comment:%s -->\n[View full results](https://github.com/acme/widgets/actions/runs/123/artifacts/%d)",
			stickyID, artifactID),
	}
}

func prWith(comments ...github.PRComment) prFetcher {
	return func(github.Options, int) (*github.PRInfo, error) {
		return &github.PRInfo{Number: 12, Comments: comments}, nil
	}
}

var _ = Describe("PR snapshot download", func() {
	It("merges every matrix shard's results into one snapshot", func() {
		fetch := prWith(
			stickyComment(10, "gavel-test", 1),
			stickyComment(20, "gavel-lint", 2),
		)
		download := func(_ github.Options, artifactID int64) ([]byte, error) {
			if artifactID == 1 {
				return []byte(`{"tests":[{"name":"Alpha","failed":true},{"name":"Beta","passed":true}]}`), nil
			}
			return []byte(`{"tests":[{"name":"Gamma","failed":true}],"lint":[{"linter":"tsc","violations":[{"file":"a.ts"}]}]}`), nil
		}

		snap, err := downloadPRSnapshot(github.Options{Repo: "acme/widgets"}, 12, fetch, download)

		Expect(err).NotTo(HaveOccurred())
		Expect(snap.Tests).To(HaveLen(3))
		Expect(snap.Lint).To(HaveLen(1))
		Expect(snap.Status.LintRun).To(BeTrue())
	})

	It("accepts the legacy bare-array artifact shape", func() {
		snap, err := downloadPRSnapshot(github.Options{}, 12,
			prWith(stickyComment(10, "gavel", 1)),
			func(github.Options, int64) ([]byte, error) {
				return []byte(`[{"name":"Alpha","failed":true}]`), nil
			})

		Expect(err).NotTo(HaveOccurred())
		Expect(snap.Tests).To(HaveLen(1))
		Expect(snap.Tests[0].Name).To(Equal("Alpha"))
	})

	DescribeTable("rebases matrix shard packages to the repository root before merging",
		func(gitRoot, shardWorkDir string) {
			payload, err := json.Marshal(testui.Snapshot{
				Git: &testui.SnapshotGit{Repo: "widgets", Root: gitRoot, SHA: "abc123"},
				Tests: []parsers.Test{{
					Name:        "./",
					PackagePath: "./.",
					WorkDir:     shardWorkDir,
					Failed:      true,
					Children: []parsers.Test{{
						Name:        "fails in the shard root",
						PackagePath: "./.",
						WorkDir:     shardWorkDir,
						File:        "external_entities_test.go",
						Framework:   parsers.Ginkgo,
						Failed:      true,
					}},
				}},
			})
			Expect(err).NotTo(HaveOccurred())

			snap, err := downloadPRSnapshot(github.Options{}, 12,
				prWith(stickyComment(10, "gavel-scrapers", 1)),
				func(github.Options, int64) ([]byte, error) { return payload, nil })

			Expect(err).NotTo(HaveOccurred())
			Expect(snap.Git).To(Equal(&testui.SnapshotGit{Repo: "widgets", Root: gitRoot, SHA: "abc123"}))
			Expect(snap.Tests).To(HaveLen(1))
			Expect(snap.Tests[0].PackagePath).To(Equal("./scrapers"))
			Expect(snap.Tests[0].WorkDir).To(Equal(gitRoot))
			Expect(snap.Tests[0].Children).To(HaveLen(1))
			Expect(snap.Tests[0].Children[0].PackagePath).To(Equal("./scrapers"))
			Expect(snap.Tests[0].Children[0].WorkDir).To(Equal(gitRoot))
			Expect(snap.Tests[0].Children[0].File).To(Equal("scrapers/external_entities_test.go"))
		},
		Entry("unix paths", "/home/runner/work/widgets/widgets", "/home/runner/work/widgets/widgets/scrapers"),
		Entry("windows paths", `D:\a\widgets\widgets`, `D:\a\widgets\widgets\scrapers`),
	)

	It("rejects artifact test paths outside the repository", func() {
		payload, err := json.Marshal(testui.Snapshot{
			Git: &testui.SnapshotGit{Repo: "widgets", Root: "/repo/widgets"},
			Tests: []parsers.Test{{
				Name:        "outside",
				PackagePath: "./.",
				WorkDir:     "/repo/other",
				Failed:      true,
			}},
		})
		Expect(err).NotTo(HaveOccurred())

		_, err = downloadPRSnapshot(github.Options{}, 12,
			prWith(stickyComment(10, "gavel", 1)),
			func(github.Options, int64) ([]byte, error) { return payload, nil })

		Expect(err).To(MatchError(ContainSubstring("outside git root")))
	})

	It("merges sticky shards from different commits without claiming one SHA", func() {
		download := func(_ github.Options, artifactID int64) ([]byte, error) {
			payload, err := json.Marshal(testui.Snapshot{
				Git: &testui.SnapshotGit{
					Repo: "widgets",
					Root: fmt.Sprintf("/runner/%d/widgets", artifactID),
					SHA:  fmt.Sprintf("sha-%d", artifactID),
				},
				Tests: []parsers.Test{{
					Name:        fmt.Sprintf("shard-%d", artifactID),
					PackagePath: "./.",
					WorkDir:     fmt.Sprintf("/runner/%d/widgets", artifactID),
					Passed:      true,
				}},
			})
			return payload, err
		}

		snap, err := downloadPRSnapshot(github.Options{}, 12,
			prWith(stickyComment(10, "gavel-core", 1), stickyComment(20, "gavel-scrapers", 2)),
			download)

		Expect(err).NotTo(HaveOccurred())
		Expect(snap.Tests).To(HaveLen(2))
		Expect(snap.Git).NotTo(BeNil())
		Expect(snap.Git.Repo).To(Equal("widgets"))
		Expect(snap.Git.SHA).To(BeEmpty())
	})

	It("fails loudly when the PR published no gavel artifacts", func() {
		_, err := downloadPRSnapshot(github.Options{}, 12,
			prWith(github.PRComment{ID: 10, Body: "looks good to me"}),
			func(github.Options, int64) ([]byte, error) { return nil, errors.New("never called") })

		Expect(err).To(MatchError(ContainSubstring("PR #12 has no gavel result artifacts")))
	})

	It("fails loudly when an artifact cannot be downloaded", func() {
		_, err := downloadPRSnapshot(github.Options{}, 12,
			prWith(stickyComment(10, "gavel", 7)),
			func(github.Options, int64) ([]byte, error) { return nil, errors.New("403 forbidden") })

		Expect(err).To(MatchError(ContainSubstring("download artifact 7 (gavel)")))
		Expect(err).To(MatchError(ContainSubstring("403 forbidden")))
	})

	It("reports when every artifact is missing its results payload", func() {
		_, err := downloadPRSnapshot(github.Options{}, 12,
			prWith(stickyComment(10, "gavel", 1)),
			func(github.Options, int64) ([]byte, error) { return nil, github.ErrArtifactResultsNotFound })

		Expect(err).To(MatchError(ContainSubstring("none of the 1 gavel artifacts contained a results payload")))
	})
})
