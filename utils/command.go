package utils

import (
	"fmt"
	"io"
	"strings"
)

// CommandRecorder is an output tee that also keeps each command run into it,
// so a viewer can show what ran apart from the output it wrote.
type CommandRecorder interface {
	RecordCommand(cwd string, argv []string)
}

// AnnounceCommand writes "▶ <command>" to tee ahead of the command's output,
// and records the command when tee is a CommandRecorder. A nil tee is a no-op.
func AnnounceCommand(tee io.Writer, cwd string, argv []string) {
	if tee == nil {
		return
	}
	if recorder, ok := tee.(CommandRecorder); ok {
		recorder.RecordCommand(cwd, argv)
	}
	fmt.Fprintf(tee, "▶ %s\n", ShellJoin(argv))
}

// ShellJoin joins argv into one line a POSIX shell parses back into argv.
func ShellJoin(argv []string) string {
	parts := make([]string, 0, len(argv))
	for _, arg := range argv {
		parts = append(parts, ShellQuote(arg))
	}
	return strings.Join(parts, " ")
}

// ShellQuote quotes s for a POSIX shell, leaving it bare when it needs none.
func ShellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\n\"'\\$`*?[]{}()&;|<>#~!") {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
