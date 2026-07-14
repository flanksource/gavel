package todos

import (
	"strings"
	"testing"
	"time"

	"github.com/flanksource/gavel/todos/types"
)

func TestCheckTODORunsPersistedVerificationFixture(t *testing.T) {
	workDir := t.TempDir()
	todo, err := ParseTODOContent("fixture check", `## Verification

### command: verification smoke

`+"```bash"+`
echo verification-ok
`+"```"+`

- contains: verification-ok
`, workDir, types.TODOFrontmatter{})
	if err != nil {
		t.Fatalf("parse TODO: %v", err)
	}
	todo.ID = "todo-check-pass"
	todo.FilePath = "todo-check-pass"

	result := CheckTODO(t.Context(), todo, CheckOptions{WorkDir: workDir, Timeout: time.Minute})
	if result.Error != nil {
		t.Fatalf("CheckTODO: %v", result.Error)
	}
	if !result.AllPassed {
		t.Fatalf("expected definition of done to pass: %+v", result)
	}
	if todo.Status != types.StatusVerified || todo.RunMode != types.ModeVerify {
		t.Fatalf("status=%q runMode=%q, want verified/internal verify", todo.Status, todo.RunMode)
	}
	if result.Output == nil || len(result.Output.Results) != 1 {
		t.Fatalf("missing fixture evidence: %+v", result.Output)
	}
	if !strings.Contains(result.Output.Results[0].Stdout, "verification-ok") {
		t.Fatalf("unexpected fixture output: %+v", result.Output.Results[0])
	}
}

func TestCheckTODOFailsWithoutDefinitionOfDone(t *testing.T) {
	todo := &types.TODO{FilePath: "todo-no-dod", TODOFrontmatter: types.TODOFrontmatter{Status: types.StatusPending}}

	result := CheckTODO(t.Context(), todo, CheckOptions{WorkDir: t.TempDir(), Timeout: time.Minute})
	if result.AllPassed || result.Error == nil {
		t.Fatalf("expected missing definition of done to fail: %+v", result)
	}
	if !strings.Contains(result.ErrorText, "no verification fixture") {
		t.Fatalf("unexpected error: %q", result.ErrorText)
	}
	if todo.Status != types.StatusUnverified {
		t.Fatalf("status=%q, want unverified", todo.Status)
	}
}
