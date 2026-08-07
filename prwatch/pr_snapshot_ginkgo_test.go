package prwatch

import (
	"errors"
	"fmt"

	"github.com/flanksource/gavel/github"
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
