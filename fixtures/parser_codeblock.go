package fixtures

import (
	"strings"
)

// extractLanguage extracts the language identifier from a code fence info string
// Example: "bash exitCode=1" returns "bash"
func extractLanguage(infoString string) string {
	infoString = strings.TrimSpace(infoString)
	if infoString == "" {
		return ""
	}

	// Language is the first token
	parts := strings.Fields(infoString)
	if len(parts) > 0 {
		return parts[0]
	}

	return ""
}

// runnerStepKind reports whether a fence info string marks a test/lint runner
// step, returning "test", "lint", or "". The marker is the last whitespace
// token; an optional leading "yaml" token (present only for editor syntax
// highlighting) is ignored. Accepts: "test", "lint", "yaml test", "yaml lint"
// (case-insensitive).
func runnerStepKind(infoString string) string {
	parts := strings.Fields(strings.ToLower(strings.TrimSpace(infoString)))
	if len(parts) == 0 {
		return ""
	}
	marker := parts[len(parts)-1]
	if marker != RunnerKindTest && marker != RunnerKindLint {
		return ""
	}
	// Reject anything between an optional leading "yaml" and the marker so a
	// genuine "yaml" config block or an unrelated fence never trips the marker.
	switch len(parts) {
	case 1:
		return marker
	case 2:
		if parts[0] == "yaml" {
			return marker
		}
	}
	return ""
}

// shouldExecuteCodeBlock determines if a code block with the given language
// should be executed based on the codeBlocks configuration
func shouldExecuteCodeBlock(language string, codeBlocks []string) bool {
	if language == "" {
		return false
	}

	// Case-insensitive comparison
	langLower := strings.ToLower(language)
	for _, allowed := range codeBlocks {
		if strings.ToLower(allowed) == langLower {
			return true
		}
	}

	return false
}

// getCodeBlocksOrDefault returns the codeBlocks list or defaults to ["bash"]
func getCodeBlocksOrDefault(frontMatter *FrontMatter) []string {
	if frontMatter == nil || len(frontMatter.CodeBlocks) == 0 {
		return []string{"bash"}
	}
	return frontMatter.CodeBlocks
}

// isExecutableLanguage returns true if the language is an executable language
// that can be used for standalone code blocks (without "command:" prefix).
// Supported languages: exec, bash, shell, sh, python, py, typescript, ts, go
func isExecutableLanguage(language string) bool {
	if language == "" {
		return false
	}

	langLower := strings.ToLower(language)
	switch langLower {
	case "exec", "bash", "shell", "sh":
		return true
	case "python", "py":
		return true
	case "typescript", "ts":
		return true
	case "go":
		return true
	default:
		return false
	}
}
