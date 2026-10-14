package runtime

import (
	"fmt"

	"github.com/flanksource/captain/pkg/captainconfig"
	"github.com/flanksource/captain/pkg/plans"
)

// planHooks are Captain's plan.revision hook prompts for workDir: the embedded
// defaults, the captain prompt dirs and <workDir>/.captain/prompts. Providers
// built without hook discovery (New, used by tests and portable import, which
// must persist plans verbatim) run none.
func (p *Provider) planHooks(workDir string) ([]plans.Hook, error) {
	if !p.discoverPlanHooks {
		return nil, nil
	}
	cfg, _, err := captainconfig.Load()
	if err != nil {
		return nil, fmt.Errorf("load Captain configuration for plan hooks: %w", err)
	}
	dirs, err := plans.HookDirs(cfg, workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve plan hook dirs: %w", err)
	}
	hooks, err := plans.DiscoverHooks(dirs...)
	if err != nil {
		return nil, fmt.Errorf("discover plan hooks: %w", err)
	}
	return hooks, nil
}
