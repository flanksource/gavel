package commit

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrBatchRequiresInteractive = errors.New("--batch requires --interactive")
	ErrBatchWithSummary         = errors.New("--batch cannot be combined with --summary")
)

func collectInteractiveBatches(ctx context.Context, opts Options) ([][]string, error) {
	var batches [][]string
	queued := map[string]struct{}{}
	for {
		picked, _, err := selectInteractiveFiles(ctx, opts, queued)
		if err != nil {
			if len(batches) > 0 && (errors.Is(err, ErrNothingStaged) || errors.Is(err, ErrInteractiveCancelled)) {
				return batches, nil
			}
			return nil, err
		}

		batch := append([]string(nil), picked.Selected...)
		batches = append(batches, batch)
		for _, path := range batch {
			queued[path] = struct{}{}
		}
		fmt.Fprintf(interactiveStdout, "queued batch %d with %d file(s)\n", len(batches), len(batch))
	}
}

func runInteractiveBatch(ctx context.Context, opts Options) (*Result, error) {
	batches, err := collectInteractiveBatches(ctx, opts)
	if err != nil {
		return nil, err
	}

	result := &Result{DryRun: opts.DryRun}
	for i, batch := range batches {
		if err := stageExplicitFiles(opts.WorkDir, batch); err != nil {
			return result, fmt.Errorf("batch %d of %d: %w", i+1, len(batches), err)
		}

		batchOpts := opts
		batchOpts.Interactive = false
		batchOpts.Stage = StageStaged
		batchOpts.Files = nil
		single, err := runSingleCommit(ctx, batchOpts)
		result = mergeResults(result, single)
		if err != nil {
			return result, fmt.Errorf("batch %d of %d: %w", i+1, len(batches), err)
		}
	}
	return result, nil
}
