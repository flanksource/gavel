package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"time"

	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/land"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// landTestProvider stands in for native storage: the land call itself is
// stubbed, so the handler only needs a provider that satisfies land.Provider.
type landTestProvider struct{ *uiTestTODOProvider }

func (landTestProvider) Captain() *captaindb.DB         { return nil }
func (landTestProvider) Repository() *native.Repository { return nil }

type landCall struct {
	todoID string
	opts   land.Options
}

func postTodoLand(body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/todos/land", bytes.NewReader([]byte(body)))
	recorder := httptest.NewRecorder()
	(&Server{}).Handler().ServeHTTP(recorder, request)
	return recorder
}

var _ = Describe("todo land API", func() {
	var (
		dir      string
		todo     *types.TODO
		calls    []landCall
		result   *native.RunLanding
		landErr  error
		provider todos.Provider
	)

	BeforeEach(func() {
		dir = filepath.Clean(GinkgoT().TempDir())
		base := uiTestProviderFor(dir)
		provider = landTestProvider{base}
		var err error
		todo, err = base.Create(GinkgoT().Context(), todos.CreateRequest{Title: "Land me", Status: types.StatusPending})
		Expect(err).NotTo(HaveOccurred())
		calls, result, landErr = nil, nil, nil

		previousProvider, previousLand := openTodoProvider, landTodoRun
		openTodoProvider = func(context.Context, string) (todos.Provider, error) { return provider, nil }
		landTodoRun = func(_ context.Context, _ land.Provider, target *types.TODO, opts land.Options) (*native.RunLanding, error) {
			calls = append(calls, landCall{todoID: target.ID, opts: opts})
			return result, landErr
		}
		DeferCleanup(func() { openTodoProvider, landTodoRun = previousProvider, previousLand })
	})

	requestFor := func(fields string) string {
		return `{"ref":"` + todo.ID + `","dir":"` + dir + `"` + fields + `}`
	}

	It("lands the referenced todo as a draft PR against the requested base and answers the landing", func() {
		prNumber := 9
		result = &native.RunLanding{
			PromptRunID: uuid.New(), Via: native.LandingPR, TargetBranch: "release", LandedSHA: "abc1234",
			PRNumber: &prNumber, PRURL: "https://github.com/acme/api/pull/9", LandedAt: time.Now().UTC(),
		}

		recorder := postTodoLand(requestFor(`,"via":"pr","base":"origin/release","draft":true`))

		Expect(recorder.Code).To(Equal(http.StatusOK), recorder.Body.String())
		Expect(calls).To(HaveLen(1))
		Expect(calls[0].todoID).To(Equal(todo.ID))
		Expect(calls[0].opts.Via).To(Equal(native.LandingPR))
		Expect(calls[0].opts.Base).To(Equal("origin/release"))
		Expect(calls[0].opts.Draft).To(BeTrue())
		Expect(calls[0].opts.Deps.CreatePR).NotTo(BeNil(), "a PR landing needs the real GitHub and content calls")
		var response struct {
			Landing map[string]any `json:"landing"`
		}
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response.Landing).To(HaveKeyWithValue("promptRunId", result.PromptRunID.String()))
		Expect(response.Landing).To(HaveKeyWithValue("via", "pr"))
		Expect(response.Landing).To(HaveKeyWithValue("prNumber", float64(prNumber)))
		Expect(response.Landing).To(HaveKeyWithValue("prUrl", "https://github.com/acme/api/pull/9"))
		Expect(response.Landing).To(HaveKey("landedAt"))
	})

	DescribeTable("rejects a request before landing", func(fields, message string) {
		recorder := postTodoLand(requestFor(fields))
		Expect(recorder.Code).To(Equal(http.StatusBadRequest), recorder.Body.String())
		Expect(recorder.Body.String()).To(ContainSubstring(message))
		Expect(calls).To(BeEmpty())
	},
		Entry("unknown via", `,"via":"rebase"`, `via \"rebase\"`),
		Entry("missing via", ``, `via \"\"`),
		Entry("base on a merge", `,"via":"merge","base":"main"`, "base and draft only apply to a pr landing"),
	)

	It("rejects a request without a ref", func() {
		recorder := postTodoLand(`{"dir":"` + dir + `","via":"merge"}`)
		Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		Expect(recorder.Body.String()).To(ContainSubstring("ref is required"))
	})

	It("answers 501 when the todo storage is not native", func() {
		provider = uiTestProviderFor(dir)
		recorder := postTodoLand(requestFor(`,"via":"merge"`))
		Expect(recorder.Code).To(Equal(http.StatusNotImplemented), recorder.Body.String())
		Expect(calls).To(BeEmpty())
	})

	DescribeTable("maps a refused landing to 409", func(err error) {
		landErr = err
		recorder := postTodoLand(requestFor(`,"via":"merge"`))
		Expect(recorder.Code).To(Equal(http.StatusConflict), recorder.Body.String())
		Expect(recorder.Body.String()).To(ContainSubstring(err.Error()))
	},
		Entry("cherry-pick conflict", &land.ConflictError{Branch: "main", Paths: []string{"feature.txt"}, Err: errors.New("exit 1")}),
		Entry("already landed", native.ErrAlreadyLanded),
		Entry("dirty worktree", land.ErrWorktreeDirty),
		Entry("no commits", land.ErrNoCommits),
		Entry("checkout not ready", land.ErrCheckoutNotReady),
	)

	It("returns the recorded landing alongside a cleanup failure", func() {
		result = &native.RunLanding{PromptRunID: uuid.New(), Via: native.LandingMerge, TargetBranch: "main", LandedSHA: "abc1234"}
		landErr = &land.CleanupError{Err: errors.New("git worktree remove: contains untracked files")}

		recorder := postTodoLand(requestFor(`,"via":"merge"`))

		Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		var response struct {
			Landing *native.RunLanding `json:"landing"`
			Error   string             `json:"error"`
		}
		Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
		Expect(response.Landing).NotTo(BeNil())
		Expect(response.Landing.PromptRunID).To(Equal(result.PromptRunID))
		Expect(response.Error).To(ContainSubstring("contains untracked files"))
	})
})
