package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flanksource/gavel/comments"
	"github.com/flanksource/gavel/git"
	"github.com/flanksource/gavel/git/proposal"
)

// CommentProposals keeps a thread's proposed patches as commits under
// refs/gavel/comments/<thread>/<reply> in the target project's repository.
type CommentProposals struct{}

var _ comments.Proposals = CommentProposals{}

func proposalRefPrefix(c comments.Comment) string {
	return "refs/gavel/comments/" + c.ID + "/"
}

func (CommentProposals) Propose(ctx context.Context, c comments.Comment, replyID string, input comments.ProposeInput) (comments.Proposal, error) {
	if c.ID == "" || replyID == "" {
		return comments.Proposal{}, fmt.Errorf("%w: a proposal needs a thread and reply id", comments.ErrInvalid)
	}
	dir, err := CommentWorkdir(ctx, c.Target)
	if err != nil {
		return comments.Proposal{}, err
	}
	ref := proposalRefPrefix(c) + replyID
	edits := make([]proposal.Edit, 0, len(input.Edits))
	for _, e := range input.Edits {
		edits = append(edits, proposal.Edit(e))
	}
	result, err := proposal.Create(ctx, proposal.CreateOptions{Dir: dir, Patch: input.Patch, Edits: edits, Message: "gavel: proposal for comment " + c.ID, Ref: ref})
	if err != nil {
		return comments.Proposal{}, fmt.Errorf("%w: %w", comments.ErrInvalid, err)
	}
	files := make([]comments.ProposalFile, 0, len(result.Files))
	for _, f := range result.Files {
		files = append(files, comments.ProposalFile{Path: f.Path, Additions: f.Additions, Deletions: f.Deletions})
	}
	return comments.Proposal{Ref: ref, Base: result.Base, Commit: result.Commit, Files: files, State: comments.ProposalProposed}, nil
}

func (CommentProposals) Discard(ctx context.Context, c comments.Comment, p comments.Proposal) error {
	if !strings.HasPrefix(p.Ref, proposalRefPrefix(c)) {
		return fmt.Errorf("%w: ref %q is not a proposal of comment %s", comments.ErrInvalid, p.Ref, c.ID)
	}
	dir, err := CommentWorkdir(ctx, c.Target)
	if err != nil {
		return err
	}
	return proposal.Delete(ctx, dir, p.Ref)
}

func (CommentProposals) Diff(ctx context.Context, c comments.Comment, p comments.Proposal) (comments.ProposalDiff, error) {
	dir, err := CommentWorkdir(ctx, c.Target)
	if err != nil {
		return comments.ProposalDiff{}, err
	}
	diff, err := git.CommitDiff(dir, git.CommitDiffOptions{Base: p.Base, Head: p.Commit})
	if err != nil {
		return comments.ProposalDiff{}, err
	}
	return comments.ProposalDiff{Diff: diff.Diff, Truncated: diff.Truncated, Binary: diff.Binary, Commit: p.Commit}, nil
}

func (CommentProposals) Apply(ctx context.Context, c comments.Comment, p comments.Proposal, stage bool) error {
	dir, err := CommentWorkdir(ctx, c.Target)
	if err != nil {
		return err
	}
	err = proposal.Apply(ctx, dir, p.Base, p.Commit, stage)
	if errors.Is(err, proposal.ErrConflict) {
		return fmt.Errorf("%w: %w", comments.ErrConflict, err)
	}
	return err
}
