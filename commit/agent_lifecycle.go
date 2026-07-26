package commit

import (
	"errors"
	"fmt"

	clickyai "github.com/flanksource/gavel/ai"
)

func closeAgent(agent clickyai.Agent, target *error) {
	if err := agent.Close(); err != nil {
		*target = errors.Join(*target, fmt.Errorf("close AI agent: %w", err))
	}
}
