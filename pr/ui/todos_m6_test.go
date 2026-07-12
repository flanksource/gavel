package ui

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/gavel/todos"
)

func TestM6LegacyProviderRejectedBeforeDatabaseOpen(t *testing.T) {
	server := &Server{}
	_, _, err := server.todoProviderContext(context.Background(), todoSource{
		Provider: todos.ProviderGrite,
		Dir:      t.TempDir(),
	})
	if !errors.Is(err, todos.ErrProviderRetired) {
		t.Fatalf("todoProviderContext(grite) error = %v, want ErrProviderRetired", err)
	}
	if !strings.Contains(err.Error(), "import-grite") {
		t.Fatalf("retired Grite error is not actionable: %v", err)
	}
	_, _, _, err = server.resolveTodoReference(context.Background(), todoSource{Provider: todos.ProviderGrite}, "legacy-ref")
	if !errors.Is(err, todos.ErrProviderRetired) {
		t.Fatalf("workspace-free legacy lookup error = %v, want ErrProviderRetired before OpenGlobal", err)
	}
}

func TestM6RetiredProviderUsesMigrationHTTPStatus(t *testing.T) {
	err := todos.ValidateRuntimeProvider(todos.ProviderFiles)
	recorder := httptest.NewRecorder()
	writeTodoError(recorder, http.StatusBadRequest, err)
	if recorder.Code != http.StatusGone {
		t.Fatalf("retired provider status = %d, want %d; body = %s", recorder.Code, http.StatusGone, recorder.Body.String())
	}
}

func TestM6ProjectCannotPinLegacyProvider(t *testing.T) {
	originalPath := projectsPath
	projectsPath = filepath.Join(t.TempDir(), "projects.json")
	t.Cleanup(func() { projectsPath = originalPath })

	err := CreateProject(Project{Name: "legacy", Dir: t.TempDir(), TodoProvider: todos.ProviderFiles})
	if !errors.Is(err, ErrProjectInvalid) || !strings.Contains(err.Error(), "import/export-only") {
		t.Fatalf("CreateProject legacy provider error = %v, want actionable ErrProjectInvalid", err)
	}
}

func TestM6ProjectSummaryDoesNotHideProviderErrorsAsEmptyCounts(t *testing.T) {
	_, err := newProjectInfo(context.Background(), Project{
		Name: "legacy", Dir: t.TempDir(), TodoProvider: todos.ProviderGrite,
	})
	if !errors.Is(err, todos.ErrProviderRetired) {
		t.Fatalf("newProjectInfo legacy provider error = %v, want ErrProviderRetired", err)
	}
}

func TestM6NativeGroupedRunRejectedDuringRequestValidation(t *testing.T) {
	if err := validateTodoRunCardinality(1); err != nil {
		t.Fatalf("single issue rejected: %v", err)
	}
	err := validateTodoRunCardinality(2)
	if err == nil || !strings.Contains(err.Error(), "one issue at a time") {
		t.Fatalf("two-issue validation error = %v, want actionable native cardinality error", err)
	}
}
