// Tests Open: local stores from config, the prefix, rejected configs, and the conformance suite
// against S3- and GCS-compatible servers named by GAVEL_BLOB_S3_TEST_ENDPOINT and _GCS_.
package blob

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/commons-db/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenLocal(t *testing.T) {
	ctx := context.Background()
	t.Run("conformance", func(t *testing.T) {
		store, err := Open(ctx, Config{Type: "local", Path: t.TempDir()})
		require.NoError(t, err)
		testStore(t, store)
	})

	t.Run("type defaults to local", func(t *testing.T) {
		dir := t.TempDir()
		store, err := Open(ctx, Config{Path: dir})
		require.NoError(t, err)
		require.NoError(t, store.Write(ctx, "a.json", strings.NewReader("{}")))
		assert.FileExists(t, filepath.Join(dir, "a.json"))
	})

	t.Run("prefix", func(t *testing.T) {
		dir := t.TempDir()
		store, err := Open(ctx, Config{Path: dir, Prefix: "gavel"})
		require.NoError(t, err)
		require.NoError(t, store.Write(ctx, "a.json", strings.NewReader("{}")))
		assert.FileExists(t, filepath.Join(dir, "gavel", "a.json"), "a prefix without a trailing slash still names a directory")
	})
}

func TestOpenRejectsIncompleteConfigs(t *testing.T) {
	for name, tc := range map[string]struct {
		config Config
		want   string
	}{
		"local without a path": {Config{Type: "local"}, "storage.path is required for local storage"},
		"unknown type":         {Config{Type: "azure", Bucket: "b"}, `storage.type "azure" is not supported; use local, s3 or gcs`},
		"s3 without a bucket":  {Config{Type: "s3"}, "storage.bucket is required for s3 storage"},
		"gcs without a bucket": {Config{Type: "gcs", GCS: GCSConfig{Credentials: types.EnvVar{ValueStatic: "{}"}}}, "storage.bucket is required for gcs storage"},
		"gcs without credentials": {Config{Type: "gcs", Bucket: "b"},
			"storage.gcs.credentials is required: gavel's GCS bucket is private, and without credentials the client connects anonymously"},
		"invalid prefix": {Config{Path: "/tmp/x", Prefix: "../up"}, `storage.prefix "../up": invalid blob key: key "../up" has a ".." segment`},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Open(context.Background(), tc.config)
			require.EqualError(t, err, tc.want)
		})
	}
}

func TestConfigIsZero(t *testing.T) {
	assert.True(t, Config{}.IsZero())
	assert.False(t, Config{Path: "/x"}.IsZero())
	assert.False(t, Config{Type: "s3", Bucket: "b"}.IsZero())
}

// uniquePrefix keeps conformance runs on a shared bucket apart, so each starts empty.
func uniquePrefix() string {
	return fmt.Sprintf("conformance-%d/", time.Now().UnixNano())
}

func TestOpenS3Conformance(t *testing.T) {
	endpoint := os.Getenv("GAVEL_BLOB_S3_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("set GAVEL_BLOB_S3_TEST_ENDPOINT to an S3-compatible server to run the S3 conformance suite")
	}
	bucket := os.Getenv("GAVEL_BLOB_S3_TEST_BUCKET")
	if bucket == "" {
		bucket = "gavel-test"
	}
	store, err := Open(context.Background(), Config{
		Type: "s3", Bucket: bucket, Prefix: uniquePrefix(),
		S3: S3Config{
			Region: "us-east-1", Endpoint: endpoint, UsePathStyle: true,
			AccessKey: types.EnvVar{ValueStatic: "test"}, SecretKey: types.EnvVar{ValueStatic: "test"},
		},
	})
	require.NoError(t, err)
	testStore(t, store)
}

func TestOpenGCSConformance(t *testing.T) {
	endpoint := os.Getenv("GAVEL_BLOB_GCS_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("set GAVEL_BLOB_GCS_TEST_ENDPOINT to a GCS-compatible server (e.g. http://127.0.0.1:4443/storage/v1/) to run the GCS conformance suite")
	}
	bucket := os.Getenv("GAVEL_BLOB_GCS_TEST_BUCKET")
	if bucket == "" {
		bucket = "gavel-test"
	}
	tokens := 0
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokens++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`))
	}))
	t.Cleanup(tokenServer.Close)

	store, err := Open(context.Background(), Config{
		Type: "gcs", Bucket: bucket, Prefix: uniquePrefix(),
		GCS: GCSConfig{Endpoint: endpoint, Credentials: types.EnvVar{ValueStatic: serviceAccountKey(t, tokenServer.URL)}},
	})
	require.NoError(t, err)
	testStore(t, store)
	assert.Positive(t, tokens, "the client signs in with the configured service account")
}

// serviceAccountKey is a service-account key file whose token endpoint is tokenURL.
func serviceAccountKey(t *testing.T, tokenURL string) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	data, err := json.Marshal(map[string]string{
		"type":           "service_account",
		"project_id":     "gavel-test",
		"private_key_id": "test",
		"private_key":    string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})),
		"client_email":   "gavel-test@gavel-test.iam.gserviceaccount.com",
		"client_id":      "1",
		"token_uri":      tokenURL,
	})
	require.NoError(t, err)
	return string(data)
}

func TestConfigLocation(t *testing.T) {
	for _, tc := range []struct {
		config Config
		want   string
	}{
		{Config{Path: "/var/lib/gavel"}, "/var/lib/gavel/runs/a.json"},
		{Config{Path: "/var/lib/gavel", Prefix: "gavel"}, "/var/lib/gavel/gavel/runs/a.json"},
		{Config{Type: "s3", Bucket: "blobs", Prefix: "gavel/"}, "s3://blobs/gavel/runs/a.json"},
		{Config{Type: "gcs", Bucket: "blobs"}, "gs://blobs/runs/a.json"},
	} {
		assert.Equal(t, tc.want, tc.config.Location("runs/a.json"))
	}
}
