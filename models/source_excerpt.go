package models

import (
	"fmt"
	"strings"

	"github.com/flanksource/clicky/api"
)

// SourceExcerpt preserves the authored source and its original line numbers.
type SourceExcerpt struct {
	Content   string `json:"content"`
	StartLine int    `json:"start_line"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
}

func (s SourceExcerpt) Pretty() api.Text {
	out := api.Text{}
	width := len(fmt.Sprint(s.StartLine + strings.Count(s.Content, "\n")))
	for i, line := range strings.Split(s.Content, "\n") {
		if i > 0 {
			out = out.NewLine()
		}
		number, marker := s.StartLine+i, " "
		if number == s.Line {
			marker = ">"
		}
		out = out.Append(fmt.Sprintf("%s %*d | ", marker, width, number), "text-muted").Append(line)
		if number == s.Line && s.Column > 0 {
			out = out.NewLine().Append(strings.Repeat(" ", width+3)+"| ", "text-muted").Append(strings.Repeat(" ", s.Column-1)+"^", "text-red-500")
		}
	}
	return out
}

func NewSourceExcerpt(content []byte, line, column int) *SourceExcerpt {
	lines := strings.Split(strings.TrimSuffix(string(content), "\n"), "\n")
	line = min(max(1, line), len(lines))
	start, end := max(1, line-3), min(len(lines), line+3)
	return &SourceExcerpt{Content: strings.Join(lines[start-1:end], "\n"), StartLine: start, Line: line, Column: column}
}
