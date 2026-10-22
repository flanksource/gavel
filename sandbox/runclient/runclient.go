// Package runclient calls gavel serve's run API from inside a sandbox job, authenticated with the
// job's run token for its own run.
package runclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ErrNotInAJob is a process outside a sandbox job, whose environment names no run.
var ErrNotInAJob = errors.New("not in a sandbox job")

// Client calls the run API for one run.
type Client struct {
	Server string
	Token  string
	Run    string
	HTTP   *http.Client
}

// FromEnv is the client for the run the job's environment names.
func FromEnv(getenv func(string) string) (*Client, error) {
	server, token, run := getenv("GAVEL_SERVER"), getenv("GAVEL_RUN_TOKEN"), getenv("GAVEL_RUN_ID")
	if server == "" || token == "" || run == "" {
		return nil, fmt.Errorf("%w: GAVEL_SERVER, GAVEL_RUN_TOKEN and GAVEL_RUN_ID must be set", ErrNotInAJob)
	}
	return &Client{Server: strings.TrimRight(server, "/"), Token: token, Run: run, HTTP: &http.Client{Timeout: 10 * time.Minute}}, nil
}

// Do sends method to /api/runs/<run>/<path>. A response that isn't 2xx is an error carrying
// serve's reason.
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.Server+"/api/runs/"+c.Run+"/"+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
		var failure struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &failure) == nil && failure.Error != "" {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, failure.Error)
		}
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return resp, nil
}
