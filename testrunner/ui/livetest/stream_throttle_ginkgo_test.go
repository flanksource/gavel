package livetest

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type sseLine struct {
	at   time.Time
	text string
}

type sseRecorder struct {
	mu    sync.Mutex
	lines []sseLine
}

func (r *sseRecorder) read(body *bufio.Scanner) {
	for body.Scan() {
		r.mu.Lock()
		r.lines = append(r.lines, sseLine{at: time.Now(), text: body.Text()})
		r.mu.Unlock()
	}
}

func (r *sseRecorder) matching(prefix string, before time.Time) []sseLine {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []sseLine
	for _, line := range r.lines {
		if strings.HasPrefix(line.text, prefix) && line.at.Before(before) {
			out = append(out, line)
		}
	}
	return out
}

var _ = Describe("Snapshot stream throttle", func() {
	It("coalesces bursts of updates but delivers the final snapshot immediately", func() {
		const updates = 30
		server := testui.NewServer()
		ch := make(chan []parsers.Test, updates)
		server.StreamFrom(ch)
		httpSrv := httptest.NewServer(server.Handler())
		defer httpSrv.Close()

		resp, err := http.Get(httpSrv.URL + "/api/tests/stream")
		Expect(err).NotTo(HaveOccurred())
		defer resp.Body.Close()
		recorder := &sseRecorder{}
		go recorder.read(bufio.NewScanner(resp.Body))

		for i := 0; i < updates; i++ {
			ch <- []parsers.Test{{Name: fmt.Sprintf("t-%d", i), Pending: true}}
			time.Sleep(20 * time.Millisecond)
		}
		time.Sleep(300 * time.Millisecond)
		burstFrames := len(recorder.matching("data:", time.Now()))
		Expect(burstFrames).To(BeNumerically(">=", 1))
		Expect(burstFrames).To(BeNumerically("<=", 2),
			"a %d-update burst should yield the initial frame plus at most one throttled frame", updates)

		close(ch)
		closed := time.Now()
		Eventually(func() []sseLine { return recorder.matching("event: done", closed.Add(400*time.Millisecond)) }).
			WithTimeout(time.Second).Should(HaveLen(1), "done must not wait for the throttle interval")
		frames := recorder.matching("data:", time.Now())
		Expect(frames[len(frames)-2].text).To(ContainSubstring(fmt.Sprintf("t-%d", updates-1)))
	})
})
