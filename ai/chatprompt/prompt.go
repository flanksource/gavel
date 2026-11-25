package chatprompt

import (
	_ "embed"

	"github.com/flanksource/gavel/prompts"
)

//go:embed chat-assistant.prompt
var Default string

func Prompts() []prompts.Prompt {
	return []prompts.Prompt{{ID: prompts.ChatAssistant, Title: "Dashboard assistant", Description: "System instructions for the workflow assistant.", ConfigPath: prompts.ChatAssistant, Default: Default, UsedBy: []string{"Dashboard chat"}}}
}
