// Builds the sandbox dispatcher for hosted gavel serve from its --sandbox flags, after checking
// the backend can run jobs safely: the container host's firewall, the Kubernetes NetworkPolicy,
// or, on agent-substrate, its default-deny egress.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/flanksource/captain/pkg/substrate"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/commons/properties"
	"github.com/flanksource/gavel/budget"
	ghcache "github.com/flanksource/gavel/github/cache"
	"github.com/flanksource/gavel/modelkeys"
	"github.com/flanksource/gavel/sandbox"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
)

// newSandboxDispatcher returns nil when --sandbox is unset. On substrate, jobs run in a checkout of
// their commit from sources, and runAPI answers what they ask of their runs over each run's
// channel.
func newSandboxDispatcher(ctx context.Context, opts DashboardOptions, db *gorm.DB, cache *ghcache.Store, secret []byte, sources *sandbox.Sources, runAPI http.Handler) (*sandbox.Dispatcher, error) {
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
		var err error
		executor, err = sandbox.NewSubstrate(sandbox.SubstrateOptions{
			Options: substrateOptions(opts), Egress: opts.SubstrateEgress, RunAPI: runAPI,
		})
		if err != nil {
			return nil, fmt.Errorf("--sandbox substrate: %w", err)
		}
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
		templates, err := orgTemplates(opts, db)
		if err != nil {
			return nil, err
		}
		dispatcher.Templates = templates
		dispatcher.Admission = &sandbox.Admission{
			Budgets: budget.Store{DB: db}, RunLifetime: dispatcher.RunLifetime(),
			Wait:    properties.Duration(15*time.Minute, "sandbox.queue.timeout"),
			Recheck: properties.Duration(30*time.Second, "sandbox.budget.interval"),
		}
		dispatcher.Sources = sources
		// Each run's requests carry its organization's keys, which the credential provider in
		// substrate's cluster fetches from this gavel by the run.
		dispatcher.Keys = modelkeys.Store{DB: db}
		dispatcher.CredentialURI = func(run uuid.UUID, provider string) string {
			return "ate-secret://gavel/" + opts.SubstrateGavelName + "/" + run.String() + "/" + provider
		}
	}
	return dispatcher, nil
}

// sandboxSweepInterval is how often serve sweeps what runs left behind.
const sandboxSweepInterval = 10 * time.Minute

// sweepSandboxEvery fails the runs dispatcher lost track of and, when sweeper is set, sweeps what
// runs older than their lifetime left behind, and, when templates is set, the organizations'
// templates no run uses any more, now and every sandboxSweepInterval until ctx ends.
func sweepSandboxEvery(ctx context.Context, dispatcher *sandbox.Dispatcher, sweeper sandbox.Sweeper, templates templateSweeper) {
	ticker := time.NewTicker(sandboxSweepInterval)
	defer ticker.Stop()
	for {
		if failed, err := dispatcher.FailExpiredRuns(ctx); err != nil && ctx.Err() == nil {
			logger.Warnf("%v", err)
		} else if failed > 0 {
			logger.Infof("failed %d sandbox runs still running past their lifetime of %s", failed, dispatcher.RunLifetime())
		}
		if sweeper != nil {
			sweepSandbox(ctx, sweeper, dispatcher.RunLifetime())
		}
		if templates != nil {
			sweepOrgTemplates(ctx, templates)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// templateSweeper deletes the organizations' templates no run uses any more, and returns their
// names.
type templateSweeper interface {
	Sweep(ctx context.Context) ([]string, error)
}

// sweepOrgTemplates deletes the organizations' templates no run uses any more, logging what it
// deleted. A sweep that fails is a warning; the next one tries again.
func sweepOrgTemplates(ctx context.Context, templates templateSweeper) {
	swept, err := templates.Sweep(ctx)
	if len(swept) > 0 {
		logger.Infof("deleted %d organization substrate templates no run uses any more: %s", len(swept), strings.Join(swept, ", "))
	}
	if err != nil && ctx.Err() == nil {
		logger.Warnf("sweep the organizations' substrate templates: %v", err)
	}
}

// sweepSandbox deletes the environments runs older than lifetime left behind, logging what it
// deleted. A sweep that fails is a warning; the next one tries again.
func sweepSandbox(ctx context.Context, sweeper sandbox.Sweeper, lifetime time.Duration) {
	swept, err := sweeper.Sweep(ctx, lifetime)
	if len(swept) > 0 {
		logger.Infof("deleted %d sandbox environments that runs older than %s left behind: %s", len(swept), lifetime, strings.Join(swept, ", "))
	}
	if err != nil && ctx.Err() == nil {
		logger.Warnf("sweep the sandbox environments runs left behind: %v", err)
	}
}

// orgTemplates gives each organization's runs their own copy of --substrate-template, at
// --substrate-cpu and --substrate-memory unless a platform admin set the organization's size.
func orgTemplates(opts DashboardOptions, db *gorm.DB) (*sandbox.OrgTemplates, error) {
	size, err := substrateSize(opts)
	if err != nil {
		return nil, err
	}
	client, err := substrate.Dial(substrateOptions(opts))
	if err != nil {
		return nil, fmt.Errorf("--sandbox substrate: %w", err)
	}
	return &sandbox.OrgTemplates{DB: db, Client: client, Atespace: opts.SubstrateAtespace, Base: opts.SubstrateTemplate, Size: size}, nil
}

// substrateSize is --substrate-cpu and --substrate-memory; a flag left empty is zero, which the
// base template's size fills.
func substrateSize(opts DashboardOptions) (substrate.Size, error) {
	var size substrate.Size
	var errs []error
	if opts.SubstrateCPU != "" {
		cpu, err := resource.ParseQuantity(opts.SubstrateCPU)
		if err != nil || cpu.Sign() <= 0 {
			errs = append(errs, fmt.Errorf("--substrate-cpu must be a positive Kubernetes quantity such as 2 or 500m, got %q", opts.SubstrateCPU))
		}
		size.CPUMillicores = cpu.MilliValue()
	}
	if opts.SubstrateMemory != "" {
		memory, err := resource.ParseQuantity(opts.SubstrateMemory)
		if err != nil || memory.Sign() <= 0 {
			errs = append(errs, fmt.Errorf("--substrate-memory must be a positive Kubernetes quantity such as 4Gi, got %q", opts.SubstrateMemory))
		}
		size.MemoryBytes = memory.Value()
	}
	return size, errors.Join(errs...)
}

// substrateOptions is how the --substrate-* flags reach the substrate jobs and agent runs execute
// on.
func substrateOptions(opts DashboardOptions) substrate.Options {
	return substrate.Options{
		EnvAPI: opts.SubstrateEnvAPI, Control: opts.SubstrateControl,
		ControlTokenFile: opts.SubstrateTokenFile, CAFile: opts.SubstrateCAFile,
		ClientCertFile: opts.SubstrateClientCert, ClientKeyFile: opts.SubstrateClientKey,
		Atespace: opts.SubstrateAtespace, Template: opts.SubstrateTemplate,
		TrustBundle: opts.SubstrateTrustBundle,
	}
}

// substrateAgentBackend is the substrate backend hosted agent runs execute on through captain's
// substrate sandbox, as its options spell it, from the --substrate-* flags jobs use too, reaching
// what jobs reach; nil without --sandbox substrate, since agent runs need it.
func substrateAgentBackend(opts DashboardOptions) (map[string]any, error) {
	if opts.Sandbox != "substrate" {
		return nil, nil
	}
	return sandbox.SubstrateBackend(substrateOptions(opts), opts.SubstrateEgress), nil
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
