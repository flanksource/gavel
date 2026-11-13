package verify

import (
	"runtime"

	"github.com/flanksource/gavel/slots"
)

func (c GavelConfig) SlotLimits() (slots.Limits, error) {
	if c.Slots == (slots.Config{}) {
		return slots.Limits{}, nil
	}
	global, err := LoadGlobalGavelConfig()
	if err != nil {
		return slots.Limits{}, err
	}
	return slots.Resolve(global.Slots, c.Slots, runtime.NumCPU())
}

func slotsSchema() map[string]any {
	value := func(description string) map[string]any {
		return map[string]any{
			"description": description + " A positive integer or a percentage (1% through 100%) of CPU cores, rounded down with a minimum of 1. Omit to leave this pool unlimited.",
			"oneOf": []any{
				map[string]any{"type": "integer", "minimum": 1},
				map[string]any{"type": "string", "pattern": `^([1-9][0-9]*|([1-9]|[1-9][0-9]|100)%)$`},
			},
		}
	}
	return object("Machine-wide limits for concurrent test, lint and fixture command subprocesses. Worktrees share a project pool identified by their absolute git common directory. Waiting is logged with pool holders and excluded from subprocess timeouts. With no slots configured, no locks are taken.", map[string]any{
		"global":         value("Maximum across all projects; read only from ~/.gavel.yaml. Project configuration cannot override it."),
		"projectDefault": value("Default maximum per repository; read only from ~/.gavel.yaml."),
		"project":        value("Project's maximum, overriding the home projectDefault; it never bypasses the global pool."),
	})
}
