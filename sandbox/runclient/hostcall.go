// The run API over a host call: a run's program in a substrate environment asks the run's caller,
// which serves the request through its run API handler, so the environment needs no token and no
// route to serve.
package runclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

// HostCallName is the host call a run's program reaches its caller's run API with.
const HostCallName = "gavel.run-api"

// hostRequest and hostResponse are a run API request and its response, as a host call carries
// them.
type hostRequest struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Body   []byte `json:"body,omitempty"`
}

type hostResponse struct {
	Status int    `json:"status"`
	Body   []byte `json:"body,omitempty"`
}

// HostCallClient calls the run API through call, the run's host call to its caller.
type HostCallClient struct {
	call func(name string, data json.RawMessage) (json.RawMessage, error)
}

// OverHostCall is the run API reached through call.
func OverHostCall(call func(name string, data json.RawMessage) (json.RawMessage, error)) *HostCallClient {
	return &HostCallClient{call: call}
}

// Do sends method to the run's path through the host call. A response that isn't 2xx is an error
// carrying the run API's reason, as over HTTP.
func (c *HostCallClient) Do(_ context.Context, method, path string, body io.Reader) (*http.Response, error) {
	request := hostRequest{Method: method, Path: path}
	if body != nil {
		data, err := io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("read the body of %s %s: %w", method, path, err)
		}
		request.Body = data
	}
	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	answer, err := c.call(HostCallName, data)
	if err != nil {
		return nil, err
	}
	var response hostResponse
	if err := json.Unmarshal(answer, &response); err != nil {
		return nil, fmt.Errorf("decode the answer to %s %s: %w", method, path, err)
	}
	if response.Status/100 != 2 {
		return nil, failed(response.Status, response.Body)
	}
	return &http.Response{StatusCode: response.Status, Body: io.NopCloser(bytes.NewReader(response.Body)), Header: http.Header{}}, nil
}

// AnswerHostCall answers run's host calls of the run API by serving each request through
// handler, a run API, as run with its token, as a request over HTTP from the run's job would be.
// Only the run's PR endpoints are reached: what the environment can't do itself.
func AnswerHostCall(handler http.Handler, run, token string) func(context.Context, json.RawMessage) (json.RawMessage, error) {
	return func(ctx context.Context, data json.RawMessage) (json.RawMessage, error) {
		var request hostRequest
		if err := json.Unmarshal(data, &request); err != nil {
			return nil, fmt.Errorf("decode a run API request: %w", err)
		}
		if !strings.HasPrefix(request.Path, "pr/") || strings.Contains(request.Path, "..") {
			return nil, fmt.Errorf("a run's program reaches only its PR's endpoints through its caller, not %q", request.Path)
		}
		req, err := http.NewRequestWithContext(ctx, request.Method, "/api/runs/"+run+"/"+request.Path, bytes.NewReader(request.Body))
		if err != nil {
			return nil, fmt.Errorf("run API request %s %s: %w", request.Method, request.Path, err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		return json.Marshal(hostResponse{Status: recorder.Code, Body: recorder.Body.Bytes()})
	}
}
