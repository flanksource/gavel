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
	"regexp"
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

// CIResults serves POST /api/ci/results, and PUT /api/ci/results/{run} that replaces a run a
// job streams while it is in progress, authenticated by a GitHub Actions OIDC token whose
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
	mux.HandleFunc("PUT /api/ci/results/{run}", c.handleUpload)
	return mux
}

// ciRunID matches the run IDs ingestRunSnapshot allocates, the only runs a PUT may replace.
var ciRunID = regexp.MustCompile(`^run-\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z(-\d+)?$`)

func (c *CIResults) handleUpload(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("run")
	if r.Method == http.MethodPut && !ciRunID.MatchString(runID) {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("%q is not a gavel run ID", runID))
		return
	}
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
	storeResults(w, r, c.Cache, repo.OrgID, repo.ID, claims.SHA, claims.Repository, runID)
}

// storeResults reads a snapshot of up to 50 MB from r and stores it as a run of repoID in org,
// answering 201 with its run ID. With runID it replaces that run, or creates it, answering 200.
func storeResults(w http.ResponseWriter, r *http.Request, store *cache.Store, org, repoID uuid.UUID, sha, what, runID string) {
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
	status := http.StatusCreated
	if runID != "" {
		status = http.StatusOK
	}
	runID, err = ingestRunSnapshot(r.Context(), store, org, repoID, sha, data, snap, runID)
	if err != nil {
		logger.Warnf("store results for %s: %v", what, err)
		writeCIError(w, http.StatusInternalServerError, "could not store the results")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
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
// indexes it in test_run_caches. sha fills in the commit when the snapshot names none. A
// non-empty runID names the run to replace; otherwise a free one is allocated. A replacement
// without a log tail keeps the replaced run's: the action's final upload carries none, and would
// otherwise erase the diagnostics the job streamed. It returns the run ID.
func ingestRunSnapshot(ctx context.Context, store *cache.Store, org, repoID uuid.UUID, sha string, data []byte, snap testui.Snapshot, runID string) (string, error) {
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
	replacing := runID != ""
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
	if replacing && snap.LogTail == "" {
		kept, err := keepLogTail(ctx, runs, key, data)
		if err != nil {
			return "", err
		}
		data = kept
	}
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

// keepLogTail returns data, a snapshot without a log tail, with the log tail of the run stored at
// key, unchanged when there is no such run or it has none.
func keepLogTail(ctx context.Context, runs blob.Store, key string, data []byte) ([]byte, error) {
	reader, err := runs.Read(ctx, key)
	if errors.Is(err, blob.ErrNotFound) {
		return data, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read the replaced run: %w", err)
	}
	defer reader.Close() //nolint:errcheck
	var stored struct {
		LogTail string `json:"log_tail"`
	}
	if err := json.NewDecoder(reader).Decode(&stored); err != nil || stored.LogTail == "" {
		return data, nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, fmt.Errorf("decode the results: %w", err)
	}
	if fields["log_tail"], err = json.Marshal(stored.LogTail); err != nil {
		return nil, err
	}
	return json.Marshal(fields)
}
