// Summarizes every organization's model calls from Captain's ledger for the platform dashboard:
// tokens and USD cost per day, model and provider, attributed through each call's session.
package platform

import (
	"context"
	"fmt"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
)

// ModelUsageDay is one UTC day's model calls of an organization with one model and provider. Org
// is nil for calls no organization of the Clerk instance owns: sessions without an owner, or owned
// by the local organization or another instance's.
type ModelUsageDay struct {
	Day              string     `json:"day"` // YYYY-MM-DD
	Org              *uuid.UUID `json:"org,omitempty"`
	OrgName          string     `json:"orgName,omitempty"`
	Model            string     `json:"model"`
	Provider         string     `json:"provider"`
	Calls            int        `json:"calls"`
	InputTokens      int64      `json:"inputTokens"`
	OutputTokens     int64      `json:"outputTokens"`
	ReasoningTokens  int64      `json:"reasoningTokens"`
	CacheReadTokens  int64      `json:"cacheReadTokens"`
	CacheWriteTokens int64      `json:"cacheWriteTokens"`
	CostUSD          float64    `json:"costUSD"`
	// OtherCurrencyCalls are priced in another currency: their tokens count, their cost doesn't,
	// since there's no rate to convert it to USD.
	OtherCurrencyCalls int `json:"otherCurrencyCalls"`
}

// ModelUsage summarizes the model calls made since since. A session's calls belong to the
// session's owner, or to its root session's owner when it has none of its own.
func (s Store) ModelUsage(ctx context.Context, since time.Time) ([]ModelUsageDay, error) {
	days := []ModelUsageDay{}
	if err := s.DB.WithContext(ctx).Raw(`
		SELECT to_char(calls.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS day, o.id AS org, o.name AS org_name,
			calls.model, COALESCE(calls.provider, '') AS provider, count(*) AS calls,
			sum(calls.input_tokens) AS input_tokens, sum(calls.output_tokens) AS output_tokens,
			sum(calls.reasoning_tokens) AS reasoning_tokens, sum(calls.cache_read_tokens) AS cache_read_tokens,
			sum(calls.cache_write_tokens) AS cache_write_tokens,
			COALESCE(sum(`+captaindb.ModelCallCostUSD+`), 0) AS cost_usd,
			count(*) FILTER (WHERE upper(calls.currency) <> 'USD') AS other_currency_calls
		FROM captain_model_calls calls
		JOIN captain_turns turns ON turns.id = calls.turn_id
		JOIN captain_sessions sessions ON sessions.id = turns.session_id
		LEFT JOIN session_owners own ON own.session_id = sessions.id
		LEFT JOIN session_owners root ON root.session_id = sessions.root_session_id
		LEFT JOIN orgs o ON o.id = COALESCE(own.org_id, root.org_id) AND o.issuer = ?
		WHERE calls.created_at >= ?
		GROUP BY 1, 2, 3, 4, 5 ORDER BY 1, 3 NULLS LAST, 4, 5`, s.Issuer, since).Scan(&days).Error; err != nil {
		return nil, fmt.Errorf("read the platform's model usage: %w", err)
	}
	return days, nil
}
