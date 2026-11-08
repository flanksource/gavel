// Transcripts of an agent job: the job uploads the transcript files its agent writes under $HOME
// as they grow, so serve can show the run while it works and keep it after the job is gone.
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/ai/history"
)

// transcriptInterval is how often a job uploads the transcripts that changed.
const transcriptInterval = 10 * time.Second

// transcriptUploader uploads the agent transcripts under home that were written since the job
// started and changed since it last looked. Transcripts older than since belong to sessions
// that ran before the job, so they are never uploaded.
type transcriptUploader struct {
	api   runAPI
	home  string
	since time.Time
	seen  map[string]transcriptStamp
}

type transcriptStamp struct {
	size    int64
	modTime time.Time
}

// transcriptRoots are where each agent writes its transcripts, under home. Upload names are
// relative to them.
var transcriptRoots = map[string]string{
	history.TranscriptSourceClaude: filepath.Join(".claude", "projects"),
	history.TranscriptSourceCodex:  filepath.Join(".codex", "sessions"),
}

// sync uploads every transcript that changed since the last sync.
func (u *transcriptUploader) sync(ctx context.Context) error {
	if u.seen == nil {
		u.seen = map[string]transcriptStamp{}
	}
	found, err := history.FindTranscripts(history.TranscriptQuery{Home: u.home, ModifiedSince: u.since})
	if err != nil {
		return fmt.Errorf("find the agent transcripts under %s: %w", u.home, err)
	}
	for _, transcript := range found {
		stamp := transcriptStamp{size: transcript.Info.Size(), modTime: transcript.Info.ModTime()}
		if u.seen[transcript.Path] == stamp {
			continue
		}
		name, err := filepath.Rel(filepath.Join(u.home, transcriptRoots[transcript.Source]), transcript.Path)
		if err != nil {
			return err
		}
		if err := u.upload(ctx, transcript.Source, filepath.ToSlash(name), transcript.Path); err != nil {
			return fmt.Errorf("upload the %s transcripts: %w", transcript.Source, err)
		}
		u.seen[transcript.Path] = stamp
	}
	return nil
}

func (u *transcriptUploader) upload(ctx context.Context, source, name, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	query := url.Values{"source": {source}, "name": {name}}
	resp, err := u.api.Do(ctx, http.MethodPost, "transcripts?"+query.Encode(), file)
	if err != nil {
		return fmt.Errorf("upload %s: %w", name, err)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.Body.Close()
}

// every syncs u each interval until ctx ends, writing what it couldn't upload to output.
func (u *transcriptUploader) every(ctx context.Context, interval time.Duration, output io.Writer) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := u.sync(ctx); err != nil && ctx.Err() == nil {
				fmt.Fprintf(output, "gavel runner: %s\n", strings.TrimSpace(err.Error()))
			}
		}
	}
}
