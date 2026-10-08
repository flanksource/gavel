package ui

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/flanksource/clicky/sse"
	"github.com/flanksource/gavel/procfile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	eventStreamAccept = "text/event-stream"
	browserFetchMode  = "cors"
	// directStreamRefusal is the body clicky's hub guard answers a refused
	// direct browser stream with, for the dashboard's default hub prefix.
	directStreamRefusal = "stream routes are served through " + sse.DefaultPrefix + "; reload the page\n"
)

func browserStreamHeader() http.Header {
	return http.Header{"Accept": {eventStreamAccept}, "Sec-Fetch-Mode": {browserFetchMode}}
}

func useProcFixture() map[string]procStatus {
	fixture := sampleWith(procfile.StatusRunning)
	original := sharedProcSampler
	sharedProcSampler = &procSampler{ttl: time.Minute, sample: func() (map[string]procStatus, error) { return fixture, nil }}
	DeferCleanup(func() { sharedProcSampler = original })
	return fixture
}

func newDashboardTestServer(s *Server) *httptest.Server {
	server := httptest.NewServer(s.Handler())
	DeferCleanup(server.Close)
	return server
}

// doRequest sends req and returns its response once the headers arrived,
// cancelling the request when the spec ends so an open stream never outlives it.
func doRequest(req *http.Request) *http.Response {
	ctx, cancel := context.WithCancel(req.Context())
	DeferCleanup(cancel)
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(resp.Body.Close)
	return resp
}

func newRequest(method, url string, header http.Header, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	Expect(err).NotTo(HaveOccurred())
	for name, values := range header {
		req.Header[name] = values
	}
	return req
}

// dashboardEvents is one open /api/events connection on the dashboard.
type dashboardEvents struct {
	base   string
	conn   string
	build  string
	frames <-chan streamFrame
}

// openDashboardEvents opens the dashboard's events stream and reads its hello.
func openDashboardEvents(base string) dashboardEvents {
	frames := openStream(base+sse.DefaultPrefix, nil)
	hello := nextStreamFrame(frames, 5*time.Second)
	Expect(hello.Event).To(Equal("__hello"))
	var payload struct {
		Conn  string `json:"conn"`
		Build string `json:"build"`
	}
	Expect(json.Unmarshal([]byte(hello.Data), &payload)).To(Succeed())
	Expect(payload.Conn).NotTo(BeEmpty())
	return dashboardEvents{base: base, conn: payload.Conn, build: payload.Build, frames: frames}
}

// subscribe POSTs a sub the way the browser bundle does, fetch metadata included.
func (e dashboardEvents) subscribe(id, path string) *http.Response {
	body, err := json.Marshal(map[string]string{"id": id, "path": path})
	Expect(err).NotTo(HaveOccurred())
	header := http.Header{"Content-Type": {"application/json"}, "Sec-Fetch-Mode": {browserFetchMode}}
	return doRequest(newRequest(http.MethodPost, e.base+sse.DefaultPrefix+"/"+e.conn+"/subs", header, strings.NewReader(string(body))))
}

// nextDataFrame skips hub pings until the next dispatched event.
func (e dashboardEvents) nextDataFrame() streamFrame {
	for {
		if frame := nextStreamFrame(e.frames, 5*time.Second); !frame.Ping {
			return frame
		}
	}
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

var _ = Describe("dashboard stream routes behind the events hub", func() {
	DescribeTable("refuses a native browser EventSource GET on a hub-servable stream route with 410",
		func(path string) {
			useProcFixture()
			server := newDashboardTestServer(&Server{})

			resp := doRequest(newRequest(http.MethodGet, server.URL+path, browserStreamHeader(), nil))

			Expect(resp.StatusCode).To(Equal(http.StatusGone))
			Expect(readBody(resp)).To(Equal(directStreamRefusal))
		},
		Entry("PR list stream", "/api/prs/stream"),
		Entry("activity stream", "/api/activity/stream"),
		Entry("test runs stream", "/api/tests/stream"),
		Entry("proc status stream", "/api/proc/status/stream"),
		Entry("clicky task runs stream", "/api/v1/tasks/runs/stream"),
	)

	It("serves the refused proc stream through a hub sub opened with browser fetch metadata", func() {
		fixture := useProcFixture()
		server := newDashboardTestServer(&Server{})
		events := openDashboardEvents(server.URL)

		Expect(events.subscribe("proc", "/api/proc/status/stream").StatusCode).To(Equal(http.StatusNoContent))

		want, err := json.Marshal(fixture)
		Expect(err).NotTo(HaveOccurred())
		frame := events.nextDataFrame()
		Expect(frame.Event).To(Equal("proc/message"))
		Expect(frame.Data).To(MatchJSON(want))
	})

	DescribeTable("serves direct GETs that are not browser stream loads",
		func(path string, header http.Header, wantContentType string) {
			useProcFixture()
			server := newDashboardTestServer(&Server{})

			resp := doRequest(newRequest(http.MethodGet, server.URL+path, header, nil))

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("Content-Type")).To(HavePrefix(wantContentType))
		},
		Entry("a CLI stream client without Sec-Fetch-Mode",
			"/api/proc/status/stream", http.Header{"Accept": {eventStreamAccept}}, eventStreamAccept),
		Entry("a browser JSON fetch of a non-stream route",
			"/api/proc/status", http.Header{"Accept": {"application/json"}, "Sec-Fetch-Mode": {browserFetchMode}}, "application/json"),
		Entry("the multiplexed events stream itself",
			sse.DefaultPrefix, browserStreamHeader(), eventStreamAccept),
	)

	It("leaves a browser POST streaming launch route to its handler", func() {
		server := newDashboardTestServer(&Server{})

		resp := doRequest(newRequest(http.MethodPost, server.URL+"/api/todos/run", browserStreamHeader(), strings.NewReader("not json")))

		Expect(resp.StatusCode).To(Equal(http.StatusBadRequest), "the launch handler's own decode error, not the guard")
		Expect(readBody(resp)).NotTo(ContainSubstring(directStreamRefusal))
	})
})

var _ = Describe("dashboard UI build identity", func() {
	metaPattern := regexp.MustCompile(`<meta name="gavel-ui-build" content="([^"]*)">`)

	It("stamps the embedded bundle hash into the events hello frame", func() {
		server := newDashboardTestServer(&Server{})

		Expect(openDashboardEvents(server.URL).build).To(Equal(expectedUIBuildID()))
	})

	DescribeTable("stamps the same hash into every served SPA page's head",
		func(path string) {
			server := newDashboardTestServer(&Server{})
			header := http.Header{"Accept": {"text/html"}, "Sec-Fetch-Mode": {"navigate"}}

			resp := doRequest(newRequest(http.MethodGet, server.URL+path, header, nil))

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			head, _, found := strings.Cut(readBody(resp), "</head>")
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
		server := newDashboardTestServer(dev)

		Expect(openDashboardEvents(server.URL).build).To(Equal(devUIBuildID))
	})

	It("stamps the dev build id into the Vite dev entry page so a dev tab never reloads", func() {
		page, err := os.ReadFile("index.html")
		Expect(err).NotTo(HaveOccurred())
		head, _, found := strings.Cut(string(page), "</head>")
		Expect(found).To(BeTrue())
		Expect(head).To(ContainSubstring(`<meta name="gavel-ui-build" content="` + devUIBuildID + `" />`))
	})
})
