package comments

import "context"

// Batch applies requests in one transaction. A reply's agent mention is
// delivered only once the transaction commits, by this (non-transaction)
// service; a batch that rolls back delivers nothing.
func (s *Service) Batch(ctx context.Context, requests []ActionRequest) ([]Comment, error) {
	store, ok := s.store.(TransactionStore)
	if !ok || len(requests) == 0 || len(requests) > 1000 {
		return nil, ErrInvalid
	}
	var out []Comment
	var mentions []*mentionDelivery
	err := store.Transaction(ctx, func(tx Store) error {
		service := *s
		service.store = tx
		for _, request := range requests {
			c, mention, err := service.actMentioning(ctx, request)
			if err != nil {
				return err
			}
			out, mentions = append(out, c), append(mentions, mention)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, mention := range mentions {
		s.deliverMention(ctx, mention)
	}
	return out, nil
}
