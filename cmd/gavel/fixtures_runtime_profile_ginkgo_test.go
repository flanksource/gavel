package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const fixtureProfileDocument = "---\nai: {}\n---\n\n# Review\n\n- [ ] The implementation has evidence.\n"

var _ = Describe("standalone fixture runtime profiles", func() {
	var cwd string
	var captured []api.Spec
	var capturedMu sync.Mutex
	var dispatched int

	BeforeEach(func() {
		cwd = GinkgoT().TempDir()
		GinkgoT().Setenv(database.EnvDisable, "off")
		GinkgoT().Setenv("XDG_CONFIG_HOME", GinkgoT().TempDir())
		Expect(database.ResetDisabledSharedForTest()).To(Succeed())
		DeferCleanup(database.ResetDisabledSharedForTest)
		captured = nil
		dispatched = 0
		original := fixtures.AIStepRunner
		fixtures.AIStepRunner = func(test fixtures.FixtureTest, options fixtures.RunOptions) fixtures.FixtureResult {
			capturedMu.Lock()
			defer capturedMu.Unlock()
			dispatched++
			if options.Spec != nil {
				captured = append(captured, *options.Spec)
			}
			return fixtures.FixtureResult{Name: test.Name, Status: task.StatusPASS, Test: test}
		}
		DeferCleanup(func() { fixtures.AIStepRunner = original })
	})

	DescribeTable("resolves the verification profile before the AI fixture dispatch",
		func(input, selection string, expectedCost float64) {
			profiles := filepath.Join(cwd, ".captain", "profiles")
			Expect(os.MkdirAll(profiles, 0o755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(profiles, "global.yaml"), []byte("name: global\nspec:\n  budget:\n    cost: 2\n"), 0o600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(profiles, "review.yaml"), []byte("name: review\nspec:\n  budget:\n    cost: 3\n    maxTurns: 7\n"), 0o600)).To(Succeed())
			config := "ai:\n  budget:\n    maxTokens: 100\ntodos:\n  runtimeProfile: global\n  verify:\n    budget:\n      maxTurns: 9\n" + selection
			Expect(os.WriteFile(filepath.Join(cwd, ".gavel.yaml"), []byte(config), 0o600)).To(Succeed())

			Expect(runProfileFixture(input, cwd, fixtureProfileDocument)).To(Succeed())
			Expect(captured).To(HaveLen(1))
			Expect(captured[0].Budget).To(Equal(api.Budget{Cost: expectedCost, MaxTokens: 100, MaxTurns: 9}))
		},
		Entry("file runner uses the global profile", "file", "", 2.0),
		Entry("report runner uses the global profile", "report", "", 2.0),
		Entry("file runner prefers the verify profile", "file", "    runtimeProfile: review\n", 3.0),
		Entry("report runner prefers the verify profile", "report", "    runtimeProfile: review\n", 3.0),
	)

	DescribeTable("does not resolve an unused profile for a command-only document", func(input string) {
		Expect(os.WriteFile(filepath.Join(cwd, ".gavel.yaml"), []byte("todos:\n  runtimeProfile: missing\n"), 0o600)).To(Succeed())
		Expect(runProfileFixture(input, cwd, verifyReportDocument)).To(Succeed())
		Expect(captured).To(BeEmpty())
	}, Entry("file runner", "file"), Entry("report runner", "report"))

	It("reports the missing profile as failed verification without dispatching AI", func() {
		Expect(os.WriteFile(filepath.Join(cwd, ".gavel.yaml"), []byte("todos:\n  verify:\n    runtimeProfile: missing\n"), 0o600)).To(Succeed())
		lines, err := runVerifyReportCommand(fixtureProfileDocument, cwd)
		Expect(err).NotTo(HaveOccurred())
		reports := reportLines(lines)
		Expect(reports).To(HaveLen(1))
		Expect(reports[0].State).To(Equal(api.VerifyStateFailed))
		Expect(reports[0].Passed).To(BeFalse())
		Expect(reports[0].Feedback).To(ContainSubstring(`runtime profile "missing" selected by default`))
		Expect(captured).To(BeEmpty())
	})

	It("dispatches a serialized snapshot without looking up the current default profile", func() {
		Expect(os.WriteFile(filepath.Join(cwd, ".gavel.yaml"), []byte("todos:\n  verify:\n    runtimeProfile: missing\n"), 0o600)).To(Succeed())
		document := "---\nai:\n  spec:\n    budget:\n      maxTurns: 12\n---\n\n# Review\n\n- [ ] The implementation has evidence.\n"
		lines, err := runVerifyReportCommand(document, cwd)
		Expect(err).NotTo(HaveOccurred())
		reports := reportLines(lines)
		Expect(reports).To(HaveLen(1))
		Expect(reports[0].State).To(Equal(api.VerifyStatePassed))
		Expect(dispatched).To(Equal(1))
		Expect(captured).To(BeEmpty())
	})
})

func runProfileFixture(input, cwd, document string) error {
	GinkgoHelper()
	if input == "report" {
		_, err := runVerifyReportCommand(document, cwd)
		return err
	}
	path := filepath.Join(cwd, "review.fixture.md")
	Expect(os.WriteFile(path, []byte(document), 0o600)).To(Succeed())
	previous := workingDir
	workingDir = cwd
	DeferCleanup(func() { workingDir = previous })
	fixturesCmd.SetContext(context.Background())
	return runFixtures(fixturesCmd, []string{path})
}
