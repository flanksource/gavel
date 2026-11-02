// Tests the provider's mutual TLS over a real gRPC connection: only a caller whose certificate
// chains to the client CA and carries the egress gateway's SPIFFE ID is served.
package credentialprovider

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flanksource/gavel/sandbox/internal/substrate/credproviderpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	gatewayIdentity = "spiffe://cluster.local/ns/ate-system/sa/atenet-egress"
	providerName    = "gavel-credential-provider.ate-system.svc"
)

type authority struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
}

func newAuthority(t *testing.T) authority {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test CA"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return authority{cert: cert, key: key}
}

// issue returns a credential bundle, a PKCS#8 key followed by the certificate, as substrate's pod
// certificates are written, for a server name or a URI SAN.
func (a authority) issue(t *testing.T, dnsName, uri string) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()), NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	if dnsName != "" {
		template.DNSNames = []string{dnsName}
	}
	if uri != "" {
		parsed, err := url.Parse(uri)
		require.NoError(t, err)
		template.URIs = []*url.URL{parsed}
	}
	der, err := x509.CreateCertificate(rand.Reader, template, a.cert, &key.PublicKey, a.key)
	require.NoError(t, err)
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	return append(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8}), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})...)
}

func writeFile(t *testing.T, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, content, 0o600))
	return path
}

func TestServerCredentialsAdmitOnlyTheEgressGateway(t *testing.T) {
	servicedns, podidentity := newAuthority(t), newAuthority(t)
	creds, err := ServerCredentials(
		writeFile(t, "server.pem", servicedns.issue(t, providerName, "")),
		writeFile(t, "trust-bundle.pem", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: podidentity.cert.Raw})),
		gatewayIdentity,
	)
	require.NoError(t, err)

	g := newGavel(t)
	server := grpc.NewServer(grpc.Creds(creds))
	newServer(t, g).Register(server)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	call := func(bundle []byte) error {
		t.Helper()
		client, err := tls.X509KeyPair(bundle, bundle)
		require.NoError(t, err)
		roots := x509.NewCertPool()
		roots.AddCert(servicedns.cert)
		conn, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			RootCAs: roots, ServerName: providerName, Certificates: []tls.Certificate{client}, MinVersion: tls.VersionTLS13,
		})))
		require.NoError(t, err)
		defer conn.Close()
		_, err = credproviderpb.NewCredentialProviderClient(conn).FetchSecret(t.Context(), &credproviderpb.FetchSecretRequest{Uri: testURI, ActorSpiffeId: testActor})
		return err
	}

	assert.NoError(t, call(podidentity.issue(t, "", gatewayIdentity)), "the egress gateway is served")
	assert.Error(t, call(podidentity.issue(t, "", "spiffe://cluster.local/ns/ate-system/sa/other")), "another workload is refused")
	assert.Error(t, call(newAuthority(t).issue(t, "", gatewayIdentity)), "the gateway's identity from another CA is refused")
	assert.EqualValues(t, 1, g.calls.Load(), "only the gateway's call reached gavel")
}

func TestServerCredentialsNeedTheirFiles(t *testing.T) {
	_, err := ServerCredentials("", "/ca.pem", gatewayIdentity)
	assert.EqualError(t, err, "--server-cred-bundle is required: the serving certificate and key")
	_, err = ServerCredentials(filepath.Join(t.TempDir(), "missing.pem"), "/ca.pem", gatewayIdentity)
	assert.ErrorContains(t, err, "read the serving credential bundle")
	_, err = ServerCredentials("/server.pem", "", gatewayIdentity)
	assert.EqualError(t, err, "--client-ca-file is required: the trust bundle the egress gateway's certificate chains to")
	_, err = ServerCredentials("/server.pem", "/ca.pem", "")
	assert.EqualError(t, err, "--injector-identity is required: the egress gateway's SPIFFE ID")
}
