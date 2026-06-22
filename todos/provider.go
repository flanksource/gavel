package todos

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/todos/types"
)

const (
	ProviderFiles = "todos"
	ProviderGrite = "grite"
)

// Provider is the persistence boundary for TODO storage.
type Provider interface {
	List(ctx context.Context, filters DiscoveryFilters) (types.TODOS, error)
	Get(ctx context.Context, ref string) (*types.TODO, error)
	UpdateState(ctx context.Context, todo *types.TODO, updates StateUpdate) error
	UpdateLatestFailure(ctx context.Context, todo *types.TODO, result *types.TestResultInfo) error
	SaveAttempt(ctx context.Context, todo *types.TODO, result *ExecutionResult) error
}

type FileProvider struct {
	Dir     string
	WorkDir string
}

func NewFileProvider(workDir, dir string) *FileProvider {
	if dir == "" {
		dir = filepath.Join(workDir, ".todos")
	}
	return &FileProvider{Dir: dir, WorkDir: workDir}
}

func (p *FileProvider) List(_ context.Context, filters DiscoveryFilters) (types.TODOS, error) {
	if _, err := os.Stat(p.Dir); os.IsNotExist(err) {
		return nil, fmt.Errorf(".todos directory not found: %s", p.Dir)
	}
	return DiscoverTODOs(p.Dir, filters)
}

func (p *FileProvider) Get(_ context.Context, ref string) (*types.TODO, error) {
	todoPath := ref
	if !filepath.IsAbs(todoPath) && !strings.Contains(todoPath, string(filepath.Separator)) {
		todoPath = filepath.Join(p.Dir, todoPath)
	}
	if _, err := os.Stat(todoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("TODO file not found: %s", todoPath)
	}
	todo, err := ParseTODO(todoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse TODO: %w", err)
	}
	todo.Provider = ProviderFiles
	return todo, nil
}

func (p *FileProvider) UpdateState(_ context.Context, todo *types.TODO, updates StateUpdate) error {
	return UpdateTODOState(todo, updates)
}

func (p *FileProvider) UpdateLatestFailure(_ context.Context, todo *types.TODO, result *types.TestResultInfo) error {
	return UpdateLatestFailure(todo, result)
}

func (p *FileProvider) SaveAttempt(_ context.Context, todo *types.TODO, result *ExecutionResult) error {
	return saveAttempt(todo, result)
}

func TODOReference(todo *types.TODO) string {
	if todo == nil {
		return ""
	}
	if todo.Provider == ProviderGrite && todo.ID != "" {
		return todo.ID
	}
	if todo.FilePath != "" {
		return todo.FilePath
	}
	return todo.ID
}
