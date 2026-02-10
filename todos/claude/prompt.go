package claude

import (
	"fmt"

	"github.com/flanksource/gavel/todos/types"
)

// BuildPrompt constructs a structured prompt from a TODO for Claude Code execution.
// The prompt includes test information, reproduction steps, implementation notes, and verification tests.
func BuildPrompt(todo *types.TODO) string {
	prompt := `You are fixing a failing test in a Go codebase.

`

	// Test information
	if todo.FileNode != nil {
		prompt += `## Test Information

`
		prompt += fmt.Sprintf("- **File:** %s\n", todo.FilePath)
		if todo.FileNode.Test != nil {
			prompt += fmt.Sprintf("- **Name:** %s\n", todo.FileNode.Test.Name)
		}
		prompt += "\n"
	}

	// Steps to reproduce the failure
	if len(todo.StepsToReproduce) > 0 {
		prompt += `## Steps to Reproduce

Run the following to reproduce the failure:

`
		for _, node := range todo.StepsToReproduce {
			if node.Test != nil {
				prompt += fmt.Sprintf("```bash\n%s\n```\n\n", node.Test.String())
			}
		}
	}

	// Implementation instructions
	if todo.Implementation != "" {
		prompt += fmt.Sprintf(`## Implementation

%s

`, todo.Implementation)
	}

	// Verification tests
	if len(todo.Verification) > 0 {
		prompt += `## Verification

After implementing your fix, verify it works by running:

`
		for _, node := range todo.Verification {
			if node.Test != nil {
				prompt += fmt.Sprintf("```bash\n%s\n```\n\n", node.Test.String())
			}
		}
	}

	// Instructions
	prompt += `## Instructions

1. Analyze the test failure and reproduction steps
2. Investigate the codebase to understand the root cause
3. Implement a fix that addresses the underlying issue
4. Run verification tests to confirm the fix works
5. Use ` + "`git add`" + ` to stage your changes

Your fix should:
- Address the root cause, not mask symptoms
- Follow existing code patterns and style
- Pass all verification tests
- Be minimal and focused
`

	return prompt
}
