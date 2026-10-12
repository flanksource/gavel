package types

import "fmt"

// VerifyTargetSource is where the commit a check verified came from.
type VerifyTargetSource string

const (
	// VerifyTargetRun is the head of the todo's newest run worktree, not yet landed.
	VerifyTargetRun VerifyTargetSource = "run"
	// VerifyTargetPR is the topic head of the pull request the run landed as.
	VerifyTargetPR VerifyTargetSource = "pr"
	// VerifyTargetCheckout is the main checkout: no run recorded a worktree, or
	// the run was merged into it.
	VerifyTargetCheckout VerifyTargetSource = "checkout"
)

// VerifyTarget is the commit a check verified.
type VerifyTarget struct {
	// Branch is the branch the verified commit is on: the run's worktree branch,
	// or the checkout's. Empty for a PR, whose topic branch the landing does not
	// record; PR names it instead.
	Branch string             `json:"branch,omitempty"`
	SHA    string             `json:"sha,omitempty"`
	Source VerifyTargetSource `json:"source"`
	// PR is the pull request URL of a PR target.
	PR string `json:"pr,omitempty"`
}

// Worktree reports whether the target is verified in a fresh worktree rather
// than the main checkout.
func (t VerifyTarget) Worktree() bool {
	return t.Source == VerifyTargetRun || t.Source == VerifyTargetPR
}

func (t VerifyTarget) String() string {
	sha := t.SHA
	if len(sha) > 7 {
		sha = sha[:7]
	}
	switch {
	case t.Source == VerifyTargetPR:
		return fmt.Sprintf("%s %s @ %s", t.Source, t.PR, sha)
	case t.Branch == "":
		return fmt.Sprintf("%s @ %s", t.Source, sha)
	}
	return fmt.Sprintf("%s %s @ %s", t.Source, t.Branch, sha)
}
