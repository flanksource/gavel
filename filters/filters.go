// Package filters stores each user's named filters, such as the projects and GitHub
// orgs they hide. A list endpoint references a filter by name (?exclude=hidden-projects)
// instead of the client sending the values. Local mode has no user: its filters belong
// to the local org with a null owner.
package filters

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/gavel/tenant"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Well-known filter names and the kinds of value they hold.
const (
	HiddenProjects   = "hidden-projects"    // Kind workspace
	HiddenGitHubOrgs = "hidden-github-orgs" // Kind github-org

	// KindWorkspace values are todo workspace roots: an absolute directory locally,
	// repo:<id> in hosted mode.
	KindWorkspace = "workspace"
	// KindGitHubOrg values are GitHub org logins.
	KindGitHubOrg = "github-org"
)

var (
	// ErrUnknownFilter is a filter name that is not one of the well-known filters.
	ErrUnknownFilter = errors.New("unknown filter")
	// ErrInvalidValue is a blank value, or a workspace value that is neither an
	// absolute directory nor repo:<id>.
	ErrInvalidValue = errors.New("invalid filter value")
)

// wellKnown maps each filter name to its kind.
var wellKnown = map[string]string{HiddenProjects: KindWorkspace, HiddenGitHubOrgs: KindGitHubOrg}

// Filter is a named set of values. Values is never nil and is sorted.
type Filter struct {
	Name   string   `json:"name"`
	Kind   string   `json:"kind"`
	Values []string `json:"values"`
}

type filterRow struct {
	Values pq.StringArray `gorm:"column:values;type:text[]"`
}

// Store reads and edits filters in user_filters.
type Store struct {
	db *gorm.DB
}

// NewStore returns the filter store on db.
func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

// ownerMatch selects the rows of scope's owner; a nil user is the org's null owner.
const ownerMatch = `org_id = ? AND user_id IS NOT DISTINCT FROM ?`

// Get returns the named filter. A well-known filter with no row is empty.
func (s *Store) Get(ctx context.Context, scope tenant.Scope, name string) (Filter, error) {
	kind, err := kindOf(name)
	if err != nil {
		return Filter{}, err
	}
	var rows []filterRow
	if err := s.db.WithContext(ctx).Raw(`SELECT "values" FROM user_filters WHERE `+ownerMatch+` AND name = ?`,
		scope.OrgID, scope.UserID, name).Scan(&rows).Error; err != nil {
		return Filter{}, fmt.Errorf("read filter %q: %w", name, err)
	}
	filter := Filter{Name: name, Kind: kind, Values: []string{}}
	if len(rows) > 0 {
		filter.Values = sorted(rows[0].Values)
	}
	return filter, nil
}

// List returns every well-known filter for scope, sorted by name.
func (s *Store) List(ctx context.Context, scope tenant.Scope) ([]Filter, error) {
	names := make([]string, 0, len(wellKnown))
	for name := range wellKnown {
		names = append(names, name)
	}
	slices.Sort(names)
	out := make([]Filter, 0, len(names))
	for _, name := range names {
		filter, err := s.Get(ctx, scope, name)
		if err != nil {
			return nil, err
		}
		out = append(out, filter)
	}
	return out, nil
}

// Add puts value in the named filter; adding a value it holds changes nothing.
func (s *Store) Add(ctx context.Context, scope tenant.Scope, name, value string) (Filter, error) {
	kind, value, err := validate(name, value)
	if err != nil {
		return Filter{}, err
	}
	err = s.db.WithContext(ctx).Exec(`
		INSERT INTO user_filters (org_id, user_id, name, kind, "values") VALUES (?, ?, ?, ?, ARRAY[?]::text[])
		ON CONFLICT (org_id, COALESCE(user_id, '00000000-0000-0000-0000-000000000000'::uuid), name) DO UPDATE
		SET "values" = CASE WHEN EXCLUDED."values"[1] = ANY (user_filters."values") THEN user_filters."values"
			ELSE user_filters."values" || EXCLUDED."values" END,
			updated_at = now()`,
		scope.OrgID, scope.UserID, name, kind, value).Error
	if err != nil {
		return Filter{}, fmt.Errorf("add %q to filter %q: %w", value, name, err)
	}
	return s.Get(ctx, scope, name)
}

// Remove takes value out of the named filter; removing a value it lacks changes nothing.
func (s *Store) Remove(ctx context.Context, scope tenant.Scope, name, value string) (Filter, error) {
	_, value, err := validate(name, value)
	if err != nil {
		return Filter{}, err
	}
	err = s.db.WithContext(ctx).Exec(`UPDATE user_filters SET "values" = array_remove("values", ?), updated_at = now()
		WHERE `+ownerMatch+` AND name = ? AND ? = ANY ("values")`,
		value, scope.OrgID, scope.UserID, name, value).Error
	if err != nil {
		return Filter{}, fmt.Errorf("remove %q from filter %q: %w", value, name, err)
	}
	return s.Get(ctx, scope, name)
}

func kindOf(name string) (string, error) {
	kind, ok := wellKnown[name]
	if !ok {
		return "", fmt.Errorf("%w %q: expected %s or %s", ErrUnknownFilter, name, HiddenProjects, HiddenGitHubOrgs)
	}
	return kind, nil
}

// validate returns the filter's kind and value normalized for it: trimmed, and a
// workspace directory cleaned.
func validate(name, value string) (string, string, error) {
	kind, err := kindOf(name)
	if err != nil {
		return "", "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", "", fmt.Errorf("%w for filter %q: value is blank", ErrInvalidValue, name)
	}
	if kind != KindWorkspace || strings.HasPrefix(value, "repo:") {
		return kind, value, nil
	}
	if !filepath.IsAbs(value) {
		return "", "", fmt.Errorf("%w for filter %q: workspace %q is neither an absolute directory nor repo:<id>", ErrInvalidValue, name, value)
	}
	return kind, filepath.Clean(value), nil
}

func sorted(values []string) []string {
	out := slices.Clone(values)
	if out == nil {
		out = []string{}
	}
	slices.Sort(out)
	return out
}
