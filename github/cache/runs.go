package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/blob"
	"github.com/flanksource/gavel/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
)

// GetCompletedRunPayload returns the JSON-encoded payload for a completed
// workflow run, or nil if not cached. Callers unmarshal into their own
// run struct so this package stays free of github-package types.
func (s *Store) GetCompletedRunPayload(runID int64) []byte {
	if s == nil || s.disabled {
		return nil
	}
	var entry WorkflowRunCache
	res := s.gorm().Where("org_id = ? AND run_id = ?", s.org, runID).First(&entry)
	if res.Error != nil {
		return nil
	}
	if entry.Status != "completed" {
		// Defensive: we never write non-completed runs, but skip if data
		// drifted somehow.
		return nil
	}
	return entry.Payload
}

// PutCompletedRun stores a workflow run, but only if status == "completed".
// In-progress runs are intentionally left to the ETag-aware HTTP cache.
func (s *Store) PutCompletedRun(repo string, runID int64, status, conclusion string, payload []byte) {
	if s == nil || s.disabled {
		return
	}
	if status != "completed" {
		return
	}
	entry := WorkflowRunCache{
		OrgID:      s.org,
		RunID:      runID,
		Repo:       repo,
		Status:     status,
		Conclusion: conclusion,
		Payload:    payload,
		FetchedAt:  time.Now(),
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	res := s.gorm().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "org_id"}, {Name: "run_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"repo", "status", "conclusion", "payload", "fetched_at"}),
	}).Create(&entry)
	if res.Error != nil {
		logger.Warnf("github cache: PutCompletedRun(%d) failed: %v", runID, res.Error)
	}
}

// jobLogs is the Store organization's part of the blob store, or nil when the Store has none.
func (s *Store) jobLogs() blob.Store {
	return s.OrgBlobs(s.org)
}

// OrgBlobs is org's part of the blob store, which holds its job logs: orgs/<org>/ of the
// configured store, or of <user cache dir>/gavel/blobs. It is nil when the Store has no blob store.
func (s *Store) OrgBlobs(org uuid.UUID) blob.Store {
	if s == nil || s.blobs == nil {
		return nil
	}
	return database.OrgBlobs(s.blobs, org)
}

// GetJobLogs returns cached logs for a job, or ("", false) on a miss.
func (s *Store) GetJobLogs(jobID int64) (string, bool) {
	if s == nil || s.disabled {
		return "", false
	}
	store := s.jobLogs()
	if store == nil {
		return "", false
	}
	var entry JobLogCache
	if err := s.gorm().Where("org_id = ? AND job_id = ?", s.org, jobID).First(&entry).Error; err != nil {
		return "", false
	}
	gz, err := readBlob(store, entry.LogKey)
	if err != nil {
		logger.Warnf("github cache: read logs for job %d: %v", jobID, err)
		return "", false
	}
	logs, err := gunzipString(gz)
	if err != nil {
		logger.Warnf("github cache: corrupt logs for job %d: %v", jobID, err)
		return "", false
	}
	return logs, true
}

// PutJobLogs stores gzip-compressed logs for a terminal job. Callers should
// only call this once the job has reached a final conclusion ("success",
// "failure", "cancelled", etc.) — in-progress jobs would just churn the
// cache.
func (s *Store) PutJobLogs(jobID int64, repo, logs string) {
	if s == nil || s.disabled {
		return
	}
	store := s.jobLogs()
	if store == nil {
		return
	}
	gz, err := gzipString(logs)
	if err != nil {
		logger.Warnf("github cache: gzip job %d logs failed: %v", jobID, err)
		return
	}
	key := database.JobLogKey(jobID)
	if err := store.Write(context.Background(), key, bytes.NewReader(gz)); err != nil {
		logger.Warnf("github cache: PutJobLogs(%d) failed: %v", jobID, err)
		return
	}
	entry := JobLogCache{
		OrgID:     s.org,
		JobID:     jobID,
		Repo:      repo,
		LogKey:    key,
		FetchedAt: time.Now(),
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	res := s.gorm().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "org_id"}, {Name: "job_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"repo", "log_key", "fetched_at"}),
	}).Create(&entry)
	if res.Error != nil {
		logger.Warnf("github cache: PutJobLogs(%d) failed: %v", jobID, res.Error)
	}
}

func readBlob(store blob.Store, key string) ([]byte, error) {
	reader, err := store.Read(context.Background(), key)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// GetWorkflowDef returns the cached YAML and path for a workflow definition
// at a specific commit SHA. The (repo, workflowID, sha) tuple is immutable.
func (s *Store) GetWorkflowDef(repo string, workflowID int64, sha string) (yaml, path string, ok bool) {
	if s == nil || s.disabled {
		return "", "", false
	}
	var entry WorkflowDefCache
	if err := s.gorm().
		Where("org_id = ? AND repo = ? AND workflow_id = ? AND sha = ?", s.org, repo, workflowID, sha).
		First(&entry).Error; err != nil {
		return "", "", false
	}
	return entry.YAML, entry.Path, true
}

// PutWorkflowDef stores a workflow definition keyed by SHA. Definitions are
// immutable per SHA so the cache never expires for this table.
func (s *Store) PutWorkflowDef(repo string, workflowID int64, sha, path, yaml string) {
	if s == nil || s.disabled {
		return
	}
	if sha == "" {
		// Without a SHA the cache key isn't stable; skip rather than risk
		// poisoning the cache.
		return
	}
	entry := WorkflowDefCache{
		OrgID:      s.org,
		Repo:       repo,
		WorkflowID: workflowID,
		SHA:        sha,
		Path:       path,
		YAML:       yaml,
		FetchedAt:  time.Now(),
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	res := s.gorm().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "org_id"}, {Name: "repo"}, {Name: "workflow_id"}, {Name: "sha"}},
		DoUpdates: clause.AssignmentColumns([]string{"path", "yaml", "fetched_at"}),
	}).Create(&entry)
	if res.Error != nil {
		logger.Warnf("github cache: PutWorkflowDef(%s/%d/%s) failed: %v", repo, workflowID, sha, res.Error)
	}
}

// MarshalJSON is a small helper so callers don't have to import encoding/json
// when they just want to round-trip a struct through PutCompletedRun.
func MarshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

func gzipString(s string) ([]byte, error) {
	if s == "" {
		return nil, nil
	}
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := io.WriteString(w, s); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func gunzipString(b []byte) (string, error) {
	if len(b) == 0 {
		return "", nil
	}
	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
