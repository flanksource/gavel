package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/commit"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

const (
	jsonContentType = "application/json"
	formContentType = "application/x-www-form-urlencoded"
)

// parentlessProvider is storage without a hierarchy: embedding the interface
// hides the test provider's SetParent.
type parentlessProvider struct{ todos.Provider }

var _ = Describe("todo parent API", func() {
	var (
		dir      string
		server   *Server
		provider *uiTestTODOProvider
		family   map[string]*types.TODO
	)

	seed := func(request todos.CreateRequest) *types.TODO {
		GinkgoHelper()
		todo, err := provider.Create(GinkgoT().Context(), request)
		Expect(err).NotTo(HaveOccurred())
		return todo
	}
	serve := func(method, path, contentType string, body io.Reader) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, path, body)
		if contentType != "" {
			request.Header.Set("Content-Type", contentType)
		}
		recorder := httptest.NewRecorder()
		server.Handler().ServeHTTP(recorder, request)
		return recorder
	}
	inWorkspace := func(path string) string { return path + "?dir=" + url.QueryEscape(dir) }
	// expand substitutes {parent}, {child} and {sibling} with that todo's id.
	expand := func(template string) string {
		for name, todo := range family {
			template = strings.ReplaceAll(template, "{"+name+"}", todo.ID)
		}
		return template
	}
	// storedFamily is the hierarchy and titles as storage holds them now.
	storedFamily := func() map[string]string {
		stored := map[string]string{}
		for _, todo := range provider.items {
			stored[todo.Title] = todo.ParentID
		}
		return stored
	}
	decodeObject := func(recorder *httptest.ResponseRecorder) map[string]any {
		GinkgoHelper()
		var object map[string]any
		Expect(json.Unmarshal(recorder.Body.Bytes(), &object)).To(Succeed(), recorder.Body.String())
		return object
	}

	BeforeEach(func() {
		dir = filepath.Clean(GinkgoT().TempDir())
		server = &Server{ghOpts: github.Options{WorkDir: dir}}
		provider = uiTestProviderFor(dir)
		parent := seed(todos.CreateRequest{Title: "Ship the release", Status: types.StatusPending})
		family = map[string]*types.TODO{
			"parent":  parent,
			"child":   seed(todos.CreateRequest{Title: "Write the changelog", Status: types.StatusCompleted, Parent: parent.ID}),
			"sibling": seed(todos.CreateRequest{Title: "Rotate the keys", Status: types.StatusPending}),
		}
		provider.createRequests = nil
	})

	It("lists children with their parentId while counting only top-level todos", func() {
		recorder := serve(http.MethodGet, inWorkspace("/api/todos"), "", nil)

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		var list struct {
			Counts todoCounts       `json:"counts"`
			Items  []map[string]any `json:"items"`
		}
		Expect(json.Unmarshal(recorder.Body.Bytes(), &list)).To(Succeed())
		Expect(list.Counts).To(Equal(todoCounts{Total: 2, Open: 2, Pending: 2}))
		Expect(list.Items).To(ConsistOf(
			And(HaveKeyWithValue("title", family["parent"].Title), Not(HaveKey("parentId"))),
			And(HaveKeyWithValue("title", family["child"].Title), HaveKeyWithValue("parentId", family["parent"].ID)),
			And(HaveKeyWithValue("title", family["sibling"].Title), Not(HaveKey("parentId"))),
		))
	})

	DescribeTable("returns parentId on the item response only for a child",
		func(name string, matchParent func() OmegaMatcher) {
			ref := url.QueryEscape(family[name].ID)
			recorder := serve(http.MethodGet, inWorkspace("/api/todos/item")+"&ref="+ref, "", nil)

			Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
			Expect(decodeObject(recorder)).To(matchParent())
		},
		Entry("a child", "child", func() OmegaMatcher { return HaveKeyWithValue("parentId", family["parent"].ID) }),
		Entry("a top-level todo", "parent", func() OmegaMatcher { return Not(HaveKey("parentId")) }),
	)

	Describe("create", func() {
		const title = "Draft the announcement"

		multipartBody := func(fields map[string]string) (io.Reader, string) {
			GinkgoHelper()
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			for key, value := range fields {
				Expect(writer.WriteField(key, value)).To(Succeed())
			}
			part, err := writer.CreateFormFile("attachment", "screen.png")
			Expect(err).NotTo(HaveOccurred())
			_, err = part.Write([]byte("png bytes"))
			Expect(err).NotTo(HaveOccurred())
			Expect(writer.Close()).To(Succeed())
			return &body, writer.FormDataContentType()
		}
		jsonBody := func(fields map[string]string) (io.Reader, string) {
			GinkgoHelper()
			body, err := json.Marshal(fields)
			Expect(err).NotTo(HaveOccurred())
			return bytes.NewReader(body), jsonContentType
		}
		// created unwraps the todo from either create endpoint's response.
		created := func(recorder *httptest.ResponseRecorder) map[string]any {
			GinkgoHelper()
			object := decodeObject(recorder)
			if wrapped, ok := object["todo"].(map[string]any); ok {
				return wrapped
			}
			return object
		}

		BeforeEach(func() {
			previous := attachmentsDir
			attachmentsDir = GinkgoT().TempDir()
			DeferCleanup(func() { attachmentsDir = previous })
		})

		DescribeTable("makes the new todo a child of the named parent",
			func(path string, encode func(map[string]string) (io.Reader, string)) {
				body, contentType := encode(map[string]string{"title": title, "parent": family["parent"].ID})

				recorder := serve(http.MethodPost, inWorkspace(path), contentType, body)

				Expect(recorder.Code).To(Equal(http.StatusCreated), recorder.Body.String())
				Expect(created(recorder)).To(HaveKeyWithValue("parentId", family["parent"].ID))
				Expect(provider.createRequests).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
					"Title":    Equal(title),
					"Parent":   Equal(family["parent"].ID),
					"NoParent": BeFalse(),
					"Origin":   BeNil(),
				})))
				Expect(storedFamily()).To(HaveKeyWithValue(title, family["parent"].ID))
			},
			Entry("JSON on /api/todos", "/api/todos", jsonBody),
			Entry("JSON on /api/todos/new", "/api/todos/new", jsonBody),
			Entry("multipart with a screenshot on /api/todos/new", "/api/todos/new", multipartBody),
		)

		DescribeTable("never takes the origin from the server's own environment",
			func(path string) {
				GinkgoT().Setenv(commit.EnvIssueID, family["parent"].ID)
				GinkgoT().Setenv(commit.EnvSessionID, "019f5b29-7890-7c11-8e7a-838e5d373e39")
				body, contentType := jsonBody(map[string]string{"title": title})

				recorder := serve(http.MethodPost, inWorkspace(path), contentType, body)

				Expect(recorder.Code).To(Equal(http.StatusCreated), recorder.Body.String())
				Expect(created(recorder)).NotTo(HaveKey("parentId"))
				Expect(provider.createRequests).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
					"Title":  Equal(title),
					"Parent": BeEmpty(),
					"Origin": BeNil(),
				})))
			},
			Entry("/api/todos", "/api/todos"),
			Entry("/api/todos/new", "/api/todos/new"),
		)

		DescribeTable("refuses a parent that cannot take children and creates nothing",
			func(path, parentRef string, status int, message string) {
				before := storedFamily()
				body, contentType := jsonBody(map[string]string{"title": title, "parent": expand(parentRef)})

				recorder := serve(http.MethodPost, inWorkspace(path), contentType, body)

				Expect(recorder.Code).To(Equal(status), recorder.Body.String())
				Expect(decodeObject(recorder)).To(HaveKeyWithValue("error", ContainSubstring(message)))
				Expect(storedFamily()).To(Equal(before))
			},
			Entry("a child as the parent on /api/todos", "/api/todos", "{child}", http.StatusBadRequest, native.ErrInvalidParent.Error()),
			Entry("a child as the parent on /api/todos/new", "/api/todos/new", "{child}", http.StatusBadRequest, native.ErrInvalidParent.Error()),
			Entry("an unknown parent on /api/todos", "/api/todos", "todo-unknown", http.StatusNotFound, native.ErrNotFound.Error()),
			Entry("an unknown parent on /api/todos/new", "/api/todos/new", "todo-unknown", http.StatusNotFound, native.ErrNotFound.Error()),
		)
	})

	Describe("patch", func() {
		patch := func(contentType, body string) *httptest.ResponseRecorder {
			return serve(http.MethodPatch, inWorkspace("/api/todos/item"), contentType, strings.NewReader(expand(body)))
		}

		DescribeTable("applies the parent field",
			func(contentType, target, body, wantParent string) {
				recorder := patch(contentType, body)

				Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
				matchParent := Not(HaveKey("parentId"))
				if wantParent != "" {
					matchParent = HaveKeyWithValue("parentId", expand(wantParent))
				}
				Expect(decodeObject(recorder)).To(matchParent)
				Expect(family[target].ParentID).To(Equal(expand(wantParent)))
			},
			Entry("JSON: a ref alone makes the todo a child", jsonContentType, "sibling", `{"ref":"{sibling}","parent":"{parent}"}`, "{parent}"),
			Entry("JSON: a ref sent with a body edit and a comment still makes the todo a child", jsonContentType, "sibling",
				`{"ref":"{sibling}","parent":"{parent}","body":"Rotate before Friday","comment":"split out of the release"}`, "{parent}"),
			Entry("JSON: an empty string detaches the child", jsonContentType, "child", `{"ref":"{child}","parent":""}`, ""),
			Entry("JSON: an absent field leaves the parent alone", jsonContentType, "child", `{"ref":"{child}","priority":"low"}`, "{parent}"),
			Entry("form: a ref alone makes the todo a child", formContentType, "sibling", `ref={sibling}&parent={parent}`, "{parent}"),
			Entry("form: an empty value detaches the child", formContentType, "child", `ref={child}&parent=`, ""),
			Entry("form: an absent field leaves the parent alone", formContentType, "child", `ref={child}&priority=low`, "{parent}"),
		)

		DescribeTable("refuses a parent the hierarchy forbids and applies none of the patch",
			func(body string, status int, message string) {
				before := storedFamily()

				recorder := patch(jsonContentType, body)

				Expect(recorder.Code).To(Equal(status), recorder.Body.String())
				Expect(decodeObject(recorder)).To(HaveKeyWithValue("error", ContainSubstring(message)))
				Expect(storedFamily()).To(Equal(before))
			},
			Entry("the todo itself", `{"ref":"{sibling}","parent":"{sibling}","title":"Renamed"}`,
				http.StatusBadRequest, native.ErrInvalidParent.Error()),
			Entry("a child as the parent", `{"ref":"{sibling}","parent":"{child}","title":"Renamed"}`,
				http.StatusBadRequest, native.ErrInvalidParent.Error()),
			Entry("a todo that has children", `{"ref":"{parent}","parent":"{sibling}","title":"Renamed"}`,
				http.StatusBadRequest, native.ErrInvalidParent.Error()),
			Entry("an unknown parent", `{"ref":"{sibling}","parent":"todo-unknown","title":"Renamed"}`,
				http.StatusNotFound, native.ErrNotFound.Error()),
		)

		It("answers 501 and applies none of the patch when storage has no hierarchy", func() {
			previous := openTodoProvider
			openTodoProvider = func(context.Context, string) (todos.Provider, error) {
				return parentlessProvider{provider}, nil
			}
			DeferCleanup(func() { openTodoProvider = previous })
			before := storedFamily()

			recorder := patch(jsonContentType, `{"ref":"{sibling}","parent":"{parent}","title":"Renamed"}`)

			Expect(recorder.Code).To(Equal(http.StatusNotImplemented), recorder.Body.String())
			Expect(decodeObject(recorder)).To(HaveKeyWithValue("error", ContainSubstring("native TODO storage")))
			Expect(storedFamily()).To(Equal(before))
		})

		It("reports a failed re-read instead of answering with the pre-write todo", func() {
			provider.rereadErr = errors.New("connection reset by peer")

			recorder := patch(jsonContentType, `{"ref":"{sibling}","parent":"{parent}","title":"Renamed"}`)

			Expect(recorder.Code).To(Equal(http.StatusInternalServerError), recorder.Body.String())
			Expect(decodeObject(recorder)).To(Equal(map[string]any{
				"error": "re-read TODO " + family["sibling"].ID + " after update: connection reset by peer",
			}))
		})

		It("names parent among the operations an empty patch could have carried", func() {
			recorder := patch(jsonContentType, `{"ref":"{sibling}"}`)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
			Expect(decodeObject(recorder)).To(HaveKeyWithValue("error",
				"status, priority, title, body, labels, parent, or comment is required"))
		})
	})

	Describe("delete", func() {
		var open *types.TODO

		remove := func(name, query string) *httptest.ResponseRecorder {
			return serve(http.MethodDelete, inWorkspace("/api/todos/item")+"&ref="+url.QueryEscape(family[name].ID)+query, "", nil)
		}

		BeforeEach(func() {
			open = seed(todos.CreateRequest{Title: "Cut the tag", Status: types.StatusPending, Parent: family["parent"].ID})
		})

		It("answers 409 with the count and the choices when the parent has open children and nobody chose", func() {
			before := storedFamily()

			recorder := remove("parent", "")

			Expect(recorder.Code).To(Equal(http.StatusConflict), recorder.Body.String())
			Expect(decodeObject(recorder)).To(Equal(map[string]any{
				"error": native.ErrOpenChildren.Error() + ": 1 under issue " + family["parent"].ID + "; " + todos.ChildrenChoice,
			}))
			Expect(storedFamily()).To(Equal(before))
		})

		DescribeTable("closes the parent once the open children are settled, leaving a closed child attached",
			func(children string, openChild func() OmegaMatcher) {
				recorder := remove("parent", "&children="+children)

				Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
				Expect(storedFamily()).To(SatisfyAll(
					Not(HaveKey(family["parent"].Title)),
					HaveKeyWithValue(family["child"].Title, family["parent"].ID),
					HaveKeyWithValue(family["sibling"].Title, ""),
					openChild(),
				))
			},
			Entry("archive deletes them too", "archive", func() OmegaMatcher { return Not(HaveKey(open.Title)) }),
			Entry("detach makes them top-level", "detach", func() OmegaMatcher { return HaveKeyWithValue(open.Title, "") }),
		)

		DescribeTable("ignores the choice for a todo with no open children",
			func(name, query string) {
				recorder := remove(name, query)

				Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
				Expect(storedFamily()).NotTo(HaveKey(family[name].Title))
			},
			Entry("a plain todo, no choice", "sibling", ""),
			Entry("a plain todo, archive", "sibling", "&children=archive"),
			Entry("a child, detach", "child", "&children=detach"),
		)

		DescribeTable("answers 400 to a choice that is neither archive nor detach and deletes nothing",
			func(name string) {
				before := storedFamily()

				recorder := remove(name, "&children=purge")

				Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
				Expect(decodeObject(recorder)).To(Equal(map[string]any{
					"error": `unknown children value "purge": use archive or detach`,
				}))
				Expect(storedFamily()).To(Equal(before))
			},
			Entry("on a parent with open children", "parent"),
			Entry("on a todo with none", "sibling"),
		)

		// The dashboard reads any 409 on this request as "ask what to do with the
		// open children" and shows the message as the question.
		DescribeTable("keeps 409 for the open-children refusal alone",
			func(err error, status int) {
				recorder := httptest.NewRecorder()

				writeTodoDeleteError(recorder, http.StatusInternalServerError, fmt.Errorf("delete: %w", err))

				Expect(recorder.Code).To(Equal(status))
				Expect(decodeObject(recorder)).To(Equal(map[string]any{"error": "delete: " + err.Error()}))
			},
			Entry("open children", native.ErrOpenChildren, http.StatusConflict),
			Entry("a stale version", native.ErrVersionConflict, http.StatusPreconditionFailed),
			Entry("an ambiguous reference", native.ErrAmbiguousReference, http.StatusPreconditionFailed),
			Entry("a todo that is gone", native.ErrNotFound, http.StatusNotFound),
			Entry("an unrelated failure", errors.New("connection reset by peer"), http.StatusInternalServerError),
		)
	})

	DescribeTable("maps a hierarchy refusal to its status whatever the handler's fallback",
		func(err error, status int) {
			recorder := httptest.NewRecorder()

			writeTodoError(recorder, http.StatusInternalServerError, fmt.Errorf("apply: %w", err))

			Expect(recorder.Code).To(Equal(status))
			Expect(decodeObject(recorder)).To(Equal(map[string]any{"error": "apply: " + err.Error()}))
		},
		Entry("an invalid parent is a bad request", native.ErrInvalidParent, http.StatusBadRequest),
		Entry("a move or delete inside a hierarchy is a conflict", native.ErrIssueInHierarchy, http.StatusConflict),
		Entry("closing a todo over its open children is a conflict", native.ErrOpenChildren, http.StatusConflict),
		Entry("an unrelated failure keeps the fallback", errors.New("connection reset by peer"), http.StatusInternalServerError),
	)
})
