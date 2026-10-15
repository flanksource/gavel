// Accepts test and lint results that GitHub Actions jobs upload with their OIDC token, and stores
// them for the organization the repository is linked to.
package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/snapshots"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// maxCIResultsBytes bounds one uploaded snapshot.
const maxCIResultsBytes = 50 << 20

// CIResults serves POST /api/ci/results, authenticated by a GitHub Actions OIDC token whose
// repository is linked to an organization.
type CIResults struct {
	DB    *gorm.DB
	OIDC  api.GitHubOIDCVerifier
	Cache *cache.Store
}

// Handler routes the CI results endpoint.
func (c *CIResults) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/ci/results", c.handleUpload)
	return mux
}

func (c *CIResults) handleUpload(w http.ResponseWriter, r *http.Request) {
	raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || raw == "" {
		writeCIError(w, http.StatusUnauthorized, "send the job's GitHub Actions OIDC token as Authorization: Bearer")
		return
	}
	claims, err := c.OIDC.Verify(r.Context(), raw)
	if err != nil {
		writeCIError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var repo struct {
		ID    uuid.UUID
		OrgID uuid.UUID
	}
	result := c.DB.WithContext(r.Context()).Raw(`SELECT id, org_id FROM repos WHERE github_repo_id = ?`, claims.RepositoryID).Scan(&repo)
	if result.Error != nil {
		logger.Warnf("look up repository %d for CI results: %v", claims.RepositoryID, result.Error)
		writeCIError(w, http.StatusInternalServerError, "could not look up the repository")
		return
	}
	if result.RowsAffected == 0 {
		writeCIError(w, http.StatusForbidden, fmt.Sprintf("%s (%d) is not linked to a gavel organization; install the gavel GitHub App on it", claims.Repository, claims.RepositoryID))
		return
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxCIResultsBytes))
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		writeCIError(w, http.StatusRequestEntityTooLarge, "results are limited to 50 MB")
		return
	}
	if err != nil {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("read the results: %v", err))
		return
	}
	var snap testui.Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("decode the results as a gavel snapshot: %v", err))
		return
	}
	runID, err := ingestRunSnapshot(r.Context(), c.Cache, repo.OrgID, repo.ID, claims.SHA, data, snap)
	if err != nil {
		logger.Warnf("store CI results for %s: %v", claims.Repository, err)
		writeCIError(w, http.StatusInternalServerError, "could not store the results")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"runId": runID})
}

func writeCIError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// RepoRunsPrefix is where a repository's uploaded runs live in its organization's blob store.
func RepoRunsPrefix(repoID uuid.UUID) string {
	return "repos/" + repoID.String() + "/runs/"
}

// RepoWorkspace is the test_run_caches workspace of a repository's uploaded runs.
func RepoWorkspace(repoID uuid.UUID) string {
	return "repo:" + repoID.String()
}

// ingestRunSnapshot stores data, the raw snapshot snap, as a run of repoID in org's blob store and
// indexes it in test_run_caches. sha fills in the commit when the snapshot names none. It
// returns the run ID.
func ingestRunSnapshot(ctx context.Context, store *cache.Store, org, repoID uuid.UUID, sha string, data []byte, snap testui.Snapshot) (string, error) {
	blobs := store.OrgBlobs(org)
	if blobs == nil {
		return "", errors.New("no blob store for run results")
	}
	runs := blob.WithPrefix(blobs, RepoRunsPrefix(repoID))
	started := time.Now().UTC()
	if snap.Metadata != nil && !snap.Metadata.Started.IsZero() {
		started = snap.Metadata.Started.UTC()
	}
	stem := snapshots.PerRunPrefix + started.Format(snapshots.PerRunTimestampLayout)
	runID := ""
	for i := 1; i < 1000 && runID == ""; i++ {
		candidate := stem
		if i > 1 {
			candidate = fmt.Sprintf("%s-%d", stem, i)
		}
		exists, err := runs.Exists(ctx, candidate+".json")
		if err != nil {
			return "", err
		}
		if !exists {
			runID = candidate
		}
	}
	if runID == "" {
		return "", fmt.Errorf("no free run id for %s", stem)
	}
	key := runID + ".json"
	if err := runs.Write(ctx, key, bytes.NewReader(data)); err != nil {
		return "", err
	}
	info := snapshots.Summarize(snap, key, RepoRunsPrefix(repoID)+key, started)
	if info.SHA == "" {
		info.SHA = sha
	}
	row := testRunRow(RepoWorkspace(repoID), info)
	if err := store.ForOrg(org).UpsertTestRuns(ctx, RepoWorkspace(repoID), []cache.TestRunCache{row}, row.StartedTS); err != nil {
		return "", fmt.Errorf("index run %s: %w", runID, err)
	}
	return runID, nil
}
