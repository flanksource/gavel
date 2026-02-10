package claude

import (
	"fmt"

	"github.com/flanksource/gavel/todos/types"
)

// BuildImplementationPrompt creates a prompt for Claude Code to implement a TODO
func BuildImplementationPrompt(todo *types.TODO, claudeMD string) string {
	prompt := fmt.Sprintf(`# TODO Implementation Request

File: %s

## Implementation Instructions

%s

`, todo.FilePath, todo.Implementation)

	// Add CLAUDE.md best practices
	if claudeMD != "" {
		prompt += fmt.Sprintf(`## Best Practices (from CLAUDE.md)

Please follow these guidelines while implementing:

%s

`, claudeMD)
	}

	// Add specific instructions
	prompt += `## Required Actions

1. Implement the changes described above
2. Follow TDD: write tests first, then implementation
3. Run tests to verify the implementation works
4. When complete, output the exact phrase: IMPLEMENTATION_COMPLETE

IMPORTANT: Do NOT commit any changes. Only implement and test.
`

	return prompt
}

// BuildRetryPrompt creates a retry prompt with failure details
func BuildRetryPrompt(todo *types.TODO, failureDetails string) string {
	return fmt.Sprintf(`# TODO Implementation - Retry Required

File: %s

## Previous Implementation Failed

The previous implementation failed verification. Details:

%s%s%s

## Implementation Instructions

%s

## Required Actions

1. Review the failure details above
2. Fix the implementation to address the failures
3. Run tests to verify the fix works
4. When complete, output the exact phrase: IMPLEMENTATION_COMPLETE

IMPORTANT: Do NOT commit any changes. Only implement and test.
`, todo.FilePath, "```\n", failureDetails, "\n```", todo.Implementation)
}

// BuildVerificationPrompt creates a prompt to verify implementation
func BuildVerificationPrompt(todo *types.TODO) string {
	return fmt.Sprintf(`# Verify Implementation

File: %s

Please verify the implementation by:

1. Running the verification tests
2. Checking that all tests pass
3. Confirming the implementation meets requirements
`, todo.FilePath)
}
