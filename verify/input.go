package verify

import (
	"fmt"
	"strings"
)

type ReviewScope struct {
	Type        string   // "diff", "range", "files"
	CommitRange string   // e.g. "main..HEAD"
	Files       []string // file paths
}

func (s ReviewScope) String() string {
	switch s.Type {
	case "range":
		return fmt.Sprintf("commits %s", s.CommitRange)
	case "files":
		return fmt.Sprintf("files [%s]", strings.Join(s.Files, ", "))
	default:
		return "uncommitted diff"
	}
}

func ResolveScope(args []string, commitRange string) ReviewScope {
	if commitRange != "" {
		return ReviewScope{Type: "range", CommitRange: commitRange}
	}
	if len(args) > 0 {
		return ReviewScope{Type: "files", Files: args}
	}
	return ReviewScope{Type: "diff"}
}
