package main

import (
	"context"
	"errors"
	"fmt"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/gavel/internal/database"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/flanksource/gavel/launch"
	"github.com/flanksource/gavel/resultstore"
	"github.com/flanksource/gavel/runconfig"
)

type launchCommandSchemas struct{}

func (launchCommandSchemas) Schema(kind string) (map[string]any, error) {
	command, err := schemaCommand(kind)
	if err != nil {
		return nil, err
	}
	schema := commandSchema(command, projectActionPositional(kind), commandSchemaOptions{})
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return nil, errors.New("native command schema has no properties")
	}
	options := reflect.TypeFor[runconfig.Test]()
	if kind == "lint" {
		options = reflect.TypeFor[runconfig.Lint]()
	}
	allowed := map[string]bool{}
	for index := range options.NumField() {
		allowed[strings.Split(options.Field(index).Tag.Get("json"), ",")[0]] = true
	}
	for _, field := range []string{"work-dir", "ui", "addr", "dev", "auto-stop", "idle-timeout", "linters", "lint", "triage", "ai-fix", "ai-fix-max-iterations", "yes"} {
		delete(allowed, field)
	}
	for field := range properties {
		if !allowed[field] {
			delete(properties, field)
		}
	}
	return map[string]any(schema), nil
}

func newLaunchService(ctx context.Context, root, backend string, localRoots []string) (*launch.Service, error) {
	db, err := database.Shared(ctx)
	if err != nil {
		return nil, err
	}
	workspaces, err := launch.NewWorkspaces(launch.WorkspacePolicy{Root: root, AllowLocal: len(localRoots) > 0, LocalRoots: localRoots, Context: func(ctx context.Context) dbcontext.Context { return dbcontext.NewContext(ctx).WithDB(db.Gorm(), nil) }})
	if err != nil {
		return nil, err
	}
	directory := filepath.Join(root, ".gavel", "runs")
	store, err := openLaunchStore(ctx, backend, directory)
	if err != nil {
		return nil, errors.Join(err, workspaces.Close())
	}
	service, err := launch.NewService(launch.ServiceOptions{Workspaces: workspaces, Store: store, Schemas: launchCommandSchemas{}, Executor: launch.NativeExecutor{Schemas: launchCommandSchemas{}}})
	if err != nil {
		return nil, errors.Join(err, store.Close(), workspaces.Close())
	}
	return service, nil
}

func openLaunchStore(ctx context.Context, backend, directory string) (resultstore.Store, error) {
	switch backend {
	case "json":
		return resultstore.OpenSnapshots(directory)
	case "sqlite":
		return resultstore.OpenSQLite(ctx, directory)
	default:
		return nil, fmt.Errorf("unknown launch result store %q; use json or sqlite", backend)
	}
}
