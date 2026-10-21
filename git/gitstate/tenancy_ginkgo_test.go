// Tests that git state is isolated per organization: another org's store neither
// sees nor changes a repository, worktree, branch or range it doesn't own.
package gitstate_test

import (
	"errors"

	gavelgit "github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/gitstate"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store organizations", func() {
	It("keeps one org's git state out of another org's reads and writes", func(ctx SpecContext) {
		r := newTrackedRepo(ctx, gitstate.Options{})
		state, err := r.tracker.State(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())
		id, err := r.tracker.Track(ctx, r.root)
		Expect(err).NotTo(HaveOccurred())
		key := gitstate.RangeKey{Base: r.base, Head: state.Branches[0].Head}
		local := r.tracker.Store()
		Expect(local.SaveRangeFiles(ctx, id, key, []gavelgit.CommitFile{{Path: "a.txt"}})).To(Succeed())
		generationOf := func() int64 {
			return r.count(`SELECT generation FROM git_repo_overview WHERE repo_id = ?`, id)
		}
		generation := generationOf()
		var refsFingerprint, statusFingerprint string
		Expect(r.db.Raw(`SELECT refs_fingerprint FROM git_repo_overview WHERE repo_id = ?`, id).Scan(&refsFingerprint).Error).To(Succeed())
		Expect(r.db.Raw(`SELECT status_fingerprint FROM git_worktree_overview WHERE repo_id = ? AND path = ?`, id, r.feature).Scan(&statusFingerprint).Error).To(Succeed())
		Expect(refsFingerprint).NotTo(BeEmpty())
		Expect(statusFingerprint).NotTo(BeEmpty())

		otherOrg := uuid.New()
		Expect(r.db.Exec(`INSERT INTO orgs (id, issuer, clerk_org_id, name) VALUES (?, 'https://clerk.example', ?, 'Other')`, otherOrg, otherOrg.String()).Error).To(Succeed())
		other := gitstate.NewStore(r.db).ForOrg(otherOrg)

		By("registering the same checkout as a separate repository")
		otherID, err := other.EnsureRepo(ctx, r.root, r.root+"/.git")
		Expect(err).NotTo(HaveOccurred())
		Expect(otherID).NotTo(Equal(id))
		againID, err := local.EnsureRepo(ctx, r.root, r.root+"/.git")
		Expect(err).NotTo(HaveOccurred())
		Expect(againID).To(Equal(id))

		By("reading nothing of the other org's repository")
		states, err := other.Load(ctx, []uuid.UUID{id})
		Expect(err).NotTo(HaveOccurred())
		Expect(states).To(BeEmpty())
		generations, err := other.Generations(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(generations).To(HaveKey(otherID))
		Expect(generations).NotTo(HaveKey(id))
		unchanged, err := other.MarkRefsUnchanged(ctx, id, refsFingerprint)
		Expect(err).NotTo(HaveOccurred())
		Expect(unchanged).To(BeFalse(), "another org's scan confirms nothing of this repository")
		unchanged, err = other.MarkStatusUnchanged(ctx, id, r.feature, statusFingerprint)
		Expect(err).NotTo(HaveOccurred())
		Expect(unchanged).To(BeFalse())
		unchanged, err = local.MarkRefsUnchanged(ctx, id, refsFingerprint)
		Expect(err).NotTo(HaveOccurred())
		Expect(unchanged).To(BeTrue(), "the owning org confirms the same fingerprint")
		_, found, err := other.WorktreeFiles(ctx, id, r.feature)
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeFalse())
		_, found, err = other.BranchHead(ctx, id, "feature")
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeFalse())
		missing, err := other.MissingRanges(ctx, id, []gitstate.RangeKey{key})
		Expect(err).NotTo(HaveOccurred())
		Expect(missing).To(Equal([]gitstate.RangeKey{key}))
		ranges, err := other.Ranges(ctx, id, []gitstate.RangeKey{key})
		Expect(err).NotTo(HaveOccurred())
		Expect(ranges).To(BeEmpty())
		single, err := other.Range(ctx, id, key)
		Expect(err).NotTo(HaveOccurred())
		Expect(single).To(BeNil())
		_, found, err = other.RangeFiles(ctx, id, key)
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeFalse())

		By("changing nothing of the other org's repository")
		for name, write := range map[string]func() error{
			"ApplyRefs": func() error { return other.ApplyRefs(ctx, id, gitstate.RefsUpdate{Fingerprint: "hijacked"}) },
			"SaveRanges": func() error {
				return other.SaveRanges(ctx, id, []gitstate.Range{{RangeKey: gitstate.RangeKey{Base: "b", Head: "h"}}})
			},
			"ApplyStatus": func() error {
				return other.ApplyStatus(ctx, id, r.feature, gitstate.StatusUpdate{Fingerprint: "hijacked"})
			},
			"RecordStatusError": func() error { return other.RecordStatusError(ctx, id, r.feature, errors.New("hijacked")) },
			"RecordError":       func() error { return other.RecordError(ctx, id, errors.New("hijacked")) },
		} {
			Expect(write()).To(MatchError(gitstate.ErrNotFound), name)
		}
		Expect(other.SaveRangeFiles(ctx, id, key, []gavelgit.CommitFile{{Path: "hijacked.txt"}})).To(Succeed())

		Expect(generationOf()).To(Equal(generation))
		Expect(r.count(`SELECT count(*) FROM git_repos WHERE id = ? AND (error IS NOT NULL OR refs_fingerprint = 'hijacked')`, id)).To(BeZero())
		Expect(r.count(`SELECT count(*) FROM git_worktrees WHERE repo_id = ? AND (status_error IS NOT NULL OR status_fingerprint = 'hijacked')`, id)).To(BeZero())
		Expect(r.count(`SELECT count(*) FROM git_range_stats WHERE base_sha = 'b'`)).To(BeZero())
		files, found, err := local.RangeFiles(ctx, id, key)
		Expect(err).NotTo(HaveOccurred())
		Expect(found).To(BeTrue())
		Expect(files).To(Equal([]gavelgit.CommitFile{{Path: "a.txt"}}))
	})
})
