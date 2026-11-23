package model

type MergeInput struct {
	Method  string `json:"method"`
	Auto    bool   `json:"auto,omitempty"`
	HeadSHA string `json:"headSha,omitempty"`
	// CommitTitle and CommitBody set the squash or merge commit's message;
	// empty keeps the host's default. A rebase ignores them.
	CommitTitle string `json:"commitTitle,omitempty"`
	CommitBody  string `json:"commitBody,omitempty"`
}
