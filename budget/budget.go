// Package budget keeps an organization's budgets for its substrate runs, which its admins set, and
// says how much of each its runs and model calls have used.
package budget

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	captainbudgets "github.com/flanksource/captain/pkg/budgets"
	"github.com/flanksource/gavel/platform"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Resource is what a budget caps.
type Resource string

const (
	// CPU is counted in core-hours, or cores at once.
	CPU Resource = "cpu"
	// Memory is counted in GiB-hours, or GiB at once.
	Memory Resource = "memory"
	// Tokens are the model calls' counted tokens: input, cache writes, output and reasoning.
	Tokens Resource = "tokens"
	// USD is the model calls' cost.
	USD Resource = "usd"
)

// Kind is how a budget caps its resource.
type Kind string

const (
	// Cumulative caps what the organization uses in a window.
	Cumulative Kind = "cumulative"
	// Concurrent caps what the organization's runs hold at once.
	Concurrent Kind = "concurrent"
)

var (
	resources = []Resource{CPU, Memory, Tokens, USD}
	kinds     = []Kind{Cumulative, Concurrent}
)

// Budget is one of an organization's budgets: at most one per resource and kind.
type Budget struct {
	Resource Resource `json:"resource"`
	Kind     Kind     `json:"kind"`
	// Amount is in core-hours, GiB-hours, tokens or USD for a cumulative budget, and cores or GiB
	// for a concurrent one.
	Amount float64 `json:"amount"`
	// Window is datemath relative to now where a cumulative budget's window starts, such as now/M
	// for this month; empty for a concurrent budget.
	Window    string     `json:"window,omitempty"`
	UpdatedBy *uuid.UUID `json:"updatedBy,omitempty"`
	UpdatedAt time.Time  `json:"updatedAt,omitzero"`
}

var (
	// ErrInvalid refuses what can't be a budget.
	ErrInvalid = errors.New("invalid budget")
	// ErrNotFound reports a budget an organization doesn't have.
	ErrNotFound = errors.New("budget not found")
)

func (b Budget) validate() error {
	invalid := func(format string, args ...any) error {
		return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
	}
	switch {
	case !slices.Contains(resources, b.Resource):
		return invalid("resource %q is not one of %s", b.Resource, joined(resources))
	case !slices.Contains(kinds, b.Kind):
		return invalid("kind %q is not one of %s", b.Kind, joined(kinds))
	case !(b.Amount > 0):
		return invalid("amount must be more than 0, got %v", b.Amount)
	case b.Kind == Concurrent && b.Resource != CPU && b.Resource != Memory:
		return invalid("only cpu and memory have concurrent budgets: %s are spent, not held", b.Resource)
	case b.Kind == Concurrent && b.Window != "":
		return invalid("a concurrent budget caps what runs hold at once and takes no window, got %q", b.Window)
	case b.Kind == Cumulative && b.Window == "":
		return invalid("a cumulative budget needs a window, such as now/M for this month")
	case b.Kind == Cumulative:
		if _, err := captainbudgets.WindowStart(b.Window, time.Now()); err != nil {
			return invalid("%v", err)
		}
	}
	return nil
}

func joined[T ~string](values []T) string {
	names := make([]string, len(values))
	for i, value := range values {
		names[i] = string(value)
	}
	return strings.Join(names, ", ")
}

// Store keeps the organizations' budgets.
type Store struct {
	DB *gorm.DB
}

// Set sets org's budget for b's resource and kind, replacing the one it had, as set by user, and
// returns it as stored.
func (s Store) Set(ctx context.Context, org uuid.UUID, user *uuid.UUID, b Budget) (Budget, error) {
	b.Window = strings.TrimSpace(b.Window)
	if err := b.validate(); err != nil {
		return Budget{}, fmt.Errorf("set organization %s's %s %s budget: %w", org, b.Kind, b.Resource, err)
	}
	var stored Budget
	if err := s.DB.WithContext(ctx).Raw(`INSERT INTO org_budgets (org_id, resource, kind, amount, "window", updated_by, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, now())
		ON CONFLICT (org_id, resource, kind) DO UPDATE SET amount = EXCLUDED.amount, "window" = EXCLUDED."window",
			updated_by = EXCLUDED.updated_by, updated_at = EXCLUDED.updated_at
		RETURNING resource, kind, amount, "window", updated_by, updated_at`,
		org, b.Resource, b.Kind, b.Amount, b.Window, user).Scan(&stored).Error; err != nil {
		return Budget{}, fmt.Errorf("set organization %s's %s %s budget: %w", org, b.Kind, b.Resource, err)
	}
	return stored, nil
}

// List is org's budgets by resource and kind.
func (s Store) List(ctx context.Context, org uuid.UUID) ([]Budget, error) {
	budgets := []Budget{}
	if err := s.DB.WithContext(ctx).Raw(`SELECT resource, kind, amount, "window", updated_by, updated_at FROM org_budgets
		WHERE org_id = ? ORDER BY resource, kind`, org).Scan(&budgets).Error; err != nil {
		return nil, fmt.Errorf("list organization %s's budgets: %w", org, err)
	}
	return budgets, nil
}

// Remove removes org's budget for resource and kind.
func (s Store) Remove(ctx context.Context, org uuid.UUID, resource Resource, kind Kind) error {
	result := s.DB.WithContext(ctx).Exec(`DELETE FROM org_budgets WHERE org_id = ? AND resource = ? AND kind = ?`, org, resource, kind)
	if result.Error != nil {
		return fmt.Errorf("remove organization %s's %s %s budget: %w", org, kind, resource, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: organization %s has no %s %s budget", ErrNotFound, org, kind, resource)
	}
	return nil
}

// Status is a budget and how much of it is used: in its window, from Since until now, for a
// cumulative budget, and held by running runs now for a concurrent one. Used is in the amount's
// unit.
type Status struct {
	Budget
	Used  float64    `json:"used"`
	Since *time.Time `json:"since,omitempty"`
}

// Reached reports whether the budget is used up.
func (s Status) Reached() bool { return s.Used >= s.Amount }

// Status is each of org's budgets with how much of it is used. runLifetime is the longest a run
// can last.
func (s Store) Status(ctx context.Context, org uuid.UUID, runLifetime time.Duration) ([]Status, error) {
	budgets, err := s.List(ctx, org)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	usages := map[string]platform.OrgUsage{}
	statuses := make([]Status, 0, len(budgets))
	for _, b := range budgets {
		since := now
		if b.Kind == Cumulative {
			if since, err = captainbudgets.WindowStart(b.Window, now); err != nil {
				return nil, fmt.Errorf("organization %s's %s %s budget: %w", org, b.Kind, b.Resource, err)
			}
		}
		usage, ok := usages[b.Window]
		if !ok {
			if usage, err = (platform.Store{DB: s.DB}).OrgUsage(ctx, org, since, runLifetime); err != nil {
				return nil, err
			}
			usages[b.Window] = usage
		}
		status := Status{Budget: b, Used: used(b, usage)}
		if b.Kind == Cumulative {
			status.Since = &since
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

// gib is a GiB in bytes.
const gib = 1 << 30

// used is how much of b usage holds, in b's unit.
func used(b Budget, usage platform.OrgUsage) float64 {
	switch {
	case b.Kind == Concurrent && b.Resource == CPU:
		return float64(usage.Compute.Running.CPUMillicores) / 1000
	case b.Kind == Concurrent && b.Resource == Memory:
		return float64(usage.Compute.Running.MemoryBytes) / gib
	case b.Resource == CPU:
		return usage.Compute.CoreSeconds / 3600
	case b.Resource == Memory:
		return usage.Compute.GiBSeconds / 3600
	case b.Resource == Tokens:
		return float64(usage.Tokens.Counted)
	default:
		return usage.CostUSD
	}
}
