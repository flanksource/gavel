package comments

import "context"

func (s *Service) ResolveTarget(ctx context.Context, action string, target Target) (Target, error) {
	p, err := s.principal(ctx)
	if err != nil {
		return target, err
	}
	if err = s.access.Authorize(ctx, p, action, target); err != nil {
		return target, err
	}
	target, err = s.targets.Resolve(ctx, p, target)
	if err != nil {
		return target, err
	}
	if target.URI == "" || target.ScopeURI == "" || target.Kind == "" {
		return target, ErrInvalid
	}
	return target, s.access.Authorize(ctx, p, action, target)
}
