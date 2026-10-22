// Transcripts of an agent job: the job uploads the transcript files its agent writes under $HOME
// as they grow, so serve can show the run while it works and keep it after the job is gone.
package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// transcriptInterval is how often a job uploads the transcripts that changed.
const transcriptInterval = 10 * time.Second

// transcriptUploader uploads the agent transcripts under home that changed since it last looked.
type transcriptUploader struct {
	api  runAPI
	home string
	seen map[string]transcriptStamp
}

type transcriptStamp struct {
	size    int64
	modTime time.Time
}

// transcriptRoots are where each agent writes its transcripts, under home.
var transcriptRoots = map[string]string{
	"claude": filepath.Join(".claude", "projects"),
	"codex":  filepath.Join(".codex", "sessions"),
}

// sync uploads every transcript that changed since the last sync.
func (u *transcriptUploader) sync(ctx context.Context) error {
	if u.seen == nil {
		u.seen = map[string]transcriptStamp{}
	}
	for source, root := range transcriptRoots {
		dir := filepath.Join(u.home, root)
		err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return filepath.SkipDir
				}
				return err
			}
			if entry.IsDir() || filepath.Ext(path) != ".jsonl" {
				return nil
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			stamp := transcriptStamp{size: info.Size(), modTime: info.ModTime()}
			if u.seen[path] == stamp {
				return nil
			}
			name, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			if err := u.upload(ctx, source, filepath.ToSlash(name), path); err != nil {
				return err
			}
			u.seen[path] = stamp
			return nil
		})
		if err != nil {
			return fmt.Errorf("upload the %s transcripts: %w", source, err)
		}
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
