// Package substratetest reads the live Substrate cluster the substrate e2e tests run against: its
// API addresses, gavel's token, the runner image, and where jobs reach the test's stand-in for serve.
package substratetest

import (
	"os"
	"testing"
)

// TrustBundle is where the gavel-runner template mounts the egress gateway's CA.
const TrustBundle = "/run/ate/egress-trust-bundle/trust-bundle.pem"

// Cluster is the cluster docs/sandbox.md sets up, as its environment variables override it.
type Cluster struct {
	EnvAPI    string
	Control   string
	TokenFile string
	Template  string
	// Image is the runner image, pinned by digest, in a registry the cluster pulls from.
	Image string
	// Listen is the host address a test serves on; ServeURL is how jobs in the cluster reach it.
	Listen   string
	ServeURL string
	// Credential is the value of the Secret CredentialURI names, which the gateway injects.
	Credential    string
	CredentialURI string
}

// FromEnv skips t unless GAVEL_SANDBOX_SUBSTRATE_TEST=1, and returns the cluster's settings.
func FromEnv(t testing.TB) Cluster {
	t.Helper()
	if os.Getenv("GAVEL_SANDBOX_SUBSTRATE_TEST") != "1" {
		t.Skip("set GAVEL_SANDBOX_SUBSTRATE_TEST=1 with a Substrate cluster set up as docs/sandbox.md describes to run the substrate e2e")
	}
	setting := func(name, fallback string) string {
		if value := os.Getenv(name); value != "" {
			return value
		}
		if fallback == "" {
			t.Fatalf("set %s", name)
		}
		return fallback
	}
	return Cluster{
		EnvAPI:        setting("GAVEL_SUBSTRATE_ENV_API", "localhost:7777"),
		Control:       setting("GAVEL_SUBSTRATE_CONTROL", "localhost:8443"),
		TokenFile:     setting("GAVEL_SUBSTRATE_TOKEN_FILE", ""),
		Template:      setting("GAVEL_SUBSTRATE_TEMPLATE", "gavel-runner"),
		Image:         setting("GAVEL_SUBSTRATE_IMAGE", ""),
		Listen:        setting("GAVEL_SUBSTRATE_SERVE_LISTEN", "172.18.0.1:18080"),
		ServeURL:      setting("GAVEL_SUBSTRATE_SERVE_URL", "http://gavel-host.default.svc.cluster.local:18080"),
		Credential:    setting("GAVEL_SUBSTRATE_CREDENTIAL", "gavel-e2e-injected-credential"),
		CredentialURI: setting("GAVEL_SUBSTRATE_CREDENTIAL_URI", "ate-secret://k8s.io/default/gavel-e2e/model/key"),
	}
}
