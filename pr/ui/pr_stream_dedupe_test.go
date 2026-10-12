package ui

import (
	"time"

	"github.com/flanksource/gavel/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The PR stream's ticker is a liveness cadence, not a change signal: before the
// dedupe it re-marshalled and re-sent the entire PR snapshot every 2s forever,
// and each frame minted a fresh object on the client that re-rendered the whole
// app — including on routes that show no PR data at all. The poller's notify()
// is the change signal, and it must reach every open tab's stream at once.
var _ = Describe("PR list stream", func() {
	const wakeWait = time.Second
	var server *Server

	BeforeEach(func() {
		// No tick lands inside a spec, so every frame after the first is a wake.
		useStreamInterval(&prStreamInterval, idleStreamCadence)
		server = &Server{
			fetchedAt: time.Date(2026, 6, 23, 10, 0, 0, 0, time.UTC),
			prs:       github.PRSearchResults{{Number: 1, Title: "first", Repo: "flanksource/gavel"}},
		}
	})

	addSecondPR := func() {
		server.mu.Lock()
		defer server.mu.Unlock()
		server.prs = append(server.prs, github.PRListItem{Number: 2, Title: "second", Repo: "flanksource/gavel"})
	}

	It("pings instead of re-sending an unchanged snapshot on a wake", func() {
		frames := serveStream(server.handleSSE, "/api/prs/stream")
		Expect(nextStreamFrame(frames, wakeWait).Data).To(ContainSubstring(`"first"`))

		server.notify()
		Expect(nextStreamFrame(frames, wakeWait)).To(Equal(streamFrame{Ping: true}))
		server.notify()
		Expect(nextStreamFrame(frames, wakeWait)).To(Equal(streamFrame{Ping: true}))
	})

	It("pushes a changed snapshot on a wake", func() {
		frames := serveStream(server.handleSSE, "/api/prs/stream")
		Expect(nextStreamFrame(frames, wakeWait).Ping).To(BeFalse())

		addSecondPR()
		server.notify()

		Expect(nextStreamFrame(frames, wakeWait).Data).To(ContainSubstring(`"second"`))
	})

	It("wakes every open stream with one notify", func() {
		first := serveStream(server.handleSSE, "/api/prs/stream")
		second := serveStream(server.handleSSE, "/api/prs/stream")
		Expect(nextStreamFrame(first, wakeWait).Ping).To(BeFalse())
		Expect(nextStreamFrame(second, wakeWait).Ping).To(BeFalse())

		addSecondPR()
		server.notify()

		Expect(nextStreamFrame(first, wakeWait).Data).To(ContainSubstring(`"second"`))
		Expect(nextStreamFrame(second, wakeWait).Data).To(ContainSubstring(`"second"`))
	})
})
