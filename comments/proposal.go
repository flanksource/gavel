package comments

import (
	"context"
	"time"
)

// A Proposal is a patch offered by a reply instead of edits to the working
// tree: a commit on Base, kept at Ref (refs/gavel/comments/<thread>/<reply>),
// stored in the reply's content under ProposalContentKey.
type Proposal struct {
	Ref       string         `json:"ref"`
	Base      string         `json:"base"`
	Commit    string         `json:"commit"`
	Files     []ProposalFile `json:"files"`
	State     ProposalState  `json:"state"`
	DecidedBy *Actor         `json:"decidedBy,omitempty"`
	DecidedAt *time.Time     `json:"decidedAt,omitempty"`
	// Feedback is the reason given when rejecting or asking for revisions.
	Feedback string `json:"feedback,omitempty"`
}

type ProposalFile struct {
	Path      string `json:"path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

type ProposalState string

const (
	ProposalProposed          ProposalState = "proposed"
	ProposalApplied           ProposalState = "applied"
	ProposalStaged            ProposalState = "staged"
	ProposalRejected          ProposalState = "rejected"
	ProposalRevisionRequested ProposalState = "revision_requested"
	ProposalSuperseded        ProposalState = "superseded"
)

// ProposalContentKey is the reply content key holding its Proposal.
const ProposalContentKey = "proposal"

// Proposal actions are service actions (reserved, never workflow actions).
const (
	ActionAcceptProposal  = "accept-proposal"
	ActionRejectProposal  = "reject-proposal"
	ActionRequestRevision = "request-revision"
)

// Accept modes for ActionAcceptProposal.
const (
	ProposalModeApply = "apply"
	ProposalModeStage = "stage"
)

// Edits policies a comment agent route or delivery sets for the agent.
const (
	EditsDirect  = "direct"
	EditsPropose = "propose"
)

// ProposalDiff is a proposal's patch against its base, shaped like the
// clicky-ui GitDiffPayload.
type ProposalDiff struct {
	Diff      string `json:"diff"`
	Truncated bool   `json:"truncated,omitempty"`
	Binary    bool   `json:"binary,omitempty"`
	Commit    string `json:"commit"`
}

// FileEdit is one Edit-tool-style change to a file a reply proposes: an
// exact OldString→NewString replacement (once, or every match with
// ReplaceAll), a whole-file Content write that also creates the file, or a
// Delete.
type FileEdit struct {
	Path       string  `json:"path"`
	OldString  string  `json:"oldString,omitempty"`
	NewString  string  `json:"newString,omitempty"`
	ReplaceAll bool    `json:"replaceAll,omitempty"`
	Content    *string `json:"content,omitempty"`
	Delete     bool    `json:"delete,omitempty"`
}

// FileEditsSchema is the JSON Schema of a []FileEdit, for agent tools and the
// HTTP API.
func FileEditsSchema() map[string]any {
	str := func(description string) map[string]any {
		return map[string]any{"type": "string", "description": description}
	}
	boolean := func(description string) map[string]any {
		return map[string]any{"type": "boolean", "description": description}
	}
	return map[string]any{
		"type":        "array",
		"minItems":    1,
		"description": "Edits applied in order to the repository's HEAD; each is a replacement {path, oldString, newString, replaceAll?}, a whole-file write {path, content}, or {path, delete: true}",
		"items": map[string]any{"type": "object", "required": []string{"path"}, "additionalProperties": false, "properties": map[string]any{
			"path":       str("File path relative to the repository root"),
			"oldString":  str("The exact text to replace, including whitespace and indentation; it must match exactly once unless replaceAll"),
			"newString":  str("The text replacing oldString; empty removes it"),
			"replaceAll": boolean("Replace every match of oldString"),
			"content":    str("The whole new file content; creates the file when it does not exist"),
			"delete":     boolean("Delete the file"),
		}},
	}
}

// ProposeInput is what a proposal reply carries: a unified diff Patch or
// Edits, never both.
type ProposeInput struct {
	Patch []byte
	Edits []FileEdit
}

// Proposals is the host's git store for proposed patches; nil on a host
// without a local repository, where proposing or accepting is ErrInvalid.
type Proposals interface {
	// Propose commits the input onto the target's repository HEAD and creates
	// the thread's ref for replyID; the returned Proposal is in ProposalProposed.
	Propose(ctx context.Context, c Comment, replyID string, input ProposeInput) (Proposal, error)
	// Discard deletes a ref Propose created whose reply was never stored.
	Discard(ctx context.Context, c Comment, p Proposal) error
	Diff(ctx context.Context, c Comment, p Proposal) (ProposalDiff, error)
	// Apply writes the patch into the working tree, and the index when stage;
	// a patch that no longer applies is ErrConflict.
	Apply(ctx context.Context, c Comment, p Proposal, stage bool) error
}
