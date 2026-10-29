// `gavel credential-provider`, deployed in substrate's cluster: serves substrate's egress gateway
// the model keys of gavel runs, asking the gavel each credential URI names.
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/sandbox/credentialprovider"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// gatewayFetchLimit is Envoy's ext_proc message timeout, after which the egress gateway abandons
// a request whose credential hasn't arrived.
const gatewayFetchLimit = 5 * time.Second

type credentialProviderOptions struct {
	ListenAddress    string
	HealthAddress    string
	ServerBundle     string
	ClientCAFile     string
	InjectorIdentity string
	Gavels           string
	FetchTimeout     time.Duration
	CacheTTL         time.Duration
}

func init() {
	opts := credentialProviderOptions{}
	cmd := &cobra.Command{
		Use:    "credential-provider",
		Short:  "Serve substrate's egress gateway the model keys of gavel runs",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()
			return runCredentialProvider(ctx, opts, nil)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&opts.ListenAddress, "listen-address", ":50051", "gRPC listen address")
	flags.StringVar(&opts.HealthAddress, "health-address", ":9090", "HTTP listen address for /healthz and /readyz")
	flags.StringVar(&opts.ServerBundle, "server-cred-bundle", "", "Credential bundle (PKCS#8 key and certificate chain) served for TLS, such as a servicedns pod certificate")
	flags.StringVar(&opts.ClientCAFile, "client-ca-file", "", "Trust bundle the egress gateway's client certificate chains to, such as the podidentity bundle")
	flags.StringVar(&opts.InjectorIdentity, "injector-identity", "", "SPIFFE ID the caller's certificate must carry: the egress gateway, e.g. spiffe://cluster.local/ns/ate-system/sa/atenet-egress")
	flags.StringVar(&opts.Gavels, "gavels", "", "YAML file naming each gavel, as in its credential URIs, with its url and tokenFile")
	flags.DurationVar(&opts.FetchTimeout, "fetch-timeout", 3*time.Second, "How long one ask of a gavel may take; under the egress gateway's 5s limit")
	flags.DurationVar(&opts.CacheTTL, "cache-ttl", time.Minute, "How long a key is reused for the same run and actor")
	rootCmd.AddCommand(cmd)
}

// runCredentialProvider serves until ctx ends, then lets in-flight fetches finish. listening, when
// set, is told the gRPC and health addresses once both serve.
func runCredentialProvider(ctx context.Context, opts credentialProviderOptions, listening func(grpcAddr, healthAddr string)) error {
	if opts.Gavels == "" {
		return errors.New("--gavels is required: the file naming each gavel's url and tokenFile")
	}
	gavels, err := credentialprovider.LoadGavels(opts.Gavels)
	if err != nil {
		return err
	}
	if opts.FetchTimeout >= gatewayFetchLimit {
		return fmt.Errorf("--fetch-timeout must be under %s, the egress gateway's limit, got %s", gatewayFetchLimit, opts.FetchTimeout)
	}
	creds, err := credentialprovider.ServerCredentials(opts.ServerBundle, opts.ClientCAFile, opts.InjectorIdentity)
	if err != nil {
		return err
	}
	server := grpc.NewServer(grpc.Creds(creds))
	credentialprovider.New(credentialprovider.Config{
		Gavels: gavels, FetchTimeout: opts.FetchTimeout, CacheTTL: opts.CacheTTL,
	}).Register(server)
	grpcListener, err := net.Listen("tcp", opts.ListenAddress)
	if err != nil {
		return fmt.Errorf("listen on --listen-address %s: %w", opts.ListenAddress, err)
	}
	healthListener, err := net.Listen("tcp", opts.HealthAddress)
	if err != nil {
		_ = grpcListener.Close()
		return fmt.Errorf("listen on --health-address %s: %w", opts.HealthAddress, err)
	}
	var ready atomic.Bool
	health := &http.Server{Handler: credentialProviderHealth(&ready), ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = health.Serve(healthListener) }()
	go func() {
		<-ctx.Done()
		ready.Store(false)
		stopped := make(chan struct{})
		go func() { server.GracefulStop(); close(stopped) }()
		select {
		case <-stopped:
		case <-time.After(5 * time.Second):
			server.Stop()
		}
		_ = health.Close()
	}()
	ready.Store(true)
	logger.Infof("gavel credential-provider serving %d gavels on %s, health on %s", len(gavels), grpcListener.Addr(), healthListener.Addr())
	if listening != nil {
		listening(grpcListener.Addr().String(), healthListener.Addr().String())
	}
	if err := server.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("serve the credential provider: %w", err)
	}
	return nil
}

// credentialProviderHealth serves /healthz, always ok so a draining pod isn't restarted, and
// /readyz, ok only while the provider serves.
func credentialProviderHealth(ready *atomic.Bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !ready.Load() {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}
