package fixtures

import (
	"strings"
)

// parseCodeFenceAttributes extracts attribute key=value pairs from a code fence info string
// Example: "bash exitCode=1 timeout=30" returns map["exitCode":"1", "timeout":"30"]
func parseCodeFenceAttributes(infoString string) map[string]string {
	attrs := make(map[string]string)

	// Split by whitespace
	parts := strings.Fields(infoString)

	// Skip first part (language)
	for i := 1; i < len(parts); i++ {
		part := parts[i]
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				attrs[kv[0]] = kv[1]
			}
		}
	}

	return attrs
}

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
// Supported languages: bash, shell, sh, python, py, typescript, ts, go
func isExecutableLanguage(language string) bool {
	if language == "" {
		return false
	}

	langLower := strings.ToLower(language)
	switch langLower {
	case "bash", "shell", "sh":
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
