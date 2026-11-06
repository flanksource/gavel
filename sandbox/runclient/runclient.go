// Package runclient calls a run's run API from where the run executes: gavel serve's, over HTTP
// from a sandbox job with the job's run token, or the run's caller's, over a host call.
package runclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
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

// requestTimeout bounds one run API request.
const requestTimeout = 10 * time.Minute

// FromEnv is the client for the run the job's environment names: through the socket its
// environment's program serves the run API on, which needs no token, or else over HTTP to serve
// with the run's token.
func FromEnv(getenv func(string) string) (*Client, error) {
	run := getenv("GAVEL_RUN_ID")
	if socket := getenv(SocketEnv); socket != "" && run != "" {
		dialer := &net.Dialer{}
		transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, "unix", socket)
		}}
		return &Client{Server: "http://run-api", Run: run, HTTP: &http.Client{Timeout: requestTimeout, Transport: transport}}, nil
	}
	server, token := getenv("GAVEL_SERVER"), getenv("GAVEL_RUN_TOKEN")
	if server == "" || token == "" || run == "" {
		return nil, fmt.Errorf("%w: GAVEL_RUN_ID, and %s or GAVEL_SERVER and GAVEL_RUN_TOKEN, must be set", ErrNotInAJob, SocketEnv)
	}
	return &Client{Server: strings.TrimRight(server, "/"), Token: token, Run: run, HTTP: &http.Client{Timeout: requestTimeout}}, nil
}

// Do sends method to /api/runs/<run>/<path>, with the run's token when the client has one. A
// response that isn't 2xx is an error carrying serve's reason.
func (c *Client) Do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.Server+"/api/runs/"+c.Run+"/"+path, body)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
		return nil, failed(resp.StatusCode, data)
	}
	return resp, nil
}

// failed is the error of a response that isn't 2xx: its status, and the run API's reason when
// its body gives one.
func failed(status int, body []byte) error {
	var failure struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &failure) == nil && failure.Error != "" {
		return fmt.Errorf("HTTP %d: %s", status, failure.Error)
	}
	return fmt.Errorf("HTTP %d", status)
}
