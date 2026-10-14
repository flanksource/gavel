package todo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/flanksource/gavel/fixtures"
)

// apiClient talks to a running gavel dashboard (`gavel serve`). Every call is
// bounded by the step's context; there is no client-side timeout of its own,
// because a launch legitimately holds its stream open until Captain admits the
// run.
type apiClient struct {
	server string
	http   *http.Client
}

// newAPIClient resolves the dashboard URL: the step's `server`, rendered against
// the fixture scope, or $GAVEL_URL from the fixture Env. Neither is an error —
// a todo step has nothing to talk to.
func newAPIClient(step *fixtures.StepContext, server string) (*apiClient, error) {
	if strings.TrimSpace(server) == "" {
		value, ok := step.Scope.Env.Get("GAVEL_URL")
		if !ok || strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("%s step %q: no `server` configured and GAVEL_URL is not set in the fixture env", step.Spec.Kind, step.Spec.ID)
		}
		server = value
	} else {
		rendered, err := step.Expand(server)
		if err != nil {
			return nil, fmt.Errorf("%s step %q: render server: %w", step.Spec.Kind, step.Spec.ID, err)
		}
		server = rendered
	}
	parsed, err := url.Parse(strings.TrimSpace(server))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("%s step %q: server %q is not an http(s) URL", step.Spec.Kind, step.Spec.ID, server)
	}
	return &apiClient{server: strings.TrimRight(parsed.String(), "/"), http: &http.Client{}}, nil
}

// todoURL is the dashboard page of a todo, exported as TODO_URL.
func (c *apiClient) todoURL(id string) string { return c.server + "/todos/" + url.PathEscape(id) }

func (c *apiClient) endpoint(path string, query url.Values) string {
	if len(query) == 0 {
		return c.server + path
	}
	return c.server + path + "?" + query.Encode()
}

func (c *apiClient) getJSON(ctx context.Context, path string, query url.Values, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(path, query), nil)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *apiClient) postJSON(ctx context.Context, path string, body, out any) error {
	req, err := c.newPost(ctx, path, body)
	if err != nil {
		return err
	}
	return c.do(req, out)
}

func (c *apiClient) newPost(ctx context.Context, path string, body any) (*http.Request, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("encode POST %s body: %w", path, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(path, nil), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *apiClient) do(req *http.Request, out any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", req.Method, req.URL.Path, err)
	}
	defer resp.Body.Close() //nolint:errcheck
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s %s: read response: %w", req.Method, req.URL.Path, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return &apiError{Method: req.Method, Path: req.URL.Path, Status: resp.StatusCode, Message: errorMessage(data)}
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("%s %s: decode response: %w", req.Method, req.URL.Path, err)
	}
	return nil
}

// apiError is a non-2xx answer; Message is the handler's {"error": …} text.
type apiError struct {
	Method, Path string
	Status       int
	Message      string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("%s %s: HTTP %d: %s", e.Method, e.Path, e.Status, e.Message)
}

// errorMessage reads pr/ui's writeTodoError body, falling back to the raw body
// for a reply that is not one (a proxy's HTML, a bare http.Error line).
func errorMessage(data []byte) string {
	var body struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(data, &body) == nil && body.Error != "" {
		return body.Error
	}
	return strings.TrimSpace(string(data))
}

// resolveRef is the todo a step addresses: its own ref, rendered, or $TODO_ID
// as exported by an earlier `todo create`.
func resolveRef(step *fixtures.StepContext, ref string) (string, error) {
	if strings.TrimSpace(ref) != "" {
		rendered, err := step.Expand(ref)
		if err != nil {
			return "", fmt.Errorf("%s step %q: render ref: %w", step.Spec.Kind, step.Spec.ID, err)
		}
		return strings.TrimSpace(rendered), nil
	}
	value, ok := step.Scope.Env.Get("TODO_ID")
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s step %q: no todo ref configured and TODO_ID is not set in the fixture env", step.Spec.Kind, step.Spec.ID)
	}
	return strings.TrimSpace(value), nil
}

// renderOptional renders an optional config string; empty stays empty.
func renderOptional(step *fixtures.StepContext, field, value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	rendered, err := step.Expand(value)
	if err != nil {
		return "", fmt.Errorf("%s step %q: render %s: %w", step.Spec.Kind, step.Spec.ID, field, err)
	}
	return strings.TrimSpace(rendered), nil
}
