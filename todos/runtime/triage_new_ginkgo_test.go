package runtime

import (
	"context"

	"github.com/flanksource/commons-db/dbtest"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/triagenew"
	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Triage new native writes", Ordered, func() {
	var provider *Provider
	BeforeAll(func(ctx SpecContext) {
		dsn := dbtest.ForGinkgo(dbtest.Options{Name: "gavel_triage_new"}).DSN()
		GinkgoT().Setenv(database.EnvDSN, dsn)
		GinkgoT().Setenv(database.EnvDisable, "")
		GinkgoT().Setenv(database.LegacyEnvDSN, "")
		GinkgoT().Setenv(database.LegacyEnvDisable, "")
		GinkgoT().Setenv("HOME", GinkgoT().TempDir())
		opened, err := database.Open(ctx, database.WithMigrations())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { Expect(opened.Close()).To(Succeed()) })
		provider, err = New(ctx, opened.Gorm(), WorkspaceOptions{Name: "triage-new", RootPath: GinkgoT().TempDir(), Repositories: []string{"acme/triage-new"}})
		Expect(err).NotTo(HaveOccurred())
	})
	create := func(ctx context.Context, title, body string) *types.TODO {
		GinkgoHelper()
		todo, err := provider.Create(ctx, todos.CreateRequest{Title: title, Body: body, Labels: []string{"pr"}})
		Expect(err).NotTo(HaveOccurred())
		return todo
	}
	result := func(action string, target *types.TODO) *types.TriageNewEnvelope {
		env := &types.TriageNewEnvelope{ResultEnvelope: types.ResultEnvelope{Summary: "Triage complete", EndStatus: types.EndCompleted}, Title: "Repair parser boundaries", Labels: []string{"bug"}, Action: action, Rationale: "Same parser scope"}
		if target != nil {
			env.Target = target.ID
		}
		return env
	}
	It("renames separate work and preserves provenance labels", func(ctx SpecContext) {
		source := create(ctx, "Raw report", "Parser stops at a boundary")
		Expect(provider.ApplyTriageNew(ctx, source, result("keep", nil), todos.TriageNewApplyOptions{})).To(Succeed())
		fresh, err := provider.Get(ctx, source.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(fresh.Title).To(Equal("Repair parser boundaries"))
		Expect(fresh.Labels).To(ConsistOf("pr", "bug"))
	})
	It("rejects a target edited since approval without changing the source", func(ctx SpecContext) {
		source := create(ctx, "Stale source", "Parser failure")
		target := create(ctx, "Stale target", "Parser work")
		version := target.Version
		changed := "Changed target scope"
		Expect(provider.Edit(ctx, target, todos.EditRequest{Title: &changed})).To(Succeed())
		Expect(provider.ApplyTriageNew(ctx, source, result("child-of", target), todos.TriageNewApplyOptions{TargetVersion: version})).To(HaveOccurred())
		fresh, err := provider.Get(ctx, source.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(fresh.Title).To(Equal("Stale source"))
		Expect(fresh.ParentID).To(BeEmpty())
	})
	It("sets a reviewed parent and refuses deeper nesting", func(ctx SpecContext) {
		parent := create(ctx, "Top parser work", "Parser work")
		source := create(ctx, "Child parser work", "Parser detail")
		Expect(provider.ApplyTriageNew(ctx, source, result("child-of", parent), todos.TriageNewApplyOptions{TargetVersion: parent.Version})).To(Succeed())
		fresh, err := provider.Get(ctx, source.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(fresh.ParentID).To(Equal(parent.ID))
		nested := create(ctx, "Too deep", "Nested detail")
		Expect(provider.ValidateTriageNew(ctx, nested, result("child-of", fresh))).To(MatchError(ContainSubstring("nested")))
	})

	It("closes a duplicate while keeping the existing target open", func(ctx SpecContext) {
		source := create(ctx, "Duplicate parser report", "Same parser work")
		target := create(ctx, "Original parser repair", "Same parser work")
		Expect(provider.ApplyTriageNew(ctx, source, result("duplicate-of", target), todos.TriageNewApplyOptions{TargetVersion: target.Version})).To(Succeed())
		fresh, err := provider.Get(ctx, source.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(fresh.Status).To(Equal(types.StatusCompleted))
		original, err := provider.Get(ctx, target.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(original.Title).To(Equal("Original parser repair"))
		Expect(original.Status).NotTo(Equal(types.StatusCompleted))
	})
	It("rejects merge content that drops an acceptance criterion or attachment", func(ctx SpecContext) {
		source := create(ctx, "Report with evidence", "## Acceptance Criteria\n\n- [ ] Handle terminal separator\n\n[evidence](/api/todos/attachments/trace.txt)")
		target := create(ctx, "Existing parser repair", "Existing parser work")
		env := result("merge-into", target)
		env.Body = "Combined parser work"
		Expect(provider.ValidateTriageNew(ctx, source, env)).To(HaveOccurred())
	})
	It("merges into the existing survivor with criteria, evidence, fixture, plan and labels", func(ctx SpecContext) {
		const body = "## Acceptance Criteria\n\n- [ ] Handle terminal separator\n\n[evidence](/api/todos/attachments/trace.txt)"
		const fixture = "| Name | Command | CEL Validation |\n|---|---|---|\n| Parser | true | exitCode == 0 |"
		const plan = "# Parser repair\n\nRetain separator boundaries."
		source := create(ctx, "Merge parser report", body)
		target := create(ctx, "Merge parser target", "Existing parser work")
		Expect(provider.Edit(ctx, source, todos.EditRequest{Verification: stringPointer(fixture)})).To(Succeed())
		updated, err := provider.SavePlanRevision(ctx, source, plan, "triage-test")
		Expect(err).NotTo(HaveOccurred())
		source = updated
		env := result("merge-into", target)
		env.Body, env.Verification, env.Plan = "Existing parser work\n\n"+body, fixture, plan
		Expect(provider.ApplyTriageNew(ctx, source, env, todos.TriageNewApplyOptions{TargetVersion: target.Version})).To(Succeed())
		fresh, err := provider.Get(ctx, target.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(fresh.MarkdownBody).To(ContainSubstring("Handle terminal separator"))
		Expect(fresh.MarkdownBody).To(ContainSubstring("/api/todos/attachments/trace.txt"))
		Expect(fresh.VerificationMarkdown).To(ContainSubstring("Parser | true"))
		Expect(fresh.Labels).To(ConsistOf("pr", "bug"))
		savedPlan, err := provider.PlanMarkdown(ctx, fresh, types.ModePlan)
		Expect(err).NotTo(HaveOccurred())
		Expect(savedPlan).To(Equal(plan))
		retired, err := provider.Get(ctx, source.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(retired.Status).To(Equal(types.StatusCompleted))
	})
	It("rejects unknown taxonomy labels without writing", func(ctx SpecContext) {
		source := create(ctx, "Unknown label source", "Parser failure")
		env := result("keep", nil)
		env.Labels = []string{"unknown-triage-label"}
		Expect(provider.ApplyTriageNew(ctx, source, env, todos.TriageNewApplyOptions{})).To(MatchError(ContainSubstring("taxonomy")))
		fresh, err := provider.Get(ctx, source.ID)
		Expect(err).NotTo(HaveOccurred())
		Expect(fresh.Title).To(Equal("Unknown label source"))
	})

	It("keeps search and reads bound to the creation workspace", func(ctx SpecContext) {
		source := create(ctx, "Source scope sentinel", "Parser scope")
		target := create(ctx, "Candidate scope sentinel", "Parser scope")
		other, err := New(ctx, provider.db, WorkspaceOptions{Name: "other-triage", RootPath: GinkgoT().TempDir(), Repositories: []string{"acme/other-triage"}})
		Expect(err).NotTo(HaveOccurred())
		foreign, err := other.Create(ctx, todos.CreateRequest{Title: "Foreign scope sentinel"})
		Expect(err).NotTo(HaveOccurred())
		tools := (&triagenew.Review{Provider: provider, SourceID: source.ID}).Tools()
		found, err := tools[0].Handler(ctx, map[string]any{"query": "scope sentinel"})
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(Equal([]map[string]any{{"id": target.ID, "title": target.Title, "status": target.Status, "labels": target.Labels, "parentId": target.ParentID}}))
		_, err = tools[1].Handler(ctx, map[string]any{"ref": foreign.ID})
		Expect(err).To(HaveOccurred())
	})
})

func stringPointer(value string) *string { return &value }
