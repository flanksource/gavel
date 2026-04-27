package fixtures

import (
	"regexp"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

var (
	colorCodeRegex   = regexp.MustCompile(`\x1b\[(3[0-7]|38;[25];|4[0-7]|48;[25];)`)
	cursorUpdateCode = regexp.MustCompile(`\x1b\[([0-9]*[ABCDHJ]|[0-9]*K|2J|\?25[hl])`)
	anyANSIEscape    = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	sgrResetCode     = regexp.MustCompile(`\x1b\[0?m`)
	cursorHideCode   = "\x1b[?25l"
	cursorShowCode   = "\x1b[?25h"
)

// stripFixtureANSI removes every ANSI SGR escape sequence from s. Used
// when measuring fixture output lengths for the compact CI one-liner.
func stripFixtureANSI(s string) string {
	return anyANSIEscape.ReplaceAllString(s, "")
}

func hasAnyANSI(s string) bool {
	return strings.Contains(s, "\x1b[")
}

func hasColorCodes(s string) bool {
	return colorCodeRegex.MatchString(s)
}

func hasCursorUpdates(s string) bool {
	return cursorUpdateCode.MatchString(s)
}

func hasCursorHide(s string) bool { return strings.Contains(s, cursorHideCode) }
func hasCursorShow(s string) bool { return strings.Contains(s, cursorShowCode) }
func hasSGRReset(s string) bool   { return sgrResetCode.MatchString(s) }

// hasStrayControls reports whether s contains C0 control bytes that clicky's
// renderer never emits intentionally: BS (0x08), VT (0x0b), FF (0x0c). BEL
// (0x07) is allowed because it legitimately terminates OSC sequences.
func hasStrayControls(s string) bool {
	return strings.ContainsAny(s, "\x08\x0b\x0c")
}

func celStringToBool(fn func(string) bool) cel.OverloadOpt {
	return cel.FunctionBinding(func(args ...ref.Val) ref.Val {
		s, ok := args[0].Value().(string)
		if !ok {
			return types.NewErr("expected string argument")
		}
		return types.Bool(fn(s))
	})
}

func ANSICelFunctions() []cel.EnvOption {
	return []cel.EnvOption{
		cel.Function("has_color", cel.Overload("has_color_string", []*cel.Type{cel.StringType}, cel.BoolType, celStringToBool(hasColorCodes))),
		cel.Function("has_ansi", cel.Overload("has_ansi_string", []*cel.Type{cel.StringType}, cel.BoolType, celStringToBool(hasAnyANSI))),
		cel.Function("has_cursor_updates", cel.Overload("has_cursor_updates_string", []*cel.Type{cel.StringType}, cel.BoolType, celStringToBool(hasCursorUpdates))),
	}
}
