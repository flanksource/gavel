package settings

import (
	"context"
	"fmt"

	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/projects"
)

// Open is the settings service in catalog's mode: database layers beside the
// projects table (gavel_settings rows reference its projects), files only
// beside the file catalog.
func Open(ctx context.Context, catalog projects.Service) (Service, error) {
	switch catalog.(type) {
	case *projects.File:
		return File{}, nil
	case *projects.DB:
		db, err := database.Require(ctx, "gavel settings")
		if err != nil {
			return nil, err
		}
		return NewDB(db)
	}
	return nil, fmt.Errorf("no settings mode for a %T project catalog", catalog)
}

// Shared is the settings service of the process's project catalog (projects.Shared).
func Shared(ctx context.Context) (Service, error) {
	catalog, err := projects.Shared(ctx)
	if err != nil {
		return nil, err
	}
	return Open(ctx, catalog)
}
