package ui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("todo bulk edit API", func() {
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

		recorder, response := postTodoBulk(todoBulkRequest{
			Items:    []todoBulkTarget{{Dir: dir, Ref: first.ID}, {Dir: dir, Ref: second.ID}},
			Priority: types.PriorityHigh,
		})

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(response.Updated).To(Equal(2))
		Expect(response.Failed).To(BeZero())
		Expect(response.Results).To(HaveLen(2))
		Expect(response.Results[0].Todo).To(PointTo(MatchFields(IgnoreExtras, Fields{
			"Title": Equal("Bulk first"), "Priority": Equal(types.PriorityHigh),
		})))
		Expect(first.Priority).To(Equal(types.PriorityHigh))
		Expect(second.Priority).To(Equal(types.PriorityHigh))
	})

	It("re-statuses and comments in one request", func() {
		todo := newTodo("Bulk status", types.StatusPending, types.PriorityMedium)

		_, response := postTodoBulk(todoBulkRequest{
			Items:   []todoBulkTarget{{Dir: dir, Ref: todo.ID}},
			Status:  types.StatusCompleted,
			Comment: "closed in bulk",
		})

		Expect(response.Updated).To(Equal(1))
		Expect(todo.Status).To(Equal(types.StatusCompleted))
		Expect(provider.comments).To(ConsistOf("closed in bulk"))
	})

	It("reports per-item failures without abandoning the rest of the batch", func() {
		ok := newTodo("Bulk survivor", types.StatusPending, types.PriorityLow)

		_, response := postTodoBulk(todoBulkRequest{
			Items:    []todoBulkTarget{{Dir: dir, Ref: "todo-does-not-exist"}, {Dir: dir, Ref: ok.ID}},
			Priority: types.PriorityHigh,
		})

		Expect(response.Updated).To(Equal(1))
		Expect(response.Failed).To(Equal(1))
		Expect(response.Results[0].Error).To(ContainSubstring("todo-does-not-exist"))
		Expect(response.Results[0].Todo).To(BeNil())
		Expect(response.Results[1].Error).To(BeEmpty())
		Expect(ok.Priority).To(Equal(types.PriorityHigh))
	})

	It("resolves a todo whose workspace the caller did not name", func() {
		todo := newTodo("Bulk global", types.StatusPending, types.PriorityLow)

		_, response := postTodoBulk(todoBulkRequest{
			Items:    []todoBulkTarget{{Ref: todo.ID}},
			Priority: types.PriorityHigh,
		})

		Expect(response.Failed).To(BeZero())
		Expect(todo.Priority).To(Equal(types.PriorityHigh))
	})

	DescribeTable("rejects a request that cannot be applied",
		func(request todoBulkRequest, message string) {
			recorder, _ := postTodoBulk(request)
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(ContainSubstring(message))
		},
		Entry("with no items", todoBulkRequest{Priority: types.PriorityHigh}, "items is required"),
		Entry("with a blank ref",
			todoBulkRequest{Items: []todoBulkTarget{{Ref: "  "}}, Priority: types.PriorityHigh},
			"items[0].ref is required"),
		Entry("with a duplicated target",
			todoBulkRequest{Items: []todoBulkTarget{{Dir: "/w", Ref: "a"}, {Dir: "/w", Ref: "a"}}, Priority: types.PriorityHigh},
			"duplicates items[0]"),
		Entry("with no operation",
			todoBulkRequest{Items: []todoBulkTarget{{Ref: "a"}}},
			"status, priority, or comment is required"),
		Entry("with a run-projected status",
			todoBulkRequest{Items: []todoBulkTarget{{Ref: "a"}}, Status: types.StatusInProgress},
			"cannot be assigned"),
		Entry("with an unknown priority",
			todoBulkRequest{Items: []todoBulkTarget{{Ref: "a"}}, Priority: types.Priority("urgent")},
			"unknown priority"),
	)

	It("leaves every todo untouched when validation rejects the request", func() {
		todo := newTodo("Bulk untouched", types.StatusPending, types.PriorityLow)

		recorder, _ := postTodoBulk(todoBulkRequest{
			Items:  []todoBulkTarget{{Dir: dir, Ref: todo.ID}},
			Status: types.StatusFailed,
		})

		Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		Expect(todo.Status).To(Equal(types.StatusPending))
	})
})

func postTodoBulk(request todoBulkRequest) (*httptest.ResponseRecorder, todoBulkResponse) {
	body, err := json.Marshal(request)
	Expect(err).NotTo(HaveOccurred())
	recorder := httptest.NewRecorder()
	(&Server{}).Handler().ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/todos/bulk", bytes.NewReader(body)))
	var response todoBulkResponse
	if recorder.Code == http.StatusOK {
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed(), "body = %q", recorder.Body.String())
	}
	return recorder, response
}
