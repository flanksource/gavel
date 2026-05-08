package aifix

import (
	"fmt"
	"io"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/clicky"
)

// NewStderrRenderer returns an OnEvent callback that writes a one-line
// summary per Claude event to w. Lines are styled with clicky tags so
// `gavel lint --ai-fix` produces the same coloured output a captain user
// would see when piping `claude -p ... | captain`.
func NewStderrRenderer(w io.Writer) func(iter int, ev captainai.Event) {
	return func(iter int, ev captainai.Event) {
		line := renderEvent(iter, ev)
		if line == "" {
			return
		}
		fmt.Fprintln(w, line)
	}
}

func renderEvent(iter int, ev captainai.Event) string {
	prefix := clicky.Text("").
		Append(fmt.Sprintf("[ai-fix iter %d] ", iter), "text-muted").ANSI()

	switch ev.Kind {
	case captainai.EventSystem:
		if ev.SessionID == "" {
			return ""
		}
		return prefix + clicky.Text("session ").
			Append(ev.SessionID, "text-blue-500").ANSI()

	case captainai.EventThinking:
		return prefix + clicky.Text("thinking: ", "text-muted").
			Append(truncateOneLine(ev.Text, 200), "text-muted").ANSI()

	case captainai.EventText:
		return prefix + clicky.Text("").
			Append(truncateOneLine(ev.Text, 200), "text-default").ANSI()

	case captainai.EventToolUse:
		return prefix + clicky.Text("").
			Append(ev.Tool, "text-green-500").Space().
			Append(toolInputSummary(ev.Input), "text-muted").ANSI()

	case captainai.EventResult:
		status := "ok"
		style := "text-green-500"
		if !ev.Success {
			status = "fail"
			style = "text-red-500"
		}
		t := clicky.Text("").
			Append("result ", "text-muted").
			Append(status, style)
		if ev.CostUSD > 0 {
			t = t.Space().Append(fmt.Sprintf("$%.4f", ev.CostUSD), "text-muted")
		}
		return prefix + t.ANSI()

	case captainai.EventError:
		return prefix + clicky.Text("error: ", "text-red-500").
			Append(truncateOneLine(ev.Error, 200), "text-red-500").ANSI()
	}
	return ""
}

func toolInputSummary(in map[string]any) string {
	if len(in) == 0 {
		return ""
	}
	for _, k := range []string{"file_path", "command", "path", "url"} {
		if v, ok := in[k].(string); ok && v != "" {
			return truncateOneLine(v, 120)
		}
	}
	return ""
}

func truncateOneLine(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) > max {
		return s[:max-1] + "…"
	}
	return s
}
