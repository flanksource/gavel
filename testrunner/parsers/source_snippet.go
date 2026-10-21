package parsers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

// maxSourceLines caps the source a failure shows; a longer test is cut with a
// note of how much more there is.
const maxSourceLines = 40

// SourceRange is the lines a test was written as, such as a fixture's YAML
// step or markdown block. A relative File resolves against WorkDir.
type SourceRange struct {
	File    string
	WorkDir string
	Line    int
	EndLine int
}

// SourceSnippet is a SourceRange's text, or the reason it could not be read.
type SourceSnippet struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	EndLine  int    `json:"end_line"`
	Language string `json:"language,omitempty"`
	Text     string `json:"text,omitempty"`
	// MoreLines counts the lines past maxSourceLines that Text leaves out.
	MoreLines int    `json:"more_lines,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Valid reports whether the range names a file and a non-empty line span.
func (r SourceRange) Valid() bool {
	return r.File != "" && r.Line > 0 && r.EndLine >= r.Line
}

func (r SourceRange) String() string {
	return fmt.Sprintf("%s:%d-%d", r.File, r.Line, r.EndLine)
}

// Read loads the range's lines; an unreadable file or a range past its end is
// reported in the snippet's Error.
func (r SourceRange) Read() SourceSnippet {
	path := r.File
	if !filepath.IsAbs(path) && r.WorkDir != "" {
		path = filepath.Join(r.WorkDir, path)
	}
	snippet := SourceSnippet{File: r.File, Line: r.Line, EndLine: r.EndLine, Language: sourceLanguage(path)}
	data, err := os.ReadFile(path)
	if err != nil {
		snippet.Error = err.Error()
		return snippet
	}
	lines := strings.Split(string(data), "\n")
	if r.EndLine > len(lines) {
		snippet.Error = fmt.Sprintf("%s has %d lines, want lines %d-%d", r.File, len(lines), r.Line, r.EndLine)
		return snippet
	}
	selected := lines[r.Line-1 : r.EndLine]
	if more := len(selected) - maxSourceLines; more > 0 {
		selected, snippet.MoreLines = selected[:maxSourceLines], more
	}
	snippet.Text = strings.Join(selected, "\n")
	return snippet
}

// Pretty renders the snippet as a code block, or says why there is none.
func (s SourceSnippet) Pretty() api.Text {
	if s.Error != "" {
		return clicky.Text("source unavailable: "+s.Error, "text-red-500")
	}
	text := clicky.Text("").Add(api.Code{Content: s.Text, Language: s.Language})
	if s.MoreLines > 0 {
		text = text.NewLine().Append(fmt.Sprintf("… %d more lines", s.MoreLines), "text-muted")
	}
	return text
}

func sourceLanguage(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".md", ".mdx":
		return "markdown"
	}
	return ""
}
