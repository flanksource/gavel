package commands

import (
	"context"

	"github.com/flanksource/gavel/externalissues/jira"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/gavel/workspace"
)

func newJiraClient(ctx context.Context, options jira.ClientOptions) (*jira.Client, error) {
	options, err := jiraClientOptions(ctx, options)
	if err != nil {
		return nil, err
	}
	return jira.NewClient(ctx, options)
}

func jiraClientOptions(ctx context.Context, options jira.ClientOptions) (jira.ClientOptions, error) {
	dirs, err := workspace.From(ctx)
	if err != nil {
		return options, err
	}
	config, err := verify.LoadGavelConfig(dirs.Invocation)
	if err != nil {
		return options, err
	}
	options.IgnoreFields = config.ExternalIssues.Jira.IgnoreFields
	return options, nil
}
