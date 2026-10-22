package runtime

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/bulk"
	"github.com/flanksource/gavel/todos/labels"
	"github.com/flanksource/gavel/todos/merge"
	"github.com/flanksource/gavel/todos/types"
)

var _ todos.MergeProvider = (*Provider)(nil)

var mergeAttachmentURLs = regexp.MustCompile(regexp.QuoteMeta(todos.AttachmentURLPrefix) + `[^\s)<>"']+`)

// MergeInto folds request.Retired into request.Survivor through merge.Apply, the
// write path `gavel todos merge` uses. Everything that could make the merge
// lose work is checked first and refuses it as a whole: TODOs linked to
// different PRs, a dropped acceptance criterion or attachment, and a fixture
// or plan one of the sources had.
func (p *Provider) MergeInto(ctx context.Context, request todos.MergeRequest) error {
	all := append([]*types.TODO{request.Survivor}, request.Retired...)
	proposal, hasPlan, err := p.mergeProposal(ctx, all, request)
	if err != nil {
		return err
	}
	hasFixture := false
	for _, todo := range all {
		hasFixture = hasFixture || strings.TrimSpace(todo.VerificationMarkdown) != ""
	}
	selection, err := proposal.Validate(request.Survivor, request.Retired, hasFixture, hasPlan, p.workDir)
	if err != nil {
		return err
	}
	if request.Preview {
		return nil
	}
	targets := make([]bulk.Target, len(all))
	for i, todo := range all {
		targets[i] = bulk.Target{Provider: p, Todo: todo}
	}
	result := merge.Apply(ctx, targets, selection, proposal, merge.Options{WorkDir: p.workDir})
	for _, item := range result.Results {
		if item.Error != "" {
			return fmt.Errorf("merge %s: %s", item.Ref, item.Error)
		}
	}
	return nil
}

// mergeProposal is the merge as merge.Apply writes it, refused when the
// combined body drops a criterion or an attachment of any TODO merged. It also
// reports whether any of them has a plan, which the proposal must then carry.
func (p *Provider) mergeProposal(ctx context.Context, all []*types.TODO, request todos.MergeRequest) (*merge.Proposal, bool, error) {
	if _, err := types.MergedPR(all); err != nil {
		return nil, false, err
	}
	criteria := map[string]bool{}
	for _, criterion := range todos.ParseAcceptanceCriteria(request.Body) {
		criteria[criterion.Text] = true
	}
	hasPlan := false
	merged := []string{}
	for _, todo := range all {
		for _, criterion := range todo.AcceptanceCriteria {
			if !criteria[criterion.Text] {
				return nil, false, fmt.Errorf("merged body drops acceptance criterion %q of %s", criterion.Text, todos.RefOf(todo))
			}
		}
		for _, attachment := range mergeAttachmentURLs.FindAllString(todo.MarkdownBody, -1) {
			if !strings.Contains(request.Body, attachment) {
				return nil, false, fmt.Errorf("merged body drops attachment %s of %s", attachment, todos.RefOf(todo))
			}
		}
		plan, err := p.PlanMarkdown(ctx, todo, types.ModePlan)
		if err != nil {
			return nil, false, err
		}
		hasPlan = hasPlan || strings.TrimSpace(plan) != ""
		merged = labels.Apply(merged, todo.Labels, nil)
	}
	return &merge.Proposal{
		Title: request.Title, Body: request.Body, Verification: request.Verification, Plan: request.Plan,
		Labels: merged, Summary: fmt.Sprintf("Merged %d TODOs into %s", len(request.Retired), todos.RefOf(request.Survivor)),
		Rationale: request.Rationale,
	}, hasPlan, nil
}
