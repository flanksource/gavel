package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/bulk"
	"github.com/flanksource/gavel/todos/run"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// These are the contracts the hand-written /api/todos/bulk and
// /api/todos/triage handlers used to carry. They now belong to the generated
// entity route, which is the single surface the CLI, the API and the dashboard
// share — so they are asserted here, over HTTP, rather than against a Go
// function no caller reaches.
//
// Selection ids ride comma-joined in the {id} path segment and the action's
// parameters ride as query params; that is the executor's wire shape, and
// getting it wrong is the most likely way this surface breaks.
func postTodoAction(action string, refs []string, params map[string]string) (*httptest.ResponseRecorder, bulk.Result) {
	query := url.Values{}
	for key, value := range params {
		query.Set(key, value)
	}
	// The route comes from the catalog rather than a convention assembled here,
	// because it is not derivable: clicky infers the method from the action's
	// name, so `delete` is a DELETE onto the entity's own /{id} path while its
	// siblings are POSTs onto /{id}/{action}. A test that hardcoded POST would
	// pass for every action except the destructive one.
	method, path := todoActionRoute(action)
	// Each ref is escaped individually so the comma stays a separator, and so a
	// ref carrying whitespace is expressible at all rather than crashing the
	// request builder.
	escaped := make([]string, 0, len(refs))
	for _, ref := range refs {
		escaped = append(escaped, url.PathEscape(ref))
	}
	path = strings.Replace(path, "{id}", strings.Join(escaped, ","), 1)
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	recorder := httptest.NewRecorder()
	(&Server{}).Handler().ServeHTTP(recorder, httptest.NewRequest(method, path, nil))

	var result bulk.Result
	if recorder.Code == http.StatusOK {
		Expect(json.Unmarshal(recorder.Body.Bytes(), &result)).
			To(Succeed(), "body = %q", recorder.Body.String())
	}
	return recorder, result
}

// todoEntityCatalog reads the published catalog — the same document the
// dashboard derives its selection toolbar from.
func todoEntityCatalog() []struct {
	Name               string `json:"name"`
	Short              string `json:"short"`
	Method             string `json:"method"`
	Path               string `json:"path"`
	SupportsFilterMode bool   `json:"supports_filter_mode"`
	ToolHints          *struct {
		Icon            string `json:"icon"`
		Group           string `json:"group"`
		DestructiveHint *bool  `json:"destructiveHint"`
	} `json:"tool_hints"`
	ParamSchema *struct {
		Properties map[string]struct {
			Enum []string `json:"enum"`
		} `json:"properties"`
		Required []string `json:"required"`
	} `json:"param_schema"`
} {
	recorder := httptest.NewRecorder()
	(&Server{}).Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/entities", nil))
	Expect(recorder.Code).To(Equal(http.StatusOK))

	var entities []struct {
		Name        string `json:"name"`
		BulkActions []struct {
			Name               string `json:"name"`
			Short              string `json:"short"`
			Method             string `json:"method"`
			Path               string `json:"path"`
			SupportsFilterMode bool   `json:"supports_filter_mode"`
			ToolHints          *struct {
				Icon            string `json:"icon"`
				Group           string `json:"group"`
				DestructiveHint *bool  `json:"destructiveHint"`
			} `json:"tool_hints"`
			ParamSchema *struct {
				Properties map[string]struct {
					Enum []string `json:"enum"`
				} `json:"properties"`
				Required []string `json:"required"`
			} `json:"param_schema"`
		} `json:"bulk_actions"`
	}
	Expect(json.Unmarshal(recorder.Body.Bytes(), &entities)).To(Succeed())
	for _, entity := range entities {
		if entity.Name == "todo" {
			return entity.BulkActions
		}
	}
	Fail("the todo entity was not published at /api/entities")
	return nil
}

func todoActionRoute(action string) (string, string) {
	for _, published := range todoEntityCatalog() {
		if published.Name == action {
			Expect(published.Method).NotTo(BeEmpty(), action)
			Expect(published.Path).NotTo(BeEmpty(), action)
			return published.Method, published.Path
		}
	}
	Fail(fmt.Sprintf("bulk action %q is not published", action))
	return "", ""
}

var _ = Describe("todo entity bulk actions", func() {
	var dir string
	var provider *uiTestTODOProvider

	newTodo := func(title string, status types.Status, priority types.Priority) *types.TODO {
		todo, err := provider.Create(GinkgoT().Context(), todos.CreateRequest{
			Title: title, Status: status, Priority: priority,
		})
		Expect(err).NotTo(HaveOccurred())
		return todo
	}

	BeforeEach(func() {
		dir = filepath.Clean(GinkgoT().TempDir())
		provider = uiTestProviderFor(dir)
	})

	It("applies one priority to every named todo", func() {
		first := newTodo("Bulk first", types.StatusPending, types.PriorityLow)
		second := newTodo("Bulk second", types.StatusPending, types.PriorityMedium)

		recorder, result := postTodoAction("priority",
			[]string{first.ID, second.ID}, map[string]string{"to": "high"})

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(result.Applied).To(Equal(2))
		Expect(result.Failed).To(BeZero())
		Expect(first.Priority).To(Equal(types.PriorityHigh))
		Expect(second.Priority).To(Equal(types.PriorityHigh))
	})

	It("re-statuses and comments in one request", func() {
		todo := newTodo("Bulk status", types.StatusPending, types.PriorityMedium)

		_, result := postTodoAction("status", []string{todo.ID},
			map[string]string{"to": "completed", "comment": "closed in bulk"})

		Expect(result.Applied).To(Equal(1))
		Expect(todo.Status).To(Equal(types.StatusCompleted))
		Expect(provider.comments).To(ConsistOf("closed in bulk"))
	})

	// The batch is the point: one archived TODO must not cost the caller the
	// other thirty-nine, so a partial batch is still a 200 with the failure
	// named inside it.
	It("reports per-item failures without abandoning the rest of the batch", func() {
		ok := newTodo("Bulk survivor", types.StatusPending, types.PriorityLow)

		recorder, result := postTodoAction("priority",
			[]string{"todo-does-not-exist", ok.ID}, map[string]string{"to": "high"})

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(result.Applied).To(Equal(1))
		Expect(result.Failed).To(Equal(1))
		Expect(result.Results[1].Error).To(ContainSubstring("todo-does-not-exist"))
		Expect(ok.Priority).To(Equal(types.PriorityHigh))
	})

	// A selection is grouped by severity or age, never by repository, so the
	// caller has no workspace to name and each ref resolves through its owner.
	It("resolves todos whose workspace the caller did not name", func() {
		here := newTodo("Bulk here", types.StatusPending, types.PriorityLow)
		otherDir := filepath.Clean(GinkgoT().TempDir())
		other, err := uiTestProviderFor(otherDir).Create(GinkgoT().Context(), todos.CreateRequest{
			Title: "Bulk elsewhere", Status: types.StatusPending, Priority: types.PriorityLow,
		})
		Expect(err).NotTo(HaveOccurred())

		_, result := postTodoAction("priority",
			[]string{here.ID, other.ID}, map[string]string{"to": "high"})

		Expect(result.Failed).To(BeZero())
		Expect(here.Priority).To(Equal(types.PriorityHigh))
		Expect(other.Priority).To(Equal(types.PriorityHigh),
			"a todo in another workspace must be written through its own provider")
	})

	// "Select all matching" is offered on workspaces of several hundred todos,
	// and the ids ride in the URL path. A selection that silently failed at some
	// unmeasured length would be the worst kind of bug: the user selected
	// everything, and only some of it was acted on.
	It("carries a whole-workspace selection in one request", func() {
		const selected = 500
		refs := make([]string, 0, selected)
		for i := 0; i < selected; i++ {
			refs = append(refs, newTodo(fmt.Sprintf("Bulk %d", i), types.StatusPending, types.PriorityLow).ID)
		}

		recorder, result := postTodoAction("priority", refs, map[string]string{"to": "high"})

		Expect(recorder.Code).To(Equal(http.StatusOK),
			"a %d-todo selection must not be rejected for its size", selected)
		Expect(result.Applied).To(Equal(selected))
		Expect(result.Failed).To(BeZero())
	})

	DescribeTable("rejects a request that cannot be applied",
		func(action string, refs []string, params map[string]string, message string) {
			recorder, _ := postTodoAction(action, refs, params)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(ContainSubstring(message))
		},
		Entry("with a duplicated target",
			"priority", []string{"a", "a"}, map[string]string{"to": "high"}, "duplicates"),
		Entry("with a blank ref",
			"priority", []string{"a", " "}, map[string]string{"to": "high"}, "blank"),
		Entry("with a run-projected status",
			"status", []string{"a"}, map[string]string{"to": "in_progress"}, "cannot be assigned"),
		Entry("with an unknown priority",
			"priority", []string{"a"}, map[string]string{"to": "urgent"}, "unknown priority"),
		Entry("with a labels change that changes nothing",
			"labels", []string{"a"}, nil, "--add or --remove"),
		Entry("with an unconfirmed delete",
			"delete", []string{"a"}, nil, "--confirm"),
	)

	// A rejected request must not have written anything: validation happens
	// before the first item is touched, not partway through the loop.
	It("leaves every todo untouched when validation rejects the request", func() {
		todo := newTodo("Bulk untouched", types.StatusPending, types.PriorityLow)

		recorder, _ := postTodoAction("status", []string{todo.ID}, map[string]string{"to": "failed"})

		Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		Expect(todo.Status).To(Equal(types.StatusPending))
	})
})

var _ = Describe("todo entity run actions", func() {
	var dir string
	var provider *uiTestTODOProvider
	var dispatched []todoRunRequest

	newTodo := func(title string) *types.TODO {
		todo, err := provider.Create(GinkgoT().Context(), todos.CreateRequest{
			Title: title, Status: types.StatusPending, Priority: types.PriorityMedium,
		})
		Expect(err).NotTo(HaveOccurred())
		return todo
	}

	BeforeEach(func() {
		dir = filepath.Clean(GinkgoT().TempDir())
		provider = uiTestProviderFor(dir)
		dispatched = nil

		original := run.Start
		run.Start = func(req todoRunRequest) (todoRunStartResult, error) {
			dispatched = append(dispatched, req)
			if len(req.Todos) == 1 && req.Todos[0].Title == "Triage explodes" {
				return todoRunStartResult{}, errors.New("driver unavailable")
			}
			return todoRunStartResult{Status: "started", SessionID: req.Options.Spec.SessionID}, nil
		}
		DeferCleanup(func() { run.Start = original })
	})

	// Grouped execution is unsupported by the PostgreSQL runtime, so a bulk
	// triage has to be N single-TODO runs rather than one grouped session.
	It("starts one run per selected todo", func() {
		first := newTodo("Triage first")
		second := newTodo("Triage second")

		recorder, result := postTodoAction("triage", []string{first.ID, second.ID}, nil)

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(result.Applied).To(Equal(2))
		Expect(result.Failed).To(BeZero())
		Expect(dispatched).To(HaveLen(2))
		for _, req := range dispatched {
			Expect(req.Todos).To(HaveLen(1))
		}
	})

	// The prompt name is the whole request: triage declares its own behaviour
	// class, so the run must resolve to plan — never a committing run.
	It("dispatches the triage prompt as a plan-class run", func() {
		todo := newTodo("Triage class")

		_, result := postTodoAction("triage", []string{todo.ID}, nil)

		Expect(result.Applied).To(Equal(1))
		Expect(dispatched).To(HaveLen(1))
		Expect(dispatched[0].Options.Prompt).To(Equal("triage"))
		Expect(dispatched[0].Options.RunMode).To(Equal(types.ModePlan))
	})

	// run, plan and triage are one code path distinguished only by the prompt
	// name, which is what makes adding a fourth prompt free.
	It("dispatches each named prompt through the same path", func() {
		todo := newTodo("Prompt axis")

		_, result := postTodoAction("plan", []string{todo.ID}, nil)

		Expect(result.Applied).To(Equal(1))
		Expect(dispatched).To(HaveLen(1))
		Expect(dispatched[0].Options.Prompt).To(Equal("plan"))
	})

	// The dashboard serves the approval endpoint, so a run started from it must
	// be admitted to ask for one — otherwise a bulk run silently behaves
	// differently from the single run started on the same page.
	It("resolves runs the way the dashboard's own single run does", func() {
		todo := newTodo("Approvable")

		_, result := postTodoAction("run", []string{todo.ID}, nil)

		Expect(result.Applied).To(Equal(1))
		Expect(dispatched).To(HaveLen(1))
		Expect(dispatched[0].Options.Spec.Mode).NotTo(BeEmpty(),
			"the dashboard's (driver, mode) catalog must have been applied")
	})

	It("reports per-item run failures without abandoning the rest of the batch", func() {
		exploding := newTodo("Triage explodes")
		survivor := newTodo("Triage survivor")

		recorder, result := postTodoAction("triage", []string{exploding.ID, survivor.ID}, nil)

		Expect(recorder.Code).To(Equal(http.StatusOK), "a partial batch is still a 200")
		Expect(result.Applied).To(Equal(1))
		Expect(result.Failed).To(Equal(1))
		Expect(result.Results[0].Error).To(ContainSubstring("driver unavailable"))
		Expect(result.Results[1].Error).To(BeEmpty())
	})

	// A duplicated target would start two runs against one TODO and race the
	// optimistic version check.
	It("rejects duplicate targets outright rather than half-dispatching", func() {
		todo := newTodo("Triage duplicate")

		recorder, _ := postTodoAction("triage", []string{todo.ID, todo.ID}, nil)

		Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		Expect(recorder.Body.String()).To(ContainSubstring("duplicates"))
		Expect(dispatched).To(BeEmpty())
	})
})

// The catalog is what the dashboard renders its selection toolbar from, so a
// registered action that never reaches /api/entities is an action the UI cannot
// offer — which is the exact drift this whole surface exists to end.
var _ = Describe("todo entity catalog", func() {
	It("publishes every bulk action with what a toolbar needs to draw and dispatch it", func() {
		actions := todoEntityCatalog()
		byName := map[string]int{}

		for i, action := range actions {
			byName[action.Name] = i
			Expect(action.Short).NotTo(BeEmpty(), action.Name)
			Expect(action.SupportsFilterMode).To(BeTrue(), action.Name+" must be reachable by filter")
			Expect(action.ToolHints).NotTo(BeNil(), action.Name)
			Expect(action.ToolHints.Icon).NotTo(BeEmpty(), action.Name)
			Expect(action.ToolHints.Group).NotTo(BeEmpty(), action.Name)
			// Without these the dashboard would have to guess the route, and
			// guessing POST 404s on delete.
			Expect(action.Method).NotTo(BeEmpty(), action.Name)
			Expect(action.Path).To(ContainSubstring("{id}"), action.Name)
		}
		for _, want := range []string{"status", "priority", "labels", "comment", "delete", "run", "plan", "triage"} {
			Expect(byName).To(HaveKey(want))
		}

		// The enum is what lets the UI render a status picker instead of
		// duplicating the assignable list in React.
		status := actions[byName["status"]]
		Expect(status.ParamSchema).NotTo(BeNil())
		Expect(status.ParamSchema.Required).To(ContainElement("to"))
		Expect(status.ParamSchema.Properties["to"].Enum).To(
			ConsistOf("draft", "pending", "verified", "completed", "skipped"))

		deleteAction := actions[byName["delete"]]
		Expect(deleteAction.ToolHints.DestructiveHint).NotTo(BeNil())
		Expect(*deleteAction.ToolHints.DestructiveHint).To(BeTrue(),
			"a UI has to know to confirm before deleting a filter-matched selection")
	})
})
