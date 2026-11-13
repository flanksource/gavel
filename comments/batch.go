package comments

import "context"

func (s *Service) Batch(ctx context.Context, requests []ActionRequest) ([]Comment, error) {
	store, ok := s.store.(TransactionStore)
	if !ok || len(requests) == 0 || len(requests) > 1000 {
		return nil, ErrInvalid
	}
	var out []Comment
	err := store.Transaction(ctx, func(tx Store) error {
		service := *s
		service.store = tx
		for _, request := range requests {
			c, err := service.Act(ctx, request)
			if err != nil {
				return err
			}
			out = append(out, c)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
