package prcreate

import (
	"fmt"
	"strings"

	"github.com/flanksource/gavel/verify"
)

// DefaultBase is the ref a PR branches from when neither the caller nor the
// project's .gavel.yaml pr.base names one.
const DefaultBase = "origin/main"

// ResolveBase picks the ref a PR branches from: an explicit override first,
// then pr.base from the .gavel.yaml governing dir, then DefaultBase.
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
	return DefaultBase, nil
}
