package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

type todoBatchListProvider struct {
	todos.Provider
	list func(context.Context, todos.DiscoveryFilters) (types.TODOS, error)
}

func (p todoBatchListProvider) List(ctx context.Context, filters todos.DiscoveryFilters) (types.TODOS, error) {
	return p.list(ctx, filters)
}

var _ = Describe("todo batch API", func() {
	var originalOpenTodoProvider func(context.Context, string) (todos.Provider, error)

	BeforeEach(func() {
		originalOpenTodoProvider = openTodoProvider
	})

	AfterEach(func() {
		openTodoProvider = originalOpenTodoProvider
	})

	It("returns an empty ordered result set without opening a provider", func() {
		var opens atomic.Int32
		openTodoProvider = func(context.Context, string) (todos.Provider, error) {
			opens.Add(1)
			return nil, fmt.Errorf("unexpected provider open")
		}

		recorder, response := postTodoBatch(context.Background(), []string{})

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(response.Results).To(BeEmpty())
		Expect(opens.Load()).To(BeZero())
		Expect(recorder.Body.String()).To(ContainSubstring(`"results":[]`))
	})

	It("returns one workspace identity, counts, and items", func() {
		dir := filepath.Clean(GinkgoT().TempDir())
		_, err := uiTestProviderFor(dir).Create(GinkgoT().Context(), todos.CreateRequest{
			Title: "Batch todo", Status: types.StatusPending, Priority: types.PriorityHigh,
		})
		Expect(err).NotTo(HaveOccurred())

		recorder, response := postTodoBatch(context.Background(), []string{dir})

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(response.Results).To(HaveLen(1))
		Expect(response.Results[0]).To(MatchFields(IgnoreExtras, Fields{
			"Dir":   Equal(dir),
			"Error": BeNil(),
			"Counts": PointTo(Equal(todoCounts{
				Total: 1, Open: 1, Pending: 1,
			})),
		}))
		Expect(response.Results[0].Items).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
			"Title":    Equal("Batch todo"),
			"Status":   Equal(types.StatusPending),
			"Priority": Equal(types.PriorityHigh),
		})))
	})

	It("bounds many workspace loads and preserves input order", func() {
		root := filepath.Clean(GinkgoT().TempDir())
		dirs := make([]string, todoBatchConcurrency+3)
		for i := range dirs {
			dirs[i] = filepath.Join(root, fmt.Sprintf("workspace-%d", i))
		}

		var active atomic.Int32
		var peak atomic.Int32
		entered := make(chan string, len(dirs))
		release := make(chan struct{})
		openTodoProvider = func(_ context.Context, dir string) (todos.Provider, error) {
			return todoBatchListProvider{
				Provider: uiTestProviderFor(dir),
				list: func(ctx context.Context, _ todos.DiscoveryFilters) (types.TODOS, error) {
					current := active.Add(1)
					updatePeak(&peak, current)
					defer active.Add(-1)
					entered <- dir
					select {
					case <-release:
					case <-ctx.Done():
						return nil, ctx.Err()
					}
					return types.TODOS{{
						ID: "todo-" + filepath.Base(dir),
						TODOFrontmatter: types.TODOFrontmatter{
							Title: filepath.Base(dir), CWD: dir,
							Status: types.StatusPending, Priority: types.PriorityMedium,
						},
					}}, nil
				},
			}, nil
		}

		recorder := httptest.NewRecorder()
		done := make(chan struct{})
		go func() {
			defer close(done)
			serveTodoBatch(recorder, context.Background(), dirs)
		}()

		for range todoBatchConcurrency {
			Eventually(entered).Should(Receive())
		}
		Consistently(entered, 100*time.Millisecond).ShouldNot(Receive())
		close(release)
		Eventually(done).Should(BeClosed())

		response := decodeTodoBatch(recorder)
		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(peak.Load()).To(Equal(int32(todoBatchConcurrency)))
		Expect(response.Results).To(HaveLen(len(dirs)))
		for i, result := range response.Results {
			Expect(result.Dir).To(Equal(dirs[i]))
			Expect(result.Error).To(BeNil())
			Expect(result.Items).To(ConsistOf(MatchFields(IgnoreExtras, Fields{
				"Title": Equal(filepath.Base(dirs[i])),
			})))
		}
	})

	It("keeps healthy results when one workspace fails", func() {
		root := filepath.Clean(GinkgoT().TempDir())
		dirs := []string{
			filepath.Join(root, "healthy-a"),
			filepath.Join(root, "broken"),
			filepath.Join(root, "healthy-b"),
		}
		openTodoProvider = func(_ context.Context, dir string) (todos.Provider, error) {
			if dir == dirs[1] {
				return nil, fmt.Errorf("open workspace: database unavailable")
			}
			return todoBatchListProvider{
				Provider: uiTestProviderFor(dir),
				list: func(context.Context, todos.DiscoveryFilters) (types.TODOS, error) {
					return types.TODOS{{
						ID: "todo-" + filepath.Base(dir),
						TODOFrontmatter: types.TODOFrontmatter{
							Title: filepath.Base(dir), CWD: dir,
							Status: types.StatusPending, Priority: types.PriorityMedium,
						},
					}}, nil
				},
			}, nil
		}

		recorder, response := postTodoBatch(context.Background(), dirs)

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(response.Results).To(HaveLen(3))
		Expect(response.Results[0].Error).To(BeNil())
		Expect(response.Results[0].Counts).To(PointTo(Equal(todoCounts{Total: 1, Open: 1, Pending: 1})))
		Expect(response.Results[1].Dir).To(Equal(dirs[1]))
		Expect(response.Results[1].Counts).To(BeNil())
		Expect(response.Results[1].Items).To(BeNil())
		Expect(response.Results[1].Error).To(PointTo(Equal(todoBatchError{
			Code: "load_failed", Message: "open workspace: database unavailable",
		})))
		Expect(response.Results[2].Error).To(BeNil())
		Expect(response.Results[2].Counts).To(PointTo(Equal(todoCounts{Total: 1, Open: 1, Pending: 1})))
	})

	DescribeTable("rejects invalid workspace sets",
		func(body string) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/api/todos/batch", strings.NewReader(body))
			(&Server{}).Handler().ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(ContainSubstring(`"error"`))
		},
		Entry("when dirs is missing", `{}`),
		Entry("when dirs is null", `{"dirs":null}`),
		Entry("when a dir is blank", `{"dirs":[" "]}`),
		Entry("when a dir is relative", `{"dirs":["relative/workspace"]}`),
		Entry("when a dir is not clean", `{"dirs":["/work/../workspace"]}`),
		Entry("when a dir is duplicated", `{"dirs":["/work/a","/work/a"]}`),
	)

	It("cancels in-flight loads and marks unscheduled workspaces explicitly", func() {
		root := filepath.Clean(GinkgoT().TempDir())
		dirs := make([]string, todoBatchConcurrency+2)
		for i := range dirs {
			dirs[i] = filepath.Join(root, fmt.Sprintf("workspace-%d", i))
		}
		started := make(chan string, len(dirs))
		finished := make(chan string, len(dirs))
		openTodoProvider = func(_ context.Context, dir string) (todos.Provider, error) {
			return todoBatchListProvider{
				Provider: uiTestProviderFor(dir),
				list: func(ctx context.Context, _ todos.DiscoveryFilters) (types.TODOS, error) {
					started <- dir
					<-ctx.Done()
					finished <- dir
					return nil, ctx.Err()
				},
			}, nil
		}

		ctx, cancel := context.WithCancel(context.Background())
		recorder := httptest.NewRecorder()
		done := make(chan struct{})
		go func() {
			defer close(done)
			serveTodoBatch(recorder, ctx, dirs)
		}()
		for range todoBatchConcurrency {
			Eventually(started).Should(Receive())
		}

		cancel()

		Eventually(done).Should(BeClosed())
		for range todoBatchConcurrency {
			Eventually(finished).Should(Receive())
		}
		response := decodeTodoBatch(recorder)
		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(response.Results).To(HaveLen(len(dirs)))
		for i, result := range response.Results {
			Expect(result.Dir).To(Equal(dirs[i]))
			Expect(result.Error).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Code": Equal("canceled"),
			})))
		}
	})
})

func postTodoBatch(ctx context.Context, dirs []string) (*httptest.ResponseRecorder, todoBatchResponse) {
	recorder := httptest.NewRecorder()
	serveTodoBatch(recorder, ctx, dirs)
	return recorder, decodeTodoBatch(recorder)
}

func serveTodoBatch(recorder *httptest.ResponseRecorder, ctx context.Context, dirs []string) {
	body, err := json.Marshal(map[string][]string{"dirs": dirs})
	Expect(err).NotTo(HaveOccurred())
	request := httptest.NewRequest(http.MethodPost, "/api/todos/batch", bytes.NewReader(body)).WithContext(ctx)
	(&Server{}).Handler().ServeHTTP(recorder, request)
}

func decodeTodoBatch(recorder *httptest.ResponseRecorder) todoBatchResponse {
	var response todoBatchResponse
	Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed(), "body = %q", recorder.Body.String())
	return response
}

func updatePeak(peak *atomic.Int32, current int32) {
	for {
		previous := peak.Load()
		if current <= previous || peak.CompareAndSwap(previous, current) {
			return
		}
	}
}
