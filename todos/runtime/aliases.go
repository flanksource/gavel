package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/githubpush"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
)

// externalIssueFrom picks the GitHub issue a TODO is linked to out of its
// aliases. A todo pushed with --force carries several; the first by alias order
// is the one shown, matching how resolveTarget reports them. An alias that no
// longer parses as an issue reference is skipped rather than failing the list —
// a malformed row must not make the whole backlog unreadable.
func externalIssueFrom(aliases []native.Alias) *types.ExternalIssue {
	for _, alias := range aliases {
		if !strings.EqualFold(alias.Kind, githubpush.AliasKind) {
			continue
		}
		ref, err := githubpush.ParseIssueRef(alias.Alias)
		if err != nil || ref.Repo == "" {
			continue
		}
		return &types.ExternalIssue{
			Kind:   githubpush.AliasKind,
			Repo:   ref.Repo,
			Number: ref.Number,
			URL:    ref.URL(),
		}
	}
	return nil
}

// Aliases lists every reference that resolves to the TODO, including the
// imported aliases native storage created for it.
func (p *Provider) Aliases(ctx context.Context, todo *types.TODO) ([]todos.TodoAlias, error) {
	id, _, err := p.mutationIdentity(todo)
	if err != nil {
		return nil, err
	}
	stored, err := p.repository.ListAliases(ctx, id)
	if err != nil {
		return nil, err
	}
	aliases := make([]todos.TodoAlias, 0, len(stored))
	for _, alias := range stored {
		aliases = append(aliases, todos.TodoAlias{Alias: alias.Alias, Kind: alias.Kind})
	}
	return aliases, nil
}

// AddAlias appends one reference to the TODO. Native storage replaces an
// issue's whole alias set, so the existing aliases are read and rewritten
// alongside the new one rather than being dropped.
func (p *Provider) AddAlias(ctx context.Context, todo *types.TODO, alias todos.TodoAlias) error {
	if strings.TrimSpace(alias.Alias) == "" {
		return fmt.Errorf("alias cannot be empty")
	}
	id, version, err := p.mutationIdentity(todo)
	if err != nil {
		return err
	}
	existing, err := p.repository.ListAliases(ctx, id)
	if err != nil {
		return err
	}
	merged := make([]native.AliasInput, 0, len(existing)+1)
	for _, stored := range existing {
		merged = append(merged, native.AliasInput{Alias: stored.Alias, Kind: stored.Kind})
	}
	merged = append(merged, native.AliasInput{Alias: alias.Alias, Kind: alias.Kind})

	issue, err := p.repository.SetAliases(ctx, id, version, merged, mutationActor)
	if err != nil {
		return err
	}
	// SetAliases bumps the issue version; refresh the caller's TODO so a
	// follow-up mutation does not fail its optimistic-concurrency check.
	return p.replaceTODO(ctx, todo, issue, p.workDir)
}
