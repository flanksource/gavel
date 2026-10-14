package model

type MergeInput struct {
	Method  string `json:"method"`
	Auto    bool   `json:"auto,omitempty"`
	HeadSHA string `json:"headSha,omitempty"`
}
