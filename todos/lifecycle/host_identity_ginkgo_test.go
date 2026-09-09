package lifecycle_test

import (
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// goldenIssue and goldenVersion pin the inputs the literal UUIDs below were
// derived from.
//
// To regenerate a value, evaluate the formula by hand:
//
//	seed      = "<issue>:<step>:<version>"
//	session   = uuid.NewSHA1(uuid.NameSpaceOID, "gavel-todo-session:"+seed)
//	promptRun = uuid.NewSHA1(uuid.NameSpaceOID, "gavel-todo-prompt-run:"+seed)
var goldenIssue = uuid.MustParse("11111111-2222-3333-4444-555555555555")

const goldenVersion = int64(7)

// The built-in steps must hash exactly as they did before the lifecycle named
// them: any change to the seed silently re-keys every in-flight run, and a
// resumed dispatch would no longer recognise its own prompt run.
var _ = Describe("dispatch identity", func() {
	DescribeTable("built-in steps keep their pre-lifecycle identities",
		func(step, wantSession, wantPromptRun string) {
			identity := lifecycle.IdentityFor(lifecycle.Seed(goldenIssue, step, goldenVersion))
			Expect(identity.SessionID.String()).To(Equal(wantSession))
			Expect(identity.PromptRunID.String()).To(Equal(wantPromptRun))
			Expect(identity.AdmissionKey).To(Equal("gavel-todo:" + goldenIssue.String() + ":" + step + ":7"))
		},
		Entry("run", "run", "676cd2c8-f81e-50dc-855f-6fe23e6ab8ea", "ffda6967-d576-5e0e-ab65-7b7e56ee8d66"),
		Entry("plan", "plan", "4fe86321-f9f8-537a-90df-53e83683f9d5", "9685e778-a709-556f-ac6d-351c1cb11941"),
		Entry("verify", "verify", "3a92756b-29ec-5009-be4c-33577760fc47", "a7894403-c2e1-5464-9883-79798192fe55"),
		// Triage is its own step now, seeded by its own name rather than as a
		// plan-class prompt: the literal is pinned so the next change is deliberate.
		Entry("triage", "triage", "3012ba60-e9bf-5546-8eba-963a7fc1f5ab", "a92f6997-2f91-538a-a2e2-150f7fa8d25f"),
	)

	It("gives two steps of one issue version distinct prompt runs", func() {
		plan := lifecycle.IdentityFor(lifecycle.Seed(goldenIssue, "plan", goldenVersion))
		triage := lifecycle.IdentityFor(lifecycle.Seed(goldenIssue, "triage", goldenVersion))
		custom := lifecycle.IdentityFor(lifecycle.Seed(goldenIssue, "security", goldenVersion))
		Expect(plan.PromptRunID).NotTo(Equal(triage.PromptRunID))
		Expect(triage.PromptRunID).NotTo(Equal(custom.PromptRunID))
	})

	It("derives a todo's identity from its native id, the step name and its version", func() {
		host := &lifecycle.Host{}
		todo := &types.TODO{ID: goldenIssue.String(), Version: goldenVersion}
		identity, err := host.Identity(todo, lifecycle.Step{Name: "run"})
		Expect(err).NotTo(HaveOccurred())
		Expect(identity.PromptRunID.String()).To(Equal("ffda6967-d576-5e0e-ab65-7b7e56ee8d66"))
	})

	It("refuses a todo without a native id", func() {
		host := &lifecycle.Host{}
		_, err := host.Identity(&types.TODO{ID: "not-a-uuid"}, lifecycle.Step{Name: "run"})
		Expect(err).To(MatchError(ContainSubstring("has no native id")))
	})
})
