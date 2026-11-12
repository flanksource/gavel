package editor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	"github.com/flanksource/gavel/projects"
	"github.com/flanksource/gavel/settings"
	"github.com/flanksource/gavel/verify"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func ptr[T any](value T) *T { return &value }

var _ = Describe("Page agents configured in the settings layers of a file", Ordered, func() {
	var (
		handler  http.Handler
		svc      *comments.Service
		bridge   *agentstest.Bridge
		slides   string
		security Document
		solution Document
	)
	folder := settings.Ref{Kind: settings.KindPath, Path: "slides"}
	call := func(method, path string, body any) *httptest.ResponseRecorder {
		data, err := json.Marshal(body)
		Expect(err).NotTo(HaveOccurred())
		req := httptest.NewRequest(method, path, strings.NewReader(string(data)))
		req.RemoteAddr = "127.0.0.1:40000"
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		return response
	}
	settingsOf := func(doc Document) fileSettings {
		GinkgoHelper()
		response := call("GET", "/api/file/"+doc.ID+"/settings", nil)
		Expect(response.Code).To(Equal(200), response.Body.String())
		var out fileSettings
		Expect(json.Unmarshal(response.Body.Bytes(), &out)).To(Succeed())
		return out
	}
	saveAgent := func(doc Document, layer settings.Ref, page bool, agent *verify.CommentAgent) *httptest.ResponseRecorder {
		return call("PUT", "/api/file/"+doc.ID+"/settings/agent", map[string]any{"layer": layer, "page": page, "agent": agent})
	}
	configured := func(name string) verify.CommentAgentsConfig {
		GinkgoHelper()
		cfg, err := verify.LoadSingleGavelConfig(filepath.Join(slides, name))
		Expect(err).NotTo(HaveOccurred())
		return cfg.Comments.Agents
	}

	BeforeAll(func(ctx SpecContext) {
		// Settings chains start at the user's home: keep the developer's own out.
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		db := dbtest.ForGinkgo(dbtest.Options{Name: "editor_page_agents"})
		Expect(database.Migrate(ctx, db.DSN())).To(Succeed())
		dir, err := projects.CanonicalPath(GinkgoT().TempDir())
		Expect(err).NotTo(HaveOccurred())
		useProject(dir)
		slides = filepath.Join(dir, "slides")
		Expect(os.MkdirAll(slides, 0o755)).To(Succeed())
		for _, name := range []string{"19-security.mdx", "09-solution.mdx"} {
			Expect(os.WriteFile(filepath.Join(slides, name), []byte("# "+name+"\n"), 0o600)).To(Succeed())
		}
		files, err := NewStore([]string{filepath.Join(slides, "19-security.mdx"), filepath.Join(slides, "09-solution.mdx")})
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
		storage, err := postgres.New(db.Gorm())
		Expect(err).NotTo(HaveOccurred())
		bridge = &agentstest.Bridge{Dir: mustTarget(security).Dir, Sessions: []captaincli.SessionRecord{}}
		delivery, err := agents.Delivery(agents.Options{Bridge: bridge, Groups: storage, Workdir: ui.CommentWorkdir})
		Expect(err).NotTo(HaveOccurred())
		catalog, err := projects.Shared(ctx)
		Expect(err).NotTo(HaveOccurred())
		host := commentHost{store: files}
		svc, err = comments.New(comments.Options{Store: storage, Access: host, Targets: host, Workflows: []comments.Workflow{comments.HumanReview()}, Blobs: storage, AttachmentLimit: 1024, Routes: agents.ProjectRoutes(catalog, settings.File{}), Delivery: delivery})
		Expect(err).NotTo(HaveOccurred())
		server := &Server{Store: files, CommentService: func(context.Context) (*comments.Service, error) { return svc, nil }, Agents: bridge}
		Expect(commentapi.Register(func(context.Context) (*comments.Service, error) { return svc, nil })).To(Succeed())
		you, err := commentapi.LocalUser()
		Expect(err).NotTo(HaveOccurred())
		server.Comments = commentapi.LocalActor(commentapi.Handler(server.OpenComments), you)
		handler = server.Handler()
	})

	It("lists the file's chain from the user's home to its folder, each position holding files only in file mode", func() {
		positions := settingsOf(security).Positions
		kinds := []string{}
		for _, position := range positions {
			Expect(position.DB).To(BeFalse(), "position %s %q", position.Kind, position.Path)
			kinds = append(kinds, string(position.Kind)+":"+position.Path)
		}
		Expect(kinds[0]).To(Equal("user:"))
		Expect(kinds[len(kinds)-2:]).To(Equal([]string{"project:", "path:slides"}))
	})

	It("saves a folder default into .gavel.yaml and a page opt-out into a new snippet, tracing each field to its file", func() {
		Expect(saveAgent(security, settings.Ref{Kind: settings.KindPath, Path: "slides", File: verify.GavelConfigFileName}, false, &verify.CommentAgent{Enabled: ptr(true), Assign: ptr(true), Prompt: "editor-comments"}).Code).To(Equal(200))
		response := saveAgent(solution, settings.Ref{Kind: settings.KindPath, Path: "slides", File: ".gavel.agents.yaml"}, true, &verify.CommentAgent{Enabled: ptr(false)})
		Expect(response.Code).To(Equal(200), response.Body.String())

		Expect([]any{configured(verify.GavelConfigFileName), configured(".gavel.agents.yaml")}).To(Equal([]any{
			verify.CommentAgentsConfig{Default: &verify.CommentAgent{Enabled: ptr(true), Assign: ptr(true), Prompt: "editor-comments"}},
			verify.CommentAgentsConfig{Pages: map[string]verify.CommentAgent{"09-solution.mdx": {Enabled: ptr(false)}}},
		}))
		main, snippet := filepath.Join(slides, verify.GavelConfigFileName), filepath.Join(slides, ".gavel.agents.yaml")
		page, other := settingsOf(security).Agent, settingsOf(solution).Agent
		Expect([]any{page.Agent.IsEnabled(), page.Provenance["enabled"], other.Agent.IsEnabled(), other.Provenance["enabled"], other.Provenance["prompt"]}).To(Equal([]any{
			true, verify.CommentAgentSource{Layer: main},
			false, verify.CommentAgentSource{Layer: snippet, Page: "09-solution.mdx"}, verify.CommentAgentSource{Layer: main},
		}))
	})

	It("starts the prompt a hand-edited page entry configures for a new human thread", func(ctx SpecContext) {
		cfg, err := verify.LoadSingleGavelConfig(filepath.Join(slides, verify.GavelConfigFileName))
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.Comments.Agents.Set("19-security.mdx", &verify.CommentAgent{Prompt: "slides-review"})).To(Succeed())
		Expect(verify.SaveGavelConfig(filepath.Join(slides, verify.GavelConfigFileName), cfg)).To(Succeed())

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

	It("removes a page's entry from the layer it was saved to", func() {
		response := saveAgent(solution, settings.Ref{Kind: settings.KindPath, Path: "slides", File: ".gavel.agents.yaml"}, true, nil)
		Expect(response.Code).To(Equal(200), response.Body.String())
		Expect([]any{configured(".gavel.agents.yaml").Pages, settingsOf(solution).Agent.Agent.IsEnabled()}).To(Equal([]any{map[string]verify.CommentAgent(nil), true}))
	})

	DescribeTable("refuses a layer it cannot write",
		func(layer settings.Ref, message string) {
			response := saveAgent(security, layer, false, &verify.CommentAgent{Enabled: ptr(true)})
			Expect([]any{response.Code, response.Body.String()}).To(ConsistOf(http.StatusBadRequest, ContainSubstring(message)))
		},
		Entry("a database layer in file-only mode", folder, "file-only settings have no database layer"),
		Entry("a folder outside the file's chain", settings.Ref{Kind: settings.KindPath, Path: "other", File: verify.GavelConfigFileName}, "is not in the target's configuration chain"),
		Entry("a file that is not a gavel config", settings.Ref{Kind: settings.KindPath, Path: "slides", File: "agents.yaml"}, "is not a .gavel.yaml"),
	)
})
