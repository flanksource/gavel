package livetest

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// streamClients is enough connections that their 200ms fallback tickers, staggered
// across one period, cannot all fire inside wakeDeadline: a client that only sees the
// update via its ticker misses the deadline.
const (
	streamClients  = 8
	fallbackTicker = 200 * time.Millisecond
	wakeDeadline   = 80 * time.Millisecond
)

func (r *sseRecorder) firstContaining(substr string) (time.Time, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, line := range r.lines {
		if strings.Contains(line.text, substr) {
			return line.at, true
		}
	}
	return time.Time{}, false
}

func connectStaggered(url string) []*sseRecorder {
	recorders := make([]*sseRecorder, streamClients)
	for i := range recorders {
		resp, err := http.Get(url)
		Expect(err).NotTo(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		DeferCleanup(resp.Body.Close)
		recorders[i] = &sseRecorder{}
		go recorders[i].read(bufio.NewScanner(resp.Body))
		time.Sleep(fallbackTicker / streamClients)
	}
	return recorders
}

func expectEveryClientWokenBy(recorders []*sseRecorder, marker string, sent time.Time) {
	for i, recorder := range recorders {
		var at time.Time
		Eventually(func() bool {
			var ok bool
			at, ok = recorder.firstContaining(marker)
			return ok
		}).WithTimeout(2*time.Second).Should(BeTrue(), "client %d never received %q", i, marker)
		Expect(at.Sub(sent)).To(BeNumerically("<", wakeDeadline),
			"client %d received %q %s after the update; a single notify must wake every stream", i, marker, at.Sub(sent))
	}
}

var _ = Describe("Concurrent stream wakes", func() {
	It("wakes every /api/tests/stream client on a single update", func() {
		server := testui.NewServer()
		ch := make(chan []parsers.Test)
		DeferCleanup(func() { close(ch) })
		server.StreamFrom(ch)
		httpSrv := httptest.NewServer(server.Handler())
		DeferCleanup(httpSrv.Close)

		recorders := connectStaggered(httpSrv.URL + "/api/tests/stream")
		time.Sleep(1100 * time.Millisecond) // let every connection's throttle window lapse

		const marker = "fanout-update"
		sent := time.Now()
		ch <- []parsers.Test{{Name: marker, Pending: true}}
		expectEveryClientWokenBy(recorders, marker, sent)
	})

	It("wakes every /api/rerun/stream client on a single output line", func() {
		server := testui.NewServer()
		output := make(chan *testui.RerunOutputBuffer, 1)
		release := make(chan struct{})
		server.SetRerunFunc(func(_ testui.RerunRequest, buf *testui.RerunOutputBuffer) error {
			output <- buf
			<-release
			return nil
		})
		httpSrv := httptest.NewServer(server.Handler())
		DeferCleanup(httpSrv.Close)
		go func() {
			defer GinkgoRecover()
			resp, err := http.Post(httpSrv.URL+"/api/rerun", "application/json", bytes.NewBufferString(`{}`))
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Body.Close()).To(Succeed())
		}()
		var buf *testui.RerunOutputBuffer
		Eventually(output).WithTimeout(2 * time.Second).Should(Receive(&buf))
		DeferCleanup(func() { close(release) })

		recorders := connectStaggered(httpSrv.URL + "/api/rerun/stream")
		const marker = "fanout-line"
		sent := time.Now()
		_, err := buf.StdoutWriter().Write([]byte(marker))
		Expect(err).NotTo(HaveOccurred())
		expectEveryClientWokenBy(recorders, marker, sent)
	})
})
