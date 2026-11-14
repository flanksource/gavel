package testui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/flanksource/gavel/linters"
)

// RerunRequest is the payload accepted by POST /api/rerun.
type RerunRequest struct {
	PackagePaths []string `json:"package_paths,omitempty"`
	WorkDir      string   `json:"work_dir,omitempty"`
	TestName     string   `json:"test_name,omitempty"`
	Suite        []string `json:"suite,omitempty"`
	Framework    string   `json:"framework,omitempty"`
	Lint         bool     `json:"lint,omitempty"`
	LintFiles    []string `json:"lint_files,omitempty"`
	LintLinters  []string `json:"lint_linters,omitempty"`
	// Params pre-supply the fixtures' params: forms; without them a rerun
	// keeps the run's own (--param / --params) and prompts for what is missing.
	Params map[string]any `json:"params,omitempty"`
	// Fixture is the fixture file (a fixture step test's File) and FromStep
	// the step id a "retry from here" rerun resumes that file at.
	Fixture  string `json:"fixture,omitempty"`
	FromStep string `json:"from_step,omitempty"`
}

// BadRerunRequest refuses a rerun request before anything runs: handleRerun
// answers it with 400 and Reason as the plain-text body.
type BadRerunRequest struct {
	Reason string
}

func (e BadRerunRequest) Error() string { return e.Reason }

// RerunFunc is invoked to rerun the tests described by req.
// The output buffer receives live process stdout/stderr for streaming to the UI.
// It runs synchronously; the handler returns once it completes.
type RerunFunc func(req RerunRequest, output *RerunOutputBuffer) error

// SetRerunFunc installs the rerun callback. Call this before serving traffic.
func (s *Server) SetRerunFunc(fn RerunFunc) {
	s.rerunFn = fn
}

// LintRerunFunc reruns the named linters over files (both as a UI rerun
// request carries them) and returns their results.
type LintRerunFunc func(ctx context.Context, linters []string, files []string) ([]*linters.LinterResult, error)

type embeddedLintRerun struct {
	claimed []string
	run     LintRerunFunc
}

// SetLintRerun lets an embedder rerun its own linters from the UI: a lint
// rerun request naming only linters in claimed calls fn, and each result it
// returns replaces the one with the same linter and work dir, keeping every
// other linter's result. A request naming no linter, or none of claimed, goes
// to the SetRerunFunc callback; one mixing claimed and other linters is
// refused. Call this before serving traffic.
func (s *Server) SetLintRerun(claimed []string, fn LintRerunFunc) {
	s.lintRerun = &embeddedLintRerun{claimed: slices.Clone(claimed), run: fn}
}

// claims reports whether req is a lint rerun of linters the embedder claimed,
// failing when it names claimed and unclaimed linters together.
func (e *embeddedLintRerun) claims(req RerunRequest) (bool, error) {
	if e == nil || !req.Lint || len(req.LintLinters) == 0 {
		return false, nil
	}
	var claimed, other []string
	for _, name := range req.LintLinters {
		if slices.Contains(e.claimed, name) {
			claimed = append(claimed, name)
		} else {
			other = append(other, name)
		}
	}
	if len(claimed) > 0 && len(other) > 0 {
		return false, fmt.Errorf("lint rerun mixes %s with %s; rerun them separately", strings.Join(claimed, ","), strings.Join(other, ","))
	}
	return len(claimed) > 0, nil
}

func (s *Server) rerunLintEmbedded(ctx context.Context, req RerunRequest) error {
	results, err := s.lintRerun.run(ctx, req.LintLinters, req.LintFiles)
	if err != nil {
		return err
	}
	for _, result := range results {
		s.UpsertLintResult(result)
	}
	return nil
}

func (s *Server) handleRerun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req RerunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	embedded, err := s.lintRerun.claims(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !embedded && s.rerunFn == nil {
		http.Error(w, "rerun not supported", http.StatusNotImplemented)
		return
	}

	if !s.rerunMu.TryLock() {
		http.Error(w, "rerun already in progress", http.StatusConflict)
		return
	}
	defer s.rerunMu.Unlock()

	command := rerunCommandLabel(req)
	buf := NewRerunOutputBuffer(command)
	s.mu.Lock()
	s.rerunOutput = buf
	s.mu.Unlock()

	if embedded {
		err = s.rerunLintEmbedded(r.Context(), req)
	} else {
		err = s.rerunFn(req, buf)
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			buf.Cancel()
			w.WriteHeader(http.StatusAccepted)
			return
		}
		buf.Finish(false)
		status := http.StatusInternalServerError
		if errors.As(err, new(BadRerunRequest)) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}
	buf.Finish(true)
	w.WriteHeader(http.StatusAccepted)
}

func rerunCommandLabel(req RerunRequest) string {
	if req.Lint {
		parts := []string{"lint"}
		if len(req.LintLinters) > 0 {
			parts = append(parts, strings.Join(req.LintLinters, ","))
		}
		if len(req.LintFiles) > 0 {
			parts = append(parts, fmt.Sprintf("(%d files)", len(req.LintFiles)))
		}
		return strings.Join(parts, " ")
	}
	parts := []string{"test"}
	if req.Framework != "" {
		parts = append(parts, req.Framework)
	}
	if req.TestName != "" {
		parts = append(parts, req.TestName)
	} else if len(req.PackagePaths) > 0 {
		parts = append(parts, strings.Join(req.PackagePaths, " "))
	}
	return strings.Join(parts, " ")
}
