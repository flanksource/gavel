package todo_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/captain/pkg/ai/provider/cmux"
	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
)

// The stub answers with the shapes pr/ui's handlers write. Attempts and session
// stats are encoded from the same captain types the handlers embed
// (todoAttemptDetail = PromptRunOverview + link fields; todoSessionStatsResponse
// = cmux.SessionStats + verification), so a renamed captain field breaks these
// specs rather than only the live dashboard.

const (
	stubWorkspace = "/work/acme"
	stubTodoID    = "6f1c2a8e-3b4d-4c5e-9f60-718293a4b5c6"
)

// attemptDetail mirrors pr/ui todoAttemptDetail.
type attemptDetail struct {
	captaindb.PromptRunOverview
	PromptRunID      uuid.UUID         `json:"promptRunId"`
	AdmissionSession uuid.UUID         `json:"admissionSessionId"`
	Ordinal          int               `json:"ordinal"`
	Step             string            `json:"step"`
	Verification     *api.VerifyReport `json:"verification"`
}

func newAttempt(runID, sessionID uuid.UUID, step string, status captaindb.PromptRunDisplayStatus, state captaindb.PromptRunState) attemptDetail {
	queued := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	return attemptDetail{
		PromptRunOverview: captaindb.PromptRunOverview{
			PromptRun: captaindb.PromptRun{
				ID: runID, SessionID: sessionID, RootSessionID: sessionID,
				Phase: captaindb.PromptRunPhaseGenerate, State: state,
				QueuedAt: queued, CreatedAt: queued, UpdatedAt: queued,
			},
			Status: status,
		},
		PromptRunID: runID, AdmissionSession: sessionID, Ordinal: 1, Step: step,
	}
}

type recordedRequest struct {
	Method string
	Path   string
	Query  map[string]string
	Accept string
	Body   map[string]any
}

type stubGavel struct {
	server *httptest.Server

	mu             sync.Mutex
	requests       []recordedRequest
	item           map[string]any
	attempts       []attemptDetail
	stats          cmux.SessionStats
	plan           map[string]any
	createResponse map[string]any
	reviewStatus   int
	reviewBody     map[string]any
	launchStatus   int
	launchType     string
	launchBody     string
}

func newStubGavel() *stubGavel {
	stub := &stubGavel{
		item: map[string]any{
			"ref": stubTodoID, "id": stubTodoID, "shortId": stubTodoID[:8], "version": 1,
			"title": "Add greeting helper", "status": string(types.StatusDraft), "priority": "medium",
			"cwd": stubWorkspace, "body": "add a greeting helper",
		},
		attempts:     []attemptDetail{},
		plan:         map[string]any{"found": false},
		reviewStatus: http.StatusOK,
		launchStatus: http.StatusOK,
		launchType:   "text/event-stream",
	}
	stub.createResponse = map[string]any{"todo": stub.item, "autoSave": false}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/todos/new", stub.handle(func(w http.ResponseWriter, _ recordedRequest) {
		writeJSON(w, http.StatusCreated, stub.createResponse)
	}))
	mux.HandleFunc("GET /api/todos/item", stub.handle(func(w http.ResponseWriter, req recordedRequest) {
		if req.Query["ref"] == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ref is required"})
			return
		}
		writeJSON(w, http.StatusOK, stub.item)
	}))
	mux.HandleFunc("GET /api/todos/session/detail", stub.handle(func(w http.ResponseWriter, req recordedRequest) {
		if req.Query["ref"] != stubTodoID || req.Query["dir"] != stubWorkspace {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "detail needs the todo's id and owning dir"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"attempts": stub.attempts})
	}))
	mux.HandleFunc("GET /api/todos/session/stats", stub.handle(func(w http.ResponseWriter, req recordedRequest) {
		if req.Query["sessionId"] == "" || req.Query["dir"] != stubWorkspace {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sessionId is required"})
			return
		}
		writeJSON(w, http.StatusOK, stub.stats)
	}))
	mux.HandleFunc("GET /api/todos/session/plan", stub.handle(func(w http.ResponseWriter, _ recordedRequest) {
		writeJSON(w, http.StatusOK, stub.plan)
	}))
	mux.HandleFunc("POST /api/todos/run", stub.handle(func(w http.ResponseWriter, _ recordedRequest) {
		w.Header().Set("Content-Type", stub.launchType)
		w.WriteHeader(stub.launchStatus)
		_, _ = io.WriteString(w, stub.launchBody)
	}))
	review := stub.handle(func(w http.ResponseWriter, _ recordedRequest) {
		writeJSON(w, stub.reviewStatus, stub.reviewBody)
	})
	mux.HandleFunc("POST /api/todos/plan/approve", review)
	mux.HandleFunc("POST /api/todos/plan/reject", review)
	stub.server = httptest.NewServer(mux)
	return stub
}

func (s *stubGavel) Close() { s.server.Close() }

func (s *stubGavel) URL() string { return s.server.URL }

// set mutates the stub's state under its lock, as the dashboard's state moves
// between observations.
func (s *stubGavel) set(fn func(*stubGavel)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s)
}

func (s *stubGavel) recorded(method, path string) []recordedRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []recordedRequest
	for _, req := range s.requests {
		if req.Method == method && req.Path == path {
			out = append(out, req)
		}
	}
	return out
}

func (s *stubGavel) handle(fn func(http.ResponseWriter, recordedRequest)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := recordedRequest{Method: r.Method, Path: r.URL.Path, Query: map[string]string{}, Accept: r.Header.Get("Accept")}
		for key := range r.URL.Query() {
			req.Query[key] = r.URL.Query().Get(key)
		}
		if r.Method == http.MethodPost {
			data, err := io.ReadAll(r.Body)
			if err != nil || json.Unmarshal(data, &req.Body) != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body: expected one JSON object"})
				return
			}
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		s.requests = append(s.requests, req)
		fn(w, req)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// sseFrames renders frames exactly as pr/ui todoLaunchStream.send does.
func sseFrames(frames ...[2]any) string {
	var out strings.Builder
	for _, frame := range frames {
		data, err := json.Marshal(frame[1])
		if err != nil {
			panic(err)
		}
		out.WriteString("event: " + frame[0].(string) + "\ndata: " + string(data) + "\n\n")
	}
	return out.String()
}

func newStep(kind, id string, env []string, config string) *fixtures.StepContext {
	return &fixtures.StepContext{
		Spec:  fixtures.StepSpec{ID: id, Kind: kind, Config: json.RawMessage(config)},
		Scope: fixtures.NewStepScope(env),
	}
}

func lookupKind(name string) fixtures.Kind {
	kind, err := fixtures.DefaultKinds.Lookup(name)
	if err != nil {
		panic(err)
	}
	return kind
}
