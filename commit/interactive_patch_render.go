package commit

import (
	"context"
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/flanksource/gavel/status"
)

var (
	patchSyntaxTheme = styles.Get("monokai")
	patchContext     = lipgloss.NewStyle().
				Foreground(lipgloss.Color(patchSyntaxTheme.Get(chroma.Text).Colour.String())).
				Background(lipgloss.Color(patchSyntaxTheme.Get(chroma.Background).Background.String()))
	patchAdded      = patchContext.Background(lipgloss.Color("#173125"))
	patchRemoved    = patchContext.Background(lipgloss.Color("#3c2024"))
	patchMetadata   = patchContext.Foreground(lipgloss.Color("#c5c5b8"))
	patchHunkHeader = patchContext.Foreground(lipgloss.Color("#66d9ef"))
)

func renderTreePatch(ctx context.Context, file status.FileStatus, patch string) (string, error) {
	newLexer := lexers.Match(file.Path)
	oldLexer := newLexer
	if file.PreviousPath != "" {
		oldLexer = lexers.Match(file.PreviousPath)
	}
	lines := strings.Split(strings.TrimSuffix(patch, "\n"), "\n")
	for i := 0; i < len(lines); i++ {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if !strings.HasPrefix(lines[i], "@@ -") {
			lines[i] = patchMetadata.Render(expandPatchTabs(lines[i]))
			continue
		}
		lines[i] = patchHunkHeader.Render(expandPatchTabs(lines[i]))
		end := i + 1
		for end < len(lines) && !strings.HasPrefix(lines[end], "@@ -") && !strings.HasPrefix(lines[end], "diff --git ") {
			end++
		}
		if err := renderTreeHunk(lines[i+1:end], oldLexer, newLexer); err != nil {
			return "", fmt.Errorf("highlight %q: %w", file.Path, err)
		}
		i = end - 1
	}
	return strings.Join(lines, "\n"), nil
}

func renderTreeHunk(lines []string, oldLexer, newLexer chroma.Lexer) error {
	before, err := patchSideTokens(lines, oldLexer, '+')
	if err != nil {
		return err
	}
	after, err := patchSideTokens(lines, newLexer, '-')
	if err != nil {
		return err
	}
	for i, line := range lines {
		switch patchLineKind(line) {
		case '+':
			lines[i] = renderPatchLine(line, after[i], patchAdded)
		case '-':
			lines[i] = renderPatchLine(line, before[i], patchRemoved)
		case ' ':
			tokens := after[i]
			if tokens == nil {
				tokens = before[i]
			}
			lines[i] = renderPatchLine(line, tokens, patchContext)
		default:
			lines[i] = patchMetadata.Render(expandPatchTabs(line))
		}
	}
	return nil
}

func patchSideTokens(lines []string, lexer chroma.Lexer, omit byte) ([][]chroma.Token, error) {
	byRow := make([][]chroma.Token, len(lines))
	if lexer == nil {
		return byRow, nil
	}
	var source strings.Builder
	var rows []int
	for i, line := range lines {
		kind := patchLineKind(line)
		if kind == omit || !strings.ContainsRune(" +-", rune(kind)) {
			continue
		}
		if line != "" {
			source.WriteString(line[1:])
		}
		source.WriteByte('\n')
		rows = append(rows, i)
	}
	if len(rows) == 0 {
		return byRow, nil
	}
	tokens, err := chroma.Tokenise(chroma.Coalesce(lexer), nil, source.String())
	if err != nil {
		return nil, err
	}
	// Some lexers rewrite their input. Use diff-only styling when spans cannot map to the patch text.
	remaining := source.String()
	for _, token := range tokens {
		if !strings.HasPrefix(remaining, token.Value) {
			return byRow, nil
		}
		remaining = remaining[len(token.Value):]
	}
	if remaining == "" {
		for i, row := range chroma.SplitTokensIntoLines(tokens) {
			byRow[rows[i]] = row
		}
	}
	return byRow, nil
}

func patchLineKind(line string) byte {
	if line == "" {
		return ' ' // Git can omit the prefix on empty context lines.
	}
	return line[0]
}

func renderPatchLine(line string, tokens []chroma.Token, base lipgloss.Style) string {
	if line == "" || len(tokens) == 0 {
		return base.Render(expandPatchTabs(line))
	}
	var out strings.Builder
	out.WriteString(base.Render(line[:1]))
	for _, token := range tokens {
		text := strings.TrimSuffix(token.Value, "\n")
		if text == "" {
			continue
		}
		style := base
		if foreground := patchSyntaxTheme.Get(token.Type).Colour; foreground.IsSet() {
			style = style.Foreground(lipgloss.Color(foreground.String()))
		}
		out.WriteString(style.Render(expandPatchTabs(text)))
	}
	return out.String()
}

func expandPatchTabs(line string) string {
	return strings.ReplaceAll(line, "\t", "    ")
}
