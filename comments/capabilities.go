package comments

import (
	"context"
	"errors"
	"sort"
)

type Capabilities struct {
	Workflow        Workflow  `json:"workflow"`
	Actions         []string  `json:"actions"`
	Providers       []string  `json:"providers"`
	Bindings        []Binding `json:"bindings"`
	AttachmentLimit int64     `json:"attachmentLimit"`
}

func (s *Service) Capabilities(ctx context.Context, id string) (Capabilities, error) {
	c, err := s.Get(ctx, id)
	if err != nil {
		return Capabilities{}, err
	}
	p, err := s.principal(ctx)
	if err != nil {
		return Capabilities{}, err
	}
	w, ok := s.workflows[c.WorkflowID]
	if !ok {
		return Capabilities{}, ErrInvalid
	}
	out := Capabilities{Workflow: copyWorkflow(w), Actions: []string{}, Providers: []string{}, Bindings: []Binding{}, AttachmentLimit: s.attachmentLimit}
	if store, ok := s.store.(ProviderStore); ok {
		out.Bindings, err = store.Bindings(ctx, p.TenantID, id)
		if err != nil {
			return out, err
		}
	}
	candidates := []string{"reply", "assign", "update", "metadata"}
	if p.Actor.Kind == "user" {
		candidates = append(candidates, "delete", "retarget")
	}
	if c.Phase != Final && len(c.Messages) > 0 {
		candidates = append(candidates, "edit-reply", "delete-reply")
	}
	if len(s.providers) > 0 && len(out.Bindings) == 0 && c.Phase != Final {
		candidates = append(candidates, "publish")
	}
	if s.blobs != nil && c.Phase != Final {
		candidates = append(candidates, "attach")
	}
	for _, binding := range out.Bindings {
		if binding.ThreadID != "" && c.Phase != Final {
			candidates = append(candidates, "publish-reply")
		}
		if binding.Resolvable {
			if c.Phase == Final {
				candidates = append(candidates, "reopen-native")
			} else {
				candidates = append(candidates, "resolve-native")
			}
		}
	}
	for action := range w.Actions {
		if _, _, err := w.Apply(c.Status, action, p.Actor); err == nil {
			candidates = append(candidates, action)
		}
	}
	for name := range s.providers {
		out.Providers = append(out.Providers, name)
	}
	for _, action := range candidates {
		if c.Phase == Final && (action == "reply" || action == "assign" || action == "update") {
			continue
		}
		if action == "assign" && (p.Actor.Kind != "user" || c.Phase != Active) {
			continue
		}
		bound := false
		for _, b := range out.Bindings {
			bound = bound || action == "update" || b.ThreadID != "" && (action == "resolve" || action == "close" || action == "reopen")
		}
		if bound {
			continue
		}
		if err = s.access.Authorize(ctx, p, action, c.Target); err != nil {
			if errors.Is(err, ErrForbidden) {
				continue
			}
			return out, err
		}
		out.Actions = append(out.Actions, action)
	}
	sort.Strings(out.Actions)
	sort.Strings(out.Providers)
	return out, nil
}
func (s *Service) Operation(ctx context.Context, id string) (Operation, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return Operation{}, err
	}
	store, ok := s.store.(OperationStore)
	if !ok {
		return Operation{}, ErrInvalid
	}
	op, err := store.Operation(ctx, p.TenantID, id)
	if err != nil {
		return op, err
	}
	_, err = s.Get(ctx, op.CommentID)
	return op, err
}
