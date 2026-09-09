package ui

import (
	"errors"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTodoSessionStream(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Todo session stream")
}

var _ = Describe("session stream errors", func() {
	It("emits the complete lookup failure as an SSE error", func() {
		recorder := httptest.NewRecorder()
		writeTodoSessionStreamError(recorder, recorder, errors.New(`captain session conflict: provider session ID "session-1" is ambiguous`))

		Expect(recorder.Header().Get("Content-Type")).To(Equal("text/event-stream"))
		Expect(recorder.Body.String()).To(And(
			ContainSubstring("event: error"),
			ContainSubstring(`"error":"captain session conflict: provider session ID \"session-1\" is ambiguous"`),
		))
	})
})
