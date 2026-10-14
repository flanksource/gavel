package native

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/flanksource/gavel/todos/types"
)

var (
	ErrPRLinkExists   = errors.New("native todo PR link already exists")
	ErrPRLinkNotFound = errors.New("native todo PR link not found")
)

// PRLink is one pull request linked to an issue. The issue's primary PR
// (todo_issues.pr) is always present as an implements link; see
// syncPrimaryPRLink.
type PRLink struct {
	WorkspaceID uuid.UUID          `json:"workspaceId"`
	IssueID     uuid.UUID          `json:"issueId"`
	Provider    string             `json:"provider"`
	Repo        string             `json:"repo"`
	Number      int                `json:"number"`
	URL         string             `json:"url"`
	Relation    types.PRRelation   `json:"relation"`
	Source      types.PRLinkSource `json:"source"`
	CreatedAt   time.Time          `json:"createdAt"`
}

// Ref is the link's PR identity.
func (l PRLink) Ref() types.PRRef {
	return types.PRRef{Provider: l.Provider, Repo: l.Repo, Number: l.Number, URL: l.URL}
}

// PRLinkInput is a request to link one PR to an issue.
type PRLinkInput struct {
	URL      string
	Relation types.PRRelation
	Source   types.PRLinkSource
	Actor    string
}

// PRLinkedIssue is an issue found through a PR, with how it is linked.
type PRLinkedIssue struct {
	Issue    *Issue
	Relation types.PRRelation
	Source   types.PRLinkSource
}

const prLinkColumns = `workspace_id, issue_id, provider, repo, number, url, relation, source, created_at`

// LinkIssuePR links a PR to an issue and records a pr_linked event. Linking a
// PR that is already linked is ErrPRLinkExists rather than a silent no-op: the
// existing link may carry a different relation, and overwriting it unasked would
// lose that.
func (r *Repository) LinkIssuePR(ctx context.Context, issueID uuid.UUID, expectedVersion int64, input PRLinkInput) (*PRLink, error) {
	ref, err := types.PRRefFromURL(input.URL)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	if relation, err := types.ParsePRRelation(string(input.Relation)); err != nil || relation != input.Relation {
		return nil, fmt.Errorf("%w: unsupported PR relation %q", ErrInvalidInput, input.Relation)
	}
	if err := input.Source.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	var created PRLink
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		result := tx.Raw(`
			INSERT INTO todo_pr_links (`+prLinkColumns+`)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, now())
			RETURNING `+prLinkColumns,
			locked.WorkspaceID, issueID, ref.Provider, ref.Repo, ref.Number, ref.URL, input.Relation, input.Source,
		).Scan(&created)
		if result.Error != nil {
			return mapUniqueError(result.Error, ErrPRLinkExists, "issue %s already links %s", issueID, ref.URL)
		}
		_, err = recordMutation(tx, locked, EventInput{Kind: "pr_linked", Actor: input.Actor, Payload: prLinkPayload(ref, map[string]any{
			"relation": input.Relation, "source": input.Source,
		})})
		return err
	})
	if err != nil {
		return nil, err
	}
	return &created, nil
}

// UnlinkIssuePR removes a PR link and records a pr_unlinked event. Removing the
// issue's primary PR also clears todo_issues.pr in the same transaction: the
// primary PR is defined as an implements link, so keeping one without the other
// would leave verification pointed at a PR the issue no longer lists.
func (r *Repository) UnlinkIssuePR(ctx context.Context, issueID uuid.UUID, expectedVersion int64, url, actor string) error {
	ref, err := types.PRRefFromURL(url)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := lockIssue(tx, issueID, expectedVersion)
		if err != nil {
			return err
		}
		result := tx.Exec(`DELETE FROM todo_pr_links WHERE issue_id = ? AND provider = ? AND repo = ? AND number = ?`,
			issueID, ref.Provider, ref.Repo, ref.Number)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("%w: issue %s does not link %s", ErrPRLinkNotFound, issueID, ref.URL)
		}
		current, err := getIssue(tx, `id = ?`, issueID)
		if err != nil {
			return err
		}
		clearedPrimary := false
		if current.PR != nil {
			primary, err := types.PRRefFromURL(current.PR.URL)
			if err != nil {
				return fmt.Errorf("issue %s primary PR: %w", issueID, err)
			}
			if clearedPrimary = primary.SamePR(ref); clearedPrimary {
				if err := tx.Exec(`UPDATE todo_issues SET pr = NULL WHERE id = ?`, issueID).Error; err != nil {
					return err
				}
			}
		}
		_, err = recordMutation(tx, locked, EventInput{Kind: "pr_unlinked", Actor: actor, Payload: prLinkPayload(ref, map[string]any{
			"clearedPrimary": clearedPrimary,
		})})
		return err
	})
}

// ListIssuePRLinks returns an issue's PR links, primary (implements) first,
// then newest first.
func (r *Repository) ListIssuePRLinks(ctx context.Context, issueID uuid.UUID) ([]PRLink, error) {
	var links []PRLink
	err := r.db.WithContext(ctx).Raw(`
		SELECT `+prLinkColumns+` FROM todo_pr_links
		WHERE issue_id = ?
		ORDER BY relation = 'implements' DESC, created_at DESC, provider, repo, number`, issueID,
	).Scan(&links).Error
	return links, err
}

// ListPRLinkedIssues returns the workspace's issues that link the PR at url.
func (r *Repository) ListPRLinkedIssues(ctx context.Context, workspaceID uuid.UUID, url string) ([]PRLinkedIssue, error) {
	ref, err := types.PRRefFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	db := r.db.WithContext(ctx)
	var links []PRLink
	if err := db.Raw(`
		SELECT `+prLinkColumns+` FROM todo_pr_links
		WHERE workspace_id = ? AND provider = ? AND repo = ? AND number = ?
		ORDER BY relation = 'implements' DESC, created_at DESC, issue_id`,
		workspaceID, ref.Provider, ref.Repo, ref.Number,
	).Scan(&links).Error; err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return []PRLinkedIssue{}, nil
	}
	ids := make([]uuid.UUID, len(links))
	for i, link := range links {
		ids[i] = link.IssueID
	}
	var records []issueRecord
	if err := db.Raw(`SELECT `+issueColumns+` FROM `+issueFrom+` WHERE issue.id IN ?`, ids).Scan(&records).Error; err != nil {
		return nil, err
	}
	issues := make(map[uuid.UUID]*Issue, len(records))
	for _, record := range records {
		issues[record.ID] = record.issue()
	}
	out := make([]PRLinkedIssue, 0, len(links))
	for _, link := range links {
		issue, ok := issues[link.IssueID]
		if !ok {
			return nil, fmt.Errorf("%w: PR link names issue %s", ErrNotFound, link.IssueID)
		}
		out = append(out, PRLinkedIssue{Issue: issue, Relation: link.Relation, Source: link.Source})
	}
	return out, nil
}

// syncPrimaryPRLink keeps the implements link in step with todo_issues.pr inside
// the transaction that changes the column: the previous primary's link goes,
// the new primary's is upserted as an implements/manual link. Upserting promotes
// a PR triage had linked as related, since naming it primary is a manual claim
// that it carries the work.
func syncPrimaryPRLink(tx *gorm.DB, workspaceID, issueID uuid.UUID, previous, next *types.PR) error {
	if previous != nil {
		ref, err := types.PRRefFromURL(previous.URL)
		if err != nil {
			return fmt.Errorf("issue %s previous primary PR: %w", issueID, err)
		}
		if err := tx.Exec(`
			DELETE FROM todo_pr_links
			WHERE issue_id = ? AND provider = ? AND repo = ? AND number = ? AND relation = ?`,
			issueID, ref.Provider, ref.Repo, ref.Number, types.PRRelationImplements,
		).Error; err != nil {
			return err
		}
	}
	if next == nil {
		return nil
	}
	ref, err := types.PRRefFromURL(next.URL)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidInput, err)
	}
	return tx.Exec(`
		INSERT INTO todo_pr_links (`+prLinkColumns+`)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, now())
		ON CONFLICT (workspace_id, issue_id, provider, repo, number)
		DO UPDATE SET url = EXCLUDED.url, relation = EXCLUDED.relation, source = EXCLUDED.source`,
		workspaceID, issueID, ref.Provider, ref.Repo, ref.Number, ref.URL, types.PRRelationImplements, types.PRLinkSourceManual,
	).Error
}

func prLinkPayload(ref types.PRRef, extra map[string]any) map[string]any {
	payload := map[string]any{"provider": ref.Provider, "repo": ref.Repo, "number": ref.Number, "url": ref.URL}
	for key, value := range extra {
		payload[key] = value
	}
	return payload
}
