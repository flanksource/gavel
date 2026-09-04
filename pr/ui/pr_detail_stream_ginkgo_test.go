package ui

import (
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/flanksource/gavel/github"
)

var _ = Describe("PR detail stream frames", func() {
	// A PR with no actionable comments leaves prDetail.Comments nil, which Go
	// marshals as `null`. The UI validates the frame shape, so a null here
	// discards the whole PR frame and renders "invalid PR update" instead of
	// the PR.
	It("emits comments as an array when the cached PR has none", func() {
		srv := &Server{detailCache: NewDetailCache(), refreshCh: make(chan struct{}, 1)}
		srv.detailCache.Put(SyncStatusKey("acme/gavel", 1), prDetail{PR: &github.PRInfo{Number: 1, Title: "Bump deps"}}, time.Now())

		rec := httptest.NewRecorder()
		srv.handleDetail(rec, httptest.NewRequest("GET", "/api/prs/detail?repo=acme%2Fgavel&number=1", nil))

		Expect(rec.Body.String()).To(And(
			ContainSubstring("event: pr"),
			ContainSubstring(`"comments":[]`),
		))
		Expect(rec.Body.String()).ToNot(ContainSubstring(`"comments":null`))
	})
})
