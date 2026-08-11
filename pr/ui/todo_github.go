package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/todos/githubpush"
	"github.com/flanksource/gavel/verify"
)

type todoGitHubPushPayload struct {
	Ref string `json:"ref"`
	Dir string `json:"dir,omitempty"`
	// BaseURL overrides the project's todos.baseUrl for this push only.
	BaseURL string `json:"baseUrl,omitempty"`
	Force   bool   `json:"force,omitempty"`
	// Update rewrites the issue the todo is already linked to; Issue names one
	// specific issue to rewrite and link.
	Update bool   `json:"update,omitempty"`
	Issue  string `json:"issue,omitempty"`
}

type todoGitHubPushResponse struct {
	Todo    todoSummary `json:"todo"`
	Repo    string      `json:"repo"`
	Number  int         `json:"number"`
	URL     string      `json:"url"`
	Alias   string      `json:"alias"`
	Updated bool        `json:"updated"`
}

// handleTodoGitHubPush opens a GitHub issue for one TODO and links the two.
func (s *Server) handleTodoGitHubPush(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload todoGitHubPushPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("invalid json"))
		return
	}
	if strings.TrimSpace(payload.Ref) == "" {
		writeTodoError(w, http.StatusBadRequest, fmt.Errorf("ref is required"))
		return
	}
	provider, src, todo, err := s.resolveTodoReference(r.Context(), todoSource{Dir: payload.Dir}, payload.Ref)
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}

	baseURL, err := s.resolveTodoPushBaseURL(payload.BaseURL, src.Dir, requestOrigin(r))
	if err != nil {
		writeTodoError(w, http.StatusBadRequest, err)
		return
	}

	// The dashboard's own Repo is the PR-search target, which is unrelated to
	// the workspace this todo lives in. Clearing it lets the push resolve the
	// repo from that workspace's own origin remote.
	ghOpts := s.ghOpts
	ghOpts.WorkDir = src.Dir
	ghOpts.Repo = ""

	result, err := githubpush.Push(r.Context(), provider, todo.ID, githubpush.Options{
		GitHub:  ghOpts,
		BaseURL: baseURL,
		Force:   payload.Force,
		Update:  payload.Update,
		Issue:   payload.Issue,
		Labels:  true,
		Plan:    true,
	})
	if err != nil {
		writeTodoError(w, todoGitHubPushStatus(err), err)
		return
	}

	json.NewEncoder(w).Encode(todoGitHubPushResponse{ //nolint:errcheck
		Todo:    summarizeTodo(result.TODO, true),
		Repo:    result.Repo,
		Number:  result.Number,
		URL:     result.URL,
		Alias:   result.Alias,
		Updated: result.Updated,
	})
}

// resolveTodoPushBaseURL prefers an explicit request value, then the project
// config, and only then the origin this request arrived on. The request origin
// is last because the dashboard is usually reached over loopback, which GitHub
// cannot fetch.
func (s *Server) resolveTodoPushBaseURL(requested, dir, origin string) (string, error) {
	var configured string
	if config, err := verify.LoadGavelConfig(dir); err == nil {
		configured = config.Todos.BaseURL
	} else {
		logger.Debugf("load .gavel.yaml in %s: %v", dir, err)
	}
	baseURL, err := githubpush.ResolveBaseURL(requested, configured, origin)
	if err != nil {
		return "", err
	}
	if githubpush.IsLoopback(baseURL) {
		logger.Warnf("base URL %s is loopback: attachment images will only render for viewers on this machine", baseURL)
	}
	return baseURL, nil
}

func todoGitHubPushStatus(err error) int {
	if errors.Is(err, githubpush.ErrAlreadyLinked) {
		return http.StatusConflict
	}
	return http.StatusBadGateway
}
