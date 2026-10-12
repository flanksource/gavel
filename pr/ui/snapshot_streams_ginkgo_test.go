package ui

import (
	"encoding/json"
	"path/filepath"
	"sync/atomic"
	"time"

	gavelctx "github.com/flanksource/gavel/context"
	"github.com/flanksource/gavel/github/activity"
	"github.com/flanksource/gavel/procfile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	// fastStreamCadence stands in for a stream's poll interval so specs observe
	// several ticks in well under a second.
	fastStreamCadence = 20 * time.Millisecond
	// idleStreamCadence is long enough that no tick lands inside a spec: any
	// frame that arrives was caused by a wake, not the ticker.
	idleStreamCadence = time.Hour
	frameWait         = 2 * time.Second
	pingTicks         = 4
)

var _ = Describe("activity stream", func() {
	BeforeEach(func() {
		activity.Shared().Reset()
		DeferCleanup(activity.Shared().Reset)
		useStreamInterval(&activityStreamInterval, fastStreamCadence)
	})

	It("sends the snapshot once, then only pings while it is unchanged", func() {
		frames := serveStream((&Server{}).handleActivityStream, "/api/activity/stream")

		got := takeStreamFrames(frames, 1+pingTicks, frameWait)

		Expect(got[0].Ping).To(BeFalse())
		Expect(got[0].Data).To(MatchJSON(`{"entries":[],"stats":{"total":0,"cacheHits":0,"errors":0,"totalBytes":0,"totalNs":0,"byKind":{}}}`))
		Expect(got[1:]).To(HaveEach(streamFrame{Ping: true}))
	})

	It("sends a new frame when a request is recorded", func() {
		const recordedURL = "https://api.example.com/repos/acme/widgets"
		frames := serveStream((&Server{}).handleActivityStream, "/api/activity/stream")
		Expect(nextStreamFrame(frames, frameWait).Ping).To(BeFalse())

		activity.Shared().Record(activity.Entry{Method: "GET", URL: recordedURL})

		Eventually(frames, frameWait).Should(Receive(WithTransform(func(f streamFrame) string { return f.Data }, ContainSubstring(recordedURL))))
	})
})

var _ = Describe("test runs stream", func() {
	It("sends the grouped runs once, then only pings while they are unchanged", func() {
		const projectName = "widgets"
		original := projectsPath
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		DeferCleanup(func() { projectsPath = original })
		dir := GinkgoT().TempDir()
		Expect(SaveProjects([]Project{{Name: projectName, Dir: dir, Repos: []string{"acme/widgets"}}})).To(Succeed())
		useStreamInterval(&testRunsStreamInterval, fastStreamCadence)

		frames := serveStream((&Server{}).handleTestRunsStream, "/api/tests/stream")
		got := takeStreamFrames(frames, 1+pingTicks, frameWait)

		var payload testRunsResponse
		Expect(json.Unmarshal([]byte(got[0].Data), &payload)).To(Succeed())
		Expect(payload.Projects).To(HaveLen(1))
		Expect(payload.Projects[0].Name).To(Equal(projectName))
		Expect(payload.Projects[0].Runs).To(BeEmpty())
		Expect(got[1:]).To(HaveEach(streamFrame{Ping: true}))
	})
})

var _ = Describe("proc status stream cadence", func() {
	// useCountingProcSampler serves status from a sampler that never caches, so
	// every load is a scan the spec can count.
	useCountingProcSampler := func(status string) *atomic.Int32 {
		var scans atomic.Int32
		original := sharedProcSampler
		sharedProcSampler = &procSampler{sample: func(gavelctx.Context) (map[string]procStatus, error) {
			scans.Add(1)
			return sampleWith(status), nil
		}}
		DeferCleanup(func() { sharedProcSampler = original })
		return &scans
	}

	BeforeEach(func() {
		useStreamInterval(&procStreamSteady, idleStreamCadence)
		useStreamInterval(&procStreamFast, fastStreamCadence)
	})

	It("re-samples at the fast cadence while a process is starting", func() {
		scans := useCountingProcSampler(procfile.StatusStarting)
		frames := serveStream((&Server{}).handleProcStatusStream, "/api/proc/status/stream")

		got := takeStreamFrames(frames, 1+pingTicks, frameWait)

		Expect(got[0].Ping).To(BeFalse())
		Expect(got[1:]).To(HaveEach(streamFrame{Ping: true}))
		Expect(scans.Load()).To(BeNumerically(">=", 1+pingTicks))
	})

	It("waits for the steady cadence while every process is settled", func() {
		scans := useCountingProcSampler(procfile.StatusRunning)
		frames := serveStream((&Server{}).handleProcStatusStream, "/api/proc/status/stream")
		Expect(nextStreamFrame(frames, frameWait).Ping).To(BeFalse())

		expectNoStreamFrame(frames, 10*fastStreamCadence)

		Expect(scans.Load()).To(Equal(int32(1)))
	})

	It("marks the dashboard as watched on every load so proc metrics keep sampling", func() {
		useCountingProcSampler(procfile.StatusRunning)
		server := &Server{}
		frames := serveStream(server.handleProcStatusStream, "/api/proc/status/stream")

		Expect(nextStreamFrame(frames, frameWait).Ping).To(BeFalse())

		server.mu.RLock()
		defer server.mu.RUnlock()
		Expect(server.lastProcPoll).To(BeTemporally("~", time.Now(), frameWait))
	})
})
