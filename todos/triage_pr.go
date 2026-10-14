package todos

import (
	"context"
	"github.com/flanksource/gavel/todos/types"
)

func preserveRetirementPR(ctx context.Context, provider Provider, retiring []triageRetirement) error {
	var sources []*types.TODO
	var survivor *types.TODO
	for _, retirement := range retiring {
		if retirement.Survivor != nil {
			survivor = retirement.Survivor
			sources = append(sources, retirement.Retired, survivor)
		}
	}
	pr, err := types.MergedPR(sources)
	if err != nil {
		return err
	}
	if survivor != nil && survivor.PR == nil && pr != nil {
		if err := provider.Edit(ctx, survivor, EditRequest{PR: &pr}); err != nil {
			return err
		}
		survivor.PR = pr
	}
	return nil
}
