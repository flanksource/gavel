package launch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons-db/shell"
	"github.com/flanksource/gavel/resultstore"
	validator "github.com/santhosh-tekuri/jsonschema/v6"
)

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	for _, kind := range []string{"", "test", "lint"} {
		path := "/api/launch"
		if kind != "" {
			path += "/" + kind
		}
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) { s.handleLaunch(w, r, kind) })
	}
	mux.HandleFunc("POST /api/launch/prepare", s.handlePrepare)
	mux.HandleFunc("GET /api/launch/openapi", func(w http.ResponseWriter, r *http.Request) {
		document := map[string]any{"info": map[string]any{"title": "Gavel launch", "version": "1.0.0"}, "paths": map[string]any{}}
		if err := s.AddOpenAPI(document); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, document)
	})
	mux.HandleFunc("GET /api/launch/prompts", s.handlePrompts)
	mux.HandleFunc("GET /api/runs", s.handleRuns)
	mux.HandleFunc("GET /api/runs/{executionId}", s.handleRun)
	mux.HandleFunc("GET /api/runs/{executionId}/stream", s.handleStream)
	mux.HandleFunc("POST /api/runs/{executionId}/stop", func(w http.ResponseWriter, r *http.Request) {
		if err := s.owned(r.Context(), r.PathValue("executionId"), requestOwner(r)); err != nil {
			writeError(w, err)
			return
		}
		if err := s.Stop(r.Context(), r.PathValue("executionId")); err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"stopping": true})
	})
	return mux
}

func (s *Service) handleLaunch(w http.ResponseWriter, r *http.Request, kind string) {
	if r.Method == http.MethodGet && r.URL.Query().Has("schema") {
		schema, err := Schema(kind, s.options.Schemas)
		if err != nil {
			writeError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/schema+json")
		writeJSON(w, http.StatusOK, schema)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "use GET ?schema or POST a launch request", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 8<<20))
	if err != nil {
		writeError(w, err)
		return
	}
	request, err := Decode(body, kind, s.options.Schemas)
	if err != nil {
		writeError(w, &InputError{Err: err})
		return
	}
	run, err := s.Start(r.Context(), request, r.Header.Get("Idempotency-Key"), requestOwner(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"executionId": run.Metadata.ExecutionID, "run": run, "snapshotUrl": "/api/runs/" + run.Metadata.ExecutionID, "streamUrl": "/api/runs/" + run.Metadata.ExecutionID + "/stream", "stopUrl": "/api/runs/" + run.Metadata.ExecutionID + "/stop"})
}

func requestOwner(r *http.Request) string { return r.Header.Get("Authorization") }

func (s *Service) handlePrepare(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Workspace *shell.Setup `json:"workspace"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeError(w, &InputError{Err: err})
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeError(w, &InputError{Err: errors.New("prepare requires exactly one JSON object")})
		return
	}
	if input.Workspace == nil {
		writeError(w, &InputError{Err: errors.New("workspace is required")})
		return
	}
	workspace, err := s.options.Workspaces.Prepare(r.Context(), *input.Workspace, requestOwner(r))
	if err != nil {
		writeError(w, &InputError{Err: err})
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

func (s *Service) handlePrompts(w http.ResponseWriter, r *http.Request) {
	ws, err := s.options.Workspaces.Get(r.URL.Query().Get("preparationId"), requestOwner(r), false)
	if err != nil {
		writeError(w, err)
		return
	}
	if reference := r.URL.Query().Get("prompt"); reference != "" {
		detail, err := captaincli.InspectPrompt(r.Context(), ws.Directory, reference)
		if err != nil {
			writeError(w, err)
			return
		}
		form, err := promptFormDetail(detail)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, form)
		return
	}
	prompts, err := captaincli.DiscoverPrompts(r.Context(), ws.Directory, captaincli.PromptListOptions{Query: r.URL.Query().Get("query"), Source: r.URL.Query().Get("source")})
	if err != nil {
		writeError(w, err)
		return
	}
	if prompts == nil {
		prompts = []captaincli.PromptSummary{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"prompts": prompts})
}

func (s *Service) handleRuns(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 200 {
			http.Error(w, "limit must be between 1 and 200", 400)
			return
		}
		limit = parsed
	}
	labels := map[string]string{}
	for key, values := range r.URL.Query() {
		if strings.HasPrefix(key, "label.") {
			labels[strings.TrimPrefix(key, "label.")] = values[0]
		}
	}
	runs, err := s.options.Store.List(r.Context(), resultstore.ListOptions{Limit: limit, Before: r.URL.Query().Get("before"), Labels: labels, Owner: s.digest([]byte(requestOwner(r)))})
	if err != nil {
		writeError(w, err)
		return
	}
	for index := range runs {
		recovered, err := s.Get(r.Context(), runs[index].Metadata.ExecutionID)
		if err != nil {
			writeError(w, err)
			return
		}
		runs[index] = recovered
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": runs})
}

func (s *Service) handleRun(w http.ResponseWriter, r *http.Request) {
	if err := s.owned(r.Context(), r.PathValue("executionId"), requestOwner(r)); err != nil {
		writeError(w, err)
		return
	}
	run, err := s.Get(r.Context(), r.PathValue("executionId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Service) handleStream(w http.ResponseWriter, r *http.Request) {
	if err := s.owned(r.Context(), r.PathValue("executionId"), requestOwner(r)); err != nil {
		writeError(w, err)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", 500)
		return
	}
	run, err := s.Get(r.Context(), r.PathValue("executionId"))
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	var revision uint64
	for {
		if run.Metadata.Revision != revision {
			data, err := json.Marshal(run)
			if err != nil {
				fmt.Fprintf(w, "event: error\ndata: %q\n\n", err.Error())
				return
			}
			fmt.Fprintf(w, "id: %d\ndata: %s\n\n", run.Metadata.Revision, data)
			flusher.Flush()
			revision = run.Metadata.Revision
		}
		if run.Terminal() {
			fmt.Fprint(w, "event: done\ndata: {}\n\n")
			flusher.Flush()
			return
		}
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
		}
		run, err = s.Get(r.Context(), r.PathValue("executionId"))
		if err != nil {
			data, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintf(w, "event: error\ndata: %s\n\n", data)
			flusher.Flush()
			return
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		return
	}
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	var input *InputError
	if errors.As(err, &input) {
		status = http.StatusUnprocessableEntity
	}
	if errors.Is(err, resultstore.ErrNotFound) {
		status = http.StatusNotFound
	}
	if errors.Is(err, ErrPreparationExpired) {
		status = http.StatusGone
	}
	if errors.Is(err, resultstore.ErrConflict) || errors.Is(err, captaincli.ErrPromptVersion) {
		status = http.StatusConflict
	}
	var validation *validator.ValidationError
	if errors.As(err, &validation) {
		status = http.StatusUnprocessableEntity
	}
	body := map[string]any{"error": err.Error()}
	if validation != nil {
		body["errors"] = validationIssues(validation)
	}
	writeJSON(w, status, body)
}

func validationIssues(err *validator.ValidationError) []map[string]string {
	if len(err.Causes) > 0 {
		var issues []map[string]string
		for _, cause := range err.Causes {
			issues = append(issues, validationIssues(cause)...)
		}
		return issues
	}
	path := ""
	for _, token := range err.InstanceLocation {
		path += "/" + strings.ReplaceAll(strings.ReplaceAll(token, "~", "~0"), "/", "~1")
	}
	return []map[string]string{{"instancePath": path, "message": err.Error()}}
}
