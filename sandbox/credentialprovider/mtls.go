// Mutual TLS for the credential provider: it serves substrate's servicedns certificate and admits
// only a caller chaining to the podidentity trust bundle with the egress gateway's SPIFFE ID.
package credentialprovider

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

// ServerCredentials serves serverBundle, a PKCS#8 key and certificate chain in one PEM file, and
// requires a client certificate that chains to clientCAFile and carries callerIdentity as a URI
// SAN. Both files are read per connection, so rotated ones are used without a restart.
func ServerCredentials(serverBundle, clientCAFile, callerIdentity string) (credentials.TransportCredentials, error) {
	switch {
	case serverBundle == "":
		return nil, errors.New("--server-cred-bundle is required: the serving certificate and key")
	case clientCAFile == "":
		return nil, errors.New("--client-ca-file is required: the trust bundle the egress gateway's certificate chains to")
	case callerIdentity == "":
		return nil, errors.New("--injector-identity is required: the egress gateway's SPIFFE ID")
	}
	if _, err := loadBundle(serverBundle); err != nil {
		return nil, err
	}
	if _, err := loadPool(clientCAFile); err != nil {
		return nil, err
	}
	return credentials.NewTLS(&tls.Config{
		MinVersion: tls.VersionTLS13,
		GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
			certificate, err := loadBundle(serverBundle)
			if err != nil {
				return nil, err
			}
			pool, err := loadPool(clientCAFile)
			if err != nil {
				return nil, err
			}
			return &tls.Config{
				MinVersion:       tls.VersionTLS13,
				Certificates:     []tls.Certificate{certificate},
				ClientAuth:       tls.RequireAndVerifyClientCert,
				ClientCAs:        pool,
				VerifyConnection: requireIdentity(callerIdentity),
			}, nil
		},
	}), nil
}

func loadBundle(path string) (tls.Certificate, error) {
	bundle, err := os.ReadFile(path)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("read the serving credential bundle: %w", err)
	}
	certificate, err := tls.X509KeyPair(bundle, bundle)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("the serving credential bundle %s must hold a key and its certificate: %w", path, err)
	}
	return certificate, nil
}

func loadPool(path string) (*x509.CertPool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read the client trust bundle: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(content) {
		return nil, fmt.Errorf("the client trust bundle %s holds no certificate", path)
	}
	return pool, nil
}

// requireIdentity admits a verified client certificate only if it carries identity as a URI SAN.
func requireIdentity(identity string) func(tls.ConnectionState) error {
	return func(state tls.ConnectionState) error {
		if len(state.PeerCertificates) == 0 {
			return errors.New("a client certificate is required")
		}
		for _, uri := range state.PeerCertificates[0].URIs {
			if uri.String() == identity {
				return nil
			}
		}
		return fmt.Errorf("the client certificate's URI SANs %v don't include %s", state.PeerCertificates[0].URIs, identity)
	}
}
