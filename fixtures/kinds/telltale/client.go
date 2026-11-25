package telltale

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/flanksource/gavel/fixtures"
)

type client struct{ server, token string }

func escape(value string) string { return url.PathEscape(value) }
func newClient(step *fixtures.StepContext, cfg config) (*client, error) {
	server := cfg.Server
	if server == "" {
		server, _ = step.Scope.Env.Get("TELLTALE_URL")
	}
	server, err := step.Expand(server)
	if err != nil {
		return nil, err
	}
	parsed, err := url.Parse(server)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("telltale server must be an http(s) URL; set server or TELLTALE_URL")
	}
	return &client{server: strings.TrimRight(server, "/"), token: cfg.Token}, nil
}

type httpError struct {
	status     int
	path, body string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("Telltale %s: HTTP %d: %s", e.path, e.status, e.body)
}
func (c *client) request(ctx context.Context, method, path string, body any) (data []byte, err error) {
	var input io.Reader
	if body != nil {
		raw, e := json.Marshal(body)
		if e != nil {
			return nil, e
		}
		input = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.server+path, input)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, resp.Body.Close()) }()
	const limit = 32 << 20
	data, err = io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, fmt.Errorf("telltale response exceeds 32 MiB")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &httpError{resp.StatusCode, path, string(data)}
	}
	return data, nil
}

type capture struct {
	client *client
	id     string
}

func (r *capture) observe(ctx context.Context) (map[string]any, error) {
	raw, err := r.client.request(ctx, "GET", "/api/traces?limit=0", nil)
	if err != nil {
		return nil, err
	}
	var list struct {
		Runs []map[string]any `json:"runs"`
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	for _, run := range list.Runs {
		if run["runId"] == r.id {
			data, err := json.Marshal(run)
			if err != nil {
				return nil, err
			}
			return map[string]any{"runId": r.id, "status": run["status"], "json": run, "stdout": string(data), "exitCode": 0}, nil
		}
	}
	return nil, fmt.Errorf("telltale capture %s disappeared", r.id)
}
func (r *capture) close(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	state, err := r.observe(ctx)
	if err != nil {
		return err
	}
	if state["status"] == "running" {
		_, err = r.client.request(ctx, "DELETE", "/api/traces/"+escape(r.id), nil)
		var conflict *httpError
		if err != nil && (!errors.As(err, &conflict) || conflict.status != 409) {
			return err
		}
	}
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		state, err = r.observe(ctx)
		if err != nil {
			return err
		}
		switch state["status"] {
		case "succeeded":
			return nil
		case "failed":
			return fmt.Errorf("telltale capture %s failed: %v", r.id, state["json"])
		case "running":
		default:
			return fmt.Errorf("unexpected Telltale capture state: %v", state["status"])
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
