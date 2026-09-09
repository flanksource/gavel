package git

import (
	"context"
	"errors"
	"fmt"

	"github.com/flanksource/captain/pkg/api"
	captaincli "github.com/flanksource/captain/pkg/cli"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/ai"
)

type preparedPromptOptions struct {
	Name         string
	Runtime      captaincli.AIRuntimeResolved
	Agent        ai.Agent
	AgentFactory func(ai.AgentConfig) (ai.Agent, error)
}

func executePreparedPrompt(ctx context.Context, options preparedPromptOptions) (response *ai.PromptResponse, err error) {
	for _, warning := range options.Runtime.Resolution.Warnings {
		logger.Warnf("%s: %s", options.Name, warning)
	}
	agent := options.Agent
	if agent == nil {
		if options.AgentFactory == nil {
			return nil, fmt.Errorf("%s: no agent or agent factory was supplied", options.Name)
		}
		agent, err = options.AgentFactory(options.Runtime.Config)
		if err != nil {
			return nil, err
		}
		if agent == nil {
			return nil, fmt.Errorf("%s: agent factory returned no agent", options.Name)
		}
		defer func() { err = errors.Join(err, agent.Close()) }()
	}
	return agent.ExecutePrompt(ctx, ai.PromptRequest{Name: options.Name, Spec: api.Spec(options.Runtime.Request)})
}
