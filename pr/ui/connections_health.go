// The connections page's health check for a hosted organization: each saved git connection's
// credentials, checked with its provider at public addresses, in the shape clicky-ui's dashboard reads.
package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	catalog "github.com/flanksource/gavel/connections"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	// maxHealthCheck is the most connections one health check names.
	maxHealthCheck = 200
	// healthCheckConcurrency is how many connections a health check checks at once.
	healthCheckConcurrency = 6
	// healthCheckBudget bounds a whole health check; a connection it outruns is "unknown".
	healthCheckBudget = 30 * time.Second
)

// connectionHealth is one connection's state as the dashboard shows it: healthy, credentials (its
// provider refused them, or it has none), unreachable, or unknown (the check outran its budget).
type connectionHealth struct {
	ID         string    `json:"id"`
	State      string    `json:"state"`
	Detail     string    `json:"detail"`
	CheckedAt  time.Time `json:"checkedAt"`
	DurationMS int64     `json:"durationMs"`
	Cached     bool      `json:"cached"`
}

// checkHostedConnectionsHealth checks the credentials of each of the organization's connections the
// request names, as saved and never hydrated. Any member may: it shows only each one's state.
func (s *Server) checkHostedConnectionsHealth(w http.ResponseWriter, r *http.Request) {
	db, scope, ok := s.connectionCatalogAccess(w, r, false)
	if !ok {
		return
	}
	var request struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&request); err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("the health check must be a JSON object: %w", err))
		return
	}
	switch {
	case len(request.IDs) == 0:
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("name the connections to check in ids"))
		return
	case len(request.IDs) > maxHealthCheck:
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("a health check names at most %d connections, not %d", maxHealthCheck, len(request.IDs)))
		return
	}
	store, closeStore, err := catalog.OpenStore(r.Context(), catalog.Options{DB: db, Scope: scope})
	if err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, fmt.Errorf("open the organization's connection catalog: %w", err))
		return
	}
	connections := make([]models.Connection, 0, len(request.IDs))
	var unknown []string
	for _, id := range request.IDs {
		if _, err := uuid.Parse(id); err != nil {
			unknown = append(unknown, id)
			continue
		}
		connection, err := store.ResolveConnection(dbcontext.NewContext(r.Context()), id)
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			unknown = append(unknown, id)
		case err != nil:
			writeJSONError(w, http.StatusServiceUnavailable, errors.Join(fmt.Errorf("read connection %s: %w", id, err), closeStore()))
			return
		default:
			connections = append(connections, *connection)
		}
	}
	if err := closeStore(); err != nil {
		writeJSONError(w, http.StatusServiceUnavailable, err)
		return
	}
	if len(unknown) > 0 {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("unknown connection ids: %s", strings.Join(unknown, ", ")))
		return
	}
	writeJSON(w, map[string]any{"results": s.connectionsHealth(r.Context(), connections), "generatedAt": time.Now().UTC()})
}

// connectionsHealth checks connections a few at a time within one budget; one failing never stops
// the rest.
func (s *Server) connectionsHealth(ctx context.Context, connections []models.Connection) []connectionHealth {
	ctx, cancel := context.WithTimeout(ctx, healthCheckBudget)
	defer cancel()
	results := make([]connectionHealth, len(connections))
	slots := make(chan struct{}, healthCheckConcurrency)
	var checks sync.WaitGroup
	for i, connection := range connections {
		results[i] = connectionHealth{ID: connection.ID.String(), State: "unknown", Detail: "the health check ran out of time before checking this connection"}
		checks.Go(func() {
			select {
			case slots <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-slots }()
			started := time.Now()
			configuration, err := s.verifyConnection(ctx, connection)
			if ctx.Err() != nil {
				return
			}
			results[i] = connectionHealth{ID: connection.ID.String(), State: healthState(configuration, err), Detail: verifiedMessage(connection.Type, err),
				CheckedAt: started.UTC(), DurationMS: time.Since(started).Milliseconds()}
		})
	}
	checks.Wait()
	return results
}

// healthState is the dashboard state of a check that ended with err: credentials when the
// connection made no client or its provider refused its token, else unreachable.
func healthState(configuration bool, err error) string {
	var refused *git.APIError
	switch {
	case err == nil:
		return "healthy"
	case configuration, errors.As(err, &refused) && (refused.StatusCode == http.StatusUnauthorized || refused.StatusCode == http.StatusForbidden):
		return "credentials"
	}
	return "unreachable"
}
