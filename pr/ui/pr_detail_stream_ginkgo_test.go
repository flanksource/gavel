package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/pr/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PR detail stream", func() {
	const (
		repo     = "acme/gavel"
		number   = 1
		oldTitle = "Old title"
		newTitle = "New title"
	)
	key := SyncStatusKey(repo, number)
	updatedAt := time.Date(2026, 11, 22, 9, 0, 0, 0, time.UTC)

	var (
		srv *Server
		gh  *detailGitHub
	)
	BeforeEach(func() {
		GinkgoT().Setenv("GAVEL_DB", "off")
		store, err := cache.Open()
		Expect(err).NotTo(HaveOccurred())
		gh = &detailGitHub{title: newTitle, updatedAt: updatedAt, runStatus: http.StatusOK}
		api := httptest.NewServer(gh)
		DeferCleanup(api.Close)
		GinkgoT().Setenv("GITHUB_API_URL", api.URL)
		srv = &Server{detailCache: NewDetailCache(), refreshCh: make(chan struct{}, 1), ghOpts: github.Options{Token: "test-token", Cache: store}}
	})
	cacheDetail := func(detail prDetail) {
		Expect(srv.detailCache.Put(key, srv.detailCache.Begin(), detail)).To(BeTrue())
	}
	stream := func(ctx context.Context, query string) []detailFrame {
		rec := httptest.NewRecorder()
		srv.handleDetail(rec, httptest.NewRequest("GET", "/api/prs/detail?repo=acme%2Fgavel&number=1"+query, nil).WithContext(ctx))
		return parseDetailFrames(rec.Body.String())
	}
	cachedTitle := func() string {
		entry, ok := srv.detailCache.Get(key)
		Expect(ok).To(BeTrue())
		return entry.Detail.PR.Title
	}

	// A PR with no actionable comments leaves prDetail.Comments nil, which Go
	// marshals as `null`. The UI validates the frame shape, so a null here
	// discards the whole PR frame and renders "invalid PR update" instead of
	// the PR. The runs and gavel frames are sent empty for the same reason the
	// UI needs them at all: it keeps what it last received.
	It("serves a current cached detail with every section, empty ones included", func() {
		cacheDetail(prDetail{PR: &model.PRInfo{Number: number, Title: oldTitle}})
		frames := stream(GinkgoT().Context(), "")
		Expect(frames).To(HaveLen(4))
		Expect(frames[0].Event).To(Equal("pr"))
		Expect(frames[0].Data).To(HaveKeyWithValue("comments", []any{}))
		Expect(frames[1]).To(Equal(detailFrame{"runs", map[string]any{"runs": map[string]any{}, "error": ""}}))
		Expect(frames[2]).To(Equal(detailFrame{"gavel", map[string]any{"gavelResults": []any{}}}))
		Expect(frames[3].Event).To(Equal("done"))
		Expect(gh.prRequests()).To(BeZero())
	})

	It("follows a stale cached detail with the current one", func() {
		cacheDetail(prDetail{PR: &model.PRInfo{Number: number, Title: oldTitle, UpdatedAt: updatedAt.Add(-time.Minute)}})
		srv.publishResults(github.PRSearchResults{{Repo: repo, Number: number, UpdatedAt: updatedAt}})
		Expect(titles(stream(GinkgoT().Context(), ""))).To(Equal([]string{oldTitle, newTitle}))
	})

	It("skips the cached detail and requests a list poll on a manual refresh", func() {
		cacheDetail(prDetail{PR: &model.PRInfo{Number: number, Title: oldTitle, UpdatedAt: updatedAt}})
		Expect(titles(stream(GinkgoT().Context(), "&refresh=1"))).To(Equal([]string{newTitle}))
		Expect(cachedTitle()).To(Equal(newTitle))
		Expect(srv.refreshCh).To(HaveLen(1))
	})

	It("keeps the last good detail and reports the failure when the PR cannot be fetched", func() {
		cacheDetail(prDetail{PR: &model.PRInfo{Number: number, Title: oldTitle}})
		gh.fail()
		frames := stream(GinkgoT().Context(), "&refresh=1")
		Expect(frames).To(HaveLen(2))
		Expect(frames[0].Event).To(Equal("error"))
		Expect(frames[0].Data["error"]).To(ContainSubstring("status 500"))
		Expect(cachedTitle()).To(Equal(oldTitle))
		status, _ := srv.detailCache.GetStatus(key)
		Expect(status.State).To(Equal(SyncError))
	})

	It("reports workflow runs that failed to load and fetches them again next time", func() {
		gh.checks, gh.runStatus = `{"__typename":"CheckRun","name":"test","status":"COMPLETED","conclusion":"SUCCESS","detailsUrl":"https://github.com/acme/gavel/actions/runs/101/job/7"}`, http.StatusForbidden
		frames := stream(GinkgoT().Context(), "&refresh=1")
		Expect(frames[1].Event).To(Equal("runs"))
		Expect(frames[1].Data["error"]).To(ContainSubstring("load workflow run 101"))
		status, _ := srv.detailCache.GetStatus(key)
		Expect(status.State).To(Equal(SyncError))

		gh.runStatus = http.StatusOK
		frames = stream(GinkgoT().Context(), "")
		Expect(frames[len(frames)-3].Data).To(And(HaveKeyWithValue("error", ""), HaveKeyWithValue("runs", HaveKey("101"))))
	})

	It("stays stale when the PR changes after it was fetched", func() {
		later := updatedAt.Add(time.Minute)
		gh.onPR = func(int) {
			srv.publishResults(github.PRSearchResults{{Repo: repo, Number: number, UpdatedAt: later}})
		}
		stream(GinkgoT().Context(), "&refresh=1")
		Expect(srv.detailCache.IsStale(key, later)).To(BeTrue())
	})

	It("does not let a cancelled refresh overwrite a newer one", func() {
		gh.title = oldTitle
		entered, release := make(chan struct{}), make(chan struct{})
		gh.onPR = func(call int) {
			if call == 1 {
				close(entered)
				<-release
			}
		}
		ctx, cancel := context.WithCancel(GinkgoT().Context())
		finished := make(chan struct{})
		go func() {
			defer close(finished)
			stream(ctx, "&refresh=1")
		}()
		Eventually(entered).Should(BeClosed())
		cancel()
		Eventually(finished).Should(BeClosed(), "the handler returns once its request is cancelled")
		gh.setTitle(newTitle)
		stream(GinkgoT().Context(), "&refresh=1")
		close(release)
		Expect(cachedTitle()).To(Equal(newTitle))
	})

	It("rejects a fetch begun before a newer one or an invalidation", func() {
		older, newer := srv.detailCache.Begin(), srv.detailCache.Begin()
		Expect(srv.detailCache.Put(key, newer, prDetail{PR: &model.PRInfo{Title: newTitle}})).To(BeTrue())
		Expect(srv.detailCache.Put(key, older, prDetail{PR: &model.PRInfo{Title: oldTitle}})).To(BeFalse())
		Expect(cachedTitle()).To(Equal(newTitle))

		inFlight := srv.detailCache.Begin()
		srv.detailCache.Invalidate(key)
		Expect(srv.detailCache.Put(key, inFlight, prDetail{PR: &model.PRInfo{Title: oldTitle}})).To(BeFalse())
	})

	It("clears the PR's gavel badge once it has no results", func() {
		srv.setGavelSummary(repo, number, &GavelResultsSummary{TestsFailed: 1})
		stream(GinkgoT().Context(), "&refresh=1")
		Expect(srv.gavelCache).To(BeEmpty())
	})

	It("marks results from a run that is not the head's outdated", func() {
		gh.comments = `{"databaseId":1,"body":"<!-- sticky-comment:gavel --> [View full results](https://github.com/acme/gavel/actions/runs/99/artifacts/4242)"}`
		frames := stream(GinkgoT().Context(), "&refresh=1")
		results := frames[2].Data["gavelResults"].([]any)
		Expect(results).To(ConsistOf(HaveKeyWithValue("outdated", true)))
	})
})

// detailGitHub answers the PR query and every workflow run request.
type detailGitHub struct {
	mu        sync.Mutex
	title     string
	updatedAt time.Time
	checks    string
	comments  string
	runStatus int
	failing   bool
	calls     int
	onPR      func(call int)
}

func (g *detailGitHub) setTitle(title string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.title = title
}

func (g *detailGitHub) fail() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.failing = true
}

func (g *detailGitHub) prRequests() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.calls
}

func (g *detailGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/graphql" {
		w.WriteHeader(g.runStatus)
		if strings.HasSuffix(r.URL.Path, "/jobs") {
			fmt.Fprint(w, `{"total_count":1,"jobs":[{"id":7,"run_id":101,"name":"test","status":"completed","conclusion":"success"}]}`)
			return
		}
		fmt.Fprint(w, `{"id":101,"name":"CI","status":"completed","conclusion":"success"}`)
		return
	}
	g.mu.Lock()
	g.calls++
	call, title, failing := g.calls, g.title, g.failing
	g.mu.Unlock()
	if failing {
		http.Error(w, "unavailable", http.StatusInternalServerError)
		return
	}
	if g.onPR != nil {
		g.onPR(call)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data":{"repository":{"pullRequest":{"id":"P","number":1,"title":%q,"state":"OPEN","updatedAt":%q,"comments":{"nodes":[%s]},"commits":{"nodes":[{"commit":{"id":"C","statusCheckRollup":{"contexts":{"nodes":[%s]}}}}]}}}}}`,
		title, g.updatedAt.Format(time.RFC3339), g.comments, g.checks)
}

type detailFrame struct {
	Event string
	Data  map[string]any
}

func parseDetailFrames(body string) []detailFrame {
	GinkgoHelper()
	var frames []detailFrame
	for block := range strings.SplitSeq(strings.TrimSpace(body), "\n\n") {
		event, data, ok := strings.Cut(block, "\ndata: ")
		Expect(ok).To(BeTrue(), block)
		frame := detailFrame{Event: strings.TrimPrefix(event, "event: ")}
		Expect(json.Unmarshal([]byte(data), &frame.Data)).To(Succeed())
		frames = append(frames, frame)
	}
	return frames
}

// titles lists the title of each pr frame.
func titles(frames []detailFrame) []string {
	var out []string
	for _, frame := range frames {
		if frame.Event == "pr" {
			out = append(out, frame.Data["pr"].(map[string]any)["title"].(string))
		}
	}
	return out
}
