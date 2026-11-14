package comments

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

func (s *Service) edit(ctx context.Context, p Principal, c *Comment, req ActionRequest, now time.Time) (Change, bool, error) {
	switch req.Action {
	case "metadata":
		if len(req.Content) == 0 || !json.Valid(req.Content) {
			return Change{}, true, ErrInvalid
		}
		c.Content = req.Content
	case "retarget":
		if p.Actor.Kind != "user" {
			return Change{}, true, ErrForbidden
		}
		if req.Target == nil || req.Target.ScopeURI != c.Target.ScopeURI {
			return Change{}, true, ErrInvalid
		}
		target, err := s.ResolveTarget(ctx, req.Action, *req.Target)
		if err != nil {
			return Change{}, true, err
		}
		c.Target = target
	case "delete", "restore":
		if p.Actor.Kind != "user" {
			return Change{}, true, ErrForbidden
		}
		c.DeletedAt = nil
		if req.Action == "delete" {
			c.DeletedAt = &now
		}
	case "edit-reply", "delete-reply":
		if req.MessageID == "" {
			return Change{}, true, ErrInvalid
		}
		if c.Phase == Final {
			return Change{}, true, ErrConflict
		}
		for i, m := range c.Messages {
			if m.ID != req.MessageID {
				continue
			}
			if p.Actor.Kind != "user" && (m.Author.ID != p.Actor.ID || m.Author.Kind != p.Actor.Kind) {
				return Change{}, true, ErrForbidden
			}
			if req.Action == "delete-reply" {
				for _, child := range c.Messages {
					if child.ParentID == m.ID {
						return Change{}, true, ErrConflict
					}
				}
				c.Messages = append(c.Messages[:i:i], c.Messages[i+1:]...)
				return Change{DeleteMessageID: m.ID}, true, nil
			}
			if strings.TrimSpace(req.Body) == "" {
				return Change{}, true, ErrInvalid
			}
			m.Body = req.Body
			if len(req.Content) > 0 {
				if !json.Valid(req.Content) {
					return Change{}, true, ErrInvalid
				}
				m.Content = req.Content
			}
			c.Messages[i] = m
			return Change{EditMessage: &m}, true, nil
		}
		return Change{}, true, ErrNotFound
	default:
		return Change{}, false, nil
	}
	return Change{}, true, nil
}
