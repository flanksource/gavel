package tools_test

import (
	"context"
	"testing"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/runtime"
	"github.com/flanksource/gavel/todos/tools"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTools(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TODO tools")
}

// The tools write when they are called, so every spec drives them against the
// store-backed provider and reads the TODOs back rather than inspecting a
// proposal.
var _ = Describe("todo tools on the native runtime", Ordered, func() {
	taxonomy := []string{"bug", "enhancement"}
	var provider *runtime.Provider

	BeforeAll(func(ctx SpecContext) {
		dsn := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_todo_tools"}).DSN()
		GinkgoT().Setenv(database.EnvDSN, dsn)
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		provider, err = runtime.New(ctx, opened.Gorm(), runtime.WorkspaceOptions{
			Name: "todo-tools", RootPath: GinkgoT().TempDir(), Repositories: []string{"acme/todo-tools"},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	create := func(ctx context.Context, request todos.CreateRequest) *types.TODO {
		GinkgoHelper()
		if request.Status == "" {
			request.Status = types.StatusPending
		}
		created, err := provider.Create(ctx, request)
		Expect(err).NotTo(HaveOccurred())
		return created
	}

	reread := func(ctx context.Context, todo *types.TODO) *types.TODO {
		GinkgoHelper()
		read, err := provider.Get(ctx, todo.ShortID)
		Expect(err).NotTo(HaveOccurred())
		return read
	}

	// toolsFor builds the todos and ask sets for a run on subject.
	toolsFor := func(subject *types.TODO, preview bool) (map[string]api.ToolDefinition, *tools.Run) {
		GinkgoHelper()
		run := &tools.Run{}
		defs, err := tools.For([]string{tools.SetTodos, tools.SetAsk}, tools.Options{
			Provider: provider, Subject: subject, WorkDir: GinkgoT().TempDir(), Labels: taxonomy, Preview: preview, Run: run,
		})
		Expect(err).NotTo(HaveOccurred())
		byName := make(map[string]api.ToolDefinition, len(defs))
		for _, def := range defs {
			byName[def.Name] = def
		}
		return byName, run
	}

	call := func(ctx context.Context, defs map[string]api.ToolDefinition, name string, input map[string]any) (any, error) {
		GinkgoHelper()
		def, ok := defs[name]
		Expect(ok).To(BeTrue(), "tool %s", name)
		return def.Handler(ctx, input)
	}

	It("asks before merging or closing and allows every other tool", func() {
		defs, _ := toolsFor(&types.TODO{ID: "policy-subject"}, false)
		policies := map[string]api.ToolPolicy{}
		for name, def := range defs {
			policies[name] = def.DefaultPermission
		}
		allow, ask := api.ToolPolicyAllow, api.ToolPolicyAsk
		Expect(policies).To(Equal(map[string]api.ToolPolicy{
			"todo_search": allow, "todo_get": allow, "todo_update": allow, "todo_labels": allow,
			"todo_link": allow, "todo_link_pr": allow, "todo_set_parent": allow, "todo_comment": allow,
			"todo_plan": allow, "todo_ask": allow, "todo_merge": ask, "todo_close": ask,
		}))
		Expect(tools.NeedsApprovals([]api.ToolDefinition{defs["todo_update"], defs["todo_get"]})).To(BeFalse())
		Expect(tools.NeedsApprovals([]api.ToolDefinition{defs["todo_get"], defs["todo_close"]})).To(BeTrue())
		Expect(tools.NeedsApprovals([]api.ToolDefinition{defs[tools.AskToolName]})).To(BeTrue(), "todo_ask raises its own question")
	})

	// The claude CLI refuses an MCP tool in plan mode unless it is read-only, and
	// triage and plan run in plan mode: every tool, writers included, says so.
	It("marks every tool read-only, so plan-mode steps can call it", func() {
		defs, _ := toolsFor(&types.TODO{ID: "hint-subject"}, false)
		var notReadOnly []string
		for name, def := range defs {
			if def.ReadOnlyHint == nil || !*def.ReadOnlyHint {
				notReadOnly = append(notReadOnly, name)
			}
		}
		Expect(defs).To(HaveLen(12))
		Expect(notReadOnly).To(BeEmpty())
	})

	It("updates and labels the subject and records each write", func(ctx SpecContext) {
		subject := create(ctx, todos.CreateRequest{Title: "Parser crashes", Body: "It crashes"})
		defs, run := toolsFor(subject, false)

		_, err := call(ctx, defs, "todo_update", map[string]any{"title": "Parser panics on empty input", "priority": "high"})
		Expect(err).NotTo(HaveOccurred())
		_, err = call(ctx, defs, "todo_labels", map[string]any{"add": []any{"bug"}})
		Expect(err).NotTo(HaveOccurred())

		read := reread(ctx, subject)
		Expect([]any{read.Title, read.Priority, read.Labels}).To(Equal([]any{"Parser panics on empty input", types.PriorityHigh, []string{"bug"}}))
		Expect(run.Writes()).To(Equal([]todos.ToolWrite{
			{Name: "todo_update", Ref: subject.ShortID}, {Name: "todo_labels", Ref: subject.ShortID},
		}))
	})

	DescribeTable("refuses a write the store would accept but the tool must not",
		func(ctx SpecContext, name string, input func(other *types.TODO) map[string]any, want string) {
			subject := create(ctx, todos.CreateRequest{Title: "Refusal subject", Body: "Body"})
			other := create(ctx, todos.CreateRequest{Title: "Finished work", Body: "Done", Status: types.StatusCompleted})
			defs, run := toolsFor(subject, false)

			_, err := call(ctx, defs, name, input(other))

			Expect(err).To(MatchError(ContainSubstring(want)))
			Expect(run.Writes()).To(BeEmpty())
			Expect(reread(ctx, subject).Labels).To(BeEmpty())
		},
		Entry("a label outside the taxonomy", "todo_labels",
			func(*types.TODO) map[string]any { return map[string]any{"add": []any{"wontfix"}} }, "not in the workspace taxonomy"),
		Entry("removing a provenance label", "todo_labels",
			func(*types.TODO) map[string]any { return map[string]any{"remove": []any{"source:todo"}} }, "namespaced"),
		Entry("an edit to a closed TODO", "todo_update",
			func(other *types.TODO) map[string]any {
				return map[string]any{"ref": other.ShortID, "title": "Reopened"}
			}, "closed TODOs take no further changes"),
		Entry("an argument the tool does not declare", "todo_update",
			func(*types.TODO) map[string]any { return map[string]any{"titel": "Typo"} }, "invalid arguments"),
	)

	It("closes a duplicate into its survivor, linked, and previews the same close without writing", func(ctx SpecContext) {
		survivor := create(ctx, todos.CreateRequest{Title: "Fix the parser", Body: "The real one"})
		duplicate := create(ctx, todos.CreateRequest{Title: "Parser is broken", Body: "Same thing"})
		input := map[string]any{"reason": "Same crash", "duplicateOf": survivor.ShortID}

		previewDefs, previewRun := toolsFor(duplicate, true)
		result, err := call(ctx, previewDefs, "todo_close", input)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(ContainSubstring("nothing was written"))
		Expect(previewRun.Writes()).To(BeEmpty())
		Expect(reread(ctx, duplicate).Status).To(Equal(types.StatusPending))

		defs, run := toolsFor(duplicate, false)
		_, err = call(ctx, defs, "todo_close", input)
		Expect(err).NotTo(HaveOccurred())

		Expect(run.Writes()).To(Equal([]todos.ToolWrite{{Name: "todo_close", Ref: duplicate.ShortID}}))
		links, err := provider.Links(ctx, duplicate)
		Expect(err).NotTo(HaveOccurred())
		Expect(todos.HasLink(links, types.RelationDuplicates, survivor.ID)).To(BeTrue())
		Expect(reread(ctx, duplicate).Status).To(Equal(types.StatusCompleted), "a closed duplicate is soft-deleted, keeping its history")
		Expect(reread(ctx, survivor).Status).To(Equal(types.StatusPending), "the survivor is untouched")
	})

	It("merges a source into the subject, and refuses a merge across different PRs", func(ctx SpecContext) {
		subject := create(ctx, todos.CreateRequest{Title: "Widget parser", Body: "Parse widgets"})
		source := create(ctx, todos.CreateRequest{Title: "Widget lexer", Body: "Lex widgets"})
		defs, run := toolsFor(subject, false)

		_, err := call(ctx, defs, "todo_merge", map[string]any{
			"sources": []any{source.ShortID}, "body": "Parse and lex widgets", "rationale": "One pipeline",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(reread(ctx, subject).MarkdownBody).To(ContainSubstring("Parse and lex widgets"))
		Expect(reread(ctx, source).Status).To(Equal(types.StatusCompleted), "the source is closed into the subject")
		Expect(run.Writes()).To(Equal([]todos.ToolWrite{
			{Name: "todo_merge", Ref: subject.ShortID}, {Name: "todo_merge", Ref: source.ShortID},
		}))

		left := create(ctx, todos.CreateRequest{Title: "Left PR work", Body: "Left",
			PR: &types.PR{Number: 1, URL: "https://github.com/acme/todo-tools/pull/1"}})
		right := create(ctx, todos.CreateRequest{Title: "Right PR work", Body: "Right",
			PR: &types.PR{Number: 2, URL: "https://github.com/acme/todo-tools/pull/2"}})
		crossDefs, crossRun := toolsFor(left, false)
		_, err = call(ctx, crossDefs, "todo_merge", map[string]any{"sources": []any{right.ShortID}, "body": "Both"})
		Expect(err).To(HaveOccurred())
		Expect(crossRun.Writes()).To(BeEmpty())
		Expect(reread(ctx, right).Status).To(Equal(types.StatusPending))
	})

	It("asks through the run's broker and returns the answers, and fails without one", func(ctx SpecContext) {
		defs, _ := toolsFor(&types.TODO{ID: "ask-subject"}, false)
		input := map[string]any{"questions": []any{map[string]any{"text": "Which database?", "options": []any{"Postgres", "SQLite"}}}}

		_, err := call(ctx, defs, tools.AskToolName, input)
		Expect(err).To(MatchError(ContainSubstring("todo_ask requires an approval broker")))

		var asked api.ApprovalRequest
		brokered := api.WithApprovalFunc(ctx, func(_ context.Context, request api.ApprovalRequest) (api.ApprovalDecision, error) {
			asked = request
			return api.ApprovalDecision{Allow: true, UpdatedInput: map[string]any{"answers": map[string]any{"Which database?": "Postgres"}}}, nil
		})
		answers, err := call(brokered, defs, tools.AskToolName, input)

		Expect(err).NotTo(HaveOccurred())
		Expect([]any{asked.Tool, asked.Kind, asked.Questions}).To(Equal([]any{
			tools.AskToolName, api.ApprovalKindQuestion,
			[]api.TerminalQuestion{{Text: "Which database?", Options: []string{"Postgres", "SQLite"}}},
		}))
		Expect(answers).To(Equal([]map[string]any{{"question": "Which database?", "answer": []string{"Postgres"}}}))
	})
})
