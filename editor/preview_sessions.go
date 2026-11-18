package editor

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/flanksource/gavel/editor/preview"
)

type PreviewOptions struct {
	HostOrigin string
	CacheRoot  string
	Directory  string
	Runtime    string
}

type previewSessions struct {
	options     PreviewOptions
	installed   []string
	defaultHost *previewHost

	mu sync.Mutex
	// Each build has its own origin so its browser worker keeps a single package profile.
	hosts map[string]*previewHost
}

func (s *Server) StartPreview(ctx context.Context, opts PreviewOptions) (func() error, error) {
	var installed []string
	selector := opts.Runtime
	if selector == "" && len(s.PreviewLocal) == 0 {
		var err error
		installed, err = preview.ListNamedRuntimes(opts.Directory)
		if err != nil {
			return nil, fmt.Errorf("list installed Facet runtimes: %w", err)
		}
		if len(installed) != 0 {
			selector = installed[0]
		}
	}
	host, err := s.startPreviewHost(ctx, opts, selector)
	if err != nil {
		return nil, err
	}
	sessions := &previewSessions{options: opts, installed: installed, defaultHost: host, hosts: map[string]*previewHost{selector: host}}
	s.previews = sessions
	return sessions.close, nil
}

func (s *Server) selectPreview(ctx context.Context, selector string) (*previewHost, error) {
	sessions := s.previews
	if sessions == nil {
		return nil, fmt.Errorf("browser runtime is not started")
	}
	sessions.mu.Lock()
	defer sessions.mu.Unlock()
	if sessions.hosts == nil {
		return nil, fmt.Errorf("browser runtime is closed")
	}
	if selector == "" || sessions.options.Runtime != "" {
		return sessions.defaultHost, nil
	}
	if host, ok := sessions.hosts[selector]; ok {
		return host, nil
	}
	if !slices.Contains(sessions.installed, selector) {
		return nil, fmt.Errorf("Facet runtime %q is unavailable in this editor session", selector)
	}
	host, err := s.startPreviewHost(ctx, sessions.options, selector)
	if err != nil {
		return nil, err
	}
	sessions.hosts[selector] = host
	return host, nil
}

func (s *previewSessions) close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	for _, host := range s.hosts {
		err = errors.Join(err, host.close())
	}
	s.hosts = nil
	return err
}
