package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
)

var testProviders = struct {
	sync.Mutex
	byDir map[string]*testTODOProvider
}{byDir: map[string]*testTODOProvider{}}

type testTODOProvider struct {
	dir   string
	items types.TODOS
}

func testProviderFor(dir string) *testTODOProvider {
	testProviders.Lock()
	defer testProviders.Unlock()
	if provider := testProviders.byDir[dir]; provider != nil {
		return provider
	}
	provider := &testTODOProvider{dir: dir}
	testProviders.byDir[dir] = provider
	return provider
}

func (p *testTODOProvider) List(_ context.Context, filters todos.DiscoveryFilters) (types.TODOS, error) {
	result := make(types.TODOS, 0, len(p.items))
	for _, todo := range p.items {
		if filters.Matches(todo) {
			result = append(result, todo)
		}
	}
	result.Sort()
	return result, nil
}

func (p *testTODOProvider) CountByStatus(_ context.Context) (map[types.Status]int, error) {
	counts := map[types.Status]int{}
	for _, todo := range p.items {
		counts[todo.Status]++
	}
	return counts, nil
}

func (p *testTODOProvider) Get(_ context.Context, ref string) (*types.TODO, error) {
	for _, todo := range p.items {
		if todo.ID == ref || todo.FilePath == ref || strings.EqualFold(todo.Title, ref) {
			return todo, nil
		}
	}
	return nil, fmt.Errorf("todo %q not found", ref)
}

func (p *testTODOProvider) Create(_ context.Context, request todos.CreateRequest) (*types.TODO, error) {
	priority := request.Priority
	if priority == "" {
		priority = types.PriorityMedium
	}
	status := request.Status
	if status == "" {
		status = types.StatusPending
	}
	created := time.Now().UTC()
	id := fmt.Sprintf("todo-%d", len(p.items)+1)
	body, bodyVerification, _ := todos.SplitVerificationFixture(request.Body)
	verification := todos.CombineVerificationFixtures(request.Verification, bodyVerification)
	parsed, err := todos.ParseTODOContent(request.Title, body, p.dir, types.TODOFrontmatter{
		Title: request.Title, Priority: priority, Status: status, CWD: p.dir, Created: &created,
	})
	if err != nil {
		return nil, err
	}
	parsed.VerificationMarkdown = verification
	parsed.Verification, err = todos.ParseVerificationMarkdown(todos.VerificationMarkdownOptions{
		Name: request.Title, Markdown: verification, SourceDir: p.dir,
		FrontMatter: parsed.FrontMatter,
	})
	if err != nil {
		return nil, err
	}
	parsed.ID = id
	parsed.FilePath = id
	parsed.Provider = todos.ProviderDB
	p.items = append(p.items, parsed)
	return parsed, nil
}

func (p *testTODOProvider) Delete(_ context.Context, todo *types.TODO) error {
	for i, candidate := range p.items {
		if candidate == todo {
			p.items = append(p.items[:i], p.items[i+1:]...)
			return nil
		}
	}
	return nil
}

func (p *testTODOProvider) Edit(_ context.Context, todo *types.TODO, edit todos.EditRequest) error {
	if edit.Title != nil {
		todo.Title = strings.TrimSpace(*edit.Title)
	}
	if edit.Body != nil || edit.Verification != nil {
		body := todo.MarkdownBody
		verification := todo.VerificationMarkdown
		var bodyVerification string
		if edit.Body != nil {
			var hasVerification bool
			body, bodyVerification, hasVerification = todos.SplitVerificationFixture(*edit.Body)
			if hasVerification {
				verification = bodyVerification
			}
		}
		if edit.Verification != nil {
			verification = todos.CombineVerificationFixtures(*edit.Verification, bodyVerification)
		}
		parsed, err := todos.ParseTODOContent(todo.Title, body, p.dir, todo.TODOFrontmatter)
		if err != nil {
			return err
		}
		parsed.VerificationMarkdown = verification
		parsed.Verification, err = todos.ParseVerificationMarkdown(todos.VerificationMarkdownOptions{
			Name: todo.Title, Markdown: verification, SourceDir: p.dir,
			FrontMatter: parsed.FrontMatter,
		})
		if err != nil {
			return err
		}
		id, path, provider := todo.ID, todo.FilePath, todo.Provider
		*todo = *parsed
		todo.ID, todo.FilePath, todo.Provider = id, path, provider
	}
	return nil
}

func (p *testTODOProvider) Comment(context.Context, *types.TODO, string) error { return nil }

func (p *testTODOProvider) UpdateState(_ context.Context, todo *types.TODO, update todos.StateUpdate) error {
	if update.Status != nil {
		todo.Status = *update.Status
	}
	if update.Priority != nil {
		todo.Priority = *update.Priority
	}
	if update.Attempts != nil {
		todo.Attempts = *update.Attempts
	}
	if update.LastRun != nil {
		todo.LastRun = update.LastRun
	}
	if update.PlanPath != nil {
		todo.PlanPath = *update.PlanPath
	}
	if update.PlanStatus != nil {
		todo.PlanStatus = *update.PlanStatus
	}
	return nil
}

func (p *testTODOProvider) UpdateLatestFailure(context.Context, *types.TODO, *types.TestResultInfo) error {
	return nil
}

func (p *testTODOProvider) SaveAttempt(context.Context, *types.TODO, *todos.ExecutionResult) error {
	return nil
}

func (p *testTODOProvider) SupportsGroupedExecution() bool { return false }
