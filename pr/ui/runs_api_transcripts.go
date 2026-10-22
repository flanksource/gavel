// An agent run's transcripts: the job uploads its agent's transcript files as they grow, and
// serve keeps them under its data directory and ingests them into Captain's store.
package ui

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/monitor"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/sandbox"
)

// TranscriptsDir is where serve keeps the transcripts agent runs upload, under dataDir.
func TranscriptsDir(dataDir string) string {
	return filepath.Join(dataDir, "transcripts")
}

// maxTranscriptBytes bounds one uploaded transcript file.
const maxTranscriptBytes = 256 << 20

// handleTranscript stores the uploaded transcript file at
// <data dir>/transcripts/<org>/<run>/<source>/<name>, replacing an earlier upload of it, ingests
// it, and records the run's organization and user as the session's owner.
func (a *RunsAPI) handleTranscript(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord) {
	if run.Spec.Kind != "agent" {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("run %s is a %s run; only agent runs upload transcripts", run.ID, run.Spec.Kind))
		return
	}
	source, name := r.URL.Query().Get("source"), r.URL.Query().Get("name")
	// The source names a directory, so it becomes one of the two fixed names rather than the
	// request's own string.
	switch source {
	case "claude":
		source = "claude"
	case "codex":
		source = "codex"
	default:
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("transcript source %q is not claude or codex", source))
		return
	}
	// No ".." at all, even one that cleans away: the name is joined under the run's directory.
	if name == "" || filepath.IsAbs(name) || strings.Contains(name, "..") || filepath.Ext(name) != ".jsonl" {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("transcript name %q must be a relative .jsonl path inside the run's transcripts", name))
		return
	}
	clean := filepath.Clean(name)
	path := filepath.Join(TranscriptsDir(a.Sources.DataDir), run.Org.String(), run.ID.String(), source, clean)
	if err := writeFileAtomically(path, http.MaxBytesReader(w, r.Body, maxTranscriptBytes)); err != nil {
		logger.Warnf("store a transcript of sandbox run %s: %v", run.ID, err)
		writeCIError(w, http.StatusInternalServerError, "could not store the transcript")
		return
	}
	captain, err := captaindb.Use(a.DB)
	if err != nil {
		logger.Warnf("open Captain's store for sandbox run %s: %v", run.ID, err)
		writeCIError(w, http.StatusInternalServerError, "could not ingest the transcript")
		return
	}
	input, err := monitor.BuildIngestInput(r.Context(), captain, captaindb.LocalHostID(), source, path)
	if err != nil {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("read transcript %s of run %s: %v", clean, run.ID, err))
		return
	}
	session, err := captain.IngestTranscript(r.Context(), input)
	if err != nil {
		logger.Warnf("ingest transcript %s of sandbox run %s: %v", clean, run.ID, err)
		writeCIError(w, http.StatusInternalServerError, "could not ingest the transcript")
		return
	}
	ids := []any{session.ID}
	if session.RootSessionID != nil {
		ids = append(ids, *session.RootSessionID)
	}
	for _, id := range ids {
		if err := a.DB.WithContext(r.Context()).Exec(`INSERT INTO session_owners (session_id, org_id, user_id) VALUES (?, ?, ?)
			ON CONFLICT (session_id) DO NOTHING`, id, run.Org, run.User).Error; err != nil {
			logger.Warnf("record the owner of a session of sandbox run %s: %v", run.ID, err)
			writeCIError(w, http.StatusInternalServerError, "could not record the session's owner")
			return
		}
	}
	w.WriteHeader(http.StatusAccepted)
}

// writeFileAtomically replaces path with what r holds, so a reader never sees half an upload.
func writeFileAtomically(path string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".upload-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := io.Copy(file, r); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
