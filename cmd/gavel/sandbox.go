// Builds the sandbox dispatcher for hosted gavel serve from its --sandbox flags, after checking
// the backend can run jobs safely: the container host's firewall, the Kubernetes NetworkPolicy,
// or, on agent-substrate, its default-deny egress.
package main

import (
	"context"
	"fmt"
	"os"
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
	case "substrate":
		if len(opts.SandboxSecretEnv) > 0 {
			return nil, fmt.Errorf("--sandbox substrate keeps secrets out of its jobs: drop --sandbox-secret-env and give model keys to substrate's egress gateway with --substrate-credential")
		}
		credentials, err := substrateCredentials(opts.SubstrateCredential)
		if err != nil {
			return nil, err
		}
		substrate, err := sandbox.NewSubstrate(sandbox.SubstrateOptions{
			EnvAPI: opts.SubstrateEnvAPI, Control: opts.SubstrateControl,
			ControlTokenFile: opts.SubstrateTokenFile, ControlCAFile: opts.SubstrateCAFile,
			Atespace: opts.SubstrateAtespace, Template: opts.SubstrateTemplate,
			Egress: opts.SubstrateEgress, Credentials: credentials, TrustBundle: opts.SubstrateTrustBundle,
		})
		if err != nil {
			return nil, fmt.Errorf("--sandbox substrate: %w", err)
		}
		executor = substrate
	default:
		return nil, fmt.Errorf("--sandbox must be docker, podman, kubernetes or substrate, got %q", opts.Sandbox)
	}
	secrets, err := sandboxSecrets(opts.SandboxSecretEnv, os.Getenv)
	if err != nil {
		return nil, err
	}
	dispatcher := &sandbox.Dispatcher{
		DB: db, Executor: executor, Image: opts.SandboxImage, Server: strings.TrimRight(opts.PublicURL, "/"),
		TokenKey: runTokenKey(secret), SecretEnv: secrets,
	}
	if cache != nil {
		dispatcher.Logs = cache.OrgBlobs
	}
	return dispatcher, nil
}

// substrateCredentials parses --substrate-credential values: host=,header=,env=,uri= and an
// optional prefix=, the model keys substrate's egress gateway injects.
func substrateCredentials(values []string) ([]sandbox.InjectedCredential, error) {
	credentials := make([]sandbox.InjectedCredential, 0, len(values))
	for _, value := range values {
		var credential sandbox.InjectedCredential
		for _, pair := range strings.Split(value, ",") {
			key, field, _ := strings.Cut(pair, "=")
			switch key {
			case "host":
				credential.Host = field
			case "header":
				credential.Header = field
			case "prefix":
				credential.Prefix = field
			case "env":
				credential.Env = field
			case "uri":
				credential.URI = field
			default:
				return nil, fmt.Errorf("--substrate-credential %q: unknown key %q; keys are host, header, prefix, env and uri", value, key)
			}
		}
		if credential.Host == "" || credential.Header == "" || credential.Env == "" || credential.URI == "" {
			return nil, fmt.Errorf("--substrate-credential %q needs host, header, env and uri", value)
		}
		credentials = append(credentials, credential)
	}
	return credentials, nil
}

// sandboxSecrets reads the --sandbox-secret-env variables from serve's environment. A variable
// that isn't set is an error: a job would run without the key it was meant to get.
func sandboxSecrets(names []string, getenv func(string) string) (map[string]string, error) {
	secrets := make(map[string]string, len(names))
	for _, name := range names {
		value := getenv(name)
		if value == "" {
			return nil, fmt.Errorf("--sandbox-secret-env %s is not set in gavel serve's environment", name)
		}
		secrets[name] = value
	}
	return secrets, nil
}
