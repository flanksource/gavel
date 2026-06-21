package aifix

import (
	"fmt"
	"io"
	"strings"

	captainai "github.com/flanksource/captain/pkg/ai"
	"github.com/flanksource/clicky"
)

// NewStderrRenderer returns an OnEvent callback that writes a one-line summary
// per Claude event to w. Each line is prefixed with `[<model> <pct>%]` where
// pct is the share of the model's context window consumed so far, taken from
// the token usage reported on each result event. When contextWindow is 0 (the
// model is unknown to captain's pricing registry) the percentage is omitted.
// Lines are styled with clicky tags so `gavel lint --ai-fix` produces the same
// coloured output a captain user would see when piping `claude -p ... | captain`.
func NewStderrRenderer(w io.Writer, model string, contextWindow int) func(iter int, ev captainai.Event) {
	usedTokens := 0
	return func(_ int, ev captainai.Event) {
		if ev.Kind == captainai.EventResult && ev.Usage != nil {
			usedTokens = ev.Usage.TotalTokens()
		}
		line := renderEvent(eventPrefix(modelLabel(model, ev), usedTokens, contextWindow), ev)
		if line == "" {
			return
		}
		fmt.Fprintln(w, line)
	}
}

// modelLabel prefers the model reported on the event, falling back to the
// configured model and finally a generic label so the prefix is never empty.
func modelLabel(fallback string, ev captainai.Event) string {
	if ev.Model != "" {
		return ev.Model
	}
	if fallback != "" {
		return fallback
	}
	return "ai-fix"
}

// eventPrefix renders the `[model x%] ` line prefix. The percentage is dropped
// when the context window is unknown (0) to avoid printing a misleading 0%.
func eventPrefix(model string, usedTokens, contextWindow int) string {
	label := model
	if contextWindow > 0 {
		pct := (usedTokens*100 + contextWindow/2) / contextWindow
		label = fmt.Sprintf("%s %d%%", model, pct)
	}
	return clicky.Text("").
		Append(fmt.Sprintf("[%s] ", label), "text-muted").ANSI()
}

func renderEvent(prefix string, ev captainai.Event) string {
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
