package todos

import (
	"errors"

	captaincli "github.com/flanksource/captain/pkg/cli"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("resolving a todo session's plan through Captain", func() {
	var captain *captaindb.DB

	BeforeEach(func(ctx SpecContext) {
		handle := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_todos_resolve_plan"})
		var err error
		captain, err = captaindb.Open(ctx, captaindb.WithDSN(handle.DSN()), captaindb.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(captain.Close()).To(Succeed()) })
	})

	It("keeps Captain's ErrNoPlan for a session that recorded no plan", func(ctx SpecContext) {
		session, err := captain.CreateOrGetSession(ctx, captaindb.CreateSessionInput{
			ID: uuid.New(), Source: "gavel", Provider: "cmux-claude", HostID: "resolve-plan-test",
		})
		Expect(err).NotTo(HaveOccurred())

		_, _, err = ResolveSessionPlan(ctx, captain, session.ID.String())

		Expect(err).To(MatchError(captaincli.ErrNoPlan))
		Expect(err).To(MatchError(ContainSubstring(session.ID.String())))
	})

	It("reports an unknown session as the resolver's own error, not a missing plan", func(ctx SpecContext) {
		unknown := uuid.NewString()

		_, _, err := ResolveSessionPlan(ctx, captain, unknown)

		Expect(err).To(MatchError(captaindb.ErrSessionNotFound))
		Expect(errors.Is(err, captaincli.ErrNoPlan)).To(BeFalse())
	})
})
