package todos

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	Create(ctx context.Context, req CreateRequest) (*types.TODO, error)
	Delete(ctx context.Context, todo *types.TODO) error
	UpdateState(ctx context.Context, todo *types.TODO, updates StateUpdate) error
	UpdateLatestFailure(ctx context.Context, todo *types.TODO, result *types.TestResultInfo) error
	SaveAttempt(ctx context.Context, todo *types.TODO, result *ExecutionResult) error
}

type CreateRequest struct {
	Title    string
	Body     string
	Priority types.Priority
	Status   types.Status
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
	if result, err := ParseFrontmatterFromFile(todoPath); err == nil {
		todo.MarkdownBody = result.MarkdownContent
	}
	todo.Provider = ProviderFiles
	return todo, nil
}

func (p *FileProvider) Create(_ context.Context, req CreateRequest) (*types.TODO, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	priority := req.Priority
	if priority == "" {
		priority = types.PriorityMedium
	}
	status := req.Status
	if status == "" {
		status = types.StatusPending
	}
	path := filepath.Join(p.Dir, uniqueTODOFilename(p.Dir, title))
	frontmatter := types.TODOFrontmatter{
		Title:    title,
		Priority: priority,
		Status:   status,
	}
	frontmatter.CWD = p.WorkDir
	todo := &types.TODO{
		FilePath:        path,
		TODOFrontmatter: frontmatter,
		Implementation:  strings.TrimSpace(req.Body),
		MarkdownBody:    strings.TrimSpace(req.Body),
		Provider:        ProviderFiles,
	}
	if err := WriteTODOFile(path, todo); err != nil {
		return nil, err
	}
	return p.Get(context.Background(), path)
}

func (p *FileProvider) Delete(_ context.Context, todo *types.TODO) error {
	if todo == nil || todo.FilePath == "" {
		return fmt.Errorf("missing TODO file path")
	}
	return os.Remove(todo.FilePath)
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

var todoSlugInvalid = regexp.MustCompile(`[^a-z0-9]+`)

func uniqueTODOFilename(dir, title string) string {
	slug := strings.Trim(todoSlugInvalid.ReplaceAllString(strings.ToLower(title), "-"), "-")
	if slug == "" {
		slug = "todo"
	}
	name := slug + ".md"
	if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
		return name
	}
	for i := 2; ; i++ {
		name = fmt.Sprintf("%s-%d.md", slug, i)
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			return name
		}
	}
}
