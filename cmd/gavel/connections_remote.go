package main

import (
	"context"
	shared "github.com/flanksource/commons-db/cmd/query/connections"
	"github.com/flanksource/commons-db/models"
	"net/url"
)

type connectionRemote struct {
	ctx    context.Context
	server gavelServer
}

func connectionPath(ref string) string { return "/api/connections/" + url.PathEscape(ref) }
func (r connectionRemote) Create(ctx context.Context, body map[string]any) (*models.Connection, error) {
	var result models.Connection
	err := r.server.call(ctx, "POST", "/api/connections", body, &result)
	return &result, err
}
func (r connectionRemote) Update(ctx context.Context, ref string, body map[string]any) (*models.Connection, error) {
	var result models.Connection
	err := r.server.call(ctx, "PATCH", connectionPath(ref), body, &result)
	return &result, err
}
func (r connectionRemote) ResolveReference(ref string) (*models.Connection, error) {
	var result models.Connection
	err := r.server.call(r.ctx, "GET", connectionPath(ref), nil, &result)
	return &result, err
}
func (r connectionRemote) Inventory(options shared.ListOptions) ([]shared.ConnectionInventory, error) {
	var result []shared.ConnectionInventory
	err := r.server.call(r.ctx, "GET", "/api/connections?type="+url.QueryEscape(options.Type), nil, &result)
	return result, err
}
func (r connectionRemote) Validate(ctx context.Context, options shared.ValidateOptions) (shared.ValidationResult, error) {
	var result shared.ValidationResult
	err := r.server.call(ctx, "POST", connectionPath(options.Reference)+"/validate", nil, &result)
	return result, err
}
func (r connectionRemote) Delete(ref string) error {
	return r.server.call(r.ctx, "DELETE", connectionPath(ref), nil, nil)
}
