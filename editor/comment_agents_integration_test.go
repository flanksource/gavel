package editor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/comments/agents"
	"github.com/flanksource/gavel/comments/agents/agentstest"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/comments/postgres"
	"github.com/flanksource/gavel/internal/database"
	ui "github.com/flanksource/gavel/pr/ui"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Page agents configured in .gavel.yaml and projected into comment routes", Ordered, func() {
	var (
		handler  http.Handler
		storage  *postgres.Store
		svc      *comments.Service
		bridge   *agentstest.Bridge
		files    *Store
		tenant   string
		slides   string
		security Document
		solution Document
	)
	call := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.RemoteAddr = "127.0.0.1:40000"
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	pageAgentFor := func(doc Document) pageAgent {
		response := call("GET", "/api/file/"+doc.ID+"/agent", "")
		Expect(response.Code).To(Equal(200), response.Body.String())
		var out pageAgent
		Expect(json.Unmarshal(response.Body.Bytes(), &out)).To(Succeed())
		return out
	}
	saveRoute := func(prefix, prompt string, enabled bool) comments.Route {
		body := `{"scopeUri":` + jsonString(mustTarget(security).Scope) + `,"targetPrefix":` + jsonString(prefix) + `,"workflowId":"human-review","target":{"kind":"run","prompt":"` + prompt + `"},"enabled":` + map[bool]string{true: "true", false: "false"}[enabled] + `,"assign":true}`
		response := call("PUT", "/api/v1/comments/routes", body)
		Expect(response.Code).To(Equal(200), response.Body.String())
		var r comments.Route
		Expect(json.Unmarshal(response.Body.Bytes(), &r)).To(Succeed())
		return r
	}
	configured := func() verify.CommentAgentsConfig {
		cfg, err := verify.LoadSingleGavelConfig(filepath.Join(slides, verify.GavelConfigFileName))
		Expect(err).NotTo(HaveOccurred())
		return cfg.Comments.Agents
	}
	projected := func(ctx context.Context) map[string]bool {
		routes, err := storage.Routes(ctx, tenant, mustTarget(security).Scope)
		Expect(err).NotTo(HaveOccurred())
		out := map[string]bool{}
		for _, r := range routes {
			out[r.TargetPrefix] = r.Enabled
		}
		return out
	}

	BeforeAll(func(ctx SpecContext) {
		db := dbtest.ForGinkgo(dbtest.Options{Name: "editor_page_agents"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		dir := GinkgoT().TempDir()
		useProject(dir)
		slides = filepath.Join(dir, "slides")
		Expect(os.MkdirAll(slides, 0o755)).To(Succeed())
		for _, name := range []string{"19-security.mdx", "09-solution.mdx"} {
			Expect(os.WriteFile(filepath.Join(slides, name), []byte("# "+name+"\n"), 0o600)).To(Succeed())
		}
		var err error
		files, err = NewStore([]string{filepath.Join(slides, "19-security.mdx"), filepath.Join(slides, "09-solution.mdx")})
		Expect(err).NotTo(HaveOccurred())
		for _, file := range files.Files() {
			doc, err := files.Read(file.ID)
			Expect(err).NotTo(HaveOccurred())
			if strings.HasSuffix(doc.Path, "19-security.mdx") {
				security = doc
			} else {
				solution = doc
			}
		}
		storage, err = postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		bridge = &agentstest.Bridge{Dir: mustTarget(security).Dir, Sessions: []captaincli.SessionRecord{}}
		delivery, err := agents.Delivery(agents.Options{Bridge: bridge, Groups: storage, Workdir: ui.CommentWorkdir})
		Expect(err).NotTo(HaveOccurred())
		host := commentHost{store: files}
		principal, err := host.Principal(ctx)
		Expect(err).NotTo(HaveOccurred())
		tenant = principal.TenantID
		svc, err = comments.New(comments.Options{Store: storage, Access: host, Targets: host, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: storage, AttachmentLimit: 1024, Routes: configRoutes{db: storage, files: files}, Delivery: delivery})
		Expect(err).NotTo(HaveOccurred())
		server := &Server{Store: files, CommentService: func(context.Context) (*comments.Service, error) { return svc, nil }, Agents: bridge}
		Expect(commentapi.Register(func(context.Context) (*comments.Service, error) { return svc, nil })).To(Succeed())
		you, err := commentapi.LocalUser()
		Expect(err).NotTo(HaveOccurred())
		server.Comments = commentapi.LocalActor(commentapi.Handler(server.OpenComments), you)
		handler = server.Handler()
	})

	It("saves a folder agent and a page opt-out into the folder's .gavel.yaml and projects both into comment_routes", func(ctx SpecContext) {
		page := pageAgentFor(security)
		folder := saveRoute(page.FolderPrefix, "editor-comments", true)
		optOut := saveRoute(pageAgentFor(solution).PagePrefix, "editor-comments", false)

		Expect(configured()).To(Equal(verify.CommentAgentsConfig{
			Default: &verify.CommentAgent{Enabled: true, Assign: true, Prompt: "editor-comments"},
			Pages:   map[string]verify.CommentAgent{"09-solution.mdx": {Enabled: false, Assign: true, Prompt: "editor-comments"}},
		}))
		Expect(projected(ctx)).To(Equal(map[string]bool{folder.TargetPrefix: true, optOut.TargetPrefix: false}))
		Expect([]any{pageAgentFor(security).Effective.ID, pageAgentFor(solution).Effective.ID, pageAgentFor(solution).Effective.Enabled}).
			To(Equal([]any{folder.ID, optOut.ID, false}))
	})

	It("re-projects a hand edit of the canonical .gavel.yaml the next time routes are read", func(ctx SpecContext) {
		Expect(verify.SaveCommentAgent(slides, "19-security.mdx", &verify.CommentAgent{Enabled: true, Prompt: "slides-review"})).To(Succeed())

		page := pageAgentFor(security)
		Expect([]any{page.Page.Target.Prompt, page.Effective.TargetPrefix, page.Folder.Target.Prompt}).
			To(Equal([]any{"slides-review", page.PagePrefix, "editor-comments"}))
		Expect(projected(ctx)).To(HaveKeyWithValue(page.PagePrefix, true))
	})

	It("starts the page's configured prompt for a new human thread on it", func(ctx SpecContext) {
		selector, err := json.Marshal(CommentSelector{Kind: "document", View: "source", Version: 1})
		Expect(err).NotTo(HaveOccurred())
		place := mustTarget(security)
		_, err = svc.Create(ctx, comments.CreateRequest{Target: comments.Target{ScopeURI: place.Scope, URI: place.URI, Kind: "document", Revision: security.Revision, Selector: selector}, Body: "Tighten the table", WorkflowID: "human-review", RequestID: "page-agent-1"})
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() []string {
			_, launched := bridge.Snapshot()
			references := []string{}
			for _, invocation := range launched {
				references = append(references, invocation.Reference)
			}
			return references
		}).WithTimeout(10 * time.Second).Should(Equal([]string{"slides-review"}))
	})

	It("deletes a page's agent from .gavel.yaml and its route from the projection", func(ctx SpecContext) {
		page := pageAgentFor(solution)
		response := call("DELETE", "/api/v1/comments/routes/"+url.PathEscape(page.Page.ID), "")
		Expect(response.Code).To(Equal(http.StatusNoContent), response.Body.String())

		Expect(configured().Pages).To(HaveKey("19-security.mdx"))
		Expect(configured().Pages).NotTo(HaveKey("09-solution.mdx"))
		Expect(projected(ctx)).NotTo(HaveKey(page.PagePrefix))
	})

	It("refuses to configure a live session, which is a viewer's own choice", func() {
		body := `{"scopeUri":` + jsonString(mustTarget(security).Scope) + `,"targetPrefix":` + jsonString(mustTarget(security).URI) + `,"workflowId":"human-review","target":{"kind":"session","sessionId":"live-1"},"enabled":true}`
		response := call("PUT", "/api/v1/comments/routes", body)
		Expect([]any{response.Code, response.Body.String()}).To(ConsistOf(http.StatusBadRequest, ContainSubstring("choose a live session per viewer")))
	})
})
