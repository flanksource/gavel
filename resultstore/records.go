package resultstore

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/commons-db/query"
	"github.com/flanksource/commons-db/recordstore"
	"github.com/flanksource/commons-db/recordstore/recordresults"
	"github.com/flanksource/commons-db/recordstore/sqlite"
)

type recordBackend struct{ backend recordstore.Backend }

func OpenSQLite(ctx context.Context, dir string) (Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	schemas := recordstore.NewSchemas()
	if err := schemas.Register("gavel_runs", []query.ColumnDef{
		{Name: "executionId", Type: query.ColumnTypeString},
		{Name: "payload", Type: query.ColumnTypeJSON},
	}, recordstore.KindOptions{Key: "executionId", OnConflict: recordstore.OnConflictReplace}); err != nil {
		return nil, err
	}
	backend, err := recordresults.OpenSharedSQLite(ctx, sqlite.Options{Path: filepath.Join(dir, "launch.sqlite"), Schema: schemas.Kind, SweepInterval: time.Minute}, "gavel-launch")
	if err != nil {
		return nil, err
	}
	return New(&recordBackend{backend: backend}, fileLock(dir)), nil
}

func (b *recordBackend) All(ctx context.Context) ([]Run, error) {
	var runs []Run
	err := b.backend.Scan(ctx, "gavel-runs", 0, func(_ int64, row recordstore.Row) error {
		data, err := json.Marshal(row["payload"])
		if err != nil {
			return err
		}
		var run Run
		if err := json.Unmarshal(data, &run); err != nil {
			return err
		}
		runs = append(runs, run)
		return nil
	})
	if errors.Is(err, recordstore.ErrNotFound) {
		return nil, nil
	}
	return runs, err
}

func (b *recordBackend) Get(ctx context.Context, id string) (Run, error) {
	runs, err := b.All(ctx)
	if err != nil {
		return Run{}, err
	}
	for _, run := range runs {
		if run.Metadata.ExecutionID == id {
			return run, nil
		}
	}
	return Run{}, ErrNotFound
}

func (b *recordBackend) Put(ctx context.Context, run Run) error {
	payload, err := recordstore.EncodeRow(run)
	if err != nil {
		return err
	}
	_, err = b.backend.Append(ctx, "gavel-runs", "gavel_runs", []recordstore.Row{{"executionId": run.Metadata.ExecutionID, "payload": payload}})
	return err
}

func (b *recordBackend) Close() error { return b.backend.Close() }
