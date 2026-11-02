package ui

import (
	"context"

	"github.com/flanksource/gavel/todos"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("hosted PR TODO lookup", func() {
	It("rejects a hosted dashboard without workspaces before opening a local provider", func() {
		previous := openGlobalTodoProvider
		openGlobalTodoProvider = func(context.Context) (todos.GlobalReferenceProvider, error) {
			Fail("a hosted request must not open the local TODO provider")
			return nil, errNoHostedWorkspaces
		}
		DeferCleanup(func() { openGlobalTodoProvider = previous })
		_, err := (&Server{hosted: true}).prTodosLinker(context.Background(), "")
		Expect(err).To(MatchError(errNoHostedWorkspaces))
	})
})
