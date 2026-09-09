package types

import "fmt"

// RunMode identifies a TODO lifecycle operation. Run and plan are public agent
// modes; verify is retained internally for fixture-only Captain prompt runs and
// historical records.
type RunMode string

const (
	// ModeRun implements the todo.
	ModeRun RunMode = "run"
	// ModePlan investigates and proposes a plan; the agent runs read-only under
	// its native plan mode and the todo moves to review/ask/pending.
	ModePlan RunMode = "plan"
	// ModeVerify executes the issue's persisted verification fixture without a
	// generation prompt. It is internal and is not accepted by ParseRunMode.
	ModeVerify RunMode = "verify"
)

// ParseRunMode resolves a CLI/API value into a RunMode. Empty defaults to
// ModeRun; any other unknown value (including the legacy inline/cmux mechanism
// names) fails loud.
func ParseRunMode(s string) (RunMode, error) {
	switch RunMode(s) {
	case "":
		return ModeRun, nil
	case ModeRun, ModePlan:
		return RunMode(s), nil
	default:
		return "", fmt.Errorf("invalid run mode %q (valid: run, plan)", s)
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
