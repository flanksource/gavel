package comments

import (
	"context"
	"encoding/json"
	"fmt"
)

// TargetSources is a Targets host that reads the current text of a thread's
// target, so a patch anchor resolves against it; only the host can read file
// content.
type TargetSources interface {
	// TargetSource is t's current text. ok is false when the host keeps no
	// file content for targets of t's kind (a web page): their threads carry
	// no resolved anchor. A target the host cannot read is an error.
	TargetSource(ctx context.Context, t Target) (source string, ok bool, err error)
}

// Anchors resolves where each thread's patch anchor lies in its target's
// current text (comments.ResolvePatch), by thread id, reading each target
// once. A thread whose selector carries no patch, or whose target the host
// has no content for, has no entry; a host without TargetSources has none.
func (s *Service) Anchors(ctx context.Context, threads []Comment) (map[string]PatchLocation, error) {
	out := map[string]PatchLocation{}
	sources, ok := s.targets.(TargetSources)
	if !ok {
		return out, nil
	}
	type content struct {
		text string
		ok   bool
	}
	read := map[[2]string]content{}
	for _, c := range threads {
		patch, err := selectorPatch(c)
		if err != nil {
			return nil, err
		}
		if patch == nil {
			continue
		}
		key := [2]string{c.Target.ScopeURI, c.Target.URI}
		got, seen := read[key]
		if !seen {
			if got.text, got.ok, err = sources.TargetSource(ctx, c.Target); err != nil {
				return nil, fmt.Errorf("reading %s to resolve its anchors: %w", c.Target.URI, err)
			}
			read[key] = got
		}
		if got.ok {
			out[c.ID] = ResolvePatch(got.text, *patch)
		}
	}
	return out, nil
}

// selectorPatch is the patch anchor c's selector carries; nil for none.
func selectorPatch(c Comment) (*PatchAnchor, error) {
	if len(c.Target.Selector) == 0 {
		return nil, nil
	}
	var selector struct {
		Patch *PatchAnchor `json:"patch"`
	}
	if err := json.Unmarshal(c.Target.Selector, &selector); err != nil {
		return nil, fmt.Errorf("decode the selector of comment %s: %w", c.ID, err)
	}
	return selector.Patch, nil
}
