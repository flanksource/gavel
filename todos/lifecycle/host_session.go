package lifecycle

import "github.com/flanksource/gavel/todos/types"

func setSessionID(todo *types.TODO, sessionID string) {
	if todo == nil || sessionID == "" {
		return
	}
	if todo.LLM == nil {
		todo.LLM = &types.LLM{}
	}
	todo.LLM.SessionId = sessionID
}

// priorSessionID is the todo's recorded agent session, the id a resume reuses.
func priorSessionID(todo *types.TODO) string {
	if todo != nil && todo.LLM != nil {
		return todo.LLM.SessionId
	}
	return ""
}
