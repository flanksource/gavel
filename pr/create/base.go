package prcreate

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/verify"
)

// ResolveBase picks the ref a PR branches from: an explicit override first,
// then pr.base from the .gavel.yaml governing dir, then origin's default branch.
func ResolveBase(dir, override string) (string, error) {
	if base := strings.TrimSpace(override); base != "" {
		return base, nil
	}
	cfg, err := verify.LoadGavelConfig(dir)
	if err != nil {
		return "", fmt.Errorf("load PR configuration: %w", err)
	}
	if base := strings.TrimSpace(cfg.PR.Base); base != "" {
		return base, nil
	}
	return defaultBase(dir)
}
