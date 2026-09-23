package ui

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/flanksource/gavel/procfile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	eventStreamAccept = "text/event-stream"
	browserFetchMode  = "cors"
)

// browserStreamRequest is what a native EventSource sends: a GET with an
// event-stream Accept and the Sec-Fetch-* metadata every browser attaches.
func browserStreamRequest(ctx context.Context, method, url string) *http.Request {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Accept", eventStreamAccept)
	req.Header.Set("Sec-Fetch-Mode", browserFetchMode)
	return req
}

func useProcFixture() map[string]procStatus {
	fixture := sampleWith(procfile.StatusRunning)
	original := sharedProcSampler
	sharedProcSampler = &procSampler{ttl: time.Minute, sample: func() (map[string]procStatus, error) { return fixture, nil }}
	DeferCleanup(func() { sharedProcSampler = original })
	return fixture
}

func newDashboardTestServer() *httptest.Server {
	server := httptest.NewServer((&Server{}).Handler())
	DeferCleanup(server.Close)
	return server
}

// doHeaders sends req and returns its response once the headers arrived,
// cancelling the request afterwards so an open stream does not outlive the spec.
func doHeaders(req *http.Request) *http.Response {
	ctx, cancel := context.WithCancel(req.Context())
	DeferCleanup(cancel)
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(resp.Body.Close)
	return resp
}

// expectedUIBuildID hashes the bundle files straight off disk, independently of
// the embed the server hashes, so the two can only agree on the same bytes.
func expectedUIBuildID() string {
	hash := sha256.New()
	for _, name := range []string{"dist/prui.js", "dist/prui.css"} {
		data, err := os.ReadFile(name)
		Expect(err).NotTo(HaveOccurred())
		hash.Write(data)
	}
	return hex.EncodeToString(hash.Sum(nil))[:12]
}

var _ = Describe("direct browser stream guard", func() {
	DescribeTable("refuses a native EventSource GET on a hub-servable stream route with 410",
		func(path string) {
			useProcFixture()
			server := newDashboardTestServer()

			resp := doHeaders(browserStreamRequest(context.Background(), http.MethodGet, server.URL+path))

			Expect(resp.StatusCode).To(Equal(http.StatusGone))
			Expect(resp.Header.Get("Content-Type")).To(HavePrefix("text/plain"))
			Expect(readBody(resp)).To(Equal(directStreamRefusal + "\n"))
		},
		Entry("proc status stream", "/api/proc/status/stream"),
		Entry("clicky task runs stream", "/api/v1/tasks/runs/stream"),
		Entry("clicky task stream", "/api/v1/tasks/stream"),
		Entry("PR list stream", "/api/prs/stream"),
	)

	It("serves the refused route through a hub sub opened with browser fetch metadata", func() {
		fixture := useProcFixture()
		server := newDashboardTestServer()
		client := openEvents(server.URL)

		body := strings.NewReader(`{"id":"proc","path":"/api/proc/status/stream"}`)
		subscribe, err := http.NewRequest(http.MethodPost, server.URL+"/api/events/"+client.conn+"/subs", body)
		Expect(err).NotTo(HaveOccurred())
		subscribe.Header.Set("Content-Type", "application/json")
		subscribe.Header.Set("Sec-Fetch-Mode", browserFetchMode)
		Expect(doHeaders(subscribe).StatusCode).To(Equal(http.StatusNoContent))

		want, err := json.Marshal(fixture)
		Expect(err).NotTo(HaveOccurred())
		frame := client.nextFrame("proc/")
		Expect(frame.Event).To(Equal("proc/message"))
		Expect(frame.Data).To(MatchJSON(want))
	})

	DescribeTable("serves direct GETs that are not browser stream loads",
		func(path string, header http.Header, wantContentType string) {
			useProcFixture()
			server := newDashboardTestServer()
			req, err := http.NewRequest(http.MethodGet, server.URL+path, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header = header

			resp := doHeaders(req)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("Content-Type")).To(HavePrefix(wantContentType))
		},
		Entry("a CLI stream client without Sec-Fetch-Mode",
			"/api/proc/status/stream", http.Header{"Accept": {eventStreamAccept}}, eventStreamAccept),
		Entry("a CLI client on a clicky task stream",
			"/api/v1/tasks/runs/stream", http.Header{"Accept": {eventStreamAccept}}, eventStreamAccept),
		Entry("a browser JSON fetch of a non-stream route",
			"/api/proc/status", http.Header{"Accept": {"application/json"}, "Sec-Fetch-Mode": {browserFetchMode}}, "application/json"),
		Entry("the multiplexed events stream itself",
			"/api/events", http.Header{"Accept": {eventStreamAccept}, "Sec-Fetch-Mode": {browserFetchMode}}, eventStreamAccept),
	)

	It("leaves a browser POST streaming launch route to its handler", func() {
		server := newDashboardTestServer()
		req := browserStreamRequest(context.Background(), http.MethodPost, server.URL+"/api/todos/run")
		req.Body = io.NopCloser(strings.NewReader("not json"))

		resp := doHeaders(req)

		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest), "the launch handler's own decode error, not the guard")
		Expect(readBody(resp)).NotTo(ContainSubstring(directStreamRefusal))
	})

	It("marks the request the hub dispatches for a sub", func() {
		server := newEventsTestServer(map[string]http.HandlerFunc{
			"GET /api/test/marker": func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", eventStreamAccept)
				fmt.Fprintf(w, "data: %t\n\n", isEventsSub(r.Context()))
			},
		})
		client := openEvents(server.URL)

		Expect(client.subscribe("m", "/api/test/marker").StatusCode).To(Equal(http.StatusNoContent))

		Expect(client.nextFrame("m/").Data).To(Equal("true"))
	})

	DescribeTable("decides per request whether the guard refuses it",
		func(method string, header http.Header, path string, hubSub bool, wantRefused bool) {
			served := false
			guard := refuseDirectBrowserStreams(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				served = true
				w.WriteHeader(http.StatusOK)
			}))
			ctx := context.Background()
			if hubSub {
				ctx = withEventsSub(ctx)
			}
			req := httptest.NewRequestWithContext(ctx, method, path, nil)
			req.Header = header
			rec := httptest.NewRecorder()

			guard.ServeHTTP(rec, req)

			Expect(served).To(Equal(!wantRefused))
			if wantRefused {
				Expect(rec.Code).To(Equal(http.StatusGone))
			}
		},
		Entry("a browser EventSource GET is refused",
			http.MethodGet, browserStreamHeader(), "/api/prs/stream", false, true),
		Entry("an Accept list naming event-stream among others is refused",
			http.MethodGet, http.Header{"Accept": {"application/json, text/event-stream"}, "Sec-Fetch-Mode": {browserFetchMode}}, "/api/prs/stream", false, true),
		Entry("a hub sub carrying every browser header is served (context marker, not a header)",
			http.MethodGet, browserStreamHeader(), "/api/prs/stream", true, false),
		Entry("a browser POST with an event-stream Accept is served",
			http.MethodPost, browserStreamHeader(), "/api/todos/run", false, false),
		Entry("a browser EventSource GET outside /api is served (results page bundle)",
			http.MethodGet, browserStreamHeader(), "/results/acme/widgets/7/api/tests/stream", false, false),
		Entry("the events stream itself is served",
			http.MethodGet, browserStreamHeader(), "/api/events", false, false),
	)
})

func browserStreamHeader() http.Header {
	return http.Header{"Accept": {eventStreamAccept}, "Sec-Fetch-Mode": {browserFetchMode}}
}

var _ = Describe("dashboard UI build identity", func() {
	metaPattern := regexp.MustCompile(`<meta name="gavel-ui-build" content="([^"]*)">`)

	It("stamps the embedded bundle hash into the events hello frame", func() {
		server := newDashboardTestServer()

		Expect(openEvents(server.URL).build).To(Equal(expectedUIBuildID()))
	})

	DescribeTable("stamps the same hash into every served SPA page's head",
		func(path string) {
			server := newDashboardTestServer()

			resp := doHeaders(navigationRequest(server.URL + path))

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			page := readBody(resp)
			head, _, found := strings.Cut(page, "</head>")
			Expect(found).To(BeTrue())
			Expect(metaPattern.FindStringSubmatch(head)).To(Equal([]string{
				`<meta name="gavel-ui-build" content="` + expectedUIBuildID() + `">`,
				expectedUIBuildID(),
			}))
		},
		Entry("the PR dashboard", "/prs"),
		Entry("the menubar page", "/menubar"),
		Entry("the processes page", "/processes"),
	)

	It("reports the dev build id in the hello frame when the UI is proxied to Vite", func() {
		dev := &Server{}
		Expect(dev.SetDevProxy("http://127.0.0.1:1")).To(Succeed())
		server := httptest.NewServer(dev.Handler())
		DeferCleanup(server.Close)

		Expect(openEvents(server.URL).build).To(Equal(devUIBuildID))
	})

	It("stamps the dev build id into the Vite dev entry page so a dev tab never reloads", func() {
		page, err := os.ReadFile("index.html")
		Expect(err).NotTo(HaveOccurred())
		head, _, found := strings.Cut(string(page), "</head>")
		Expect(found).To(BeTrue())
		Expect(head).To(ContainSubstring(`<meta name="gavel-ui-build" content="` + devUIBuildID + `" />`))
	})
})

// navigationRequest is a browser page load: an HTML Accept and navigate mode.
func navigationRequest(url string) *http.Request {
	req := browserStreamRequest(context.Background(), http.MethodGet, url)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	return req
}
