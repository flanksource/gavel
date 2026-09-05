package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/flanksource/captain/pkg/api"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/verify"
)

func fixtureSpecResolver(cwd string, cfg verify.GavelConfig) func(context.Context) (api.Spec, error) {
	var once sync.Once
	var spec api.Spec
	var err error
	return func(ctx context.Context) (api.Spec, error) {
		once.Do(func() {
			host := &lifecycle.Host{Config: cfg, WorkDir: cwd, Kind: lifecycle.HostCLI}
			host.Catalog = func(ctx context.Context) (*runtimeprofiles.Catalog, error) {
				db, err := database.Shared(ctx)
				if err != nil {
					return nil, fmt.Errorf("open fixture runtime catalog database: %w", err)
				}
				options := runtimeprofiles.DefaultCatalogOptions{Cwd: cwd}
				if !db.Disabled() {
					options.Read = func(context.Context) (*captaindb.DB, error) {
						return captaindb.Use(db.Gorm())
					}
				}
				return runtimeprofiles.NewDefaultCatalog(ctx, options)
			}
			var defaults verify.PromptSpec
			defaults, err = host.StepDefaults(ctx, lifecycle.Step{Name: lifecycle.StepVerify})
			spec = defaults.Spec
		})
		return spec, err
	}
}
