package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/todos/run"
	"github.com/ghodss/yaml"
)

type todoLaunchStream struct {
	w       http.ResponseWriter
	started bool
}

func newTodoLaunchStream(w http.ResponseWriter, r *http.Request) *todoLaunchStream {
	if !strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		return nil
	}
	return &todoLaunchStream{w: w}
}

func (stream *todoLaunchStream) resolved(prepared *run.Prepared) error {
	if stream == nil {
		return nil
	}
	specYAML, err := yaml.Marshal(prepared.Resolution.Spec)
	if err != nil {
		return fmt.Errorf("render resolved run spec: %w", err)
	}
	return stream.send("resolved", struct {
		Spec     api.Spec `json:"spec"`
		SpecYAML string   `json:"specYaml"`
		Step     string   `json:"step"`
		Reason   string   `json:"reason,omitempty"`
	}{Spec: prepared.Resolution.Spec, SpecYAML: string(specYAML), Step: prepared.Step.Name, Reason: prepared.Reason})
}

func (stream *todoLaunchStream) send(event string, value any) error {
	if stream == nil {
		return nil
	}
	flusher, ok := stream.w.(http.Flusher)
	if !ok {
		return fmt.Errorf("launch progress response cannot flush %s event", event)
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode %s launch event: %w", event, err)
	}
	stream.w.Header().Set("Content-Type", "text/event-stream")
	stream.w.Header().Set("Cache-Control", "no-cache")
	stream.w.Header().Set("X-Accel-Buffering", "no")
	stream.started = true
	if _, err := fmt.Fprintf(stream.w, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return fmt.Errorf("write %s launch event: %w", event, err)
	}
	flusher.Flush()
	return nil
}

func writeTodoLaunchError(w http.ResponseWriter, stream *todoLaunchStream, status int, err error) {
	if stream != nil && stream.started {
		stream.failed(status, err)
		return
	}
	writeTodoError(w, status, err)
}

func (stream *todoLaunchStream) failed(status int, err error) {
	_ = stream.send("failed", struct {
		Error  string `json:"error"`
		Status int    `json:"status"`
	}{Error: err.Error(), Status: status})
}
