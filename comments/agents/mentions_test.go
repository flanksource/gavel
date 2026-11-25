package agents

import (
	"encoding/json"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/gavel/comments"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mentionable names", func() {
	names := func(found []comments.Mentionable) map[string]string {
		out := map[string]string{}
		for _, m := range found {
			out[m.ID] = m.Name
		}
		return out
	}
	reserved := map[string]bool{comments.MentionAgent: true, "claude": true, "codex": true}

	It("names sessions by slug, a managed session without one by its label, and others by id", func() {
		sessions := []comments.AgentSession{
			{ID: "live-1", Title: "rename-pass"},
			{ID: "managed-1", Title: "Write the changelog!", Status: "ended", Managed: true, Resumable: true},
			{ID: "managed-2", Title: "managed-2", Status: "ended", Managed: true, Resumable: true},
			{ID: "live-3", Title: "Fix the footer"},
		}
		Expect(names(sessionMentionables(sessions, map[string]string{"live-1": "rename-pass"}, reserved))).To(Equal(map[string]string{
			"session:live-1": "rename-pass", "session:managed-1": "write-the-changelog", "session:managed-2": "managed-2", "session:live-3": "live-3",
		}))
	})

	It("leaves out a session whose name a runtime or @agent holds, and lists two sessions sharing a slug by slug and short id", func() {
		sessions := []comments.AgentSession{
			{ID: "aaaaaaaa-1111", Title: "x"}, {ID: "bbbbbbbb-2222", Title: "y"},
			{ID: "claude-live", Title: "z"}, {ID: "agent-live", Title: "w"}, {ID: "solo", Title: "v"},
		}
		slugs := map[string]string{"aaaaaaaa-1111": "review", "bbbbbbbb-2222": "Review", "claude-live": "Claude", "agent-live": "agent", "solo": "docs"}
		Expect(names(sessionMentionables(sessions, slugs, reserved))).To(Equal(map[string]string{
			"session:aaaaaaaa-1111": "review-aaaaaaaa", "session:bbbbbbbb-2222": "Review-bbbbbbbb", "session:solo": "docs",
		}))
	})

	It("starts a runtime mention from the rule's spec, replacing only its model", func() {
		rule := comments.DeliverTarget{Kind: "run", Prompt: "slides-review", Edits: comments.EditsPropose,
			RunOptions: json.RawMessage(`{"instructions":"Be brief","spec":{"model":"other-model","budget":{"maxTurns":7}}}`)}
		families := []api.RuntimeFamily{{Family: "claude", Modes: []api.RuntimeModeEntry{{Mode: string(api.ModeAgent), DefaultModel: "claude-default"}}}}
		found, err := runtimeMentionables(&rule, families)
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(HaveLen(1))
		target := found[0].Route.Target
		options, err := decodeRunOptions(target.RunOptions)
		Expect(err).NotTo(HaveOccurred())
		Expect([]any{target.Prompt, target.Edits, target.Fresh, options.Instructions, options.Spec.Name, options.Spec.Mode, options.Spec.Budget.MaxTurns}).
			To(Equal([]any{"slides-review", comments.EditsPropose, true, "Be brief", "claude-default", api.ModeAgent, 7}))
	})
})
