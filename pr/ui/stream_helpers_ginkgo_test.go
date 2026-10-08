package ui

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// streamFrame is one frame read off an event stream: a dispatched event (Event
// and Data) or a ": ping" comment (Ping).
type streamFrame struct {
	Event string
	Data  string
	Ping  bool
}

// openStream GETs url and returns its frames in arrival order. The request is
// cancelled when the spec ends, so the stream never outlives it.
func openStream(url string, header http.Header) <-chan streamFrame {
	ctx, cancel := context.WithCancel(context.Background())
	DeferCleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	Expect(err).NotTo(HaveOccurred())
	for name, values := range header {
		req.Header[name] = values
	}
	resp, err := http.DefaultClient.Do(req)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	Expect(resp.Header.Get("Content-Type")).To(Equal("text/event-stream"))
	frames := make(chan streamFrame, 1024)
	go readStreamFrames(resp.Body, frames)
	return frames
}

// serveStream serves handler on a real test server and opens a stream on it.
func serveStream(handler http.HandlerFunc, path string) <-chan streamFrame {
	server := httptest.NewServer(handler)
	DeferCleanup(server.Close)
	return openStream(server.URL+path, nil)
}

// readStreamFrames is an independent SSE reader (not clicky's parser, so a
// framing bug cannot hide behind its own oracle).
func readStreamFrames(body io.ReadCloser, out chan<- streamFrame) {
	defer body.Close()
	defer close(out)
	reader := bufio.NewReaderSize(body, 1<<20)
	var frame streamFrame
	var data []string
	seen := false
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSuffix(line, "\n")
		if line == "" {
			if seen {
				frame.Data = strings.Join(data, "\n")
				out <- frame
			}
			frame, data, seen = streamFrame{}, nil, false
			continue
		}
		seen = true
		if strings.HasPrefix(line, ":") {
			frame.Ping = true
			continue
		}
		field, value, _ := strings.Cut(line, ": ")
		switch field {
		case "event":
			frame.Event = value
		case "data":
			data = append(data, value)
		}
	}
}

// nextStreamFrame waits up to within for the next frame.
func nextStreamFrame(frames <-chan streamFrame, within time.Duration) streamFrame {
	select {
	case frame, ok := <-frames:
		Expect(ok).To(BeTrue(), "stream closed while waiting for a frame")
		return frame
	case <-time.After(within):
		Fail("timed out waiting for a stream frame")
	}
	return streamFrame{}
}

// takeStreamFrames reads the next n frames.
func takeStreamFrames(frames <-chan streamFrame, n int, within time.Duration) []streamFrame {
	got := make([]streamFrame, 0, n)
	for range n {
		got = append(got, nextStreamFrame(frames, within))
	}
	return got
}

// expectNoStreamFrame asserts nothing arrives for the given window.
func expectNoStreamFrame(frames <-chan streamFrame, window time.Duration) {
	select {
	case frame := <-frames:
		Fail("unexpected stream frame: " + frame.Event + " " + frame.Data)
	case <-time.After(window):
	}
}

// useStreamInterval sets a stream cadence for the spec and restores it after.
func useStreamInterval(target *time.Duration, value time.Duration) {
	original := *target
	*target = value
	DeferCleanup(func() { *target = original })
}

func readBody(resp *http.Response) string {
	body, err := io.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	return string(body)
}
