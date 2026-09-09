package commit

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/verify"
)

const (
	CheckModePrompt = "prompt"
	CheckModeFail   = "fail"
	CheckModeSkip   = "skip"
	CheckModeFalse  = "false"
)

func normalizeCheckMode(raw, name string) (string, error) {
	mode := strings.ToLower(strings.TrimSpace(raw))
	if mode == "" {
		return CheckModePrompt, nil
	}

	switch mode {
	case CheckModePrompt, CheckModeFail, CheckModeSkip:
		return mode, nil
	case CheckModeFalse:
		return CheckModeSkip, nil
	default:
		return "", fmt.Errorf("unknown %s mode: %q", name, raw)
	}
}

// resolvePrecommitMode resolves the pre-commit gate mode (gitignore, file-size,
// and linked-dependency checks) from the --precommit flag, else the
// commit.precommit.mode config, defaulting to prompt.
func resolvePrecommitMode(raw string, cfg verify.CommitConfig) (string, error) {
	for _, candidate := range []string{raw, cfg.Precommit.Mode.String()} {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		return normalizeCheckMode(candidate, "--precommit")
	}
	return CheckModePrompt, nil
}

func shouldRunPrecommitChecks(mode string) bool {
	return mode != CheckModeSkip
}
