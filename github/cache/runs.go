package cache

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"time"

	"github.com/flanksource/commons/logger"
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
	res := s.gorm().Where("run_id = ?", runID).First(&entry)
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
		Columns:   []clause.Column{{Name: "run_id"}},
		UpdateAll: true,
	}).Create(&entry)
	if res.Error != nil {
		logger.Warnf("github cache: PutCompletedRun(%d) failed: %v", runID, res.Error)
	}
}

// GetJobLogs returns cached logs for a job, or ("", false) on a miss.
func (s *Store) GetJobLogs(jobID int64) (string, bool) {
	if s == nil || s.disabled {
		return "", false
	}
	var entry JobLogCache
	if err := s.gorm().Where("job_id = ?", jobID).First(&entry).Error; err != nil {
		return "", false
	}
	logs, err := gunzipString(entry.LogsGz)
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
	gz, err := gzipString(logs)
	if err != nil {
		logger.Warnf("github cache: gzip job %d logs failed: %v", jobID, err)
		return
	}
	entry := JobLogCache{
		JobID:     jobID,
		Repo:      repo,
		LogsGz:    gz,
		FetchedAt: time.Now(),
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	res := s.gorm().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "job_id"}},
		UpdateAll: true,
	}).Create(&entry)
	if res.Error != nil {
		logger.Warnf("github cache: PutJobLogs(%d) failed: %v", jobID, res.Error)
	}
}

// GetWorkflowDef returns the cached YAML and path for a workflow definition
// at a specific commit SHA. The (repo, workflowID, sha) tuple is immutable.
func (s *Store) GetWorkflowDef(repo string, workflowID int64, sha string) (yaml, path string, ok bool) {
	if s == nil || s.disabled {
		return "", "", false
	}
	var entry WorkflowDefCache
	if err := s.gorm().
		Where("repo = ? AND workflow_id = ? AND sha = ?", repo, workflowID, sha).
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
		Columns:   []clause.Column{{Name: "repo"}, {Name: "workflow_id"}, {Name: "sha"}},
		UpdateAll: true,
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
