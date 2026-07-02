package types

import "fmt"

// RunMode selects which built-in prompt a todo agent run executes: implement
// the todo (run), propose a plan for review (plan), or score committed work
// against acceptance criteria (verify).
type RunMode string

const (
	// ModeRun implements the todo.
	ModeRun RunMode = "run"
	// ModePlan investigates and proposes a plan; the agent runs read-only under
	// its native plan mode and the todo moves to review/ask/pending.
	ModePlan RunMode = "plan"
	// ModeVerify scores the committed work against the todo's acceptance
	// criteria via the verify engine (no agent executor is constructed).
	ModeVerify RunMode = "verify"
)

// ParseRunMode resolves a CLI/API value into a RunMode. Empty defaults to
// ModeRun; any other unknown value (including the legacy inline/cmux mechanism
// names) fails loud.
func ParseRunMode(s string) (RunMode, error) {
	switch RunMode(s) {
	case "":
		return ModeRun, nil
	case ModeRun, ModePlan, ModeVerify:
		return RunMode(s), nil
	default:
		return "", fmt.Errorf("invalid run mode %q (valid: run, plan, verify)", s)
	}
}

// Valid reports whether m is one of the supported run modes.
func (m RunMode) Valid() bool {
	switch m {
	case ModeRun, ModePlan, ModeVerify:
		return true
	}
	return false
}
