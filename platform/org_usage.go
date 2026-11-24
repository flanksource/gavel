// Measures what one organization used in a window, for its usage page and its budgets: the CPU and
// memory its substrate runs were sized at for as long as they ran, and its model calls' tokens and cost.
package platform

import (
	"context"
	"fmt"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
)

// OrgUsage is what an organization used from Since until now.
type OrgUsage struct {
	Since   time.Time `json:"since"`
	Compute Compute   `json:"compute"`
	Tokens  Tokens    `json:"tokens"`
	// CostUSD is its model calls' cost: what the provider reported, else the list price.
	CostUSD float64 `json:"costUSD"`
	Calls   int     `json:"calls"`
	// OtherCurrencyCalls are priced in another currency: their tokens count, their cost doesn't,
	// since there's no rate to convert it to USD.
	OtherCurrencyCalls int `json:"otherCurrencyCalls"`
}

// Compute is the CPU and memory an organization's runs were sized at, times how long each ran in
// the window: what substrate held for them, whatever they used of it. A run left running when serve
// stopped counts only as long as a run can live.
type Compute struct {
	CoreSeconds float64 `json:"coreSeconds"`
	GiBSeconds  float64 `json:"gibSeconds"`
	// Runs are the sized runs that ran in the window.
	Runs int `json:"runs"`
	// Running are its runs running now, and their size together.
	Running Running `json:"running"`
}

// Running is the organization's runs running now and their size together.
type Running struct {
	Runs          int   `json:"runs"`
	CPUMillicores int64 `json:"cpuMillicores"`
	MemoryBytes   int64 `json:"memoryBytes"`
}

// Tokens are an organization's model calls' tokens by bucket; the buckets don't overlap. Counted
// is what a token budget counts: input, cache writes, output and reasoning, leaving out cache
// reads, as the providers' input limits do.
type Tokens struct {
	Input      int64 `json:"input"`
	Output     int64 `json:"output"`
	Reasoning  int64 `json:"reasoning"`
	CacheRead  int64 `json:"cacheRead"`
	CacheWrite int64 `json:"cacheWrite"`
	Counted    int64 `json:"counted"`
}

// modelCallOwners joins each of Captain's model calls, aliased calls, to the organizations that
// own its session (own) and its root session (root); modelCallOwner is the one it belongs to.
const (
	modelCallOwners = `captain_model_calls calls
		JOIN captain_turns turns ON turns.id = calls.turn_id
		JOIN captain_sessions sessions ON sessions.id = turns.session_id
		LEFT JOIN session_owners own ON own.session_id = sessions.id
		LEFT JOIN session_owners root ON root.session_id = sessions.root_session_id`
	modelCallOwner = `COALESCE(own.org_id, root.org_id)`
)

// OrgUsage is what org used from since until now. runLifetime is the longest a run can last.
func (s Store) OrgUsage(ctx context.Context, org uuid.UUID, since time.Time, runLifetime time.Duration) (OrgUsage, error) {
	usage := OrgUsage{Since: since}
	db := s.DB.WithContext(ctx)
	var compute struct {
		CoreSeconds float64 `gorm:"column:core_seconds"`
		GiBSeconds  float64 `gorm:"column:gib_seconds"`
		Runs        int     `gorm:"column:runs"`
	}
	if err := db.Raw(`
		WITH spans AS (
			SELECT cpu_millicores, memory_bytes, GREATEST(started_at, ?) AS span_start,
				LEAST(COALESCE(finished_at, now()), started_at + ? * interval '1 second') AS span_end
			FROM sandbox_runs
			WHERE org_id = ? AND cpu_millicores IS NOT NULL AND memory_bytes IS NOT NULL AND started_at IS NOT NULL
		)
		SELECT COALESCE(sum(cpu_millicores / 1000.0 * EXTRACT(EPOCH FROM span_end - span_start)), 0) AS core_seconds,
			COALESCE(sum(memory_bytes / 1073741824.0 * EXTRACT(EPOCH FROM span_end - span_start)), 0) AS gib_seconds,
			count(*) AS runs
		FROM spans WHERE span_end > span_start`, since, runLifetime.Seconds(), org).Scan(&compute).Error; err != nil {
		return OrgUsage{}, fmt.Errorf("read the CPU and memory organization %s's runs used: %w", org, err)
	}
	usage.Compute.CoreSeconds, usage.Compute.GiBSeconds, usage.Compute.Runs = compute.CoreSeconds, compute.GiBSeconds, compute.Runs
	if err := db.Raw(`
		SELECT count(*) AS runs, COALESCE(sum(cpu_millicores), 0) AS cpu_millicores, COALESCE(sum(memory_bytes), 0) AS memory_bytes
		FROM sandbox_runs
		WHERE org_id = ? AND status = 'running' AND finished_at IS NULL AND cpu_millicores IS NOT NULL
			AND started_at > now() - ? * interval '1 second'`, org, runLifetime.Seconds()).Scan(&usage.Compute.Running).Error; err != nil {
		return OrgUsage{}, fmt.Errorf("read organization %s's running runs: %w", org, err)
	}
	var calls struct {
		Tokens
		CostUSD            float64
		Calls              int
		OtherCurrencyCalls int
	}
	if err := db.Raw(`
		SELECT COALESCE(sum(calls.input_tokens), 0) AS input, COALESCE(sum(calls.output_tokens), 0) AS output,
			COALESCE(sum(calls.reasoning_tokens), 0) AS reasoning, COALESCE(sum(calls.cache_read_tokens), 0) AS cache_read,
			COALESCE(sum(calls.cache_write_tokens), 0) AS cache_write,
			COALESCE(sum(`+captaindb.ModelCallCostUSD+`), 0) AS cost_usd, count(*) AS calls,
			count(*) FILTER (WHERE upper(calls.currency) <> 'USD') AS other_currency_calls
		FROM `+modelCallOwners+`
		WHERE `+modelCallOwner+` = ? AND calls.created_at >= ?`, org, since).Scan(&calls).Error; err != nil {
		return OrgUsage{}, fmt.Errorf("read organization %s's model calls: %w", org, err)
	}
	usage.Tokens = calls.Tokens
	usage.Tokens.Counted = calls.Input + calls.CacheWrite + calls.Output + calls.Reasoning
	usage.CostUSD, usage.Calls, usage.OtherCurrencyCalls = calls.CostUSD, calls.Calls, calls.OtherCurrencyCalls
	return usage, nil
}
