package todos

import "strings"

// ComposeIssueMarkdown joins a TODO body with its verification fixture under a
// `# Verification` heading. It is the portable `.todos` document body without
// the YAML frontmatter, shared by the portable exporter and by external issue
// trackers that render markdown but have nowhere to put frontmatter.
func ComposeIssueMarkdown(body, verification string) string {
	verification = strings.TrimSpace(verification)
	if verification == "" {
		return body
	}
	body = strings.TrimSpace(body)
	if body != "" {
		body += "\n\n"
	}
	return body + "# Verification\n\n" + verification
}
