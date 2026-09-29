package todoprojection

import (
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("projection steps for a Captain row change", func() {
	var (
		rowID     = uuid.MustParse("00000000-0000-0000-0000-000000000001")
		sessionID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
		rootID    = uuid.MustParse("00000000-0000-0000-0000-000000000003")
		runID     = uuid.MustParse("00000000-0000-0000-0000-000000000004")
	)
	change := func(table captaindb.RowChangeTable, op captaindb.RowChangeOp, session, root, run *uuid.UUID) captaindb.RowChange {
		return captaindb.RowChange{Table: table, Op: op, ID: rowID, SessionID: session, RootSessionID: root, PromptRunID: run}
	}

	DescribeTable("maps each change onto the TODOs it can move",
		func(input captaindb.RowChange, want []step) {
			got, err := stepsFor(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(want))
		},
		Entry("a written session re-projects its agent family and its TODO root tree",
			change(captaindb.RowChangeSessions, captaindb.RowChangeUpdate, &rowID, &rootID, nil),
			[]step{{projectSession, rowID}, {projectRootSession, rowID}}),
		Entry("a deleted descendant session re-projects the surviving family root and touches the TODO it belonged to",
			change(captaindb.RowChangeSessions, captaindb.RowChangeDelete, &rowID, &rootID, nil),
			[]step{{projectSession, rootID}, {touchIssue, rootID}}),
		Entry("a deleted root session has no surviving row to project from",
			change(captaindb.RowChangeSessions, captaindb.RowChangeDelete, &rowID, &rowID, nil),
			[]step{{touchIssue, rowID}}),
		Entry("any prompt-run write re-projects the run onto its active TODO",
			change(captaindb.RowChangePromptRuns, captaindb.RowChangeInsert, &sessionID, &rootID, &rowID),
			[]step{{projectPromptRun, rowID}}),
		Entry("a written turn request carries its own timestamps, then re-projects run and session",
			change(captaindb.RowChangeTurnRequests, captaindb.RowChangeUpdate, &sessionID, nil, &runID),
			[]step{{projectTurnRequest, rowID}, {projectPromptRun, runID}, {projectSession, sessionID}}),
		Entry("a turn request outside any run only re-projects its session",
			change(captaindb.RowChangeTurnRequests, captaindb.RowChangeInsert, &sessionID, nil, nil),
			[]step{{projectTurnRequest, rowID}, {projectSession, sessionID}}),
		Entry("a deleted turn request touches its run's TODO at observation time",
			change(captaindb.RowChangeTurnRequests, captaindb.RowChangeDelete, &sessionID, nil, &runID),
			[]step{{touchPromptRun, runID}, {projectSession, sessionID}}),
		Entry("a written iteration re-projects its run",
			change(captaindb.RowChangePromptRunIterations, captaindb.RowChangeInsert, nil, nil, &runID),
			[]step{{projectPromptRun, runID}}),
		Entry("a deleted iteration touches its run's TODO at observation time",
			change(captaindb.RowChangePromptRunIterations, captaindb.RowChangeDelete, nil, nil, &runID),
			[]step{{touchPromptRun, runID}}),
	)

	DescribeTable("rejects a change that breaks Captain's payload contract",
		func(input captaindb.RowChange, wantErr string) {
			_, err := stepsFor(input)
			Expect(err).To(MatchError(ContainSubstring(wantErr)))
		},
		Entry("session without a family root",
			change(captaindb.RowChangeSessions, captaindb.RowChangeUpdate, &rowID, nil, nil), "rootSessionId"),
		Entry("turn request without a session",
			change(captaindb.RowChangeTurnRequests, captaindb.RowChangeInsert, nil, nil, &runID), "sessionId"),
		Entry("iteration without a prompt run",
			change(captaindb.RowChangePromptRunIterations, captaindb.RowChangeUpdate, nil, nil, nil), "promptRunId"),
		Entry("a table Captain does not notify",
			change("captain_plans", captaindb.RowChangeUpdate, nil, nil, nil), "captain_plans"),
	)
})
