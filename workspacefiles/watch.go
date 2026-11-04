package workspacefiles

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type Lease struct {
	ID               string    `json:"id"`
	Expires          time.Time `json:"expires"`
	IdleMilliseconds int64     `json:"idleMilliseconds"`
	File             File      `json:"file"`
}
type Event struct {
	Error   string `json:"error,omitempty"`
	Expired bool   `json:"expired,omitempty"`
}
type lease struct {
	Lease
	owner  string
	path   string
	events chan Event
}

func (s *Service) Acquire(ctx context.Context, input FileRequest, owner string) (Lease, error) {
	file, err := s.Read(ctx, input)
	if err != nil {
		return Lease{}, err
	}
	path := filepath.Join(input.Root, filepath.FromSlash(input.Path))
	s.mu.Lock()
	defer s.mu.Unlock()
	directory := filepath.Dir(path)
	if s.directories[directory] == 0 {
		if err := s.watcher.Add(directory); err != nil {
			return Lease{}, err
		}
	}
	s.directories[directory]++
	entry := &lease{Lease: Lease{ID: uuid.NewString(), Expires: time.Now().Add(s.options.IdleTimeout), IdleMilliseconds: s.options.IdleTimeout.Milliseconds(), File: file}, owner: owner, path: path, events: make(chan Event, 1)}
	s.leases[entry.ID] = entry
	return entry.Lease, nil
}

func (s *Service) Renew(id, owner string) (Lease, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := s.leases[id]
	if entry == nil || entry.owner != owner || time.Now().After(entry.Expires) {
		return Lease{}, ErrExpired
	}
	entry.Expires = time.Now().Add(s.options.IdleTimeout)
	return entry.Lease, nil
}

func (s *Service) Subscribe(id, owner string) (<-chan Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := s.leases[id]
	if entry == nil || entry.owner != owner || time.Now().After(entry.Expires) {
		return nil, ErrExpired
	}
	return entry.events, nil
}

func (s *Service) Release(id, owner string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := s.leases[id]
	if entry == nil {
		return nil
	}
	if entry.owner != owner {
		return ErrExpired
	}
	return s.release(entry)
}

func notify(entry *lease, event Event) {
	select {
	case entry.events <- event:
	default:
	}
}

func (s *Service) release(entry *lease) error {
	delete(s.leases, entry.ID)
	directory := filepath.Dir(entry.path)
	s.directories[directory]--
	if s.directories[directory] > 0 {
		return nil
	}
	delete(s.directories, directory)
	return s.watcher.Remove(directory)
}

func (s *Service) expire(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, entry := range s.leases {
		if now.Before(entry.Expires) {
			continue
		}
		event := Event{Expired: true}
		if err := s.release(entry); err != nil {
			event.Error = err.Error()
		}
		// Expiration must replace a queued filesystem notification.
		select {
		case <-entry.events:
		default:
		}
		notify(entry, event)
	}
}

func (s *Service) watchCount() int { s.mu.Lock(); defer s.mu.Unlock(); return len(s.directories) }

func (s *Service) watch() {
	defer close(s.done)
	ticker := time.NewTicker(min(time.Second, s.options.IdleTimeout))
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case now := <-ticker.C:
			s.expire(now)
		case change, ok := <-s.watcher.Events:
			if !ok {
				return
			}
			s.mu.Lock()
			for _, entry := range s.leases {
				if filepath.Clean(change.Name) == entry.path {
					notify(entry, Event{})
				}
			}
			s.mu.Unlock()
		case err, ok := <-s.watcher.Errors:
			if !ok {
				return
			}
			if err == nil {
				err = errors.New("filesystem watcher failed")
			}
			s.mu.Lock()
			for _, entry := range s.leases {
				notify(entry, Event{Error: err.Error()})
			}
			s.mu.Unlock()
		}
	}
}
