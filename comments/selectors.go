package comments

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

type CodeSelector struct {
	Kind       string `json:"kind"`
	Version    int    `json:"version"`
	Path       string `json:"path"`
	Revision   string `json:"revision,omitempty"`
	BaseCommit string `json:"baseCommit,omitempty"`
	HeadCommit string `json:"headCommit,omitempty"`
	Side       string `json:"side,omitempty"`
	StartLine  int    `json:"startLine"`
	EndLine    int    `json:"endLine"`
	Quote      string `json:"quote,omitempty"`
}

func ParseCodeSelector(data json.RawMessage) (CodeSelector, error) {
	var s CodeSelector
	if err := json.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("%w: invalid code selector: %v", ErrInvalid, err)
	}
	if s.Version != 1 || !fs.ValidPath(s.Path) || strings.Contains(s.Path, "\\") || s.Path == "." || s.StartLine < 1 || s.EndLine < s.StartLine {
		return s, fmt.Errorf("%w: invalid code range", ErrInvalid)
	}
	switch s.Kind {
	case "code-range":
		if s.Revision == "" {
			return s, fmt.Errorf("%w: code revision is required", ErrInvalid)
		}
	case "diff-range":
		if s.BaseCommit == "" || s.HeadCommit == "" || (s.Side != "old" && s.Side != "new") {
			return s, fmt.Errorf("%w: diff commits and side are required", ErrInvalid)
		}
	default:
		return s, fmt.Errorf("%w: unknown code selector %s", ErrInvalid, s.Kind)
	}
	return s, nil
}
