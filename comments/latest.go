package comments

import (
	"context"
	"errors"
	"fmt"
)

// latestAttempts bounds how often WithLatest re-reads a thread that keeps moving.
const latestAttempts = 5

// WithLatest runs fn on the thread's latest version and, when fn conflicts
// because a concurrent writer moved the thread, re-reads it and runs fn again.
// A conflict at an unchanged version is not a race and is returned as is, and a
// thread that moves on every attempt fails with ErrConflict.
func (s *Service) WithLatest(ctx context.Context, id string, fn func(Comment) (Comment, error)) (Comment, error) {
	c, err := s.Get(ctx, id)
	if err != nil {
		return Comment{}, err
	}
	for attempt := 1; ; attempt++ {
		out, err := fn(c)
		if !errors.Is(err, ErrConflict) {
			return out, err
		}
		latest, getErr := s.Get(ctx, id)
		switch {
		case getErr != nil:
			return Comment{}, errors.Join(err, getErr)
		case latest.Version == c.Version:
			return Comment{}, err
		case attempt == latestAttempts:
			return Comment{}, fmt.Errorf("thread %s changed on each of %d attempts: %w", id, latestAttempts, err)
		}
		c = latest
	}
}
