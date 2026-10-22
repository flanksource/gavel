package livetest

import (
	"context"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"github.com/flanksource/gavel/testrunner/parsers"
	testui "github.com/flanksource/gavel/testrunner/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type countedDetail struct{ calls atomic.Int32 }

func (d *countedDetail) MarshalJSON() ([]byte, error) {
	d.calls.Add(1)
	return []byte(`{"message":"large detail"}`), nil
}

var _ = Describe("Snapshot encoding cache", func() {
	It("does not re-encode the result tree on idle SSE ticks", func() {
		detail := &countedDetail{}
		server := testui.NewServer()
		server.LoadSnapshot(testui.Snapshot{Status: testui.SnapshotStatus{Running: true}, Tests: []parsers.Test{{Name: "result", Detail: detail}}})
		ctx, cancel := context.WithTimeout(context.Background(), 650*time.Millisecond)
		defer cancel()
		response := httptest.NewRecorder()
		server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/api/tests/stream", nil).WithContext(ctx))
		Expect(detail.calls.Load()).To(Equal(int32(1)))
		Expect(response.Body.String()).To(ContainSubstring("large detail"))
	})
	It("encodes unchanged results once and invalidates when metadata changes", func() {
		detail := &countedDetail{}
		server := testui.NewServer()
		server.LoadSnapshot(testui.Snapshot{Tests: []parsers.Test{{Name: "result", Detail: detail}}})
		read := func() string {
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, httptest.NewRequest("GET", "/api/tests", nil))
			Expect(response.Code).To(Equal(200))
			return response.Body.String()
		}
		first := read()
		Expect(read()).To(Equal(first))
		Expect(detail.calls.Load()).To(Equal(int32(1)))
		server.SetVersion("updated")
		Expect(read()).To(ContainSubstring("updated"))
		Expect(detail.calls.Load()).To(Equal(int32(2)))
	})
})
