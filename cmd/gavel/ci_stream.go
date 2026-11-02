// Streams a `gavel test` run's in-flight results and diagnostics to a hosted gavel while the
// run is in progress, so a GitHub Actions job killed before the action's upload step still
// leaves its latest state on the server.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/commons/logger"
	testui "github.com/flanksource/gavel/testrunner/ui"
	"github.com/golang-jwt/jwt/v5"
)

const (
	// ciStreamServerEnv names the hosted gavel to stream to; the action sets it from its
	// gavel-server input.
	ciStreamServerEnv = "GAVEL_CI_SERVER"
	// ciStreamInterval is how often a changed in-flight snapshot is uploaded.
	ciStreamInterval = 15 * time.Second
	// ciStreamMaxDiagnostics bounds the diagnostics capture carried in each upload.
	ciStreamMaxDiagnostics = 256 << 10
)

// ciStream uploads the latest in-flight snapshot every ciStreamInterval when it changed: the
// first upload creates a run on the server, later ones replace it. Failures are logged and
// never fail the run.
type ciStream struct {
	server string
	getenv func(string) string
	client *http.Client

	mu          sync.Mutex
	pending     *testui.Snapshot
	diagnostics string
	dirty       bool

	// runID, token and tokenExpiry are only touched by the upload loop and Stop's final flush,
	// which runs after the loop ended.
	runID       string
	token       string
	tokenExpiry time.Time
	failed      bool

	stop chan struct{}
	done chan struct{}
}

// startCIStream starts streaming when GAVEL_CI_SERVER is set in a GitHub Actions job that
// can request an OIDC token; otherwise it returns nil, whose methods do nothing.
func startCIStream(getenv func(string) string) *ciStream {
	server := strings.TrimRight(getenv(ciStreamServerEnv), "/")
	if server == "" {
		return nil
	}
	if getenv("ACTIONS_ID_TOKEN_REQUEST_URL") == "" {
		logger.Warnf("%s is set but the job can't request an OIDC token (it needs permissions: id-token: write); not streaming results", ciStreamServerEnv)
		return nil
	}
	s := &ciStream{
		server: server,
		getenv: getenv,
		client: &http.Client{Timeout: 30 * time.Second},
		stop:   make(chan struct{}),
		done:   make(chan struct{}),
	}
	go s.loop()
	return s
}

// Update records snap as the run's latest in-flight state.
func (s *ciStream) Update(snap *testui.Snapshot) {
	if s == nil || snap == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending, s.dirty = snap, true
}

// Diagnostics records text, the latest diagnostics capture, to be sent with the snapshot as
// its log tail, and uploads it on the next tick even when no test changed.
func (s *ciStream) Diagnostics(text string) {
	if s == nil {
		return
	}
	if len(text) > ciStreamMaxDiagnostics {
		text = text[len(text)-ciStreamMaxDiagnostics:]
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.diagnostics, s.dirty = text, true
}

// Stop ends the periodic uploads and sends whatever changed since the last one.
func (s *ciStream) Stop() {
	if s == nil {
		return
	}
	select {
	case <-s.stop:
		return
	default:
		close(s.stop)
	}
	<-s.done
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	s.flush(ctx)
}

func (s *ciStream) loop() {
	defer close(s.done)
	ticker := time.NewTicker(ciStreamInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			s.flush(ctx)
			cancel()
		}
	}
}

// flush uploads the pending snapshot when it changed since the last upload. A failed upload
// stays pending for the next tick.
func (s *ciStream) flush(ctx context.Context) {
	s.mu.Lock()
	if !s.dirty || (s.pending == nil && s.diagnostics == "") {
		s.mu.Unlock()
		return
	}
	snap := testui.Snapshot{Status: testui.SnapshotStatus{Running: true}}
	if s.pending != nil {
		snap = *s.pending
	}
	snap.LogTail = s.diagnostics
	s.dirty = false
	s.mu.Unlock()

	if err := s.upload(ctx, snap); err != nil {
		s.mu.Lock()
		s.dirty = true
		s.mu.Unlock()
		if !s.failed {
			logger.Warnf("stream results to %s: %v (retrying every %s)", s.server, err, ciStreamInterval)
			s.failed = true
		} else {
			logger.V(1).Infof("stream results to %s: %v", s.server, err)
		}
		return
	}
	s.failed = false
}

func (s *ciStream) upload(ctx context.Context, snap testui.Snapshot) error {
	data, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("encode the snapshot: %w", err)
	}
	if s.token == "" || time.Until(s.tokenExpiry) < time.Minute {
		if s.token, err = requestActionsOIDCToken(ctx, s.client, s.getenv, s.server); err != nil {
			return err
		}
		s.tokenExpiry = oidcTokenExpiry(s.token)
	}
	runID, err := sendCIResults(ctx, s.client, s.server, s.token, s.runID, data)
	if err != nil {
		return err
	}
	if s.runID == "" {
		s.runID = runID
		logger.Infof("Streaming results to %s as run %s", s.server, runID)
		s.publishRunID(runID)
	}
	return nil
}

// publishRunID writes ci_run_id to the step's outputs, so the action's upload step replaces
// the streamed run with the final results instead of adding a second run.
func (s *ciStream) publishRunID(runID string) {
	path := s.getenv("GITHUB_OUTPUT")
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		logger.Warnf("write ci_run_id to GITHUB_OUTPUT: %v", err)
		return
	}
	defer f.Close()
	if _, err := fmt.Fprintf(f, "ci_run_id=%s\n", runID); err != nil {
		logger.Warnf("write ci_run_id to GITHUB_OUTPUT: %v", err)
	}
}

// oidcTokenExpiry reads raw's exp without verifying it (the server does that), so the token
// is renewed before it lapses. A token without one is renewed on the next upload.
func oidcTokenExpiry(raw string) time.Time {
	claims := jwt.RegisteredClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(raw, &claims); err != nil || claims.ExpiresAt == nil {
		return time.Time{}
	}
	return claims.ExpiresAt.Time
}
