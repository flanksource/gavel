package postgres_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	api "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	registerAPI sync.Once
	apiService  atomic.Pointer[comments.Service]
)

// commentsAPI serves svc through the generated comments routes. The entity is
// registered once per process (a second registration is skipped), so it
// resolves whichever service the running spec handed here last.
func commentsAPI(svc *comments.Service) http.Handler {
	GinkgoHelper()
	apiService.Store(svc)
	resolve := func(context.Context) (*comments.Service, error) { return apiService.Load(), nil }
	registerAPI.Do(func() { Expect(api.Register(resolve)).To(Succeed()) })
	return api.Handler(resolve)
}

// sourceAccess is access whose host reads file text: sources by target URI,
// no file content for web pages, and an unreadable target.
type sourceAccess struct {
	access
	sources    map[string]string
	unreadable string
}

func (a sourceAccess) TargetSource(_ context.Context, t comments.Target) (string, bool, error) {
	if t.URI == a.unreadable {
		return "", false, errors.New("disk offline")
	}
	source, ok := a.sources[t.URI]
	return source, ok, nil
}

var _ = Describe("comment list and create contracts over HTTP", Ordered, func() {
	const scope = "git://repo"
	var (
		svc     *comments.Service
		handler http.Handler
		host    *sourceAccess
	)
	BeforeAll(func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "comments_contracts_http"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		store, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		host = &sourceAccess{access: access{tenant: "tenant-contracts"}, sources: map[string]string{}}
		svc, err = comments.New(comments.Options{Store: store, Access: host, Targets: host, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: store, AttachmentLimit: 64})
		Expect(err).NotTo(HaveOccurred())
		handler = commentsAPI(svc)
	})
	call := func(method, path string, body any) *httptest.ResponseRecorder {
		data, err := json.Marshal(body)
		Expect(err).NotTo(HaveOccurred())
		r := httptest.NewRequest(method, path, bytes.NewReader(data))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w
	}
	type item struct {
		ID     string                  `json:"id"`
		Anchor *comments.PatchLocation `json:"anchor"`
	}
	list := func(target string) (int, []item) {
		w := call(http.MethodGet, "/api/v1/comments?scope="+url.QueryEscape(scope)+"&target="+url.QueryEscape(target)+"&limit=20", nil)
		var listed struct {
			Items []item `json:"data"`
		}
		if w.Code == http.StatusOK {
			Expect(json.Unmarshal(w.Body.Bytes(), &listed)).To(Succeed(), w.Body.String())
		}
		return w.Code, listed.Items
	}
	create := func(ctx context.Context, uri, requestID string, selector any) comments.Comment {
		raw, err := json.Marshal(selector)
		Expect(err).NotTo(HaveOccurred())
		c, err := svc.Create(ctx, comments.CreateRequest{Target: comments.Target{URI: uri, ScopeURI: scope, Kind: "code", Selector: raw}, Body: "Check " + requestID, WorkflowID: "human-review", RequestID: requestID})
		Expect(err).NotTo(HaveOccurred())
		return c
	}

	It("lists each patch-anchored thread's anchor resolved against the target's current text, and none where there is no patch or no file content", func(ctx SpecContext) {
		const file, page = scope + "/widget.go", scope + "/pages/home"
		original := "package widget\n\nfunc A() {}\n\nfunc B() {}\n"
		host.sources[file] = original
		patch, err := comments.CapturePatch(original, 5, 5)
		Expect(err).NotTo(HaveOccurred())
		moved := create(ctx, file, "anchor-moved", map[string]any{"patch": patch})
		plain := create(ctx, file, "anchor-none", map[string]any{"kind": "document"})
		web := create(ctx, page, "anchor-web", map[string]any{"patch": patch})
		anchorAfter := func(source string) *comments.PatchLocation {
			host.sources[file] = source
			code, items := list(file)
			Expect(code).To(Equal(http.StatusOK))
			Expect(items).To(HaveLen(2))
			for _, i := range items {
				if i.ID == plain.ID {
					Expect(i.Anchor).To(BeNil(), "a thread without a patch has no anchor")
				}
			}
			for _, i := range items {
				if i.ID == moved.ID {
					return i.Anchor
				}
			}
			return nil
		}
		Expect([]*comments.PatchLocation{
			anchorAfter("package widget\n\n// added\n\nfunc A() {}\n\nfunc B() {}\n"),
			anchorAfter("package widget\n\nfunc A() {}\n\nfunc C() {}\n"),
			anchorAfter("\nfunc A() {}\n\nfunc B() {}\n\nfunc A() {}\n\nfunc B() {}\n"),
		}).To(Equal([]*comments.PatchLocation{{State: "matched", StartLine: 7, EndLine: 7}, {State: "outdated"}, {State: "ambiguous"}}))
		_, items := list(page)
		Expect(items).To(Equal([]item{{ID: web.ID}}), "a web page has no file content, so no anchor")
	})

	It("fails the list when the host cannot read a patch-anchored thread's target", func(ctx SpecContext) {
		const broken = scope + "/broken.go"
		patch, err := comments.CapturePatch("one\ntwo", 1, 1)
		Expect(err).NotTo(HaveOccurred())
		create(ctx, broken, "anchor-broken", map[string]any{"patch": patch})
		host.unreadable = broken
		DeferCleanup(func() { host.unreadable = "" })
		code, _ := list(broken)
		Expect(code).To(Equal(http.StatusInternalServerError))
	})

	It("stores a create's base64 attachments with the thread, and refuses invalid base64, media type or size before writing anything", func() {
		const file = scope + "/shots.go"
		body := func(requestID string, attachment map[string]any) map[string]any {
			return map[string]any{"target": comments.Target{URI: file, ScopeURI: scope, Kind: "code"}, "bodyMarkdown": "The button is clipped", "workflowId": "human-review", "requestId": requestID, "attachments": []any{attachment}}
		}
		w := call(http.MethodPost, "/api/v1/comments", body("attach-ok", map[string]any{"filename": "clipped.png", "contentType": "image/png", "data": "cG5nIGJ5dGVz"}))
		Expect(w.Code).To(Equal(http.StatusOK), w.Body.String())
		var created comments.Comment
		Expect(json.Unmarshal(w.Body.Bytes(), &created)).To(Succeed())
		Expect([]any{created.Version, len(created.Attachments), created.Attachments[0].Name, created.Attachments[0].Size}).To(Equal([]any{int64(2), 1, "clipped.png", int64(9)}))

		codes := []int{}
		for i, invalid := range []map[string]any{
			{"filename": "x.png", "contentType": "image/png", "data": "not base64!"},
			{"filename": "x.png", "contentType": "not a type", "data": "eA=="},
			{"filename": "x.png", "contentType": "image/png", "data": bytes.Repeat([]byte("x"), 65)},
			{"filename": "../x.png", "contentType": "image/png", "data": "eA=="},
		} {
			codes = append(codes, call(http.MethodPost, "/api/v1/comments", body("attach-bad-"+string(rune('a'+i)), invalid)).Code)
		}
		_, items := list(file)
		Expect([]any{codes, len(items)}).To(Equal([]any{[]int{400, 400, 400, 400}, 1}))
	})

	It("pages a target's threads by id, so a thread deleted mid-way skips none after it", func(ctx SpecContext) {
		const file = scope + "/keyset.go"
		for _, id := range []string{"a", "b", "c"} {
			create(ctx, file, "keyset-"+id, map[string]any{"kind": "document"})
		}
		first, err := svc.Page(ctx, comments.Filter{ScopeURI: scope, TargetURI: file, Limit: 1})
		Expect(err).NotTo(HaveOccurred())
		all, err := svc.Page(ctx, comments.Filter{ScopeURI: scope, TargetURI: file, Limit: 3})
		Expect(err).NotTo(HaveOccurred())
		_, err = svc.Act(ctx, comments.ActionRequest{ID: first.Items[0].ID, ExpectedVersion: first.Items[0].Version, Action: "delete"})
		Expect(err).NotTo(HaveOccurred())
		next, err := svc.Page(ctx, comments.Filter{ScopeURI: scope, TargetURI: file, Limit: 1, After: first.Items[0].ID})
		Expect(err).NotTo(HaveOccurred())
		Expect([]string{next.Items[0].ID}).To(Equal([]string{all.Items[1].ID}))
		threads, _, err := svc.TargetThreads(ctx, comments.ReadTarget{ScopeURI: scope, URI: file})
		Expect(err).NotTo(HaveOccurred())
		Expect(threads).To(HaveLen(2))
		_, err = svc.Page(ctx, comments.Filter{ScopeURI: scope, TargetURI: file, Limit: 1, Offset: 1, After: first.Items[0].ID})
		Expect(err).To(MatchError(comments.ErrInvalid))
	})
})
