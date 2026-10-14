package ui

import (
	"encoding/json"
	"time"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/todos/native"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("todo session detail attempts", func() {
	It("returns the recorded workspace per attempt", func() {
		issueID, runID, verifyID := uuid.New(), uuid.New(), uuid.New()
		created := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
		links := []native.PromptRunLink{
			{IssueID: issueID, PromptRunID: verifyID, StepKind: native.StepVerify, CreatedAt: created.Add(time.Minute)},
			{IssueID: issueID, PromptRunID: runID, StepKind: native.StepRun, CreatedAt: created},
		}
		workspace := &api.WorkspaceRecord{
			Cwd: "/repo/.worktrees/shell-289107d9",
			Worktree: &api.WorktreeState{
				Repo: "/repo", Path: "/repo/.worktrees/shell-289107d9", Branch: "shell/289107d9",
				Base: "ba5e000", Setup: "5e7a000", Head: "d0b5c96", Kept: true, KeptReason: "uncommitted changes",
				Dirty: []string{"go.mod"},
			},
			Commits: []api.CommitRecord{{SHA: "d0b5c96", Message: "feat: record the workspace"}},
		}
		overviews := []captaindb.PromptRunOverview{
			{PromptRun: captaindb.PromptRun{ID: runID, Workspace: workspace}},
			{PromptRun: captaindb.PromptRun{ID: verifyID}},
		}

		attempts, err := assembleTodoAttempts(todoAttemptRecords{Links: links, Overviews: overviews})
		Expect(err).NotTo(HaveOccurred())
		encoded, err := json.Marshal(todoSessionDetailResponse{Attempts: attempts})
		Expect(err).NotTo(HaveOccurred())

		var decoded struct {
			Attempts []struct {
				Step      string           `json:"step"`
				Workspace *json.RawMessage `json:"workspace"`
			} `json:"attempts"`
		}
		Expect(json.Unmarshal(encoded, &decoded)).To(Succeed())
		Expect(decoded.Attempts).To(HaveLen(2))
		Expect(decoded.Attempts[0].Step).To(Equal("verify"))
		Expect(decoded.Attempts[0].Workspace).To(BeNil())
		Expect(decoded.Attempts[1].Step).To(Equal("run"))
		Expect(decoded.Attempts[1].Workspace).NotTo(BeNil())
		Expect(string(*decoded.Attempts[1].Workspace)).To(MatchJSON(`{
			"cwd": "/repo/.worktrees/shell-289107d9",
			"worktree": {
				"repo": "/repo", "path": "/repo/.worktrees/shell-289107d9", "branch": "shell/289107d9",
				"base": "ba5e000", "setup": "5e7a000", "head": "d0b5c96",
				"kept": true, "keptReason": "uncommitted changes", "dirty": ["go.mod"]
			},
			"commits": [{"sha": "d0b5c96", "message": "feat: record the workspace"}]
		}`))
	})

	It("returns the recorded landing per attempt", func() {
		issueID, landedRun, openRun := uuid.New(), uuid.New(), uuid.New()
		created := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
		links := []native.PromptRunLink{
			{IssueID: issueID, PromptRunID: landedRun, StepKind: native.StepRun, CreatedAt: created},
			{IssueID: issueID, PromptRunID: openRun, StepKind: native.StepRun, CreatedAt: created.Add(time.Hour)},
		}
		overviews := []captaindb.PromptRunOverview{
			{PromptRun: captaindb.PromptRun{ID: landedRun}},
			{PromptRun: captaindb.PromptRun{ID: openRun}},
		}
		prNumber := 42
		landing := native.RunLanding{
			PromptRunID: landedRun, Via: native.LandingPR, TargetBranch: "main",
			LandedSHA: "0123456789abcdef0123456789abcdef01234567", PRNumber: &prNumber,
			PRURL: "https://github.com/acme/widgets/pull/42", CommitCount: 2, LandedAt: created.Add(2 * time.Hour),
		}

		attempts, err := assembleTodoAttempts(todoAttemptRecords{Links: links, Overviews: overviews, Landings: []native.RunLanding{landing}})
		Expect(err).NotTo(HaveOccurred())
		encoded, err := json.Marshal(todoSessionDetailResponse{Attempts: attempts})
		Expect(err).NotTo(HaveOccurred())

		var decoded struct {
			Attempts []struct {
				PromptRunID uuid.UUID        `json:"promptRunId"`
				Landing     *json.RawMessage `json:"landing"`
			} `json:"attempts"`
		}
		Expect(json.Unmarshal(encoded, &decoded)).To(Succeed())
		Expect(decoded.Attempts).To(HaveLen(2))
		Expect(decoded.Attempts[0].PromptRunID).To(Equal(openRun))
		Expect(decoded.Attempts[0].Landing).To(BeNil())
		Expect(decoded.Attempts[1].PromptRunID).To(Equal(landedRun))
		Expect(decoded.Attempts[1].Landing).NotTo(BeNil())
		Expect(string(*decoded.Attempts[1].Landing)).To(MatchJSON(`{
			"promptRunId": "` + landedRun.String() + `", "via": "pr", "targetBranch": "main",
			"landedSha": "0123456789abcdef0123456789abcdef01234567",
			"prNumber": 42, "prUrl": "https://github.com/acme/widgets/pull/42", "commitCount": 2,
			"landedAt": "2026-09-01T12:00:00Z"
		}`))
	})

	It("fails loudly when a landing names a run that is not linked", func() {
		issueID, runID := uuid.New(), uuid.New()
		links := []native.PromptRunLink{{IssueID: issueID, PromptRunID: runID, StepKind: native.StepRun}}
		overviews := []captaindb.PromptRunOverview{{PromptRun: captaindb.PromptRun{ID: runID}}}
		stray := native.RunLanding{PromptRunID: uuid.New(), Via: native.LandingMerge, TargetBranch: "main", LandedSHA: "abc1234"}
		_, err := assembleTodoAttempts(todoAttemptRecords{Links: links, Overviews: overviews, Landings: []native.RunLanding{stray}})
		Expect(err).To(MatchError(ContainSubstring(stray.PromptRunID.String())))
	})

	It("fails loudly when a linked prompt run has no overview", func() {
		links := []native.PromptRunLink{{IssueID: uuid.New(), PromptRunID: uuid.New(), StepKind: native.StepRun}}
		_, err := assembleTodoAttempts(todoAttemptRecords{Links: links})
		Expect(err).To(MatchError(captaindb.ErrPromptRunNotFound))
	})

	// A run Captain finished can still fail once gavel persists its outcome; the
	// lifecycle records why on a failed lifecycle_outcome event, and the attempt
	// reads as failed with that error while its state stays Captain's.
	It("reports an attempt whose outcome could not be persisted as failed with the error", func() {
		issueID, unpersistedRun, laterRun := uuid.New(), uuid.New(), uuid.New()
		created := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
		const persistErr = "persist TODO attempt: persist Captain plan revision: save Captain plan: plan hook plan-tldr failed"
		links := []native.PromptRunLink{
			{IssueID: issueID, PromptRunID: unpersistedRun, StepKind: native.StepPlan, CreatedAt: created},
			{IssueID: issueID, PromptRunID: laterRun, StepKind: native.StepPlan, CreatedAt: created.Add(time.Hour)},
		}
		succeeded := func(id uuid.UUID) captaindb.PromptRunOverview {
			return captaindb.PromptRunOverview{
				PromptRun: captaindb.PromptRun{ID: id, State: captaindb.PromptRunStateSucceeded, Phase: captaindb.PromptRunPhaseFinished},
				Status:    captaindb.PromptRunStatusCompleted,
			}
		}
		outcomeEvent := func(sequence int64, runID uuid.UUID, payload string) native.Event {
			return native.Event{
				IssueID: issueID, Sequence: sequence, Kind: lifecycle.EventLifecycleOutcome,
				Payload: json.RawMessage(`{"promptRunId":"` + runID.String() + `",` + payload + `}`),
			}
		}
		events := []native.Event{
			outcomeEvent(1, unpersistedRun, `"status":"failed","error":"`+persistErr+`"`),
			{IssueID: issueID, Sequence: 2, Kind: "comment", Payload: json.RawMessage(`{}`)},
			outcomeEvent(3, laterRun, `"status":"review"`),
		}
		unpersisted, err := unpersistedOutcomes(events)
		Expect(err).NotTo(HaveOccurred())

		attempts, err := assembleTodoAttempts(todoAttemptRecords{
			Links: links, Overviews: []captaindb.PromptRunOverview{succeeded(unpersistedRun), succeeded(laterRun)},
			Unpersisted: unpersisted,
		})
		Expect(err).NotTo(HaveOccurred())
		encoded, err := json.Marshal(todoSessionDetailResponse{Attempts: attempts})
		Expect(err).NotTo(HaveOccurred())

		type attemptJSON struct {
			PromptRunID  uuid.UUID `json:"promptRunId"`
			State        string    `json:"state"`
			Status       string    `json:"status"`
			Error        string    `json:"error"`
			OutcomeError string    `json:"outcomeError"`
		}
		var decoded struct {
			Attempts []attemptJSON `json:"attempts"`
		}
		Expect(json.Unmarshal(encoded, &decoded)).To(Succeed())
		Expect(decoded.Attempts).To(Equal([]attemptJSON{
			{PromptRunID: laterRun, State: "succeeded", Status: "completed"},
			{PromptRunID: unpersistedRun, State: "succeeded", Status: "failed", Error: persistErr, OutcomeError: persistErr},
		}))
	})

	It("lets a run's later persisted outcome clear an earlier unpersisted one", func() {
		issueID, runID := uuid.New(), uuid.New()
		events := []native.Event{
			{IssueID: issueID, Sequence: 1, Kind: lifecycle.EventLifecycleOutcome,
				Payload: json.RawMessage(`{"promptRunId":"` + runID.String() + `","status":"failed","error":"update TODO state: version conflict"}`)},
			{IssueID: issueID, Sequence: 2, Kind: lifecycle.EventLifecycleOutcome,
				Payload: json.RawMessage(`{"promptRunId":"` + runID.String() + `","status":"pending"}`)},
		}
		unpersisted, err := unpersistedOutcomes(events)
		Expect(err).NotTo(HaveOccurred())
		Expect(unpersisted).To(BeEmpty())
	})
})
