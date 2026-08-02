package headless

import (
	captainai "github.com/flanksource/captain/pkg/ai"
	commitpkg "github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/todos/types"
)

// commitHooks builds the run's commit hooks from its spec. The adapter onto
// gavel's commit pipeline lives in commit.AgentHooks, shared with the other
// agent loops (`pr status --ai-fix`); here it only supplies the todo's identity.
//
// A spec with no commit policies returns no hooks and makes no commit. A todo run
// never pushes — the branch is pushed when the PR is opened, not per turn.
func commitHooks(req captainai.Request, todosInGroup []*types.TODO, sessionID string) []any {
	if req.Workflow == nil {
		return nil
	}
	return commitpkg.AgentHooks(commitpkg.AgentHooksOptions{
		Commits: req.Workflow.Commits,
		Meta:    commitpkg.AgentRunMetadata{IssueID: issueIDOf(todosInGroup), SessionID: sessionID},
	})
}

// lastCommitSHA returns the final commit a run's hooks recorded, which is the
// squashed one when the chain collapsed. Empty when the run declared no commit
// policy or no stageable changes were produced.
func lastCommitSHA(resp *captainai.Response) string {
	if resp == nil || resp.Workspace == nil || len(resp.Workspace.Commits) == 0 {
		return ""
	}
	return resp.Workspace.Commits[len(resp.Workspace.Commits)-1].SHA
}

func issueIDOf(todosInGroup []*types.TODO) string {
	for _, todo := range todosInGroup {
		if todo != nil {
			return todo.ID
		}
	}
	return ""
}
