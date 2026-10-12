package todos

import (
	"context"
	"errors"
	"fmt"

	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var errFamilyOpenChildren = errors.New("open children need a decision")

// familyProvider is storage with the single-level hierarchy and its rules: a
// TODO with open children cannot be deleted, and only an open top-level TODO
// takes children. It records every write that landed, in order.
type familyProvider struct {
	Provider
	items  types.TODOS
	writes []string
	// failOn is the one write storage refuses, e.g. "delete:ccc222".
	failOn string
}

func (p *familyProvider) write(op string) error {
	if op == p.failOn {
		return fmt.Errorf("%s: storage unavailable", op)
	}
	p.writes = append(p.writes, op)
	return nil
}

func (p *familyProvider) List(_ context.Context, filters DiscoveryFilters) (types.TODOS, error) {
	var listed types.TODOS
	for _, todo := range p.items {
		if filters.Matches(todo) {
			listed = append(listed, todo)
		}
	}
	return listed, nil
}

func (p *familyProvider) Get(_ context.Context, ref string) (*types.TODO, error) {
	for _, todo := range p.items {
		if todo.ID == ref || todo.ShortID == ref {
			return todo, nil
		}
	}
	return nil, fmt.Errorf("no TODO matched %q", ref)
}

func (p *familyProvider) Delete(ctx context.Context, todo *types.TODO) error {
	open, err := p.List(ctx, DiscoveryFilters{ParentID: todo.ID, ExcludeStatuses: []types.Status{types.StatusCompleted}})
	if err != nil {
		return err
	}
	if len(open) > 0 {
		return fmt.Errorf("%w: %d under %s", errFamilyOpenChildren, len(open), todo.ShortID)
	}
	if err := p.write("delete:" + todo.ShortID); err != nil {
		return err
	}
	todo.Status = types.StatusCompleted
	return nil
}

func (p *familyProvider) SetParent(ctx context.Context, todo *types.TODO, parentRef string) error {
	parent := &types.TODO{}
	if parentRef != "" {
		var err error
		if parent, err = p.Get(ctx, parentRef); err != nil {
			return err
		}
		if parent.ParentID != "" || parent.Status == types.StatusCompleted {
			return fmt.Errorf("%s cannot take children", parent.ShortID)
		}
	}
	if err := p.write("parent:" + todo.ShortID + "->" + parent.ShortID); err != nil {
		return err
	}
	todo.ParentID = parent.ID
	return nil
}

var _ = Describe("closing a TODO that has children", func() {
	var (
		provider                                  *familyProvider
		parent, first, second, finished, survivor *types.TODO
	)

	member := func(short string, status types.Status, under *types.TODO) *types.TODO {
		todo := &types.TODO{
			ID: "id-" + short, ShortID: short,
			TODOFrontmatter: types.TODOFrontmatter{Title: "TODO " + short, Status: status},
		}
		if under != nil {
			todo.ParentID = under.ID
		}
		provider.items = append(provider.items, todo)
		return todo
	}
	parents := func() map[string]string {
		stored := map[string]string{}
		for _, todo := range provider.items {
			stored[todo.ShortID] = todo.ParentID
		}
		return stored
	}

	BeforeEach(func() {
		provider = &familyProvider{}
		parent = member("aaa111", types.StatusPending, nil)
		first = member("ccc111", types.StatusPending, parent)
		second = member("ccc222", types.StatusInProgress, parent)
		finished = member("ddd111", types.StatusCompleted, parent)
		survivor = member("sss111", types.StatusPending, nil)
	})

	Describe("Archive", func() {
		It("passes on storage's refusal and writes nothing when nobody chose what happens to the open children", func(ctx SpecContext) {
			_, err := Archive(ctx, provider, parent, ArchiveOptions{})

			Expect(err).To(MatchError(errFamilyOpenChildren))
			Expect(provider.writes).To(BeEmpty())
			Expect(parent.Status).To(Equal(types.StatusPending))
		})

		It("archives the open children before the parent and leaves a closed child alone", func(ctx SpecContext) {
			outcome, err := Archive(ctx, provider, parent, ArchiveOptions{Children: ChildrenArchive})

			Expect(err).NotTo(HaveOccurred())
			Expect(provider.writes).To(Equal([]string{"delete:ccc111", "delete:ccc222", "delete:aaa111"}))
			Expect([]types.Status{first.Status, second.Status, parent.Status}).To(HaveEach(types.StatusCompleted))
			Expect(finished.ParentID).To(Equal(parent.ID))
			Expect(outcome.String()).To(Equal("open children archived too: ccc111, ccc222"))
		})

		It("makes the open children top-level before closing the parent and leaves a closed child attached", func(ctx SpecContext) {
			outcome, err := Archive(ctx, provider, parent, ArchiveOptions{Children: ChildrenDetach})

			Expect(err).NotTo(HaveOccurred())
			Expect(provider.writes).To(Equal([]string{"parent:ccc111->", "parent:ccc222->", "delete:aaa111"}))
			Expect(parents()).To(Equal(map[string]string{
				"aaa111": "", "ccc111": "", "ccc222": "", "ddd111": parent.ID, "sss111": "",
			}))
			Expect(outcome.String()).To(Equal("open children made top-level TODOs: ccc111, ccc222"))
		})

		DescribeTable("leaves the parent open when a child cannot be settled",
			func(ctx SpecContext, children ChildrenDisposition, failOn string, landed []string) {
				provider.failOn = failOn

				_, err := Archive(ctx, provider, parent, ArchiveOptions{Children: children})

				Expect(err).To(MatchError(ContainSubstring(failOn + ": storage unavailable")))
				Expect(err).To(MatchError(ContainSubstring("ccc222")), "the error names the child that was not settled")
				Expect(provider.writes).To(Equal(landed))
				Expect(parent.Status).To(Equal(types.StatusPending))
			},
			Entry("archiving", ChildrenArchive, "delete:ccc222", []string{"delete:ccc111"}),
			Entry("detaching", ChildrenDetach, "parent:ccc222->", []string{"parent:ccc111->"}),
		)

		DescribeTable("ignores the choice for a TODO with no open children",
			func(ctx SpecContext, children ChildrenDisposition) {
				outcome, err := Archive(ctx, provider, survivor, ArchiveOptions{Children: children})

				Expect(err).NotTo(HaveOccurred())
				Expect(provider.writes).To(Equal([]string{"delete:sss111"}))
				Expect(outcome.String()).To(BeEmpty())
			},
			Entry("archive", ChildrenArchive),
			Entry("detach", ChildrenDetach),
			Entry("none", ChildrenDisposition("")),
		)

		It("refuses a choice it does not know before settling any child", func(ctx SpecContext) {
			_, err := Archive(ctx, provider, parent, ArchiveOptions{Children: "purge"})

			Expect(err).To(MatchError(ContainSubstring(`unknown children value "purge"`)))
			Expect(provider.writes).To(BeEmpty())
		})

		It("refuses to look for the children of a TODO that has no id", func(ctx SpecContext) {
			_, err := Archive(ctx, provider, &types.TODO{ShortID: "zzz999"}, ArchiveOptions{Children: ChildrenArchive})

			Expect(err).To(MatchError(ContainSubstring("has no id")))
			Expect(provider.writes).To(BeEmpty())
		})
	})

	Describe("SettleChildren for a TODO retired in favour of a survivor", func() {
		retire := ArchiveOptions{Children: ChildrenDetach}

		It("moves the open children under a top-level survivor and closes nothing", func(ctx SpecContext) {
			retire.Survivor = survivor

			outcome, err := SettleChildren(ctx, provider, parent, retire)

			Expect(err).NotTo(HaveOccurred())
			Expect(provider.writes).To(Equal([]string{"parent:ccc111->sss111", "parent:ccc222->sss111"}))
			Expect(finished.ParentID).To(Equal(parent.ID))
			Expect(outcome.String()).To(Equal("open children moved to sss111: ccc111, ccc222"))
		})

		It("moves them under the survivor's parent when the survivor is a child", func(ctx SpecContext) {
			retire.Survivor = member("sss222", types.StatusPending, survivor)

			outcome, err := SettleChildren(ctx, provider, parent, retire)

			Expect(err).NotTo(HaveOccurred())
			Expect(provider.writes).To(Equal([]string{"parent:ccc111->sss111", "parent:ccc222->sss111"}))
			Expect(outcome.String()).To(Equal("open children moved to sss111: ccc111, ccc222"))
		})

		DescribeTable("makes them top-level when nothing open can take them",
			func(ctx SpecContext, pick func() *types.TODO) {
				retire.Survivor = pick()

				outcome, err := SettleChildren(ctx, provider, parent, retire)

				Expect(err).NotTo(HaveOccurred())
				Expect(provider.writes).To(Equal([]string{"parent:ccc111->", "parent:ccc222->"}))
				Expect(outcome.String()).To(Equal("open children made top-level TODOs: ccc111, ccc222"))
			},
			Entry("no survivor", func() *types.TODO { return nil }),
			Entry("a closed survivor", func() *types.TODO { return member("sss333", types.StatusCompleted, nil) }),
			Entry("a survivor under a closed parent", func() *types.TODO {
				return member("sss555", types.StatusPending, member("sss444", types.StatusCompleted, nil))
			}),
		)

		It("promotes a survivor that is a child of the retired TODO and moves its siblings under it", func(ctx SpecContext) {
			retire.Survivor = first

			outcome, err := SettleChildren(ctx, provider, parent, retire)

			Expect(err).NotTo(HaveOccurred())
			Expect(provider.writes).To(Equal([]string{"parent:ccc111->", "parent:ccc222->ccc111"}))
			Expect(outcome.String()).To(Equal("open children moved to ccc111: ccc222"))
		})
	})

	DescribeTable("ParseChildrenDisposition",
		func(raw string, want ChildrenDisposition, wantErr string) {
			got, err := ParseChildrenDisposition(raw)

			if wantErr != "" {
				Expect(err).To(MatchError(wantErr))
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(want))
		},
		Entry("nothing chosen", "", ChildrenDisposition(""), ""),
		Entry("archive", "archive", ChildrenArchive, ""),
		Entry("detach, whatever its case and padding", " Detach ", ChildrenDetach, ""),
		Entry("anything else", "purge", ChildrenDisposition(""), `unknown children value "purge": use archive or detach`),
	)
})
