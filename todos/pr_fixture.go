package todos

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
)

func PRVerificationFixture(pr *types.PR, fixture string) (string, error) {
	if pr == nil {
		return fixture, nil
	}
	if err := pr.Validate(); err != nil {
		return "", err
	}
	fixture = strings.TrimSpace(fixture)
	front, body, err := fixtures.SplitFrontMatter(fixture)
	if err != nil {
		return "", err
	}
	header := ""
	if front != nil {
		for i, line := range strings.SplitAfter(fixture, "\n") {
			header += line
			if i > 0 && strings.TrimSpace(line) == "---" {
				break
			}
		}
	}
	command := fmt.Sprintf("gavel pr status '%s' --whole-pr --head-sha HEAD --follow --fail-fast --logs --tail-logs 40 --json", strings.ReplaceAll(pr.URL, "'", "'\"'\"'"))
	block := "### Linked PR\n\n| Name | Command | CWD | Exit Code | OS | Arch | Skip |\n|------|---------|-----|-----------|----|------|------|\n| Entire PR at the verified revision | " + command + " | {{.GIT_ROOT_DIR}} | 0 | " + runtime.GOOS + " | " + runtime.GOARCH + " | false |"
	return header + "\n" + ExtractVerificationFixture(UpsertMarkedVerificationBlock("## Verification\n\n"+body, "gavel:pr-status-verification", block)), nil
}
