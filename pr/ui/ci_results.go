// Accepts test and lint results that GitHub Actions, GitLab CI and Azure Pipelines jobs upload with
// their own credentials, and stores them for every organization the repository is linked to.
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

	"github.com/flanksource/commons-db/git"
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
// job streams while it is in progress, authenticated by the job's own credential: a GitHub Actions
// OIDC token, a GitLab CI ID token, or an Azure Pipelines job's access token for its build.
type CIResults struct {
	DB   *gorm.DB
	OIDC api.GitHubOIDCVerifier
	// GitLab verifies ID tokens from gitlab.com and the GitLabs organizations sync repositories from.
	GitLab api.GitLabOIDCVerifier
	// ClientOptions configures the Azure DevOps client an Azure job's build is read with.
	ClientOptions git.ClientOptions
	Cache         *cache.Store
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
		writeCIError(w, http.StatusUnauthorized, "send the job's GitHub Actions OIDC token, GitLab CI ID token or Azure Pipelines System.AccessToken as Authorization: Bearer")
		return
	}
	job, err := c.identify(r, raw)
	var refused ciRefusal
	if errors.As(err, &refused) {
		writeCIError(w, refused.status, refused.message)
		return
	}
	if err != nil {
		writeCIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	storeResults(w, r, c.Cache, job.holders, job.sha, job.what, runID)
}

// resultsRepo is a repository a run's results are stored in: its ID and its organization's.
type resultsRepo struct {
	ID    uuid.UUID
	OrgID uuid.UUID
}

// storeResults reads a snapshot of up to 50 MB from r and stores it as a run of each of repos,
// answering 201 with its run ID: the first repository allocates it and the rest take the same one.
// With runID it replaces that run, or creates it, answering 200.
func storeResults(w http.ResponseWriter, r *http.Request, store *cache.Store, repos []resultsRepo, sha, what, runID string) {
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
	for _, repo := range repos {
		if runID, err = ingestRunSnapshot(r.Context(), store, repo.OrgID, repo.ID, sha, data, snap, runID); err != nil {
			logger.Warnf("store results for %s in org %s: %v", what, repo.OrgID, err)
			writeCIError(w, http.StatusInternalServerError, "could not store the results")
			return
		}
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
