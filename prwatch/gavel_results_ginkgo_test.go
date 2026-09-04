package prwatch

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/flanksource/gavel/testrunner/parsers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gavel artifact results", func() {
	It("hides HTML comments while preserving fenced examples in review output", func() {
		result := PRWatchResult{
			PR: &github.PRInfo{
				Number: 57, Title: "comments", Author: github.PRAuthor{Login: "alice"},
				HeadRefName: "feature", BaseRefName: "main",
			},
			Comments: []github.PRComment{{
				ID: 20,
				Body: strings.Join([]string{
					"Visible finding",
					"<!-- hidden workflow metadata -->",
					"```html",
					"<!-- visible fenced example -->",
					"```",
				}, "\n"),
			}},
		}

		ansi := result.Pretty().ANSI()
		Expect(ansi).To(ContainSubstring("Visible finding"))
		Expect(ansi).NotTo(ContainSubstring("hidden workflow metadata"))
		Expect(ansi).To(ContainSubstring("<!-- visible fenced example -->"))
	})

	It("renders structured failures and lint violations in PR status", func() {
		result := PRWatchResult{
			PR: &github.PRInfo{
				Number: 57, Title: "artifact failures", Author: github.PRAuthor{Login: "alice"},
				HeadRefName: "feature", BaseRefName: "main",
			},
			GavelResults: []*GavelResultsSummary{{
				StickyID:       "gavel",
				ArtifactID:     456,
				ArtifactURL:    "https://github.com/acme/widgets/actions/runs/123/artifacts/456",
				TestsPassed:    9,
				TestsFailed:    2,
				TestsTotal:     11,
				LintViolations: 1,
				LintLinters:    1,
				Failures: []parsers.Test{
					{Name: "Combo", File: "combo.test.tsx", Line: 42, Message: "missing API radio", Failed: true},
					{Name: "Advanced Config", Message: "missing Model field", Failed: true},
				},
				Lint: []*linters.LinterResult{{
					Linter:  "tsc",
					Success: true,
					Violations: []models.Violation{{
						File: "config.ts", Line: 7,
						Rule:    &models.Rule{Method: "TS2322"},
						Message: models.StringPtr("type mismatch"),
					}},
				}},
			}},
		}

		// Plain text, so the assertions read against the shared renderers'
		// output rather than their ANSI styling.
		plain := result.Pretty().String()
		Expect(plain).To(ContainSubstring("Gavel Results:"))
		// parsers.TestSummary.Pretty's roll-up, the same line `gavel test` ends on.
		Expect(plain).To(ContainSubstring("passed: 9"))
		Expect(plain).To(ContainSubstring("failed: 2"))
		Expect(plain).To(ContainSubstring("total: 11"))
		Expect(plain).To(ContainSubstring("Test failures (2)"))
		Expect(plain).To(ContainSubstring("combo.test.tsx:42"))
		Expect(plain).To(ContainSubstring("Combo"))
		Expect(plain).To(ContainSubstring("Advanced Config"))
		Expect(plain).To(ContainSubstring("View full results"))
	})

	It("prints the gavel commands that re-run a failing shard", func() {
		results := []*GavelResultsSummary{
			{StickyID: "gavel-test", TestsFailed: 2, TestsTotal: 11},
			{StickyID: "gavel-lint", LintViolations: 3, LintLinters: 1},
			{StickyID: "gavel-e2e", TestsPassed: 4, TestsTotal: 4},
		}

		annotateReproduceCommands(results, "", 57)

		Expect(results[0].Commands).To(Equal([]string{"gavel test --pr 57"}))
		Expect(results[1].Commands).To(Equal([]string{"gavel lint --pr 57"}))
		Expect(results[2].Commands).To(BeEmpty())

		plain := PRWatchResult{PR: &github.PRInfo{Number: 57}, GavelResults: results}.Pretty().String()
		Expect(plain).To(ContainSubstring("Reproduce locally"))
		Expect(plain).To(ContainSubstring("$ gavel test --pr 57"))
		Expect(plain).To(ContainSubstring("$ gavel lint --pr 57"))
	})

	It("suggests both commands when one shard runs tests and lint", func() {
		// A broken linter reports no violations, so the lint command must be
		// driven by the same signal statusIcon uses, not by the count alone.
		results := []*GavelResultsSummary{{
			StickyID:    "gavel",
			TestsFailed: 1,
			Lint: []*linters.LinterResult{{
				Linter: "golangci-lint", Success: false,
				Error: "golangci-lint execution failed: exit status 3",
			}},
		}}

		annotateReproduceCommands(results, "acme/widgets", 12)

		Expect(results[0].Commands).To(Equal([]string{
			"gavel test --pr acme/widgets#12",
			"gavel lint --pr acme/widgets#12",
		}))
	})

	It("renders lint findings through the shared gavel lint summary tree", func() {
		result := PRWatchResult{
			PR: &github.PRInfo{Number: 57},
			GavelResults: []*GavelResultsSummary{{
				StickyID:       "gavel-lint",
				LintLinters:    1,
				LintViolations: 3,
				Lint: []*linters.LinterResult{{
					Linter:  "tsc",
					Success: true,
					Violations: []models.Violation{{
						File: "config.ts", Line: 7,
						Rule:    &models.Rule{Method: "TS2322"},
						Message: models.StringPtr("type mismatch"),
					}},
				}},
			}},
		}

		plain := result.Pretty().String()
		// The linter → rule → file hierarchy is linters.SummaryView's, not a
		// bespoke one, and the header reports the artifact's true total.
		Expect(plain).To(ContainSubstring("Lint summary: 1 of 3 violations"))
		Expect(plain).To(ContainSubstring("tsc (1 violations)"))
		Expect(plain).To(ContainSubstring("TS2322 (1) — type mismatch"))
		Expect(plain).To(ContainSubstring("config.ts:7"))
	})

	It("does not report a shard as passing when a linter failed to run", func() {
		// A linter that crashed produces no violations, so counts alone say
		// "clean". The carried LinterResult is the only evidence.
		summary := GavelResultsSummary{
			StickyID:    "gavel-lint",
			LintLinters: 1,
			Lint: []*linters.LinterResult{{
				Linter:  "golangci-lint",
				Success: false,
				Error:   "golangci-lint execution failed: exit status 3",
			}},
		}

		Expect(summary.statusIcon()).To(Equal(icons.Warning))
		plain := summary.Pretty().String()
		Expect(plain).To(ContainSubstring(icons.Warning.Unicode))
	})

	It("removes only comments backed by rendered artifacts", func() {
		comments := []github.PRComment{
			{ID: 10, Body: "<!-- sticky-comment:gavel -->\nresults"},
			{ID: 20, Body: "review finding"},
		}
		results := []*GavelResultsSummary{{CommentID: 10, ArtifactID: 456}}

		filtered := removeRenderedArtifactComments(comments, results)

		Expect(filtered).To(Equal([]github.PRComment{{ID: 20, Body: "review finding"}}))
	})

	It("scopes artifact results with workflow action filters", func() {
		result := &PRWatchResult{
			PR: &github.PRInfo{StatusCheckRollup: github.StatusChecks{
				{Name: "gavel", Status: "COMPLETED", Conclusion: "FAILURE", DetailsURL: "https://github.com/acme/widgets/actions/runs/123/job/1"},
				{Name: "analyze", Status: "COMPLETED", Conclusion: "SUCCESS", DetailsURL: "https://github.com/acme/widgets/actions/runs/456/job/2"},
			}},
			Runs: map[int64]*github.WorkflowRun{
				123: {DatabaseID: 123, Name: "CI", Jobs: []github.Job{{Name: "gavel"}}},
				456: {DatabaseID: 456, Name: "CodeQL", Jobs: []github.Job{{Name: "analyze"}}},
			},
			GavelResults: []*GavelResultsSummary{{RunID: 123, ArtifactID: 789}},
		}

		newResultFilters(nil, []string{"CodeQL"}).apply(result)

		Expect(result.GavelResults).To(BeEmpty())
	})

	It("computes the same bounded summary consumed by the PR UI", func() {
		failures := []string{
			`{"name":"F1","failed":true}`,
			`{"name":"F2","failed":true}`,
			`{"name":"F3","failed":true}`,
			`{"name":"F4","failed":true}`,
			`{"name":"F5","failed":true}`,
			`{"name":"F6","failed":true}`,
		}
		summary := ComputeGavelSummary(
			[]byte("["+strings.Join(failures, ",")+"]"),
			github.GavelArtifact{StickyID: "gavel", RunID: 123, ArtifactID: 456, CommentID: 10},
		)

		Expect(summary.TestsFailed).To(Equal(6))
		Expect(summary.Failures).To(HaveLen(5))
		Expect(summary.Failures[0].Name).To(Equal("F1"))
		Expect(summary.Failures[4].Name).To(Equal("F5"))
		Expect(summary.RunID).To(Equal(int64(123)))
		Expect(summary.CommentID).To(Equal(int64(10)))
	})

	It("surfaces download errors and omits artifacts without result JSON", func() {
		artifacts := []github.GavelArtifact{
			{StickyID: "gavel-test", RunID: 100, ArtifactID: 1, CommentID: 10},
			{StickyID: "gavel-empty", RunID: 100, ArtifactID: 2, CommentID: 20},
			{StickyID: "gavel-lint", RunID: 100, ArtifactID: 3, CommentID: 30},
		}
		results := fetchGavelArtifacts(github.Options{Repo: "acme/widgets"}, artifacts,
			func(_ github.Options, artifactID int64) ([]byte, error) {
				switch artifactID {
				case 1:
					return []byte(`[{"name":"passes","passed":true}]`), nil
				case 2:
					return nil, github.ErrArtifactResultsNotFound
				default:
					return nil, errors.New("artifact unavailable")
				}
			})

		Expect(results).To(HaveLen(2))
		Expect(results[0].StickyID).To(Equal("gavel-test"))
		Expect(results[0].TestsPassed).To(Equal(1))
		Expect(results[1].StickyID).To(Equal("gavel-lint"))
		Expect(results[1].Error).To(Equal("artifact unavailable"))
		Expect(results[1].CommentID).To(Equal(int64(30)))
	})

	It("exports artifact summaries without internal filtering metadata", func() {
		payload, err := json.Marshal(PRWatchResult{
			PR: &github.PRInfo{Number: 57},
			GavelResults: []*GavelResultsSummary{{
				StickyID: "gavel", RunID: 123, CommentID: 10, ArtifactID: 456,
			}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(string(payload)).To(ContainSubstring(`"gavelResults"`))
		Expect(string(payload)).To(ContainSubstring(`"artifactId":456`))
		Expect(string(payload)).NotTo(ContainSubstring("runId"))
		Expect(string(payload)).NotTo(ContainSubstring("commentId"))
	})
})
