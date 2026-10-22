// Tests a PR's client and source branch inside a sandbox job: with no GitHub credentials there,
// status, thread resolution and pushes go to serve's run API for the job's run.
package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fakeRunAPI is serve's run API for one run's PR: its head, its status, and what was resolved
// and pushed.
type fakeRunAPI struct {
	mu       sync.Mutex
	head     string
	resolved []string
	pushes   []string
}

func (f *fakeRunAPI) handler() http.Handler {
	mux := http.NewServeMux()
	auth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer run-token" {
				http.Error(w, `{"error":"bad token"}`, http.StatusUnauthorized)
				return
			}
			next(w, r)
		}
	}
	mux.HandleFunc("GET /api/runs/run-1/pr/status", auth(func(w http.ResponseWriter, r *http.Request) {
		Expect(r.URL.Query().Get("number")).To(Equal("7"))
		Expect(r.URL.Query().Get("threads")).To(Equal("true"))
		f.mu.Lock()
		defer f.mu.Unlock()
		_ = json.NewEncoder(w).Encode(model.StatusSnapshot{PR: &model.PRInfo{Number: 7, State: "OPEN", HeadRefOID: f.head}})
	}))
	mux.HandleFunc("GET /api/runs/run-1/pr/head", auth(func(w http.ResponseWriter, _ *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]string{"sha": f.head})
	}))
	mux.HandleFunc("POST /api/runs/run-1/pr/threads/{thread}/resolve", auth(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.resolved = append(f.resolved, r.PathValue("thread"))
		w.WriteHeader(http.StatusAccepted)
	}))
	mux.HandleFunc("POST /api/runs/run-1/pr/push", auth(func(w http.ResponseWriter, r *http.Request) {
		bundle, _ := io.ReadAll(r.Body)
		dir := GinkgoT().TempDir()
		Expect(os.WriteFile(filepath.Join(dir, "push.bundle"), bundle, 0o600)).To(Succeed())
		heads, err := targetGit(context.Background(), dir, "bundle", "list-heads", "push.bundle")
		Expect(err).NotTo(HaveOccurred())
		sha, _, _ := strings.Cut(heads, " ")
		f.mu.Lock()
		defer f.mu.Unlock()
		Expect(r.URL.Query().Get("base")).To(Equal(f.head), "the push names the head it continues")
		f.pushes = append(f.pushes, sha)
		f.head = sha
		_ = json.NewEncoder(w).Encode(map[string]string{"sha": sha})
	}))
	return mux
}

var _ = Describe("a PR inside a sandbox job", func() {
	var (
		fake *fakeRunAPI
		work string
		ctx  context.Context
	)
	BeforeEach(func() {
		ctx = context.Background()
		work = GinkgoT().TempDir()
		for _, args := range [][]string{{"init", "-q", "-b", "topic"}, {"config", "user.name", "Fixture User"}, {"config", "user.email", "fixture@example.com"}, {"commit", "-q", "--allow-empty", "-m", "first"}} {
			_, err := targetGit(ctx, work, args...)
			Expect(err).NotTo(HaveOccurred())
		}
		first, err := Revision(ctx, work, "HEAD")
		Expect(err).NotTo(HaveOccurred())
		fake = &fakeRunAPI{head: first}
		server := httptest.NewServer(fake.handler())
		DeferCleanup(server.Close)
		GinkgoT().Setenv("GAVEL_SERVER", server.URL)
		GinkgoT().Setenv("GAVEL_RUN_TOKEN", "run-token")
		GinkgoT().Setenv("GAVEL_RUN_ID", "run-1")
	})

	It("reads the PR's status and resolves its threads through serve", func() {
		client, number, err := ResolveURL(Options{Context: ctx}, "https://github.com/acme/app/pull/7")
		Expect(err).NotTo(HaveOccurred())
		Expect(number).To(Equal(7))
		snapshot, err := client.Status(ctx, number, model.StatusOptions{Threads: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(snapshot.PR.HeadRefOID).To(Equal(fake.head))

		Expect(client.ResolveThread(ctx, number, "thread-1")).To(Succeed())
		Expect(fake.resolved).To(Equal([]string{"thread-1"}))

		Expect(client.MergePR(ctx, number, model.MergeInput{})).To(MatchError("merging a PR isn't available inside a sandbox job"))
	})

	It("pushes the source branch through serve, continuing the head it saw", func() {
		target := &RunTarget{RepositoryURL: "https://github.com/fork/app.git", Branch: "topic", HeadSHA: fake.head}
		Expect(target.FetchHead(ctx, work)).To(Succeed())

		_, err := targetGit(ctx, work, "commit", "-q", "--allow-empty", "-m", "second")
		Expect(err).NotTo(HaveOccurred())
		second, err := Revision(ctx, work, "HEAD")
		Expect(err).NotTo(HaveOccurred())
		pushed, err := target.PushHead(ctx, work)
		Expect(err).NotTo(HaveOccurred())
		Expect(pushed).To(Equal(second))
		Expect(fake.pushes).To(Equal([]string{second}))
		Expect(target.HeadSHA).To(Equal(second))

		fake.head = strings.Repeat("a", 40)
		Expect(target.FetchHead(ctx, work)).To(MatchError(ContainSubstring("PR source branch advanced from " + second)))
	})
})
