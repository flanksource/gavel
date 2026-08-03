package record

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// caValidity bounds the generated authority. A fixture run is minutes, so an
// hour is generous; the point of the short window is that a certificate which
// escapes the run — copied out of .gavel/recordings, pasted into a bug report —
// stops being usable almost immediately.
const caValidity = time.Hour

// caSkew backdates the certificate so a child whose clock runs slightly behind
// the runner's does not reject a CA that was valid the moment it was issued.
const caSkew = 5 * time.Minute

// CA is the ephemeral authority a mitm recorder signs its per-host certificates
// with. It is generated per recorder and never written to a system trust store:
// the private key lives only in this process, and only the certificate is
// materialised on disk for children to trust.
type CA struct {
	cert    tls.Certificate
	certPEM []byte
}

// newCA generates an ECDSA P-256 authority. ECDSA rather than RSA because
// goproxy derives each host key from the CA key deterministically and a 2048-bit
// RSA generation per host is slow enough to show up as fixture latency.
func newCA() (*CA, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("record: generate ca key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("record: generate ca serial: %w", err)
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"Gavel"},
			CommonName:   "Gavel fixture recorder (ephemeral)",
		},
		NotBefore:             now.Add(-caSkew),
		NotAfter:              now.Add(caValidity),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, key.Public(), key)
	if err != nil {
		return nil, fmt.Errorf("record: create ca certificate: %w", err)
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, fmt.Errorf("record: parse ca certificate: %w", err)
	}

	return &CA{
		// Leaf is set because goproxy's signer reparses the DER on every host
		// otherwise, once per connection.
		cert:    tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf},
		certPEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
	}, nil
}

// trustEnv is the environment that points a child's TLS stack at path. There is
// no single variable that works everywhere, hence the list — and no guarantee
// that a given binary reads any of them, which is what `requireEntries:` is for.
//
// SSL_CERT_FILE and CURL_CA_BUNDLE *replace* the system roots rather than adding
// to them. Under mitm that is intended: every certificate the child sees comes
// from this CA, because every connection it makes goes through the proxy.
func trustEnv(path string) map[string]string {
	return map[string]string{
		"SSL_CERT_FILE":       path, // OpenSSL, Go, Python's ssl
		"CURL_CA_BUNDLE":      path, // curl built against OpenSSL/GnuTLS
		"NODE_EXTRA_CA_CERTS": path, // Node — additive, unlike the others
		"REQUESTS_CA_BUNDLE":  path, // python-requests
		"GIT_SSL_CAINFO":      path,
		"AWS_CA_BUNDLE":       path,
	}
}
