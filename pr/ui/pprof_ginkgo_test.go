package ui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/flanksource/captain/pkg/monitor"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("runtime profiler endpoint", func() {
	mux := http.NewServeMux()
	registerPprof(mux)

	status := func(path, remoteAddr string) int {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.RemoteAddr = remoteAddr
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, request)
		return recorder.Code
	}

	It("serves profiles to this host and hides them from everyone else", func() {
		Expect(status("/debug/pprof/", "127.0.0.1:54321")).To(Equal(http.StatusOK))
		Expect(status("/debug/pprof/heap", "[::1]:54321")).To(Equal(http.StatusOK))
		// A profile is a workload fingerprint plus the process command line, so
		// an off-host caller must not be able to tell the endpoint exists.
		Expect(status("/debug/pprof/", "10.1.2.3:54321")).To(Equal(http.StatusNotFound))
		Expect(status("/debug/pprof/cmdline", "10.1.2.3:54321")).To(Equal(http.StatusNotFound))
	})
})

var _ = Describe("ingest counter endpoint", func() {
	get := func(read func() (monitor.IngestStats, bool), remoteAddr string) *httptest.ResponseRecorder {
		mux := http.NewServeMux()
		registerIngestStats(mux, read)
		request := httptest.NewRequest(http.MethodGet, "/debug/ingest", nil)
		request.RemoteAddr = remoteAddr
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, request)
		return recorder
	}

	running := func() (monitor.IngestStats, bool) {
		return monitor.IngestStats{
			FilesConsidered: 120, FilesIngested: 3,
			MessagesParsed: 900, MessagesOffered: 9,
			WriteDuration: 250 * time.Millisecond,
		}, true
	}

	It("reports the counters with the ratio the numbers are read for", func() {
		recorder := get(running, "127.0.0.1:54321")
		Expect(recorder.Code).To(Equal(http.StatusOK))

		var payload map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &payload)).To(Succeed())
		Expect(payload).To(SatisfyAll(
			HaveKeyWithValue("filesConsidered", BeNumerically("==", 120)),
			HaveKeyWithValue("filesIngested", BeNumerically("==", 3)),
			HaveKeyWithValue("messagesParsed", BeNumerically("==", 900)),
			HaveKeyWithValue("messagesOffered", BeNumerically("==", 9)),
			// Served precomputed: the ratio is the reason to open the endpoint,
			// and it should not have to be worked out by hand at 3am.
			HaveKeyWithValue("offerRatio", BeNumerically("==", 0.01)),
		))
	})

	It("says the monitor is not running rather than reporting zeroes as fact", func() {
		absent := func() (monitor.IngestStats, bool) { return monitor.IngestStats{}, false }
		Expect(get(absent, "127.0.0.1:54321").Code).To(Equal(http.StatusServiceUnavailable))
	})

	It("hides the counters from off-host callers", func() {
		Expect(get(running, "10.1.2.3:54321").Code).To(Equal(http.StatusNotFound))
	})
})
