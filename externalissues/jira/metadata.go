package jira

import (
	"context"
	"fmt"

	"github.com/flanksource/clicky/task"
)

type CatalogOptions struct {
	NoCache bool
	Store   CatalogStore
}

func (c *Client) Catalog(ctx context.Context, options CatalogOptions) (map[string]FieldDefinition, error) {
	return runJiraTask(ctx, "Load Jira field catalog", func(ctx context.Context, progress *task.Task) (map[string]FieldDefinition, error) {
		progress.SetDescription("Resolving Jira account")
		if c.account == "" {
			var identity struct {
				ID string `json:"accountId"`
			}
			if err := c.json(ctx, requestOptions{Path: "/myself"}, &identity); err != nil {
				return nil, err
			}
			if identity.ID == "" {
				return nil, fmt.Errorf("Jira myself response has no account ID")
			}
			c.account = identity.ID
		}
		progress.SetDescription("Reading field catalog cache")
		fields, err := options.Store.Resolve(ctx, CatalogScope{Site: c.Site, Account: c.account}, options.NoCache, func(ctx context.Context) ([]FieldDefinition, error) {
			progress.SetDescription("Refreshing field catalog from Jira")
			var fields []FieldDefinition
			if err := c.json(ctx, requestOptions{Path: "/field"}, &fields); err != nil {
				return nil, err
			}
			if fields == nil {
				return nil, fmt.Errorf("Jira field response requires an array")
			}
			return fields, nil
		})
		if err != nil {
			return nil, err
		}
		progress.SetDescription(fmt.Sprintf("Loaded %d field definitions", len(fields)))
		return fields, nil
	})
}
