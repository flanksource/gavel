// Runs sandbox jobs on agent-substrate: each job is an environment created from the runner image,
// whose one process is the job, with egress limited to an allowlist and model keys injected outside.
package sandbox

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/gavel/sandbox/internal/substrate/ateapipb"
	"github.com/flanksource/gavel/sandbox/internal/substrate/ateenvv1alpha"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

// DefaultSubstrateEgress are the hosts a job reaches besides gavel serve and the model APIs: the
// Go, npm and Python registries, and GitHub. They are passed through without decryption.
var DefaultSubstrateEgress = []string{
	"proxy.golang.org", "sum.golang.org",
	"registry.npmjs.org",
	"pypi.org", "files.pythonhosted.org",
	"github.com", "codeload.github.com", "objects.githubusercontent.com",
}

// substrateCredentialPlaceholder is what a job sends in a header the egress gateway replaces with
// the real credential, which never enters the environment.
const substrateCredentialPlaceholder = "injected-by-the-substrate-egress-gateway"

// InjectedCredential is a credential the egress gateway puts into a job's requests to Host: it
// replaces Header's value with Prefix and the credential URI names. Env is the variable the job's
// client reads the credential from, set to a placeholder so the client sends the header.
type InjectedCredential struct {
	Host   string
	Header string
	Prefix string
	// URI names the credential for Substrate's credential provider, such as
	// ate-secret://k8s.io/default/<namespace>/<secret>/<key>.
	URI string
	Env string
}

// SubstrateOptions configures the agent-substrate executor.
type SubstrateOptions struct {
	// EnvAPI is the ate-env-api gRPC address, such as ate-env-api.ate-env.svc:7777.
	EnvAPI string
	// Control is the ate-api-server gRPC address, which egress policies are set through.
	Control string
	// ControlTokenFile holds the bearer token the control plane accepts, such as a projected
	// ServiceAccount token with the api.ate-system.svc audience. It is read on every call.
	ControlTokenFile string
	// ControlCAFile holds the CA the control plane's certificate chains to; empty skips
	// verification, as a port-forwarded control plane needs.
	ControlCAFile string
	// Atespace is the atespace environments are created in; empty is ate-env.
	Atespace string
	// Template is the ActorTemplate environments derive from, its container replaced by the job's
	// image; empty is env's default-template.
	Template string
	// Egress are hosts jobs may reach besides DefaultSubstrateEgress.
	Egress []string
	// Credentials are the model API keys the egress gateway injects.
	Credentials []InjectedCredential
	// TrustBundle is the path, inside the environment, of the egress gateway's CA, which a job
	// must trust to reach a host whose requests carry an injected credential.
	TrustBundle string
}

type substrateExecutor struct {
	options   SubstrateOptions
	envs      ateenvv1alpha.EnvironmentServiceClient
	processes ateenvv1alpha.ProcessServiceClient
	control   ateapipb.ControlClient
}

// NewSubstrate returns an executor that runs jobs as agent-substrate environments.
func NewSubstrate(options SubstrateOptions) (Executor, error) {
	if options.EnvAPI == "" || options.Control == "" {
		return nil, errors.New("the substrate executor needs the ate-env-api and ate-api-server addresses")
	}
	if options.Atespace == "" {
		options.Atespace = "ate-env"
	}
	// ate-env-api serves cleartext gRPC inside the cluster.
	envConn, err := grpc.NewClient(options.EnvAPI, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect to ate-env-api %s: %w", options.EnvAPI, err)
	}
	tlsConfig := &tls.Config{InsecureSkipVerify: options.ControlCAFile == ""} //nolint:gosec // a port-forwarded control plane serves a cluster-internal certificate
	if options.ControlCAFile != "" {
		pem, err := os.ReadFile(options.ControlCAFile)
		if err != nil {
			return nil, fmt.Errorf("read the control plane's CA: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("the control plane's CA file %s holds no certificate", options.ControlCAFile)
		}
		tlsConfig.RootCAs = pool
	}
	dial := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))}
	if options.ControlTokenFile != "" {
		dial = append(dial, grpc.WithPerRPCCredentials(tokenFile(options.ControlTokenFile)))
	}
	conn, err := grpc.NewClient(options.Control, dial...)
	if err != nil {
		return nil, fmt.Errorf("connect to ate-api-server %s: %w", options.Control, err)
	}
	return &substrateExecutor{
		options:   options,
		envs:      ateenvv1alpha.NewEnvironmentServiceClient(envConn),
		processes: ateenvv1alpha.NewProcessServiceClient(envConn),
		control:   ateapipb.NewControlClient(conn),
	}, nil
}

func (e *substrateExecutor) Name() string { return "substrate" }

func (e *substrateExecutor) Submit(ctx context.Context, job Job) (Handle, error) {
	if err := substrateImage(job.Image); err != nil {
		return nil, err
	}
	process, err := substrateProcess(job, e.options)
	if err != nil {
		return nil, err
	}
	policy, err := substrateEgressPolicy(e.options, e.options.Atespace, job.Env["GAVEL_SERVER"])
	if err != nil {
		return nil, err
	}
	id := jobName(job.RunID)
	create := &ateenvv1alpha.CreateEnvironmentRequest{Id: id, Atespace: e.options.Atespace, RootfsImage: job.Image}
	if e.options.Template != "" {
		create.Template = &ateenvv1alpha.Template{Name: e.options.Template, Atespace: e.options.Atespace}
	}
	if _, err := e.envs.CreateEnvironment(ctx, create); err != nil {
		return nil, fmt.Errorf("create substrate environment %s: %w", id, err)
	}
	handle := &substrateHandle{id: id, atespace: e.options.Atespace, envs: e.envs, processes: e.processes}
	if _, err := e.control.CreateActorEgressPolicy(ctx, &ateapipb.CreateActorEgressPolicyRequest{
		Actor: &ateapipb.ObjectRef{Atespace: e.options.Atespace, Name: id}, EgressPolicy: policy,
	}); err != nil {
		_ = handle.remove(context.WithoutCancel(ctx))
		return nil, fmt.Errorf("set the egress policy of substrate environment %s: %w", id, err)
	}
	started, err := e.processes.StartProcess(handle.guest(ctx), process)
	if err != nil {
		_ = handle.remove(context.WithoutCancel(ctx))
		return nil, fmt.Errorf("start the job in substrate environment %s: %w", id, err)
	}
	handle.process = started.GetProcessId()
	return handle, nil
}

// substrateImage refuses an image that isn't pinned by digest, which Substrate requires.
func substrateImage(image string) error {
	if !regexp.MustCompile(`@sha256:[0-9a-f]{64}$`).MatchString(image) {
		return fmt.Errorf("substrate runs images pinned by digest: %q must be repo@sha256:<digest>", image)
	}
	return nil
}

// substrateProcess is the job's process: its command and settings, its run token, and a
// placeholder for each injected credential. Any other secret would enter the environment, so it
// is refused.
func substrateProcess(job Job, options SubstrateOptions) (*ateenvv1alpha.StartProcessRequest, error) {
	env := map[string]string{}
	for key, value := range job.Env {
		env[key] = value
	}
	for key, value := range job.SecretEnv {
		if key != "GAVEL_RUN_TOKEN" {
			return nil, fmt.Errorf("secret %s would enter the substrate environment; give it to substrate's egress gateway as an injected credential instead", key)
		}
		env[key] = value
	}
	for _, credential := range options.Credentials {
		if credential.Env != "" {
			env[credential.Env] = substrateCredentialPlaceholder
		}
	}
	if options.TrustBundle != "" && len(options.Credentials) > 0 {
		env["NODE_EXTRA_CA_CERTS"] = options.TrustBundle
	}
	return &ateenvv1alpha.StartProcessRequest{
		Command: job.Command,
		Env:     env,
		Timeout: durationpb.New(job.Limits.withDefaults().Timeout),
	}, nil
}

// substrateEgressPolicy allows a job to reach gavel serve at server, the hosts it needs, and the
// model APIs, whose credentials the gateway injects. Everything else is denied.
func substrateEgressPolicy(options SubstrateOptions, atespace, server string) (*ateapipb.EgressPolicy, error) {
	serve, err := url.Parse(server)
	if err != nil || serve.Hostname() == "" {
		return nil, fmt.Errorf("substrate egress names hosts: GAVEL_SERVER %q isn't a URL", server)
	}
	if net.ParseIP(serve.Hostname()) != nil {
		return nil, fmt.Errorf("substrate egress names hosts: GAVEL_SERVER %q must name serve by DNS name, not an IP address", server)
	}
	port := uint64(80)
	if serve.Scheme == "https" {
		port = 443
	}
	if serve.Port() != "" {
		// url.Parse checks a port is digits, not that it is a port.
		port, err = strconv.ParseUint(serve.Port(), 10, 16)
		if err != nil || port == 0 {
			return nil, fmt.Errorf("substrate egress: GAVEL_SERVER %q has port %s, outside 1-65535", server, serve.Port())
		}
	}
	ports := &ateapipb.Ports{Numbers: []int32{int32(port)}} //nolint:gosec // at most 65535
	rules := []*ateapipb.EgressRule{}
	if serve.Scheme == "https" {
		rules = append(rules, &ateapipb.EgressRule{TlsPassthrough: &ateapipb.TlsPassthroughRule{Hostnames: []string{serve.Hostname()}, Ports: ports}})
	} else {
		rules = append(rules, &ateapipb.EgressRule{Http: &ateapipb.HttpRule{Hostnames: []string{serve.Hostname()}, Ports: ports}})
	}
	hosts := append(append([]string{}, DefaultSubstrateEgress...), options.Egress...)
	rules = append(rules, &ateapipb.EgressRule{TlsPassthrough: &ateapipb.TlsPassthroughRule{Hostnames: hosts, Ports: &ateapipb.Ports{Numbers: []int32{443}}}})
	for _, credential := range options.Credentials {
		rules = append(rules, &ateapipb.EgressRule{Https: &ateapipb.HttpsRule{
			Hostnames: []string{credential.Host},
			Effects: &ateapipb.HttpRuleEffects{ReplaceHeaders: []*ateapipb.CredentialHeader{{
				Header: credential.Header, Prefix: credential.Prefix, CredentialUri: credential.URI,
			}}},
		}})
	}
	return &ateapipb.EgressPolicy{
		Metadata: &ateapipb.ResourceMetadata{Atespace: atespace, Name: "default"},
		Rules:    rules,
	}, nil
}

type substrateHandle struct {
	id, atespace string
	envs         ateenvv1alpha.EnvironmentServiceClient
	processes    ateenvv1alpha.ProcessServiceClient
	process      string
}

func (h *substrateHandle) Ref() string { return h.id }

// guest addresses ctx's calls to the processes of h's environment, as ate-env-api routes them.
func (h *substrateHandle) guest(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "x-env-id", h.id, "x-env-atespace", h.atespace)
}

// Logs follows the job's stdout and stderr until it exits. Substrate's router ends a long stream,
// so it reconnects from the offsets it has read.
func (h *substrateHandle) Logs(ctx context.Context) (io.ReadCloser, error) {
	reader, writer := io.Pipe()
	go func() {
		var stdout, stderr int64
		for {
			stream, err := h.processes.StreamProcessOutput(h.guest(ctx), &ateenvv1alpha.StreamProcessOutputRequest{
				ProcessId: h.process, Follow: true, StdoutOffset: stdout, StderrOffset: stderr,
			})
			if err != nil {
				writer.CloseWithError(fmt.Errorf("follow the output of %s: %w", h.id, err))
				return
			}
			for {
				message, err := stream.Recv()
				if err != nil {
					break
				}
				var chunk []byte
				switch {
				case message.GetExit() != nil:
					writer.Close()
					return
				case message.GetStdout() != nil:
					chunk = message.GetStdout()
					stdout += int64(len(chunk))
				default:
					chunk = message.GetStderr()
					stderr += int64(len(chunk))
				}
				if _, err := writer.Write(chunk); err != nil {
					return
				}
			}
			if ctx.Err() != nil {
				writer.CloseWithError(ctx.Err())
				return
			}
		}
	}()
	return reader, nil
}

// Wait returns the job's exit code once it exits, polling, since a single stream doesn't outlast
// Substrate's router, then removes its environment.
func (h *substrateHandle) Wait(ctx context.Context) (Result, error) {
	for {
		process, err := h.processes.GetProcess(h.guest(ctx), &ateenvv1alpha.GetProcessRequest{ProcessId: h.process})
		if err != nil && ctx.Err() != nil {
			return Result{}, ctx.Err()
		}
		if err == nil && process.GetState() == ateenvv1alpha.ProcessState_PROCESS_STATE_EXITED {
			result := Result{ExitCode: int(process.GetExitCode())}
			return result, h.remove(context.WithoutCancel(ctx))
		}
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

// Cancel kills the job and removes its environment.
func (h *substrateHandle) Cancel(ctx context.Context) error {
	if h.process != "" {
		_, err := h.processes.SignalProcess(h.guest(ctx), &ateenvv1alpha.SignalProcessRequest{ProcessId: h.process, Signal: ateenvv1alpha.Signal_SIGNAL_KILL})
		if err != nil && !processExited(err) {
			return fmt.Errorf("kill the job in substrate environment %s: %w", h.id, err)
		}
	}
	return h.remove(ctx)
}

// processExited reports whether err refuses a signal because the process already exited, which
// then needs no killing.
func processExited(err error) bool {
	return status.Code(err) == codes.FailedPrecondition && strings.Contains(status.Convert(err).Message(), "has exited")
}

func (h *substrateHandle) remove(ctx context.Context) error {
	if _, err := h.envs.DeleteEnvironment(ctx, &ateenvv1alpha.DeleteEnvironmentRequest{Id: h.id, Atespace: h.atespace}); err != nil {
		return fmt.Errorf("delete substrate environment %s: %w", h.id, err)
	}
	return nil
}

// tokenFile sends the bearer token in a file, read on every call so a rotated token is used.
type tokenFile string

func (t tokenFile) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	token, err := os.ReadFile(string(t))
	if err != nil {
		return nil, fmt.Errorf("read the control plane token: %w", err)
	}
	return map[string]string{"authorization": "Bearer " + strings.TrimSpace(string(token))}, nil
}

func (tokenFile) RequireTransportSecurity() bool { return true }
