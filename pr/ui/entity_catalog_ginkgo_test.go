package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

// serveEntityRequest drives the dashboard's full handler, so the routes asserted
// here are the ones the chat's context picker and the assistant actually reach.
func serveEntityRequest(path, accept string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	if accept != "" {
		request.Header.Set("Accept", accept)
	}
	(&Server{}).Handler().ServeHTTP(recorder, request)
	return recorder
}

type generatedOperation struct {
	Parameters []struct {
		Name string `json:"name"`
		In   string `json:"in"`
	} `json:"parameters"`
	Clicky struct {
		Surface string `json:"surface"`
		Verb    string `json:"verb"`
		Scope   string `json:"scope"`
	} `json:"x-clicky"`
}

type generatedDocument struct {
	Clicky struct {
		Surfaces []struct {
			Key string `json:"key"`
		} `json:"surfaces"`
	} `json:"x-clicky"`
	Paths map[string]map[string]generatedOperation `json:"paths"`
}

func (d generatedDocument) surfaceKeys() []string {
	keys := make([]string, 0, len(d.Clicky.Surfaces))
	for _, surface := range d.Clicky.Surfaces {
		keys = append(keys, surface.Key)
	}
	return keys
}

func (o generatedOperation) parameterNames() []string {
	names := make([]string, 0, len(o.Parameters))
	for _, parameter := range o.Parameters {
		names = append(names, parameter.Name)
	}
	return names
}

// clickyTableIDs reads the hidden _id cell of every row of a clicky table
// document, which is what the context picker selects rows by.
func clickyTableIDs(body []byte) []string {
	var document struct {
		Node struct {
			Kind string `json:"kind"`
			Rows []struct {
				Cells map[string]struct {
					Text  string `json:"text"`
					Plain string `json:"plain"`
				} `json:"cells"`
			} `json:"rows"`
		} `json:"node"`
	}
	Expect(json.Unmarshal(body, &document)).To(Succeed(), "body = %s", body)
	Expect(document.Node.Kind).To(Equal("table"), "body = %s", body)
	ids := make([]string, 0, len(document.Node.Rows))
	for _, row := range document.Node.Rows {
		cell := row.Cells["_id"]
		id := cell.Plain
		if id == "" {
			id = cell.Text
		}
		ids = append(ids, id)
	}
	return ids
}

var _ = Describe("generated entity catalog", func() {
	var firstDir, secondDir string
	var first, second *types.TODO

	seed := func(dir, title string) *types.TODO {
		todo, err := uiTestProviderFor(dir).Create(GinkgoT().Context(), todos.CreateRequest{
			Title: title, Status: types.StatusPending, Priority: types.PriorityMedium,
		})
		Expect(err).NotTo(HaveOccurred())
		return todo
	}

	BeforeEach(func() {
		originalPath, originalCounts := projectsPath, projectsTodoCounts
		projectsPath = filepath.Join(GinkgoT().TempDir(), "projects.json")
		projectsTodoCounts = func(_ context.Context, projects []Project) []todoCountsResult {
			return make([]todoCountsResult, len(projects))
		}
		DeferCleanup(func() {
			projectsPath = originalPath
			projectsTodoCounts = originalCounts
		})

		firstDir, secondDir = filepath.Clean(GinkgoT().TempDir()), filepath.Clean(GinkgoT().TempDir())
		Expect(SaveProjects([]Project{
			{Name: "Alpha Work", Dir: firstDir},
			{Name: "beta", Dir: secondDir},
		})).To(Succeed())
		first = seed(firstDir, "Alpha work")
		second = seed(secondDir, "Beta work")
	})

	It("serves the todo and project surfaces at /api/v1/openapi.json for the context picker", func() {
		recorder := serveEntityRequest("/api/v1/openapi.json", "")
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())

		var document generatedDocument
		Expect(json.Unmarshal(recorder.Body.Bytes(), &document)).To(Succeed())
		Expect(document.surfaceKeys()).To(ContainElements("todo", "project"))

		for _, key := range []string{"todo", "project"} {
			list := document.Paths["/api/v1/"+key]["get"]
			Expect(list.Clicky).To(MatchFields(IgnoreExtras, Fields{
				"Surface": Equal(key), "Verb": Equal("list"), "Scope": Equal("collection"),
			}), key)
			get := document.Paths["/api/v1/"+key+"/{id}"]["get"]
			Expect(get.Clicky).To(MatchFields(IgnoreExtras, Fields{
				"Surface": Equal(key), "Verb": Equal("get"), "Scope": Equal("entity"),
			}), key)
		}
		Expect(document.Paths["/api/v1/todo"]["get"].parameterNames()).To(ContainElements("project", "dir", "limit", "offset"))
	})

	It("lists the registered projects as rows the picker selects by short name", func() {
		recorder := serveEntityRequest("/api/v1/project", "application/json+clicky")

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		Expect(clickyTableIDs(recorder.Body.Bytes())).To(Equal([]string{"alpha-work", "beta"}))
	})

	It("gets one project by short name and answers 404 for anything else", func() {
		recorder := serveEntityRequest("/api/v1/project/alpha-work", "application/json")
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var project projectInfo
		Expect(json.Unmarshal(recorder.Body.Bytes(), &project)).To(Succeed())
		Expect(project).To(MatchFields(IgnoreExtras, Fields{
			"Name": Equal("Alpha Work"), "Short": Equal("alpha-work"), "Dir": Equal(firstDir),
		}))

		for _, missing := range []string{"nope", url.PathEscape("Alpha Work")} {
			recorder := serveEntityRequest("/api/v1/project/"+missing, "application/json")
			Expect(recorder.Code).To(Equal(http.StatusNotFound), "%s: %s", missing, recorder.Body.String())
		}
	})

	It("lists todos across every registered project as compact rows naming their project by short name", func() {
		recorder := serveEntityRequest("/api/v1/todo", "application/json")
		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())

		var page struct {
			Data []map[string]any `json:"data"`
			Page struct {
				Limit int   `json:"limit"`
				Total int64 `json:"total"`
			} `json:"page"`
		}
		Expect(json.Unmarshal(recorder.Body.Bytes(), &page)).To(Succeed())
		row := func(todo *types.TODO, project string) map[string]any {
			return map[string]any{
				"_id": todo.DisplayID(), "id": todo.DisplayID(), "title": todo.Title,
				"status": "pending", "priority": "medium", "project": project,
			}
		}
		Expect(page.Data).To(ConsistOf(row(first, "alpha-work"), row(second, "beta")))
		Expect(page.Page.Limit).To(Equal(50), "the default page size")
		Expect(page.Page.Total).To(BeEquivalentTo(2))
	})

	It("lists todo rows the picker selects by short id, filtered by project short name", func() {
		recorder := serveEntityRequest("/api/v1/todo?project=alpha-work", "application/json+clicky")

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		Expect(clickyTableIDs(recorder.Body.Bytes())).To(Equal([]string{first.DisplayID()}))
	})

	DescribeTable("rejects a scope that names no single workspace",
		func(query url.Values, message string) {
			recorder := serveEntityRequest("/api/v1/todo?"+query.Encode(), "application/json")

			Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
			Expect(recorder.Body.String()).To(ContainSubstring(message))
		},
		Entry("an unknown project", url.Values{"project": {"nope"}}, `unknown project \"nope\"; registered projects: alpha-work, beta`),
		Entry("a project's full name", url.Values{"project": {"Alpha Work"}}, `unknown project \"Alpha Work\"`),
		Entry("a dir and a project together", url.Values{"project": {"beta"}, "dir": {"/somewhere"}}, "either dir or project"),
	)

	It("refuses short-name listings when two projects share a short name, but keeps the dashboard's list", func() {
		Expect(SaveProjects([]Project{
			{Name: "Alpha Work", Dir: firstDir},
			{Name: "alpha-work", Dir: secondDir},
		})).To(Succeed())

		for _, path := range []string{"/api/v1/project", "/api/v1/todo"} {
			recorder := serveEntityRequest(path, "application/json")
			Expect(recorder.Code).NotTo(Equal(http.StatusOK), path)
			Expect(recorder.Body.String()).To(ContainSubstring(`share the short name \"alpha-work\"`), path)
		}
		Expect(serveEntityRequest("/api/projects", "application/json").Code).To(Equal(http.StatusOK))
	})

	It("exposes project and todo listing to the assistant as tools that run without asking", func() {
		root, err := entityCommandRoot()
		Expect(err).NotTo(HaveOccurred())
		provider, err := newGavelChatTools(root)
		Expect(err).NotTo(HaveOccurred())

		tools, err := provider.ToolSet(GinkgoT().Context())
		Expect(err).NotTo(HaveOccurred())

		byName := map[string]string{}
		var todoList map[string]any
		for _, entry := range tools.Catalog {
			byName[entry.Name] = entry.Method
			if entry.Name == "todo" {
				todoList = entry.InputSchema
			}
		}
		Expect(byName).To(HaveKeyWithValue("project", http.MethodGet))
		Expect(byName).To(HaveKeyWithValue("project_get", http.MethodGet))
		Expect(byName).To(HaveKeyWithValue("todo", http.MethodGet))
		Expect(byName).To(HaveKeyWithValue("todo_get", http.MethodGet))
		Expect(todoList).To(HaveKeyWithValue("properties", And(HaveKey("project"), HaveKey("limit"))))
	})
})
