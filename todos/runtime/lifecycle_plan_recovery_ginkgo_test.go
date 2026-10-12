package runtime

import (
	"context"
	"errors"
	"fmt"

	captaincli "github.com/flanksource/captain/pkg/cli"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/native"
	"github.com/flanksource/gavel/todos/types"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// A failed plan run need not have left a plan: recovering one from its session
// tolerates exactly Captain's "no plan" answer. Any other resolver error is the
// attempt's to report.
var _ = Describe("recovering the plan of a failed plan run", func() {
	var active *activeRun

	BeforeEach(func() {
		active = &activeRun{
			issue: &native.Issue{ID: uuid.New()},
			run:   &captaindb.PromptRun{ID: uuid.New(), SessionID: uuid.New()},
		}
	})

	resolveWith := func(err error) {
		original := resolveSessionPlan
		resolveSessionPlan = func(context.Context, *captaindb.DB, string) (string, string, error) {
			return "", "", err
		}
		DeferCleanup(func() { resolveSessionPlan = original })
	}

	failedPlanRun := &todos.ExecutionResult{Success: false, EndStatus: types.EndFailed, ErrorMessage: "agent crashed"}

	It("keeps the attempt when Captain reports the session has no plan", func(ctx SpecContext) {
		resolveWith(fmt.Errorf("resolve plan of session: %w: session has no plan", captaincli.ErrNoPlan))

		kept, err := (&Provider{}).persistPlanAttempt(ctx, &types.TODO{}, active, failedPlanRun)

		Expect(err).NotTo(HaveOccurred())
		Expect(kept).To(BeIdenticalTo(active))
	})

	It("reports a resolver failure instead of reading it as a missing plan", func(ctx SpecContext) {
		storeFailure := errors.New("list persisted plans: connection refused")
		resolveWith(storeFailure)

		_, err := (&Provider{}).persistPlanAttempt(ctx, &types.TODO{}, active, failedPlanRun)

		Expect(err).To(MatchError(storeFailure))
		Expect(err).To(MatchError(ContainSubstring(active.run.SessionID.String())), "the error names the session it resolved")
	})
})
