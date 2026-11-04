// An agent run's transcripts: its agent's transcript files come to serve as they grow, and serve
// keeps them under its data directory and ingests them into Captain's store.
package ui

import (
	"context"
	"errors"
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
	"gorm.io/gorm"
)

// TranscriptsDir is where serve keeps the transcripts agent runs upload, under dataDir.
func TranscriptsDir(dataDir string) string {
	return filepath.Join(dataDir, "transcripts")
}

// maxTranscriptBytes bounds one uploaded transcript file.
const maxTranscriptBytes = 256 << 20

// handleTranscript stores the uploaded transcript file and ingests it, as RunTranscripts.Keep does.
func (a *RunsAPI) handleTranscript(w http.ResponseWriter, r *http.Request, run sandbox.RunRecord) {
	if run.Spec.Kind != "agent" {
		writeCIError(w, http.StatusBadRequest, fmt.Sprintf("run %s is a %s run; only agent runs upload transcripts", run.ID, run.Spec.Kind))
		return
	}
	transcripts := RunTranscripts{DB: a.DB, DataDir: a.Sources.DataDir}
	err := transcripts.Keep(r.Context(), run, r.URL.Query().Get("source"), r.URL.Query().Get("name"), http.MaxBytesReader(w, r.Body, maxTranscriptBytes))
	switch {
	case errors.As(err, new(*TranscriptRefusedError)):
		writeCIError(w, http.StatusBadRequest, err.Error())
	case err != nil:
		logger.Warnf("keep a transcript of sandbox run %s: %v", run.ID, err)
		writeCIError(w, http.StatusInternalServerError, "could not keep the transcript")
	default:
		w.WriteHeader(http.StatusAccepted)
	}
}

// TranscriptRefusedError is a transcript that isn't one: an unknown source, a name outside the
// run's transcripts, or a file that isn't a transcript.
type TranscriptRefusedError struct{ Reason string }

func (e *TranscriptRefusedError) Error() string { return e.Reason }

// RunTranscripts keeps agent runs' transcripts under serve's data directory and ingests them into
// Captain's store.
type RunTranscripts struct {
	DB      *gorm.DB
	DataDir string
}

// Keep stores run's transcript file of source (claude or codex) at
// <data dir>/transcripts/<org>/<run>/<source>/<name>, replacing an earlier copy of it, ingests it,
// and records the run's organization and user as the session's owner.
func (t RunTranscripts) Keep(ctx context.Context, run sandbox.RunRecord, source, name string, data io.Reader) error {
	// The source names a directory, so it becomes one of the two fixed names rather than the
	// caller's own string.
	switch source {
	case "claude":
		source = "claude"
	case "codex":
		source = "codex"
	default:
		return &TranscriptRefusedError{Reason: fmt.Sprintf("transcript source %q is not claude or codex", source)}
	}
	// No ".." at all, even one that cleans away: the name is joined under the run's directory.
	if name == "" || filepath.IsAbs(name) || strings.Contains(name, "..") || filepath.Ext(name) != ".jsonl" {
		return &TranscriptRefusedError{Reason: fmt.Sprintf("transcript name %q must be a relative .jsonl path inside the run's transcripts", name)}
	}
	clean := filepath.Clean(name)
	path := filepath.Join(TranscriptsDir(t.DataDir), run.Org.String(), run.ID.String(), source, clean)
	if err := writeFileAtomically(path, data); err != nil {
		return fmt.Errorf("store transcript %s: %w", clean, err)
	}
	captain, err := captaindb.Use(t.DB)
	if err != nil {
		return fmt.Errorf("open Captain's store: %w", err)
	}
	input, err := monitor.BuildIngestInput(ctx, captain, captaindb.LocalHostID(), source, path)
	if err != nil {
		return &TranscriptRefusedError{Reason: fmt.Sprintf("read transcript %s of run %s: %v", clean, run.ID, err)}
	}
	session, err := captain.IngestTranscript(ctx, input)
	if err != nil {
		return fmt.Errorf("ingest transcript %s: %w", clean, err)
	}
	ids := []any{session.ID}
	if session.RootSessionID != nil {
		ids = append(ids, *session.RootSessionID)
	}
	for _, id := range ids {
		if err := t.DB.WithContext(ctx).Exec(`INSERT INTO session_owners (session_id, org_id, user_id) VALUES (?, ?, ?)
			ON CONFLICT (session_id) DO NOTHING`, id, run.Org, run.User).Error; err != nil {
			return fmt.Errorf("record the owner of a session: %w", err)
		}
	}
	return nil
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
