package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
	captaindb "github.com/flanksource/captain/pkg/database"
	"github.com/flanksource/captain/pkg/runtimeprofiles"
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/todos/lifecycle"
	"github.com/flanksource/gavel/verify"
)

func fixtureSpecResolver(cwd string, cfg verify.GavelConfig) func(context.Context) (api.ResolveSpecOptions, error) {
	var once sync.Once
	var runtime api.ResolveSpecOptions
	var err error
	return func(ctx context.Context) (api.ResolveSpecOptions, error) {
		once.Do(func() {
			saved, _, loadErr := captainconfig.Load()
			if loadErr != nil {
				err = fmt.Errorf("load fixture saved Captain config: %w", loadErr)
				return
			}
			host := &lifecycle.Host{Config: cfg, WorkDir: cwd, Kind: lifecycle.HostCLI, Saved: &saved}
			host.Catalog = func(ctx context.Context) (*runtimeprofiles.Catalog, error) {
				db, err := database.Shared(ctx)
				if err != nil {
					return nil, fmt.Errorf("open fixture runtime catalog database: %w", err)
				}
				options := runtimeprofiles.DefaultCatalogOptions{Cwd: cwd, Config: host.Saved}
				if !db.Disabled() {
					options.Read = func(context.Context) (*captaindb.DB, error) {
						return captaindb.Use(db.Gorm())
					}
				}
				return runtimeprofiles.NewDefaultCatalog(ctx, options)
			}
			var layers runtimeprofiles.LayerResult
			layers, err = host.StepLayers(ctx, lifecycle.Step{Name: lifecycle.StepVerify})
			if err != nil {
				return
			}
			runtime = api.ResolveSpecOptions{Layers: layers.Layers, Saved: &host.Saved.AI}
		})
		return runtime, err
	}
}
