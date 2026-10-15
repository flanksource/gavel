// Builds the sandbox dispatcher for hosted gavel serve from its --sandbox flags, after checking
// the backend can run jobs safely: the container host's firewall, or the Kubernetes NetworkPolicy.
package main

import (
	"context"
	"fmt"
	"strings"

	ghcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/sandbox"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/labels"
)

// newSandboxDispatcher returns nil when --sandbox is unset.
func newSandboxDispatcher(ctx context.Context, opts DashboardOptions, db *gorm.DB, cache *ghcache.Store, secret []byte) (*sandbox.Dispatcher, error) {
	var executor sandbox.Executor
	switch opts.Sandbox {
	case "":
		return nil, nil
	case "docker", "podman":
		check := sandbox.HostCheck{Binary: opts.Sandbox, Network: opts.SandboxNetwork, ServeAddr: opts.SandboxServeAddr}
		if err := check.Run(ctx); err != nil {
			return nil, fmt.Errorf("--sandbox %s: %w", opts.Sandbox, err)
		}
		executor = sandbox.NewContainer(opts.Sandbox, sandbox.ContainerOptions{Network: opts.SandboxNetwork})
	case "kubernetes":
		selector, err := labels.ConvertSelectorToLabelsMap(opts.ServePodSelector)
		if err != nil {
			return nil, fmt.Errorf("--serve-pod-selector %q must be labels like app.kubernetes.io/name=gavel: %w", opts.ServePodSelector, err)
		}
		kubernetes, err := sandbox.NewKubernetes(ctx, sandbox.KubernetesOptions{
			Namespace: opts.SandboxNamespace, ServePodSelector: selector, ServePort: opts.Port, ExtraDeny: opts.SandboxExtraDeny,
		})
		if err != nil {
			return nil, err
		}
		if err := kubernetes.EnsureNetworkPolicy(ctx); err != nil {
			return nil, err
		}
		executor = kubernetes
	default:
		return nil, fmt.Errorf("--sandbox must be docker, podman or kubernetes, got %q", opts.Sandbox)
	}
	dispatcher := &sandbox.Dispatcher{
		DB: db, Executor: executor, Image: opts.SandboxImage, Server: strings.TrimRight(opts.PublicURL, "/"),
		TokenKey: runTokenKey(secret),
	}
	if cache != nil {
		dispatcher.Logs = cache.OrgBlobs
	}
	return dispatcher, nil
}
