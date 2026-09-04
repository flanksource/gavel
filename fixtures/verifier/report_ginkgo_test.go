package verifier_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/flanksource/captain/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/verifier"
)

const scopedDocument = `---
codeBlocks: [bash]
---

# Definition of done

### command: only the changed package matters

` + "```bash\necho scoped\n```" + `

- cel: 'pkg/a.go' in changed_files
`

var _ = Describe("fixture verifier changed-file scope", func() {
	// captain hands the verifier the files the change touched when the workflow
	// is scoped to them. An exec step cannot be narrowed to a file, so it runs
	// as written — but the set reaches its expectation as `changed_files`, and
	// the node records the scope it was judged under.
	It("binds the changed set as the CEL root changed_files and records it on the node", func() {
		v, err := verifier.NewVerifier(scopedDocument)
		Expect(err).ToNot(HaveOccurred())

		verdict, err := v.Verify(context.Background(), GinkgoT().TempDir(), []string{"pkg/a.go"})

		Expect(err).ToNot(HaveOccurred())
		Expect(verdict.OK).To(BeTrue(), verdict.Feedback)
		Expect(verdict.Report.Validate()).To(Succeed())
		var detail map[string]any
		Expect(json.Unmarshal(verdict.Report.Tests[0].Detail, &detail)).To(Succeed())
		Expect(detail).To(HaveKeyWithValue("changed_files", []any{"pkg/a.go"}))
	})

	It("leaves changed_files undefined on an unscoped run rather than empty", func() {
		v, err := verifier.NewVerifier(scopedDocument)
		Expect(err).ToNot(HaveOccurred())

		verdict, err := v.Verify(context.Background(), GinkgoT().TempDir(), nil)

		Expect(err).ToNot(HaveOccurred())
		Expect(verdict.OK).To(BeFalse(), "an expectation over a scope that was never set must not pass")
		Expect(verdict.Report.Tests[0].Detail).To(BeNil())
	})
})

var _ = Describe("verifier.RunningReport", func() {
	leaf := func(name string, state fixtures.ExecutionState) *fixtures.ExecutionNode {
		return &fixtures.ExecutionNode{Key: name, Name: name, Kind: fixtures.ExecutionKindCommand, State: state}
	}
	snapshotOf := func(leaves ...*fixtures.ExecutionNode) fixtures.ExecutionSnapshot {
		return fixtures.ExecutionSnapshot{Root: &fixtures.ExecutionNode{
			Key: "fixtures", Name: "Fixtures", Kind: fixtures.ExecutionKindRoot, Children: leaves,
		}}
	}

	It("reports a tree with work in flight as running and valid", func() {
		report, live := verifier.RunningReport(snapshotOf(
			leaf("done", fixtures.ExecutionPassed), leaf("busy", fixtures.ExecutionRunning)))

		Expect(live).To(BeTrue())
		Expect(report.State).To(Equal(api.VerifyStateRunning))
		Expect(report.Passed).To(BeFalse())
		Expect(report.Validate()).To(Succeed())
	})

	It("reports a tree nothing has started as queued", func() {
		report, live := verifier.RunningReport(snapshotOf(leaf("later", fixtures.ExecutionQueued)))

		Expect(live).To(BeTrue())
		Expect(report.State).To(Equal(api.VerifyStateQueued))
		Expect(report.Ran).To(BeFalse())
		Expect(report.Validate()).To(Succeed())
	})

	// A finished tree adds up to a verdict. Publishing it as a running report
	// would say `passed` with passed=false — the one contradiction a reader
	// cannot resolve — so it is not a running report at all.
	It("declines a tree with nothing left to run", func() {
		_, live := verifier.RunningReport(snapshotOf(
			leaf("done", fixtures.ExecutionPassed), leaf("also done", fixtures.ExecutionPassed)))

		Expect(live).To(BeFalse())
	})
})

var _ = Describe("fixture verifier progress stream", func() {
	It("never publishes a verdict state as progress", func() {
		var mu sync.Mutex
		var states []api.VerifyState
		v, err := verifier.NewVerifier(greenDocument)
		Expect(err).ToNot(HaveOccurred())
		v.SetProgress(func(report api.VerifyReport) {
			mu.Lock()
			defer mu.Unlock()
			states = append(states, report.State)
		})

		verdict, err := v.Verify(context.Background(), GinkgoT().TempDir(), nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(verdict.OK).To(BeTrue())

		mu.Lock()
		defer mu.Unlock()
		Expect(states).ToNot(BeEmpty())
		for _, state := range states {
			Expect(state).To(BeElementOf(api.VerifyStateQueued, api.VerifyStateRunning),
				"a progress report must not carry the verdict's state")
		}
	})
})

var _ = Describe("verifier.Feedback", func() {
	failing := func(n int) []api.VerifyNode {
		nodes := make([]api.VerifyNode, 0, n)
		for i := 0; i < n; i++ {
			nodes = append(nodes, api.VerifyNode{Name: fmt.Sprintf("check-%02d", i), Failed: true})
		}
		return nodes
	}
	unmet := func(n int) []api.VerifyChecklistItem {
		no := false
		items := make([]api.VerifyChecklistItem, 0, n)
		for i := 0; i < n; i++ {
			items = append(items, api.VerifyChecklistItem{Item: fmt.Sprintf("criterion-%02d", i), Passed: &no})
		}
		return items
	}

	// The digest is bounded for the agent's context, so the bound has to cover
	// everything that goes into it: fifty failures and three unmet criteria are
	// fifty lines and a count of three, not fifty-three lines.
	It("caps failing nodes and unmet criteria under one bound, noting the rest last", func() {
		lines := strings.Split(verifier.Feedback(failing(48), unmet(5)), "\n")

		Expect(lines).To(HaveLen(51))
		Expect(lines[47]).To(Equal("- check-47"))
		Expect(lines[48]).To(Equal("- criterion not met: criterion-00"))
		Expect(lines[49]).To(Equal("- criterion not met: criterion-01"))
		Expect(lines[50]).To(Equal("… (3 more failures truncated)"))
	})

	It("lists every unmet criterion after the failures when nothing is cut", func() {
		lines := strings.Split(verifier.Feedback(failing(2), unmet(2)), "\n")

		Expect(lines).To(Equal([]string{
			"- check-00", "- check-01",
			"- criterion not met: criterion-00", "- criterion not met: criterion-01",
		}))
	})
})
