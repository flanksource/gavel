package postgres_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	api "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("page read state over HTTP", func() {
	It("POST /comments/read marks one page's threads read for the reader, leaving other pages unread", func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comments_read_http"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		a := access{tenant: "tenant-read"}
		svc, err := comments.New(comments.Options{Store: store, Access: a, Targets: a, Workflows: []comments.Workflow{comments.HumanReview()}})
		Expect(err).NotTo(HaveOccurred())
		router := commentsAPI(svc)
		const scope, home, about = "https://example.test/site", "https://example.test/site/home", "https://example.test/site/about"
		for i, page := range []string{home, home, about} {
			_, err := svc.Create(ctx, comments.CreateRequest{Target: comments.Target{URI: page, ScopeURI: scope, Kind: "document"}, Body: "Review " + page, WorkflowID: "human-review", RequestID: "read-" + string(rune('a'+i))})
			Expect(err).NotTo(HaveOccurred())
		}
		read := func(body any) (int, string) {
			data, err := json.Marshal(body)
			Expect(err).NotTo(HaveOccurred())
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/comments/read", bytes.NewReader(data)))
			return w.Code, w.Body.String()
		}
		// unread is, per page, whether GET /api/v1/comments lists any of its threads unread.
		unread := func() map[string]bool {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/comments?scope="+url.QueryEscape(scope)+"&limit=20", nil))
			Expect(w.Code).To(Equal(http.StatusOK), w.Body.String())
			var listed struct {
				Items []struct {
					Target comments.Target `json:"target"`
					Unread bool            `json:"unread"`
				} `json:"data"`
			}
			Expect(json.Unmarshal(w.Body.Bytes(), &listed)).To(Succeed(), w.Body.String())
			out := map[string]bool{}
			for _, c := range listed.Items {
				out[c.Target.URI] = out[c.Target.URI] || c.Unread
			}
			return out
		}
		target := map[string]any{"target": comments.ReadTarget{ScopeURI: scope, URI: home}}

		Expect(unread()).To(Equal(map[string]bool{home: true, about: true}))
		code, body := read(target)
		Expect([]any{code, body}).To(Equal([]any{http.StatusOK, `{"marked":2}` + "\n"}))
		Expect(unread()).To(Equal(map[string]bool{home: false, about: true}))
		code, body = read(target)
		Expect([]any{code, body}).To(Equal([]any{http.StatusOK, `{"marked":0}` + "\n"}), "already read threads are not marked again")

		codes := []int{}
		for _, invalid := range []any{
			map[string]any{"target": comments.ReadTarget{ScopeURI: scope}},
			map[string]any{"target": comments.ReadTarget{URI: home}},
			map[string]any{"target": comments.ReadTarget{ScopeURI: scope, URI: home}, "extra": true},
		} {
			code, _ := read(invalid)
			codes = append(codes, code)
		}
		Expect(codes).To(Equal([]int{http.StatusBadRequest, http.StatusBadRequest, http.StatusBadRequest}))
		Expect(api.HTTPPaths("/api/v1")).To(HaveKey("/api/v1/comments/read"))
	})
})
