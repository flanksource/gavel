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
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ todos.TriageNewProvider = (*Provider)(nil)

var triageAttachmentURLs = regexp.MustCompile(regexp.QuoteMeta(todos.AttachmentURLPrefix) + `[^\s)<>"']+`)

func (p *Provider) ValidateTriageNew(ctx context.Context, todo *types.TODO, env *types.TriageNewEnvelope) error {
	if err := env.Validate(); err != nil {
		return err
	}
	if env.EndStatus != types.EndCompleted {
		return nil
	}
	if closedTriageTODO(todo) {
		return fmt.Errorf("triage.new source %s is closed", todo.ID)
	}
	definitions, err := p.LabelDefinitions(ctx)
	if err != nil {
		return fmt.Errorf("triage.new taxonomy: %w", err)
	}
	for _, label := range env.Labels {
		if !labels.Contains(definitions.Names(), label) {
			return fmt.Errorf("triage.new label %q is outside the workspace taxonomy", label)
		}
	}
	if env.Action == "keep" {
		return nil
	}
	target, err := p.Get(ctx, env.Target)
	if err != nil {
		return err
	}
	if target.ID == todo.ID || target.CWD != todo.CWD {
		return fmt.Errorf("triage.new target must be another TODO in the same workspace")
	}
	if closedTriageTODO(target) {
		return fmt.Errorf("triage.new target %s is closed", target.ID)
	}
	if env.Action == "child-of" {
		children, err := p.List(ctx, todos.DiscoveryFilters{ParentID: todo.ID})
		if err != nil {
			return err
		}
		if target.ParentID != "" || len(children) > 0 {
			return fmt.Errorf("triage.new cannot create nested children")
		}
		return nil
	}
	if env.Action == "merge-into" {
		proposal, hasPlan, err := p.triageNewMerge(ctx, todo, target, env)
		if err != nil {
			return err
		}
		_, err = proposal.Validate(target, []*types.TODO{todo}, todo.VerificationMarkdown != "" || target.VerificationMarkdown != "", hasPlan, p.workDir)
		return err
	}
	return nil
}

func (p *Provider) ApplyTriageNew(ctx context.Context, todo *types.TODO, env *types.TriageNewEnvelope, opts todos.TriageNewApplyOptions) error {
	if err := env.Validate(); err != nil {
		return err
	}
	if env.EndStatus != types.EndCompleted {
		return nil
	}
	if _, err := p.requireWorkspace(ctx); err != nil {
		return err
	}
	ids := []string{todo.ID}
	if env.Action != "keep" {
		if opts.TargetVersion <= 0 {
			return fmt.Errorf("triage.new requires the reviewed target version")
		}
		ids = append(ids, env.Target)
	}
	for _, id := range ids {
		if _, err := uuid.Parse(id); err != nil {
			return fmt.Errorf("triage.new issue ID: %w", err)
		}
	}
	var updated *types.TODO
	err := p.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked []struct{ ID string }
		if err := tx.Table("todo_issues").Select("id").Where("workspace_id = ? AND id IN ?", p.workspace.ID, ids).Order("id").Clauses(clause.Locking{Strength: "UPDATE"}).Find(&locked).Error; err != nil {
			return err
		}
		if len(locked) != len(ids) {
			return fmt.Errorf("triage.new source or target is missing from this workspace")
		}
		global, err := NewGlobal(tx)
		if err != nil {
			return err
		}
		writer := global.withWorkspace(p.workspace, p.workDir)
		updated, err = writer.Get(ctx, todo.ID)
		if err != nil {
			return err
		}
		if updated.Version != todo.Version {
			return fmt.Errorf("triage.new source changed since review")
		}
		if env.Action != "keep" {
			target, err := writer.Get(ctx, env.Target)
			if err != nil {
				return err
			}
			if target.Version != opts.TargetVersion {
				return fmt.Errorf("triage.new target changed since review")
			}
		}
		return writer.applyTriageNew(ctx, updated, env)
	})
	if err != nil {
		return err
	}
	*todo = *updated
	return nil
}

func (p *Provider) applyTriageNew(ctx context.Context, todo *types.TODO, env *types.TriageNewEnvelope) error {
	if err := p.ValidateTriageNew(ctx, todo, env); err != nil {
		return err
	}
	if env.EndStatus != types.EndCompleted {
		return nil
	}
	if env.Action == "merge-into" {
		target, err := p.Get(ctx, env.Target)
		if err != nil {
			return err
		}
		proposal, _, err := p.triageNewMerge(ctx, todo, target, env)
		if err != nil {
			return err
		}
		result := merge.Apply(ctx, []bulk.Target{{Provider: p, Todo: target}, {Provider: p, Todo: todo}}, &merge.Selection{Survivor: target, Retired: []*types.TODO{todo}}, proposal, merge.Options{WorkDir: p.workDir})
		for _, item := range result.Results {
			if item.Error != "" {
				return fmt.Errorf("triage.new merge %s: %s", item.Ref, item.Error)
			}
		}
		return nil
	}
	nextLabels := labels.Apply(todo.Labels, env.Labels, nil)
	if err := p.Edit(ctx, todo, todos.EditRequest{Title: &env.Title, Labels: &nextLabels}); err != nil {
		return err
	}
	switch env.Action {
	case "duplicate-of":
		return todos.ApplyTriage(ctx, p, todo, &types.TriageEnvelope{ResultEnvelope: env.ResultEnvelope, Verdict: types.VerdictDuplicateOf, DuplicateOf: env.Target, Comment: env.Rationale}, todos.TriageOptions{WorkDir: p.workDir})
	case "child-of":
		if err := p.SetParent(ctx, todo, env.Target); err != nil {
			return err
		}
		return p.Comment(ctx, todo, todos.CommentRequest{Body: fmt.Sprintf("Made child of %s. %s", env.Target, env.Rationale)})
	}
	return nil
}

func (p *Provider) triageNewMerge(ctx context.Context, source, target *types.TODO, env *types.TriageNewEnvelope) (*merge.Proposal, bool, error) {
	hasPlan := false
	criteria := map[string]bool{}
	for _, criterion := range todos.ParseAcceptanceCriteria(env.Body) {
		criteria[criterion.Text] = true
	}
	for _, todo := range []*types.TODO{source, target} {
		for _, criterion := range todo.AcceptanceCriteria {
			if !criteria[criterion.Text] {
				return nil, false, fmt.Errorf("triage.new merge dropped acceptance criterion %q from %s", criterion.Text, todo.ID)
			}
		}
		for _, attachment := range triageAttachmentURLs.FindAllString(todo.MarkdownBody, -1) {
			if !strings.Contains(env.Body, attachment) {
				return nil, false, fmt.Errorf("triage.new merge dropped attachment %s from %s", attachment, todo.ID)
			}
		}
		plan, err := p.PlanMarkdown(ctx, todo, types.ModePlan)
		if err != nil {
			return nil, false, err
		}
		hasPlan = hasPlan || strings.TrimSpace(plan) != ""
	}
	return &merge.Proposal{Title: env.Title, Body: env.Body, Verification: env.Verification, Plan: env.Plan, Labels: labels.Apply(labels.Apply(target.Labels, source.Labels, nil), env.Labels, nil), Summary: env.Summary, Rationale: env.Rationale}, hasPlan, nil
}

func closedTriageTODO(todo *types.TODO) bool {
	return todo.Status == types.StatusCompleted || todo.Status == types.StatusVerified || todo.Status == types.StatusSkipped
}
