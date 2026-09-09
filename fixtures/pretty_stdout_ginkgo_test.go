package fixtures_test

import (
	"strings"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fixture JSON stdout", func() {
	It("renders a transported Clicky document instead of the raw JSON envelope", func() {
		result := fixtures.FixtureResult{
			Status: task.StatusERR,
			Test:   fixtures.FixtureTest{Name: "TBILLEQ"},
			Stdout: `{
  "result": 0.051336,
  "trace": {"durationMs": 0.03},
  "tracePretty": {
    "version": 1,
    "node": {
      "kind": "table",
      "plain": "expression  value\nTBILLEQ()  0.051336"
    }
  }
}`,
			Display: &fixtures.DisplayOptions{
				ShowPassed: true,
				ShowStdout: fixtures.OutputAlways,
				ShowStderr: fixtures.OutputOnFailure,
			},
		}

		pretty := result.Pretty()
		for _, output := range []string{pretty.String(), pretty.Markdown()} {
			Expect(output).To(ContainSubstring("TBILLEQ()"))
			Expect(output).To(ContainSubstring("0.051336"))
			Expect(strings.ToLower(output)).NotTo(ContainSubstring("durationms"))
			Expect(output).NotTo(ContainSubstring("tracePretty"))
		}
	})

	It("preserves ordinary JSON output without a Clicky document", func() {
		result := fixtures.FixtureResult{
			Status: task.StatusERR,
			Test:   fixtures.FixtureTest{Name: "plain JSON"},
			Stdout: `{"result": 0.051336}`,
			Display: &fixtures.DisplayOptions{
				ShowPassed: true,
				ShowStdout: fixtures.OutputAlways,
				ShowStderr: fixtures.OutputOnFailure,
			},
		}

		Expect(result.Pretty().String()).To(ContainSubstring(`{"result": 0.051336}`))
	})
})
