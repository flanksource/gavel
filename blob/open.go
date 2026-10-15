// Opens the Store a `storage:` config section describes: a local directory, an S3-compatible
// bucket or a GCS bucket, optionally under a key prefix.
package blob

import (
	"context"
	"fmt"
	"strings"

	"github.com/flanksource/commons-db/connection"
	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/types"
)

// Config is the `storage:` section of the serve host's ~/.gavel.yaml.
type Config struct {
	// Type is local, s3 or gcs; empty means local.
	Type string `yaml:"type,omitempty" json:"type,omitempty"`
	// Path is the directory local storage writes to.
	Path   string `yaml:"path,omitempty" json:"path,omitempty"`
	Bucket string `yaml:"bucket,omitempty" json:"bucket,omitempty"`
	// Prefix is prepended to every key; a trailing slash is implied.
	Prefix string    `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	S3     S3Config  `yaml:"s3,omitempty" json:"s3,omitempty"`
	GCS    GCSConfig `yaml:"gcs,omitempty" json:"gcs,omitempty"`
}

// S3Config reaches an S3 or S3-compatible bucket. Without keys the AWS default credential chain
// applies.
type S3Config struct {
	Region       string       `yaml:"region,omitempty" json:"region,omitempty"`
	Endpoint     string       `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	UsePathStyle bool         `yaml:"usePathStyle,omitempty" json:"usePathStyle,omitempty"`
	AccessKey    types.EnvVar `yaml:"accessKey,omitempty" json:"accessKey,omitempty"`
	SecretKey    types.EnvVar `yaml:"secretKey,omitempty" json:"secretKey,omitempty"`
}

// GCSConfig reaches a GCS bucket with a service-account key. Endpoint points at an emulator.
type GCSConfig struct {
	Endpoint    string       `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Credentials types.EnvVar `yaml:"credentials,omitempty" json:"credentials,omitempty"`
}

// IsZero reports a config with no storage set, so callers keep their own default.
func (c Config) IsZero() bool {
	return c == Config{}
}

// Open returns the Store c describes.
func Open(ctx context.Context, c Config) (Store, error) {
	prefix := c.Prefix
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	if prefix != "" {
		if err := validateKey(strings.TrimSuffix(prefix, "/")); err != nil {
			return nil, fmt.Errorf("storage.prefix %q: %w", c.Prefix, err)
		}
	}
	store, err := open(ctx, c)
	if err != nil {
		return nil, err
	}
	if prefix != "" {
		store = WithPrefix(store, prefix)
	}
	return store, nil
}

func open(ctx context.Context, c Config) (Store, error) {
	switch c.Type {
	case "", "local":
		if c.Path == "" {
			return nil, fmt.Errorf("storage.path is required for local storage")
		}
		return Local(c.Path), nil
	case "s3":
		if c.Bucket == "" {
			return nil, fmt.Errorf("storage.bucket is required for s3 storage")
		}
		conn := connection.S3Connection{
			AWSConnection: connection.AWSConnection{
				Region: c.S3.Region, Endpoint: c.S3.Endpoint, AccessKey: c.S3.AccessKey, SecretKey: c.S3.SecretKey,
			},
			Bucket:       c.Bucket,
			UsePathStyle: c.S3.UsePathStyle,
		}
		filesystem, err := conn.Filesystem(dbcontext.NewContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("open s3 bucket %s: %w", c.Bucket, err)
		}
		return &fsStore{fs: filesystem}, nil
	case "gcs":
		if c.Bucket == "" {
			return nil, fmt.Errorf("storage.bucket is required for gcs storage")
		}
		if c.GCS.Credentials.IsEmpty() {
			return nil, fmt.Errorf("storage.gcs.credentials is required: gavel's GCS bucket is private, and without credentials the client connects anonymously")
		}
		credentials := c.GCS.Credentials
		conn := connection.GCSConnection{
			GCPConnection: connection.GCPConnection{Endpoint: c.GCS.Endpoint, Credentials: &credentials},
			Bucket:        c.Bucket,
		}
		filesystem, err := conn.Filesystem(dbcontext.NewContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("open gcs bucket %s: %w", c.Bucket, err)
		}
		return &fsStore{fs: filesystem, literalListing: true}, nil
	default:
		return nil, fmt.Errorf("storage.type %q is not supported; use local, s3 or gcs", c.Type)
	}
}
