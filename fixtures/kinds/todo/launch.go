package todo

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// runPayload is pr/ui todoRunPayload as a fixture sends it.
type runPayload struct {
	Dir         string          `json:"dir,omitempty"`
	Ref         string          `json:"ref"`
	Step        string          `json:"step"`
	Spec        json.RawMessage `json:"spec,omitempty"`
	Resume      bool            `json:"resume,omitempty"`
	Force       bool            `json:"force,omitempty"`
	ReuseBranch bool            `json:"reuseBranch,omitempty"`
}

// launchResult is what POST /api/todos/run answered. With Accept:
// text/event-stream the handler streams `resolved` (the folded spec), then
// `admitted` (pr/ui todoRunResponse) or `failed` ({error, status}); a request
// refused before the stream starts is a plain JSON error with its HTTP status.
type launchResult struct {
	Resolved map[string]any
	Admitted map[string]any
	Failed   bool
	Error    string
	Status   int
}

// outputs is the `launch` CEL root. It is present on failure too, so a
// `retry.when` can read launch.failed / launch.error / launch.status.
func (l launchResult) outputs() map[string]any {
	return map[string]any{
		"resolved":    nonNil(l.Resolved),
		"admitted":    nonNil(l.Admitted),
		"failed":      l.Failed,
		"error":       l.Error,
		"status":      l.Status,
		"sessionId":   stringField(l.Admitted, "sessionId"),
		"promptRunId": stringField(l.Admitted, "promptRunId"),
	}
}

func (l *launchResult) fail(status int, message string) {
	l.Failed, l.Status, l.Error = true, status, message
}

// launch starts a lifecycle step and returns once the dashboard admitted or
// refused it. Every failure is recorded on the result as well as returned.
func (c *apiClient) launch(ctx context.Context, payload runPayload) (launchResult, error) {
	var result launchResult
	req, err := c.newPost(ctx, "/api/todos/run", payload)
	if err != nil {
		return result, err
	}
	req.Header.Set("Accept", "text/event-stream")
	resp, err := c.http.Do(req)
	if err != nil {
		result.fail(0, err.Error())
		return result, fmt.Errorf("POST /api/todos/run: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		data, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			data = []byte(fmt.Sprintf("(reading the error body failed: %v)", readErr))
		}
		result.fail(resp.StatusCode, errorMessage(data))
		return result, &apiError{Method: http.MethodPost, Path: "/api/todos/run", Status: resp.StatusCode, Message: result.Error}
	}
	if contentType := resp.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "text/event-stream") {
		result.fail(resp.StatusCode, fmt.Sprintf("expected a text/event-stream launch reply, got %q", contentType))
		return result, errors.New("POST /api/todos/run: " + result.Error)
	}
	if err := readEvents(resp.Body, result.apply); err != nil {
		result.fail(resp.StatusCode, err.Error())
		return result, fmt.Errorf("POST /api/todos/run: %w", err)
	}
	return result, result.verdict()
}

func (l *launchResult) apply(event string, data []byte) error {
	var frame map[string]any
	if err := json.Unmarshal(data, &frame); err != nil {
		return fmt.Errorf("decode %s frame: %w", event, err)
	}
	switch event {
	case "resolved":
		l.Resolved = frame
	case "admitted":
		l.Admitted = frame
	case "failed":
		status, _ := frame["status"].(float64)
		l.fail(int(status), stringField(frame, "error"))
	default:
		return fmt.Errorf("unexpected launch event %q", event)
	}
	return nil
}

func (l *launchResult) verdict() error {
	switch {
	case l.Failed:
		return fmt.Errorf("POST /api/todos/run: launch failed (HTTP %d): %s", l.Status, l.Error)
	case l.Admitted == nil:
		l.fail(0, "launch stream ended without an admitted or failed frame")
		return errors.New("POST /api/todos/run: " + l.Error)
	case stringField(l.Admitted, "sessionId") == "" || stringField(l.Admitted, "promptRunId") == "":
		l.fail(0, fmt.Sprintf("admitted frame names no sessionId/promptRunId: %v", l.Admitted))
		return errors.New("POST /api/todos/run: " + l.Error)
	}
	return nil
}

// readEvents parses a server-sent event stream, calling handle per dispatched
// event. Multi-line data is joined with newlines; comment lines are ignored.
func readEvents(r io.Reader, handle func(event string, data []byte) error) error {
	reader := bufio.NewReader(r)
	var event string
	var data []string
	dispatch := func() error {
		if event == "" && len(data) == 0 {
			return nil
		}
		name := event
		if name == "" {
			name = "message"
		}
		err := handle(name, []byte(strings.Join(data, "\n")))
		event, data = "", nil
		return err
	}
	for {
		line, readErr := reader.ReadString('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return fmt.Errorf("read launch stream: %w", readErr)
		}
		line = strings.TrimRight(line, "\r\n")
		field, value, _ := strings.Cut(line, ":")
		value = strings.TrimPrefix(value, " ")
		var err error
		switch {
		case line == "":
			err = dispatch()
		case field == "event":
			event = value
		case field == "data":
			data = append(data, value)
		}
		if err != nil {
			return err
		}
		if readErr != nil {
			return dispatch()
		}
	}
}

func nonNil(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	return m
}
