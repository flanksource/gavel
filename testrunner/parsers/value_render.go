package parsers

import (
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// ValueRenderer formats a gomega value of a known Go type into a compact,
// human-readable form. Implementations should be lossy in favour of
// readability — the raw form is always reachable via Test.Message and the
// "Full message" collapsible. Body is the dedented value block with the
// outer "<type>: " marker already stripped.
type ValueRenderer interface {
	Render(body string) string
}

// ValueRendererFunc adapts a function to ValueRenderer.
type ValueRendererFunc func(body string) string

func (f ValueRendererFunc) Render(body string) string { return f(body) }

var (
	valueRendererMu sync.RWMutex
	valueRenderers  = map[string]ValueRenderer{}
)

// RegisterValueRenderer wires a renderer for a Go type, identified by the
// string gomega prints in its `<type | 0xADDR>:` marker. Type names are
// matched exactly (e.g. "*pgconn.PgError"); the optional " | 0x..." pointer
// suffix is stripped before lookup.
//
// Built-ins for common error types are registered in init(); test code can
// register custom renderers as needed.
func RegisterValueRenderer(typeName string, r ValueRenderer) {
	valueRendererMu.Lock()
	defer valueRendererMu.Unlock()
	valueRenderers[typeName] = r
}

// renderGomegaValue applies the registered renderer for typeName when one
// exists, otherwise returns prettyDefault(body) — generic cleanup that strips
// pointer noise and trailing commas without parsing the body.
func renderGomegaValue(typeName, body string) string {
	if typeName != "" {
		valueRendererMu.RLock()
		r, ok := valueRenderers[typeName]
		valueRendererMu.RUnlock()
		if ok {
			out := r.Render(body)
			if out != "" {
				return prettyDefault(out)
			}
		}
	}
	return prettyDefault(body)
}

var (
	pointerSuffixRe   = regexp.MustCompile(`<\s*([^>|]+?)\s*\|\s*0x[0-9a-fA-F]+\s*>`)
	pointerOnlyLineRe = regexp.MustCompile(`^\s*0x[0-9a-fA-F]+\s*,?\s*$`)
	trailingCommaRe   = regexp.MustCompile(`,(\s*\n\s*\})`)
)

// prettyDefault is the fallback applied to every gomega value, before or after
// a registered renderer runs. Strips inline pointer suffixes
// (<*pkg.T | 0x...> -> <*pkg.T>), removes lines that are only a hex pointer,
// and strips trailing commas before closing braces. Indentation is left
// intact (the caller has already dedented).
func prettyDefault(body string) string {
	if body == "" {
		return ""
	}
	body = pointerSuffixRe.ReplaceAllString(body, "<$1>")
	out := make([]string, 0, strings.Count(body, "\n")+1)
	prevBlank := false
	for _, line := range strings.Split(body, "\n") {
		if pointerOnlyLineRe.MatchString(line) {
			continue
		}
		blank := strings.TrimSpace(line) == ""
		if blank && prevBlank {
			continue
		}
		prevBlank = blank
		out = append(out, line)
	}
	joined := strings.Join(out, "\n")
	joined = trailingCommaRe.ReplaceAllString(joined, "$1")
	return strings.TrimRight(joined, "\n ")
}

// init registers built-in renderers for the error types we see most often in
// real failures.
func init() {
	RegisterValueRenderer("*pgconn.PgError", ValueRendererFunc(renderPgError))
	RegisterValueRenderer("*url.Error", ValueRendererFunc(renderURLError))
	RegisterValueRenderer("*fmt.wrapError", ValueRendererFunc(renderWrapError))
	RegisterValueRenderer("*errors.errorString", ValueRendererFunc(renderErrorString))
	RegisterValueRenderer("*errors.joinError", ValueRendererFunc(renderJoinError))
	RegisterValueRenderer("*net.OpError", ValueRendererFunc(renderLeadingMessage))
	RegisterValueRenderer("*os.SyscallError", ValueRendererFunc(renderLeadingMessage))
}

// structFields parses the brace-block at the top level of body and returns a
// map of FieldName -> raw value text. Nested braces are returned as-is. The
// returned map preserves insertion order via the keys slice.
//
// Gomega's struct dumps follow a strict shape:
//
//	Type {
//	    FieldA: value,
//	    FieldB: <*pkg.T | 0xADDR>{nested},
//	}
//
// We only need top-level fields, so a tiny brace-counting walker is enough.
func structFields(body string) (map[string]string, []string) {
	open := strings.Index(body, "{")
	close := strings.LastIndex(body, "}")
	if open < 0 || close <= open {
		return nil, nil
	}
	inner := body[open+1 : close]
	fields := map[string]string{}
	var keys []string
	depth := 0
	field := ""
	value := strings.Builder{}
	collectingValue := false
	flush := func() {
		f := strings.TrimSpace(field)
		v := strings.TrimRight(strings.TrimSpace(value.String()), ",")
		if f != "" {
			if _, exists := fields[f]; !exists {
				keys = append(keys, f)
			}
			fields[f] = v
		}
		field = ""
		value.Reset()
		collectingValue = false
	}
	for _, line := range strings.Split(inner, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if collectingValue {
				value.WriteString("\n")
			}
			continue
		}
		if depth == 0 && !collectingValue {
			if i := strings.Index(line, ":"); i > 0 {
				field = strings.TrimSpace(line[:i])
				rest := strings.TrimSpace(line[i+1:])
				value.WriteString(rest)
				depth += strings.Count(rest, "{") + strings.Count(rest, "[")
				depth -= strings.Count(rest, "}") + strings.Count(rest, "]")
				if depth <= 0 {
					depth = 0
					flush()
				} else {
					collectingValue = true
				}
				continue
			}
		}
		if collectingValue {
			value.WriteString("\n")
			value.WriteString(line)
			depth += strings.Count(trimmed, "{") + strings.Count(trimmed, "[")
			depth -= strings.Count(trimmed, "}") + strings.Count(trimmed, "]")
			if depth <= 0 {
				depth = 0
				flush()
			}
		}
	}
	if field != "" {
		flush()
	}
	return fields, keys
}

// unquote strips one pair of surrounding double quotes if present.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		if u, err := strconv.Unquote(s); err == nil {
			return u
		}
	}
	return s
}

// renderPgError produces a compact view of a *pgconn.PgError dump. Drops the
// noisy implementation fields (SeverityUnlocalized, Position, InternalPosition,
// SchemaName/TableName/etc., File/Line/Routine) — keeps the parts a developer
// reads to diagnose: Code, Severity, Message, Detail, Hint, Where.
func renderPgError(body string) string {
	fields, _ := structFields(body)
	if len(fields) == 0 {
		return ""
	}
	get := func(k string) string { return unquote(fields[k]) }
	severity := get("Severity")
	code := get("Code")
	msg := get("Message")
	if msg == "" {
		return ""
	}
	head := msg
	switch {
	case code != "" && severity != "":
		head = severity + " " + code + ": " + msg
	case code != "":
		head = code + ": " + msg
	case severity != "":
		head = severity + ": " + msg
	}
	out := []string{head}
	if v := get("Detail"); v != "" {
		out = append(out, "Detail: "+v)
	}
	if v := get("Hint"); v != "" {
		out = append(out, "Hint: "+v)
	}
	if v := get("Where"); v != "" {
		out = append(out, "Where: "+v)
	}
	return strings.Join(out, "\n")
}

// renderURLError collapses a *url.Error dump to "<Op> <URL>: <Err>", which is
// what the error's Error() method returns anyway. When Err is itself a struct
// dump (gomega's typical case), recover the textual error from the leading
// line of body (which gomega always emits as the error's String() output).
func renderURLError(body string) string {
	fields, _ := structFields(body)
	if len(fields) == 0 {
		return firstLineOf(body)
	}
	op := unquote(fields["Op"])
	url := unquote(fields["URL"])
	errVal := strings.TrimSpace(fields["Err"])
	// If Err is a marker-prefixed struct, the readable error is the suffix
	// of the leading body line (gomega prints `Get "URL": <err message>`
	// before the struct). Fall through to the leading line below.
	if strings.HasPrefix(errVal, "<") || strings.Contains(errVal, "{") {
		errVal = ""
	} else {
		errVal = unquote(errVal)
	}
	if errVal == "" {
		// The leading line is exactly url.Error.Error(), shaped as
		// `Op "URL": tail`. Strip the `Op "URL": ` prefix to recover tail.
		lead := firstLineOf(body)
		prefix := op + " \"" + url + "\": "
		if op != "" && url != "" && strings.HasPrefix(lead, prefix) {
			return op + " " + url + ": " + lead[len(prefix):]
		}
		return lead
	}
	if op != "" && url != "" {
		return op + " " + url + ": " + condense(errVal)
	}
	return firstLineOf(body)
}

// renderWrapError pulls the readable message line from a *fmt.wrapError dump.
// The struct body just repeats the message + a nested err, which is noise.
func renderWrapError(body string) string {
	if msg := messageField(body); msg != "" {
		return msg
	}
	return firstLineOf(body)
}

// renderErrorString unwraps a *errors.errorString dump to its `s:` field.
func renderErrorString(body string) string {
	fields, _ := structFields(body)
	if s := unquote(fields["s"]); s != "" {
		return s
	}
	return firstLineOf(body)
}

// renderJoinError emits a bullet list of the wrapped errors. The dump shape
// is `{ errs: [<err1>, <err2>, ...] }` where each err is itself a struct.
func renderJoinError(body string) string {
	fields, _ := structFields(body)
	errs, ok := fields["errs"]
	if !ok || errs == "" {
		return firstLineOf(body)
	}
	// errs is "[<*errors.errorString | 0x...>{s: "..."}, ...]". Pull each
	// nested struct's leading message via the same prettyDefault path.
	parts := splitTopLevel(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(errs), "["), "]"))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		typeName := extractGomegaType(p)
		body := stripLeadingMarker(p)
		rendered := renderGomegaValue(typeName, body)
		out = append(out, "• "+condense(rendered))
	}
	if len(out) == 0 {
		return firstLineOf(body)
	}
	return strings.Join(out, "\n")
}

// renderLeadingMessage is the simple fallback used by *net.OpError /
// *os.SyscallError where the leading non-brace line already contains the
// useful diagnostic ("dial tcp ...: connect: connection refused").
func renderLeadingMessage(body string) string {
	return firstLineOf(body)
}

// firstLineOf returns the first line of body that is not blank and not the
// opening of a struct ("{").
func firstLineOf(body string) string {
	for _, line := range strings.Split(body, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || t == "{" {
			continue
		}
		return t
	}
	return strings.TrimSpace(body)
}

// messageField pulls the "msg:" struct field, often used by fmt.wrapError.
func messageField(body string) string {
	fields, _ := structFields(body)
	if m := unquote(fields["msg"]); m != "" {
		return m
	}
	return ""
}

// stripLeadingMarker drops a leading "<type | 0x...>" marker from s, in
// either the "<type>:" header form or the "<type>{...}" inline form. Returns
// the rest with surrounding whitespace trimmed. Used when descending into
// nested error values.
func stripLeadingMarker(s string) string {
	s = strings.TrimSpace(s)
	if loc := gomegaAnyMarkerRe.FindStringIndex(s); loc != nil && loc[0] == 0 {
		s = s[loc[1]:]
		s = strings.TrimPrefix(strings.TrimSpace(s), ":")
	}
	return strings.TrimSpace(s)
}

// condense collapses a multi-line value to a single line for inline use.
func condense(s string) string {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "\n") {
		return s
	}
	return strings.Join(strings.Fields(strings.ReplaceAll(s, "\n", " ")), " ")
}

// splitTopLevel splits a comma-separated list at depth 0 (commas inside
// {} or [] are ignored). Used by renderJoinError to split the errs slice.
func splitTopLevel(s string) []string {
	var parts []string
	depth := 0
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{', '[':
			depth++
		case '}', ']':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	return parts
}
