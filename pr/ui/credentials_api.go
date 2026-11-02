// Serves /api/credentials/fetch to the credential provider in substrate's cluster, which asks
// for an organization's model key on behalf of substrate's egress gateway and a run's environment.
package ui

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/modelkeys"
	"github.com/flanksource/gavel/sandbox"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CredentialsAPI answers the credential provider. It is mounted outside the Clerk middleware.
type CredentialsAPI struct {
	DB   *gorm.DB
	Keys modelkeys.Store
	// Atespace is where substrate creates run environments.
	Atespace string
	// RunLifetime is how long after it was created a run may still use its keys.
	RunLifetime time.Duration
	// Token is the credential provider's bearer token.
	Token string
}

// Handler routes the credential fetch.
func (a *CredentialsAPI) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/credentials/fetch", a.handleFetch)
	return mux
}

// handleFetch returns the key of the run's organization for provider, only while the run is
// active and recent, and only to the run's own environment.
func (a *CredentialsAPI) handleFetch(w http.ResponseWriter, r *http.Request) {
	raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || raw == "" {
		writeCIError(w, http.StatusUnauthorized, "send the credential provider's token as Authorization: Bearer")
		return
	}
	if subtle.ConstantTimeCompare([]byte(raw), []byte(a.Token)) != 1 {
		writeCIError(w, http.StatusUnauthorized, "this is not the credential provider's token")
		return
	}
	var body struct {
		Run      uuid.UUID `json:"run"`
		Provider string    `json:"provider"`
		Actor    string    `json:"actor"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&body); err != nil {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("invalid JSON: %v", err))
		return
	}
	refuse := func(status int, message string) {
		logger.Infof("credential fetch for run %s's %s key refused: %s", body.Run, body.Provider, message)
		writeCIError(w, status, message)
	}
	run, err := sandbox.LoadRun(r.Context(), a.DB, body.Run)
	if errors.Is(err, sandbox.ErrRunNotFound) {
		refuse(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		logger.Warnf("credential fetch: load sandbox run %s: %v", body.Run, err)
		writeCIError(w, http.StatusInternalServerError, "could not read the run")
		return
	}
	switch {
	case run.Status != "queued" && run.Status != "running":
		refuse(http.StatusForbidden, fmt.Sprintf("run %s is %s, not active", run.ID, run.Status))
		return
	case time.Since(run.CreatedAt) > a.RunLifetime:
		refuse(http.StatusForbidden, fmt.Sprintf("run %s started over %s ago, longer than a run lasts", run.ID, a.RunLifetime))
		return
	case body.Actor != sandbox.SubstrateActor(a.Atespace, run.ID):
		refuse(http.StatusForbidden, fmt.Sprintf("actor %q isn't run %s's environment", body.Actor, run.ID))
		return
	}
	value, err := a.Keys.Value(r.Context(), run.Org, body.Provider)
	if errors.Is(err, modelkeys.ErrNotFound) {
		refuse(http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		logger.Errorf("credential fetch for run %s: %v", run.ID, err)
		writeCIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	logger.Infof("credential fetch for run %s: organization %s's %s key", run.ID, run.Org, body.Provider)
	writeJSON(w, map[string]string{"value": value})
}
