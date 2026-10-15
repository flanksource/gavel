// Package githubtest runs a fake GitHub API for one GitHub App: installation tokens, the
// installations and repositories it grants, and which installations each user token sees.
package githubtest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/gavel/github"
)

// TB is the part of testing.TB the server uses, so Ginkgo's GinkgoT() works as well.
type TB interface {
	Helper()
	Fatal(args ...any)
	Cleanup(func())
	Setenv(key, value string)
	TempDir() string
}

// AppID is the fake App's ID.
const AppID = 123

// Server is the fake GitHub API. It sets GITHUB_API_URL for the test.
type Server struct {
	*httptest.Server
	// KeyPath is a PEM file with the App's private key, for github.LoadApp.
	KeyPath string

	mu            sync.Mutex
	installations map[int64]installation
	users         map[string][]int64
	pulls         []PullRequest
	searches      []Search
}

// PullRequest is an open pull request the GraphQL search returns.
type PullRequest struct {
	Repo   string
	Number int
	Title  string
	Author string
}

// Search is one GraphQL PR search: the token it used and its search string.
type Search struct {
	Token, Query string
}

// AddPullRequest makes pr show up in searches that name its repository.
func (s *Server) AddPullRequest(pr PullRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pulls = append(s.pulls, pr)
}

// Searches lists the PR searches served, in order.
func (s *Server) Searches() []Search {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Search(nil), s.searches...)
}

type installation struct {
	account github.InstallationAccount
	repos   []github.InstallationRepo
}

// New starts the server for the test and points the GitHub client at it.
func New(t TB) *Server {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(t.TempDir(), "github-app.pem")
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}), 0o600); err != nil {
		t.Fatal(err)
	}
	s := &Server{KeyPath: keyPath, installations: map[int64]installation{}, users: map[string][]int64{}}
	s.Server = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.Close)
	t.Setenv("GITHUB_API_URL", s.URL)
	return s
}

// App loads the fake App from KeyPath.
func (s *Server) App(t TB) *github.App {
	t.Helper()
	app, err := github.LoadApp(AppID, "gavel-test", s.KeyPath)
	if err != nil {
		t.Fatal(err)
	}
	return app
}

// Install records an installation on account that grants repos.
func (s *Server) Install(id int64, account github.InstallationAccount, repos ...github.InstallationRepo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.installations[id] = installation{account: account, repos: repos}
}

// GrantUser lets a user token see installations, as GET /user/installations does.
func (s *Server) GrantUser(userToken string, installations ...int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[userToken] = installations
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	bearer := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	switch {
	case r.Method == http.MethodPost && len(parts) == 4 && parts[0] == "app" && parts[3] == "access_tokens":
		id, _ := strconv.ParseInt(parts[2], 10, 64)
		if _, ok := s.installations[id]; !ok {
			notFound(w)
			return
		}
		writeJSON(w, map[string]any{"token": fmt.Sprintf("ghs_%d", id), "expires_at": time.Now().Add(time.Hour)})
	case r.Method == http.MethodGet && len(parts) == 3 && parts[0] == "app" && parts[1] == "installations":
		id, _ := strconv.ParseInt(parts[2], 10, 64)
		installation, ok := s.installations[id]
		if !ok {
			notFound(w)
			return
		}
		writeJSON(w, github.Installation{ID: id, Account: installation.account})
	case r.Method == http.MethodGet && r.URL.Path == "/installation/repositories":
		id, _ := strconv.ParseInt(strings.TrimPrefix(bearer, "ghs_"), 10, 64)
		repos := s.installations[id].repos
		if r.URL.Query().Get("page") != "1" {
			repos = nil
		}
		writeJSON(w, map[string]any{"total_count": len(s.installations[id].repos), "repositories": emptyIfNil(repos)})
	case r.Method == http.MethodGet && r.URL.Path == "/user/installations":
		ids, ok := s.users[bearer]
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			writeJSON(w, map[string]string{"message": "Bad credentials"})
			return
		}
		installs := []map[string]any{}
		for _, id := range ids {
			installs = append(installs, map[string]any{"id": id, "app_id": AppID, "account": s.installations[id].account})
		}
		if r.URL.Query().Get("page") != "1" {
			installs = []map[string]any{}
		}
		writeJSON(w, map[string]any{"total_count": len(ids), "installations": installs})
	case r.Method == http.MethodPost && r.URL.Path == "/graphql":
		s.search(w, r, bearer)
	default:
		notFound(w)
	}
}

// search answers a PR search with the added pull requests of every repo:<owner>/<name>
// qualifier in its query.
func (s *Server) search(w http.ResponseWriter, r *http.Request, token string) {
	var body struct {
		Variables struct {
			Query string `json:"query"`
		} `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.searches = append(s.searches, Search{Token: token, Query: body.Variables.Query})
	nodes := []map[string]any{}
	for _, pr := range s.pulls {
		if !strings.Contains(" "+body.Variables.Query+" ", " repo:"+pr.Repo+" ") {
			continue
		}
		nodes = append(nodes, map[string]any{
			"number": pr.Number, "title": pr.Title, "author": map[string]any{"login": pr.Author},
			"repository": map[string]any{"nameWithOwner": pr.Repo, "owner": map[string]any{}},
			"state":      "OPEN", "url": fmt.Sprintf("https://github.com/%s/pull/%d", pr.Repo, pr.Number),
			"updatedAt": time.Now().UTC().Format(time.RFC3339), "commits": map[string]any{"nodes": []any{}},
		})
	}
	writeJSON(w, map[string]any{"data": map[string]any{"search": map[string]any{"issueCount": len(nodes), "nodes": nodes}}})
}

func emptyIfNil(repos []github.InstallationRepo) []github.InstallationRepo {
	if repos == nil {
		return []github.InstallationRepo{}
	}
	return repos
}

func notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	writeJSON(w, map[string]string{"message": "Not Found"})
}

func writeJSON(w http.ResponseWriter, value any) {
	_ = json.NewEncoder(w).Encode(value)
}
