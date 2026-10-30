// Builds the sandbox dispatcher for hosted gavel serve from its --sandbox flags, after checking
// the backend can run jobs safely: the container host's firewall, the Kubernetes NetworkPolicy,
// or, on agent-substrate, its default-deny egress.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/flanksource/commons/logger"
	ghcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/modelkeys"
	"github.com/flanksource/gavel/sandbox"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/labels"
)

// newSandboxDispatcher returns nil when --sandbox is unset.
func newSandboxDispatcher(ctx context.Context, opts DashboardOptions, db *gorm.DB, cache *ghcache.Store, secret []byte) (*sandbox.Dispatcher, error) {
	if warning := sandboxKeysWarning(opts); warning != "" {
		logger.Warnf("%s", warning)
	}
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
			return nil, fmt.Errorf("--sandbox substrate keeps secrets out of its jobs: drop --sandbox-secret-env; organizations set their own model keys, which substrate's egress gateway injects")
		}
		substrate, err := sandbox.NewSubstrate(sandbox.SubstrateOptions{
			EnvAPI: opts.SubstrateEnvAPI, Control: opts.SubstrateControl,
			ControlTokenFile: opts.SubstrateTokenFile, CAFile: opts.SubstrateCAFile,
			ClientCertFile: opts.SubstrateClientCert, ClientKeyFile: opts.SubstrateClientKey,
			Atespace: opts.SubstrateAtespace, Template: opts.SubstrateTemplate,
			Egress: opts.SubstrateEgress, TrustBundle: opts.SubstrateTrustBundle,
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
	if opts.Sandbox == "substrate" {
		// Each run's requests carry its organization's keys, which the credential provider in
		// substrate's cluster fetches from this gavel by the run.
		dispatcher.Keys = modelkeys.Store{DB: db}
		dispatcher.CredentialURI = func(run uuid.UUID, provider string) string {
			return "ate-secret://gavel/" + opts.SubstrateGavelName + "/" + run.String() + "/" + provider
		}
	}
	return dispatcher, nil
}

// sandboxKeysWarning says which model keys a container or Kubernetes sandbox's runs use: serve's
// --sandbox-secret-env keys, the same for every organization, since organization keys need
// substrate's egress gateway. Those sandboxes are for development.
func sandboxKeysWarning(opts DashboardOptions) string {
	switch opts.Sandbox {
	case "", "substrate":
		return ""
	}
	if len(opts.SandboxSecretEnv) == 0 {
		return fmt.Sprintf("--sandbox %s: runs get no model keys, since --sandbox-secret-env is unset; organization model keys need --sandbox substrate", opts.Sandbox)
	}
	return fmt.Sprintf("--sandbox %s: every organization's runs use serve's --sandbox-secret-env keys (%s), not their own; organization model keys need --sandbox substrate",
		opts.Sandbox, strings.Join(opts.SandboxSecretEnv, ", "))
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
