package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

var uiTestProviders = struct {
	sync.Mutex
	byDir map[string]*uiTestTODOProvider
}{byDir: map[string]*uiTestTODOProvider{}}

var uiTestTodoSequence atomic.Uint64

func init() {
	openTodoProvider = func(_ context.Context, dir string) (todos.Provider, error) {
		return uiTestProviderFor(dir), nil
	}
	openGlobalTodoProvider = func(context.Context) (todos.GlobalReferenceProvider, error) {
		return uiTestGlobalProvider{}, nil
	}
}

type uiTestGlobalProvider struct{}

func (uiTestGlobalProvider) GetGlobal(ctx context.Context, ref string) (*types.TODO, error) {
	uiTestProviders.Lock()
	providers := make([]*uiTestTODOProvider, 0, len(uiTestProviders.byDir))
	for _, provider := range uiTestProviders.byDir {
		providers = append(providers, provider)
	}
	uiTestProviders.Unlock()
	for _, provider := range providers {
		if todo, err := provider.Get(ctx, ref); err == nil {
			return todo, nil
		}
	}
	return nil, fmt.Errorf("%w: todo %q", native.ErrNotFound, ref)
}

func (uiTestGlobalProvider) GetGlobalBySession(_ context.Context, ref string) (*types.TODO, string, error) {
	uiTestProviders.Lock()
	providers := make([]*uiTestTODOProvider, 0, len(uiTestProviders.byDir))
	for _, provider := range uiTestProviders.byDir {
		providers = append(providers, provider)
	}
	uiTestProviders.Unlock()
	for _, provider := range providers {
		for _, todo := range provider.items {
			if todo.LLM != nil && strings.EqualFold(strings.TrimSpace(todo.LLM.SessionId), strings.TrimSpace(ref)) {
				return todo, todo.LLM.SessionId, nil
			}
		}
	}
	return nil, "", fmt.Errorf("%w: session %q", native.ErrNotFound, ref)
}

func uiTestProviderFor(dir string) *uiTestTODOProvider {
	uiTestProviders.Lock()
	defer uiTestProviders.Unlock()
	if provider := uiTestProviders.byDir[dir]; provider != nil {
		return provider
	}
	provider := &uiTestTODOProvider{dir: dir}
	uiTestProviders.byDir[dir] = provider
	return provider
}

type uiTestTODOProvider struct {
	dir   string
	items types.TODOS
	// activeRun is the Captain prompt run backing the current attempt, as the
	// native PostgreSQL runtime would report it. nil means no run history.
	activeRun *captaindb.PromptRun
	comments  []string
	links     []uiTestLink
}

type uiTestLink struct {
	issueID  string
	targetID string
	relation types.RelationKind
}

func (p *uiTestTODOProvider) ActivePromptRun(context.Context, *types.TODO) (*captaindb.PromptRun, error) {
	return p.activeRun, nil
}

func (p *uiTestTODOProvider) List(_ context.Context, filters todos.DiscoveryFilters) (types.TODOS, error) {
	result := make(types.TODOS, 0, len(p.items))
	for _, todo := range p.items {
		if filters.Matches(todo) {
			result = append(result, todo)
		}
	}
	result.Sort()
	return result, nil
}

func (p *uiTestTODOProvider) CountByStatus(_ context.Context) (map[types.Status]int, error) {
	counts := map[types.Status]int{}
	for _, todo := range p.items {
		counts[todo.Status]++
	}
	return counts, nil
}

func (p *uiTestTODOProvider) Get(_ context.Context, ref string) (*types.TODO, error) {
	for _, todo := range p.items {
		if todo.ID == ref || todo.FilePath == ref || strings.EqualFold(todo.Title, ref) {
			return todo, nil
		}
	}
	return nil, fmt.Errorf("todo %q not found", ref)
}

func (p *uiTestTODOProvider) Create(_ context.Context, request todos.CreateRequest) (*types.TODO, error) {
	priority := request.Priority
	if priority == "" {
		priority = types.PriorityMedium
	}
	status := request.Status
	if status == "" {
		status = types.StatusPending
	}
	created := time.Now().UTC()
	id := fmt.Sprintf("todo-%d", uiTestTodoSequence.Add(1))
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
	parsed.Labels = append([]string(nil), request.Labels...)
	p.items = append(p.items, parsed)
	return parsed, nil
}

func (p *uiTestTODOProvider) Delete(_ context.Context, todo *types.TODO) error {
	for i, candidate := range p.items {
		if candidate == todo {
			p.items = append(p.items[:i], p.items[i+1:]...)
			return nil
		}
	}
	return nil
}

func (p *uiTestTODOProvider) Edit(_ context.Context, todo *types.TODO, edit todos.EditRequest) error {
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

func (p *uiTestTODOProvider) Comment(_ context.Context, todo *types.TODO, body string) error {
	p.comments = append(p.comments, body)
	updated := strings.TrimSpace(todo.MarkdownBody) + "\n\n## Comments\n\n" + strings.TrimSpace(body) + "\n"
	return p.Edit(context.Background(), todo, todos.EditRequest{Body: &updated})
}

func (p *uiTestTODOProvider) UpdateState(_ context.Context, todo *types.TODO, update todos.StateUpdate) error {
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
	if update.SessionID != nil {
		if todo.LLM == nil {
			todo.LLM = &types.LLM{}
		}
		todo.LLM.SessionId = *update.SessionID
	}
	if update.PlanPath != nil {
		todo.PlanPath = *update.PlanPath
	}
	if update.PlanStatus != nil {
		todo.PlanStatus = *update.PlanStatus
	}
	if update.RunMode != nil {
		todo.RunMode = *update.RunMode
	}
	if update.Questions != nil {
		todo.Questions = append([]types.AgentQuestion(nil), (*update.Questions)...)
	}
	return nil
}

func (p *uiTestTODOProvider) UpdateLatestFailure(context.Context, *types.TODO, *types.TestResultInfo) error {
	return nil
}
func (p *uiTestTODOProvider) SaveAttempt(context.Context, *types.TODO, *todos.ExecutionResult) error {
	return nil
}
func (p *uiTestTODOProvider) SupportsGroupedExecution() bool { return false }

func (p *uiTestTODOProvider) MoveTo(ctx context.Context, todo *types.TODO, target todos.Provider) (*types.TODO, error) {
	created, err := target.Create(ctx, todos.CreateRequest{
		Title: todo.Title, Body: todo.MarkdownBody, Verification: todo.VerificationMarkdown,
		Priority: todo.Priority, Status: todo.Status, Labels: todo.Labels,
	})
	if err != nil {
		return nil, err
	}
	if err := p.Delete(ctx, todo); err != nil {
		return nil, err
	}
	return created, nil
}

// Link/Unlink/Links mirror the native repository's perspective flip: an
// incoming depends_on is reported to the target as the derived blocks relation.
func (p *uiTestTODOProvider) Link(ctx context.Context, todo *types.TODO, targetRef string, relation types.RelationKind) (*todos.Link, error) {
	target, err := p.Get(ctx, targetRef)
	if err != nil {
		return nil, err
	}
	if target.ID == todo.ID {
		return nil, fmt.Errorf("a TODO cannot be linked to itself")
	}
	for _, existing := range p.links {
		if existing.issueID == todo.ID && existing.targetID == target.ID && existing.relation == relation {
			return nil, fmt.Errorf("link already exists")
		}
	}
	p.links = append(p.links, uiTestLink{issueID: todo.ID, targetID: target.ID, relation: relation})
	return &todos.Link{
		Relation: relation, TargetID: target.ID, TargetShortID: target.ShortID,
		TargetTitle: target.Title, TargetStatus: target.Status,
	}, nil
}

func (p *uiTestTODOProvider) Unlink(ctx context.Context, todo *types.TODO, targetRef string, relation types.RelationKind) error {
	target, err := p.Get(ctx, targetRef)
	if err != nil {
		return err
	}
	for i, existing := range p.links {
		if existing.issueID == todo.ID && existing.targetID == target.ID && existing.relation == relation {
			p.links = append(p.links[:i], p.links[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("link not found")
}

func (p *uiTestTODOProvider) Links(ctx context.Context, todo *types.TODO) ([]todos.Link, error) {
	links := make([]todos.Link, 0, len(p.links))
	for _, existing := range p.links {
		relation, targetID := existing.relation, existing.targetID
		switch {
		case existing.issueID == todo.ID:
		case existing.targetID == todo.ID:
			targetID = existing.issueID
			if relation == types.RelationDependsOn {
				relation = types.RelationBlocks
			}
		default:
			continue
		}
		target, err := p.Get(ctx, targetID)
		if err != nil {
			return nil, err
		}
		links = append(links, todos.Link{
			Relation: relation, TargetID: target.ID, TargetShortID: target.ShortID,
			TargetTitle: target.Title, TargetStatus: target.Status,
		})
	}
	return links, nil
}

func (p *uiTestTODOProvider) ApprovePlan(ctx context.Context, todo *types.TODO, _, _ string) (*types.TODO, error) {
	pending := types.StatusPending
	if err := p.UpdateState(ctx, todo, todos.StateUpdate{Status: &pending}); err != nil {
		return nil, err
	}
	return todo, nil
}

func (p *uiTestTODOProvider) RejectPlan(ctx context.Context, todo *types.TODO, _, _ string) (*types.TODO, error) {
	pending := types.StatusPending
	path := ""
	status := types.PlanStatus("")
	if err := p.UpdateState(ctx, todo, todos.StateUpdate{Status: &pending, PlanPath: &path, PlanStatus: &status}); err != nil {
		return nil, err
	}
	return todo, nil
}

func (p *uiTestTODOProvider) RequestPlanRevision(_ context.Context, todo *types.TODO, _, _ string) (*types.TODO, error) {
	return todo, nil
}

func (p *uiTestTODOProvider) PlanMarkdown(_ context.Context, todo *types.TODO, _ types.RunMode) (string, error) {
	if strings.TrimSpace(todo.PlanPath) == "" {
		return "", nil
	}
	data, err := os.ReadFile(todo.PlanPath)
	if os.IsNotExist(err) {
		return "", nil
	}
	return string(data), err
}
