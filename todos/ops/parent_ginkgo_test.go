package ops

import (
	"context"

	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// recordingProvider stores nothing: it records the order of the writes an edit
// makes. Methods an edit must not reach are left to the nil embedded interface.
type recordingProvider struct {
	todos.Provider
	calls []string
}

func (p *recordingProvider) Edit(context.Context, *types.TODO, todos.EditRequest) error {
	p.calls = append(p.calls, "edit")
	return nil
}

func (p *recordingProvider) UpdateState(context.Context, *types.TODO, todos.StateUpdate) error {
	p.calls = append(p.calls, "state")
	return nil
}

type parentingProvider struct{ *recordingProvider }

func (p parentingProvider) SetParent(_ context.Context, todo *types.TODO, parentRef string) error {
	p.calls = append(p.calls, "parent:"+parentRef)
	todo.ParentID = parentRef
	return nil
}

var _ = Describe("editing a TODO's parent", func() {
	const parentRef = "0a1b2c3d"

	DescribeTable("BuildEdit carries the parent change",
		func(flags EditFlags, want *string) {
			changes, err := BuildEdit(flags)
			Expect(err).NotTo(HaveOccurred())
			Expect(changes.Parent).To(Equal(want))
		},
		Entry("--parent names the new parent", EditFlags{Parent: " " + parentRef + " "}, ptr(parentRef)),
		Entry("--no-parent detaches with an empty ref", EditFlags{NoParent: true}, ptr("")),
		Entry("neither leaves the parent alone", EditFlags{Status: "draft"}, nil),
	)

	It("BuildEdit refuses --parent together with --no-parent", func() {
		_, err := BuildEdit(EditFlags{Parent: parentRef, NoParent: true})
		Expect(err).To(MatchError(ContainSubstring("--no-parent cannot be combined with --parent")))
	})

	It("BuildEdit names the parent flags when there is nothing to edit", func() {
		_, err := BuildEdit(EditFlags{})
		Expect(err).To(MatchError(SatisfyAll(ContainSubstring("--parent"), ContainSubstring("--no-parent"))))
	})

	It("ApplyEdit sets the parent before the other writes", func(ctx SpecContext) {
		provider := parentingProvider{&recordingProvider{}}
		todo := &types.TODO{ID: "todo-under-edit"}
		changes, err := BuildEdit(EditFlags{Parent: parentRef, Title: ptr("Renamed"), Status: "draft"})
		Expect(err).NotTo(HaveOccurred())

		edited, err := ApplyEdit(ctx, provider, todo, changes)

		Expect(err).NotTo(HaveOccurred())
		Expect(edited.ParentID).To(Equal(parentRef))
		Expect(provider.calls).To(Equal([]string{"parent:" + parentRef, "edit", "state"}))
	})

	It("ApplyEdit detaches with an empty parent ref", func(ctx SpecContext) {
		provider := parentingProvider{&recordingProvider{}}
		changes, err := BuildEdit(EditFlags{NoParent: true})
		Expect(err).NotTo(HaveOccurred())

		_, err = ApplyEdit(ctx, provider, &types.TODO{ParentID: parentRef}, changes)

		Expect(err).NotTo(HaveOccurred())
		Expect(provider.calls).To(Equal([]string{"parent:"}))
	})

	It("ApplyEdit writes nothing when the provider cannot store parents", func(ctx SpecContext) {
		provider := &recordingProvider{}
		changes, err := BuildEdit(EditFlags{Parent: parentRef, Title: ptr("Renamed")})
		Expect(err).NotTo(HaveOccurred())

		_, err = ApplyEdit(ctx, provider, &types.TODO{}, changes)

		Expect(err).To(MatchError(ContainSubstring("does not support parents")))
		Expect(provider.calls).To(BeEmpty())
	})
})

var _ = Describe("OriginFromEnv", func() {
	const (
		issueID   = "5f0e9d8c-7b6a-4f5e-8d4c-3b2a1f0e9d8c"
		sessionID = "session-91ab"
	)

	DescribeTable("reads the run a command was started inside",
		func(issue, session string, want *todos.CreateOrigin) {
			GinkgoT().Setenv(commit.EnvIssueID, issue)
			GinkgoT().Setenv(commit.EnvSessionID, session)
			Expect(OriginFromEnv()).To(Equal(want))
		},
		Entry("outside a run there is no origin", "", "", nil),
		Entry("a run exports its TODO and session", issueID, sessionID,
			&todos.CreateOrigin{IssueID: issueID, SessionID: sessionID}),
		Entry("whitespace is not an id", "  ", " \t", nil),
		Entry("a session alone is still an origin", "", sessionID, &todos.CreateOrigin{SessionID: sessionID}),
		Entry("values are trimmed", " "+issueID+" ", "", &todos.CreateOrigin{IssueID: issueID}),
	)
})
