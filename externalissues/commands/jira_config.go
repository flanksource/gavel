package commands

import (
	"context"

	"github.com/flanksource/gavel/externalissues/jira"
	"github.com/flanksource/gavel/verify"
	"github.com/flanksource/gavel/workspace"
)

func newJiraClient(ctx context.Context, options jira.ClientOptions) (*jira.Client, error) {
	dirs, err := workspace.From(ctx)
	if err != nil {
		return nil, err
	}
	config, err := verify.LoadGavelConfig(dirs.Invocation)
	if err != nil {
		return nil, err
	}
	options.IgnoreFields = config.ExternalIssues.Jira.IgnoreFields
	return jira.NewClient(ctx, options)
}
