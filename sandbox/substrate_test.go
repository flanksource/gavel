// Tests what the agent-substrate executor asks of Substrate for a job: the environment's egress
// allowlist, with model keys injected outside it, and the job's process.
package sandbox

import (
	"testing"
	"time"

	"github.com/flanksource/gavel/sandbox/internal/substrate/ateapipb"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testCredential = InjectedCredential{
	Host: "api.anthropic.com", Header: "x-api-key", URI: "ate-secret://k8s.io/default/gavel/anthropic/key", Env: "ANTHROPIC_API_KEY",
}

func TestSubstrateEgressAllowsOnlyWhatJobsNeed(t *testing.T) {
	opts := SubstrateOptions{Egress: []string{"goproxy.example.com"}, Credentials: []InjectedCredential{testCredential}}
	policy, err := substrateEgressPolicy(opts, "ate-env", "http://gavel.gavel.svc.cluster.local:8080")
	require.NoError(t, err)
	assert.Equal(t, "ate-env", policy.GetMetadata().GetAtespace())
	assert.Equal(t, "default", policy.GetMetadata().GetName())

	var serve *ateapipb.HttpRule
	var passthrough []string
	var injected *ateapipb.HttpsRule
	for _, rule := range policy.GetRules() {
		switch {
		case rule.GetHttp() != nil:
			serve = rule.GetHttp()
		case rule.GetTlsPassthrough() != nil:
			passthrough = append(passthrough, rule.GetTlsPassthrough().GetHostnames()...)
			assert.Equal(t, []int32{443}, rule.GetTlsPassthrough().GetPorts().GetNumbers(), "Substrate requires a passed-through rule's ports")
		case rule.GetHttps() != nil:
			injected = rule.GetHttps()
		}
	}
	require.NotNil(t, serve, "jobs reach serve's run API")
	assert.Equal(t, []string{"gavel.gavel.svc.cluster.local"}, serve.GetHostnames())
	assert.Equal(t, []int32{8080}, serve.GetPorts().GetNumbers())
	assert.Subset(t, passthrough, []string{"proxy.golang.org", "sum.golang.org", "registry.npmjs.org", "pypi.org", "files.pythonhosted.org", "github.com", "codeload.github.com", "objects.githubusercontent.com", "goproxy.example.com"})
	require.NotNil(t, injected, "the model API is reached with its key injected by the gateway")
	assert.Equal(t, []string{"api.anthropic.com"}, injected.GetHostnames())
	require.Len(t, injected.GetEffects().GetReplaceHeaders(), 1)
	header := injected.GetEffects().GetReplaceHeaders()[0]
	assert.Equal(t, "x-api-key", header.GetHeader())
	assert.Equal(t, testCredential.URI, header.GetCredentialUri())

	tls, err := substrateEgressPolicy(SubstrateOptions{}, "ate-env", "https://gavel.example.com")
	require.NoError(t, err)
	assert.Contains(t, tls.GetRules(), &ateapipb.EgressRule{TlsPassthrough: &ateapipb.TlsPassthroughRule{
		Hostnames: []string{"gavel.example.com"}, Ports: &ateapipb.Ports{Numbers: []int32{443}},
	}}, "an HTTPS serve is passed through")

	_, err = substrateEgressPolicy(SubstrateOptions{}, "ate-env", "http://10.0.0.5:8080")
	require.EqualError(t, err, `substrate egress names hosts: GAVEL_SERVER "http://10.0.0.5:8080" must name serve by DNS name, not an IP address`)
}

func TestSubstrateProcessRunsTheJobWithItsRunTokenOnly(t *testing.T) {
	run := uuid.New()
	job := Job{
		RunID:     run,
		Image:     "ghcr.io/flanksource/gavel-runner@sha256:" + strings64("a"),
		Command:   []string{"gavel", "runner", "exec", "--run", run.String()},
		Env:       map[string]string{"GAVEL_SERVER": "http://gavel.gavel.svc.cluster.local:8080", "CI": "1"},
		SecretEnv: map[string]string{"GAVEL_RUN_TOKEN": "run-token"},
		Limits:    Limits{Timeout: 20 * time.Minute},
	}
	opts := SubstrateOptions{Credentials: []InjectedCredential{testCredential}, TrustBundle: "/ate/trust/egress-mitm.pem"}
	process, err := substrateProcess(job, opts)
	require.NoError(t, err)
	assert.Equal(t, job.Command, process.GetCommand())
	assert.Equal(t, 20*time.Minute, process.GetTimeout().AsDuration())
	assert.Equal(t, map[string]string{
		"GAVEL_SERVER":        "http://gavel.gavel.svc.cluster.local:8080",
		"CI":                  "1",
		"GAVEL_RUN_TOKEN":     "run-token",
		"ANTHROPIC_API_KEY":   substrateCredentialPlaceholder,
		"NODE_EXTRA_CA_CERTS": "/ate/trust/egress-mitm.pem",
	}, process.GetEnv())

	job.SecretEnv["ANTHROPIC_API_KEY"] = "sk-ant-secret"
	_, err = substrateProcess(job, opts)
	require.EqualError(t, err, "secret ANTHROPIC_API_KEY would enter the substrate environment; give it to substrate's egress gateway as an injected credential instead")
	assert.NotContains(t, err.Error(), "sk-ant-secret")
}

func TestSubstrateImagesArePinnedByDigest(t *testing.T) {
	require.NoError(t, substrateImage("ghcr.io/flanksource/gavel-runner@sha256:"+strings64("0")))
	require.EqualError(t, substrateImage("ghcr.io/flanksource/gavel-runner:1"),
		`substrate runs images pinned by digest: "ghcr.io/flanksource/gavel-runner:1" must be repo@sha256:<digest>`)
}

func strings64(c string) string {
	out := ""
	for range 64 {
		out += c
	}
	return out
}

func TestSubstrateEgressRefusesAPortOutOfRange(t *testing.T) {
	for _, port := range []string{"0", "65536", "4294967297"} {
		server := "http://gavel.example.com:" + port
		_, err := substrateEgressPolicy(SubstrateOptions{}, "ate-env", server)
		require.EqualError(t, err, `substrate egress: GAVEL_SERVER "`+server+`" has port `+port+`, outside 1-65535`)
	}
}

func TestSubstrateActorIsTheSPIFFEIDOfTheRunsEnvironment(t *testing.T) {
	run := uuid.MustParse("6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f")
	assert.Equal(t, "spiffe://substrate-actor.local/actor/ate-env/gavel-run-6f1c2d3e-4a5b-6c7d-8e9f-0a1b2c3d4e5f", SubstrateActor("ate-env", run))
}
