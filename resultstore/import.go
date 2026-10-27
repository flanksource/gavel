package resultstore

import (
	"context"
	"encoding/json"
	"errors"
)

func (s *adapter) Import(ctx context.Context, run Run) (bool, error) {
	if !run.Terminal() || run.Metadata.Revision == 0 {
		return false, errors.New("only terminal revisioned executions can be imported")
	}
	created := false
	err := s.lock(func() error {
		previous, err := s.backend.Get(ctx, run.Metadata.ExecutionID)
		if err == nil {
			left, err := json.Marshal(previous)
			if err != nil {
				return err
			}
			right, err := json.Marshal(run)
			if err != nil {
				return err
			}
			if string(left) != string(right) {
				return ErrConflict
			}
			return nil
		}
		if !errors.Is(err, ErrNotFound) {
			return err
		}
		all, err := s.backend.All(ctx)
		if err != nil {
			return err
		}
		for _, existing := range all {
			if existing.Admission == run.Admission && existing.Metadata.Workspace == run.Metadata.Workspace {
				return ErrConflict
			}
		}
		if err := s.backend.Put(ctx, run); err != nil {
			return err
		}
		created = true
		return nil
	})
	return created, err
}
