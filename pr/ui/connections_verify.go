// Checks a hosted organization's git connections for the connections page: Test checks the form's
// credentials, and the health check each saved connection's, with its provider at public addresses.
package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/git"
	"github.com/flanksource/commons-db/models"
	catalog "github.com/flanksource/gavel/connections"
	"github.com/flanksource/gavel/tenant"
	"gorm.io/gorm"
)

// connectionTestTimeout bounds checking a connection's credentials with its provider.
const connectionTestTimeout = 15 * time.Second

// connectionCheck is what the connections page's Test shows: whether the provider accepted the
// connection's credentials, and why not.
type connectionCheck struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// testHostedConnection checks the credentials of the git connection the request's form holds with
// its provider, as the form has them and never hydrated. An edit form leaves a saved connection's
// token blank, so the organization's saved token stands in for it.
func (s *Server) testHostedConnection(w http.ResponseWriter, r *http.Request) {
	db, scope, ok := s.connectionCatalogAccess(w, r, true)
	if !ok {
		return
	}
	var body map[string]any
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil || body == nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Errorf("the connection to test must be a JSON object: %v", err))
		return
	}
	if err := hostedConnectionType(body["type"], true); err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	saved, _ := body["id"].(string)
	delete(body, "id")
	draft, err := connectionDraft(body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err)
		return
	}
	if saved != "" && draft.Password == "" {
		if err := savedConnectionToken(r.Context(), db, scope, saved, &draft); errors.Is(err, gorm.ErrRecordNotFound) {
			writeJSONError(w, http.StatusNotFound, err)
			return
		} else if err != nil {
			writeJSONError(w, http.StatusServiceUnavailable, err)
			return
		}
	}
	_, err = s.verifyConnection(r.Context(), draft)
	writeJSON(w, connectionCheck{OK: err == nil, Message: verifiedMessage(draft.Type, err)})
}

// verifyConnection checks connection's credentials with its provider, through a client that
// reaches only public addresses unless the server was given other options. configuration is true
// when the connection can't make a client at all, such as one with no token.
func (s *Server) verifyConnection(ctx context.Context, connection models.Connection) (configuration bool, err error) {
	options := git.ClientOptions{Defensive: true}
	if s.connectionClientOptions != nil {
		options = *s.connectionClientOptions
	}
	ctx, cancel := context.WithTimeout(ctx, connectionTestTimeout)
	defer cancel()
	client, err := git.FromConnection(dbcontext.NewContext(ctx), connection, options)
	if err != nil {
		return true, err
	}
	return false, client.Provider.Verify(ctx)
}

// verifiedMessage says whether a connection of kind passed its check, and if not, why.
func verifiedMessage(kind string, err error) string {
	if err != nil {
		return fmt.Sprintf("%s didn't confirm the connection's credentials: %v", kind, err)
	}
	return kind + " accepted the connection's credentials"
}

// connectionDraft is the connection body holds.
func connectionDraft(body map[string]any) (models.Connection, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return models.Connection{}, fmt.Errorf("encode the connection to test: %w", err)
	}
	var draft models.Connection
	if err := json.Unmarshal(data, &draft); err != nil {
		return models.Connection{}, fmt.Errorf("the connection to test is not a connection: %w", err)
	}
	return draft, nil
}

// savedConnectionToken fills draft's blank token and certificate from scope's saved connection
// reference.
func savedConnectionToken(ctx context.Context, db *gorm.DB, scope tenant.Scope, reference string, draft *models.Connection) error {
	store, closeStore, err := catalog.OpenStore(ctx, catalog.Options{DB: db, Scope: scope})
	if err != nil {
		return fmt.Errorf("open the organization's connection catalog: %w", err)
	}
	stored, err := store.ResolveConnection(dbcontext.NewContext(ctx), reference)
	if err := errors.Join(err, closeStore()); err != nil {
		return fmt.Errorf("read the saved connection to test: %w", err)
	}
	draft.Password = stored.Password
	if draft.Certificate == "" {
		draft.Certificate = stored.Certificate
	}
	return nil
}
