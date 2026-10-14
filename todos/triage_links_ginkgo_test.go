package todos

import (
	"context"

	"github.com/flanksource/gavel/todos/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// LinkPR and its siblings make triageRecorder a PRLinkProvider. A PR link is
// recorded in links as "pr:<relation>:<source>:<url>" so a test reads every
// link ApplyTriage wrote, in order, from one list.
func (p *triageRecorder) LinkPR(_ context.Context, todo *types.TODO, request PRLinkRequest) (*PRLink, error) {
	p.links = append(p.links, "pr:"+string(request.Relation)+":"+string(request.Source)+":"+request.URL)
	p.record("prlink", todo)
	return &PRLink{PRRef: types.PRRef{URL: request.URL}, Relation: request.Relation, Source: request.Source}, nil
}

func (p *triageRecorder) UnlinkPR(context.Context, *types.TODO, string) error { return nil }

func (p *triageRecorder) PRLinks(context.Context, *types.TODO) ([]PRLink, error) { return nil, nil }

func (p *triageRecorder) TodosForPR(context.Context, string) ([]PRLinkedTODO, error) {
	return nil, nil
}

// linkedTriageRecorder is a recorder whose TODO already carries links, the
// state a second triage of the same TODO starts from.
type linkedTriageRecorder struct {
	*triageRecorder
	existing   []Link
	existingPR []PRLink
}

func (p *linkedTriageRecorder) Links(context.Context, *types.TODO) ([]Link, error) {
	return p.existing, nil
}

func (p *linkedTriageRecorder) PRLinks(context.Context, *types.TODO) ([]PRLink, error) {
	return p.existingPR, nil
}

var _ = Describe("triage duplicate and PR links", func() {
	const (
		prURL      = "https://github.com/acme/widgets/pull/12"
		otherPRURL = "https://github.com/acme/widgets/pull/30"
	)

	ready := func(mutate func(*types.TriageEnvelope)) *types.TriageEnvelope {
		return completedTriage(func(e *types.TriageEnvelope) {
			e.Verdict = types.VerdictReady
			e.Body = ""
			mutate(e)
		})
	}

	It("links duplicates and PRs without closing anything or touching the TODO's own PR", func() {
		provider := &triageRecorder{}
		todo := triageTODO()
		env := ready(func(e *types.TriageEnvelope) {
			e.Related = []string{"cc3344"}
			e.Duplicates = []string{"ff0011", "FF0011", "ab12cd"}
			e.PRs = []string{prURL, "https://github.com/ACME/widgets/pull/12", otherPRURL}
		})

		Expect(ApplyTriage(context.Background(), provider, todo, env, TriageOptions{})).To(Succeed())
		Expect(provider.links).To(Equal([]string{
			"related_to:cc3344",
			"duplicates:ff0011",
			"pr:related:triage:" + prURL,
			"pr:related:triage:" + otherPRURL,
		}))
		Expect(provider.deletes).To(BeEmpty())
		Expect(provider.edits).To(BeEmpty(), "linking a PR never rewrites todo.PR")
	})

	It("writes the links before the retirement", func() {
		survivor := backlogTODO("ff0011", "Fix the parser panic", nil)
		provider := &triageRecorder{known: map[string]*types.TODO{"ff0011": survivor}}
		env := completedTriage(func(e *types.TriageEnvelope) {
			e.Verdict = types.VerdictDuplicateOf
			e.Body = ""
			e.DuplicateOf = "ff0011"
			e.Comment = "ff0011 already covers it"
			e.Duplicates = []string{"cc3344"}
			e.PRs = []string{prURL}
		})

		Expect(ApplyTriage(context.Background(), provider, triageTODO(), env, TriageOptions{})).To(Succeed())
		Expect(provider.writes).To(Equal([]string{
			"link:ab12cd", "prlink:ab12cd", "comment:ab12cd", "comment:ab12cd", "link:ab12cd", "delete:ab12cd",
		}))
		Expect(provider.links).To(Equal([]string{"duplicates:cc3344", "pr:related:triage:" + prURL, "duplicates:ff0011"}))
	})

	It("skips links the TODO already carries, so a second triage does not fail", func() {
		provider := &linkedTriageRecorder{
			triageRecorder: &triageRecorder{},
			existing:       []Link{{Relation: types.RelationDuplicates, TargetID: "id-ff0011", TargetShortID: "ff0011"}},
			existingPR:     []PRLink{{PRRef: types.PRRef{Provider: types.PRProviderGitHub, Repo: "acme/widgets", Number: 12, URL: prURL}}},
		}
		env := ready(func(e *types.TriageEnvelope) {
			e.Duplicates = []string{"ff0011", "cc3344"}
			e.PRs = []string{"https://github.com/Acme/Widgets/pull/12", otherPRURL}
		})

		Expect(ApplyTriage(context.Background(), provider, triageTODO(), env, TriageOptions{})).To(Succeed())
		Expect(provider.links).To(Equal([]string{"duplicates:cc3344", "pr:related:triage:" + otherPRURL}))
	})

	It("closes a duplicate whose duplicates link an earlier triage already recorded", func() {
		survivor := backlogTODO("ff0011", "Fix the parser panic", nil)
		provider := &linkedTriageRecorder{
			triageRecorder: &triageRecorder{known: map[string]*types.TODO{"ff0011": survivor}},
			existing:       []Link{{Relation: types.RelationDuplicates, TargetID: survivor.ID, TargetShortID: "ff0011"}},
		}
		env := completedTriage(func(e *types.TriageEnvelope) {
			e.Verdict = types.VerdictDuplicateOf
			e.Body = ""
			e.DuplicateOf = "ff0011"
			e.Comment = "ff0011 already covers it"
		})

		Expect(ApplyTriage(context.Background(), provider, triageTODO(), env, TriageOptions{})).To(Succeed())
		Expect(provider.links).To(BeEmpty())
		Expect(provider.deletes).To(Equal([]string{"ab12cd"}))
	})

	It("fails loudly when the provider cannot link PRs", func() {
		env := ready(func(e *types.TriageEnvelope) { e.PRs = []string{prURL} })
		err := ApplyTriage(context.Background(), &noLinkProvider{}, triageTODO(), env, TriageOptions{})
		Expect(err).To(MatchError(ContainSubstring("does not support PR links")))
	})

	It("reports the new fields as applied", func() {
		env := ready(func(e *types.TriageEnvelope) {
			e.Duplicates = []string{"ff0011"}
			e.PRs = []string{prURL}
		})
		Expect(TriageEnvelopeFields(env)).To(Equal([]string{"duplicates", "prs"}))
		Expect(RenderTriageApplied(triageTODO(), env)).To(Equal(`triage verdict "ready" applied to ab12cd: duplicates, prs`))
	})
})
