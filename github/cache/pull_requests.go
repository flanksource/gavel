package cache

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

// ErrPRIndexUnavailable is returned by PR index reads on a disabled store.
// Writes stay a no-op there, like every other cache write, but a search over an
// index that does not exist must not pass for "nothing matched".
var ErrPRIndexUnavailable = errors.New("PR index unavailable: the Gavel database is not configured")

// ErrInvalidPRQuery wraps a search the caller phrased wrongly, as opposed to a
// search the database failed to run.
var ErrInvalidPRQuery = errors.New("invalid PR search")

// MaxPRSearchLimit caps one search. The index backs a picker and a triage
// lookup; neither reads more than a page.
const MaxPRSearchLimit = 200

// PullRequest is one row of the persisted PR index, written by the PR
// dashboard poller. Repo keeps the host's display spelling; lookups compare it
// case-insensitively.
type PullRequest struct {
	OrgID     uuid.UUID  `gorm:"primaryKey;type:uuid" json:"-"`
	UserID    *uuid.UUID `gorm:"type:uuid" json:"-"`
	Provider  string     `gorm:"primaryKey" json:"provider"`
	Repo      string     `gorm:"primaryKey" json:"repo"`
	Number    int        `gorm:"primaryKey" json:"number"`
	URL       string     `json:"url"`
	Title     string     `json:"title"`
	Author    string     `json:"author"`
	State     string     `json:"state"`
	IsDraft   bool       `json:"isDraft"`
	Head      string     `json:"head,omitempty"`
	Base      string     `json:"base,omitempty"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime:false" json:"updatedAt"`
	SyncedAt  time.Time  `json:"syncedAt"`
}

func (PullRequest) TableName() string { return "pull_requests" }

// PRKey identifies one indexed PR. Repo is compared case-insensitively.
type PRKey struct {
	Provider string
	Repo     string
	Number   int
}

func (k PRKey) normalized() PRKey {
	return PRKey{Provider: k.Provider, Repo: strings.ToLower(k.Repo), Number: k.Number}
}

// Key returns the row's case-insensitive identity, the shape GetPRs keys by.
func (p PullRequest) Key() PRKey {
	return PRKey{Provider: p.Provider, Repo: p.Repo, Number: p.Number}.normalized()
}

// PRQuery filters a PR index search. Empty Text lists the most recently
// updated PRs; State is open, closed, merged or all (empty means all).
type PRQuery struct {
	Text  string
	Repo  string
	State string
	Limit int
}

// UpsertPRs writes the poller's current results. A disabled store is a no-op:
// without a database the dashboard keeps its in-memory list, as it always has.
func (s *Store) UpsertPRs(prs []PullRequest) error {
	if s.Disabled() || len(prs) == 0 {
		return nil
	}
	now := time.Now()
	rows := make([]PullRequest, len(prs))
	for i, pr := range prs {
		if pr.Provider == "" || pr.Repo == "" || pr.Number <= 0 || pr.URL == "" {
			return fmt.Errorf("index PR %s %s#%d: provider, repo, number and url are required", pr.Provider, pr.Repo, pr.Number)
		}
		pr.OrgID = s.org
		pr.SyncedAt = now
		rows[i] = pr
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	err := s.gorm().Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "org_id"}, {Name: "provider"}, {Name: "repo"}, {Name: "number"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"url", "title", "author", "state", "is_draft", "head", "base", "updated_at", "synced_at",
		}),
	}).Create(&rows).Error
	if err != nil {
		return fmt.Errorf("index %d PRs: %w", len(rows), err)
	}
	return nil
}

var prSearchToken = regexp.MustCompile(`[\p{L}\p{N}_]+`)

// SearchPRs finds indexed PRs. Text that is a PR number ("123" or "#123")
// matches that number; any other text matches titles word by word, each word as
// a prefix, so a picker narrows as the user types.
func (s *Store) SearchPRs(query PRQuery) ([]PullRequest, error) {
	if s.Disabled() {
		return nil, ErrPRIndexUnavailable
	}
	if query.Limit <= 0 || query.Limit > MaxPRSearchLimit {
		return nil, fmt.Errorf("%w: limit %d must be between 1 and %d", ErrInvalidPRQuery, query.Limit, MaxPRSearchLimit)
	}
	db := s.gorm().Model(&PullRequest{}).Where("org_id = ?", s.org)
	switch state := strings.ToLower(strings.TrimSpace(query.State)); state {
	case "", "all":
	case "open", "closed", "merged":
		db = db.Where("lower(state) = ?", state)
	default:
		return nil, fmt.Errorf("%w: unknown PR state %q; valid states: open, closed, merged, all", ErrInvalidPRQuery, query.State)
	}
	if repo := strings.TrimSpace(query.Repo); repo != "" {
		db = db.Where("lower(repo) = ?", strings.ToLower(repo))
	}
	text := strings.TrimSpace(query.Text)
	if number, err := strconv.Atoi(strings.TrimPrefix(text, "#")); err == nil && number > 0 {
		db = db.Where("number = ?", number)
	} else if tokens := prSearchToken.FindAllString(text, -1); len(tokens) > 0 {
		db = db.Where("to_tsvector('simple'::regconfig, title) @@ to_tsquery('simple', ?)",
			strings.Join(tokens, ":* & ")+":*")
	} else if text != "" {
		return nil, fmt.Errorf("%w: text %q has no searchable words", ErrInvalidPRQuery, query.Text)
	}
	var rows []PullRequest
	if err := db.Order("updated_at DESC NULLS LAST, number DESC").Limit(query.Limit).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("search PR index: %w", err)
	}
	return rows, nil
}

// GetPRs returns the indexed rows for keys, keyed by PullRequest.Key. A key the
// poller has not synced is simply absent.
func (s *Store) GetPRs(keys []PRKey) (map[PRKey]PullRequest, error) {
	if s.Disabled() {
		return nil, ErrPRIndexUnavailable
	}
	out := map[PRKey]PullRequest{}
	if len(keys) == 0 {
		return out, nil
	}
	tuples := make([][]any, 0, len(keys))
	for _, key := range keys {
		key = key.normalized()
		tuples = append(tuples, []any{key.Provider, key.Repo, key.Number})
	}
	var rows []PullRequest
	if err := s.gorm().Where("org_id = ? AND (provider, lower(repo), number) IN ?", s.org, tuples).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("load indexed PRs: %w", err)
	}
	for _, row := range rows {
		out[row.Key()] = row
	}
	return out, nil
}
