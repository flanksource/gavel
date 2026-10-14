package runtime

import (
	"context"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

var _ todos.PRLinkProvider = (*Provider)(nil)

// LinkPR links a PR to todo. URL validation, relation and source checks, and
// duplicate detection are the repository's, inside one transaction.
func (p *Provider) LinkPR(ctx context.Context, todo *types.TODO, request todos.PRLinkRequest) (*todos.PRLink, error) {
	id, version, err := p.mutationIdentity(todo)
	if err != nil {
		return nil, err
	}
	created, err := p.repository.LinkIssuePR(ctx, id, version, native.PRLinkInput{
		URL: request.URL, Relation: request.Relation, Source: request.Source, Actor: mutationActor,
	})
	if err != nil {
		return nil, err
	}
	if err := p.reloadTODO(ctx, todo, p.workDir); err != nil {
		return nil, err
	}
	link := prLinkFromNative(*created)
	return &link, nil
}

// UnlinkPR removes the link to the PR at url. Removing the primary PR's link
// also clears todo.PR; the reload carries that back to the caller.
func (p *Provider) UnlinkPR(ctx context.Context, todo *types.TODO, url string) error {
	id, version, err := p.mutationIdentity(todo)
	if err != nil {
		return err
	}
	if err := p.repository.UnlinkIssuePR(ctx, id, version, url, mutationActor); err != nil {
		return err
	}
	return p.reloadTODO(ctx, todo, p.workDir)
}

// PRLinks returns todo's PR links, the primary (implements) link first.
func (p *Provider) PRLinks(ctx context.Context, todo *types.TODO) ([]todos.PRLink, error) {
	id, _, err := p.mutationIdentity(todo)
	if err != nil {
		return nil, err
	}
	stored, err := p.repository.ListIssuePRLinks(ctx, id)
	if err != nil {
		return nil, err
	}
	links := make([]todos.PRLink, 0, len(stored))
	for _, link := range stored {
		links = append(links, prLinkFromNative(link))
	}
	return links, nil
}

// TodosForPR lists this workspace's TODOs that link the PR at url.
func (p *Provider) TodosForPR(ctx context.Context, url string) ([]todos.PRLinkedTODO, error) {
	linked, err := p.repository.ListPRLinkedIssues(ctx, p.workspace.ID, url)
	if err != nil {
		return nil, err
	}
	out := make([]todos.PRLinkedTODO, 0, len(linked))
	for _, entry := range linked {
		id := entry.Issue.ID.String()
		out = append(out, todos.PRLinkedTODO{
			ID:       id,
			ShortID:  id[:8],
			Title:    entry.Issue.Title,
			Status:   todoStatus(entry.Issue.Status, entry.Issue.ExecutionState),
			Priority: todoPriority(entry.Issue.Priority),
			Relation: entry.Relation,
			Source:   entry.Source,
		})
	}
	return out, nil
}

func prLinkFromNative(link native.PRLink) todos.PRLink {
	return todos.PRLink{
		PRRef:     link.Ref(),
		Relation:  link.Relation,
		Source:    link.Source,
		CreatedAt: link.CreatedAt,
	}
}
