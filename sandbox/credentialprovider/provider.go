// Package credentialprovider serves substrate's CredentialProvider gRPC API for gavel: it resolves
// ate-secret://gavel/<gavel>/<run>/<provider> by asking that gavel for the run's organization's key.
package credentialprovider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/sandbox/internal/substrate/credproviderpb"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/yaml"
)

// Name is the provider name substrate's egress gateway knows this provider by: the host of the
// URIs it serves.
const Name = "gavel"

// Gavel is a gavel the provider asks for keys: its URL and the file holding its token.
type Gavel struct {
	URL       string `json:"url"`
	TokenFile string `json:"tokenFile"`
}

// Config configures a Server.
type Config struct {
	// Gavels are the gavels by the name their URIs use.
	Gavels map[string]Gavel
	// FetchTimeout bounds one ask of a gavel; keep it under the gateway's 5s ext_proc timeout.
	FetchTimeout time.Duration
	// CacheTTL is how long a key gavel gave is reused for the same URI and actor.
	CacheTTL time.Duration
	HTTP     *http.Client
}

// Server answers FetchSecret.
type Server struct {
	credproviderpb.UnimplementedCredentialProviderServer
	config Config

	mu    sync.Mutex
	cache map[cacheKey]cached
}

type cacheKey struct{ uri, actor string }

type cached struct {
	value   []byte
	expires time.Time
}

// Register serves s as the CredentialProvider service of server.
func (s *Server) Register(server *grpc.Server) {
	credproviderpb.RegisterCredentialProviderServer(server, s)
}

// New returns a Server for config.
func New(config Config) *Server {
	if config.HTTP == nil {
		config.HTTP = &http.Client{}
	}
	return &Server{config: config, cache: map[cacheKey]cached{}}
}

var (
	dnsLabel        = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
	providerSegment = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
)

// LoadGavels reads the gavels file: each gavel's name with its url and tokenFile.
func LoadGavels(path string) (map[string]Gavel, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read the gavels file: %w", err)
	}
	gavels := map[string]Gavel{}
	if err := yaml.UnmarshalStrict(content, &gavels); err != nil {
		return nil, fmt.Errorf("parse the gavels file %s: %w", path, err)
	}
	if len(gavels) == 0 {
		return nil, fmt.Errorf("the gavels file %s names no gavel", path)
	}
	for name, gavel := range gavels {
		switch {
		case !dnsLabel.MatchString(name):
			return nil, fmt.Errorf("gavel name %q must be a DNS label, as in its URIs", name)
		case gavel.URL == "" || gavel.TokenFile == "":
			return nil, fmt.Errorf("gavel %q needs url and tokenFile", name)
		}
		if parsed, err := url.Parse(gavel.URL); err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return nil, fmt.Errorf("gavel %q url %q must be http or https", name, gavel.URL)
		}
	}
	return gavels, nil
}

// credentialURI is a parsed ate-secret://gavel/<gavel>/<run>/<provider>.
type credentialURI struct {
	gavel    string
	run      uuid.UUID
	provider string
}

const uriFormat = "ate-secret://gavel/<gavel>/<run id>/<provider>"

func parseURI(raw string) (credentialURI, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "ate-secret" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.EscapedPath() != u.Path {
		return credentialURI{}, fmt.Errorf("%q must be %s", raw, uriFormat)
	}
	if u.Host != Name {
		return credentialURI{}, fmt.Errorf("%q names provider %q, not %s", raw, u.Host, Name)
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) != 3 || !dnsLabel.MatchString(parts[0]) || !providerSegment.MatchString(parts[2]) {
		return credentialURI{}, fmt.Errorf("%q must be %s", raw, uriFormat)
	}
	run, err := uuid.Parse(parts[1])
	if err != nil {
		return credentialURI{}, fmt.Errorf("%q: %q is not a run ID", raw, parts[1])
	}
	return credentialURI{gavel: parts[0], run: run, provider: parts[2]}, nil
}

// FetchSecret resolves req's URI to the key of the run's organization, asking the gavel it names
// on behalf of req's actor.
func (s *Server) FetchSecret(ctx context.Context, req *credproviderpb.FetchSecretRequest) (*credproviderpb.FetchSecretResponse, error) {
	uri, err := parseURI(req.GetUri())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	gavel, ok := s.config.Gavels[uri.gavel]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "%q: no gavel named %q is configured", req.GetUri(), uri.gavel)
	}
	key := cacheKey{uri: req.GetUri(), actor: req.GetActorSpiffeId()}
	if value, ok := s.cached(key); ok {
		return &credproviderpb.FetchSecretResponse{OpaqueBytes: value}, nil
	}
	value, err := s.ask(ctx, gavel, uri, req.GetActorSpiffeId())
	if err != nil {
		logger.Warnf("credential fetch %s for %s: %v", req.GetUri(), req.GetActorSpiffeId(), err)
		return nil, err
	}
	s.store(key, value)
	return &credproviderpb.FetchSecretResponse{OpaqueBytes: value}, nil
}

// ask posts the fetch to gavel and maps its answer to the codes the gateway acts on: Unavailable
// and DeadlineExceeded are retried, every other code refuses the request.
func (s *Server) ask(ctx context.Context, gavel Gavel, uri credentialURI, actor string) ([]byte, error) {
	token, err := os.ReadFile(gavel.TokenFile)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "read gavel %s's token: %v", uri.gavel, err)
	}
	body, _ := json.Marshal(map[string]string{"run": uri.run.String(), "provider": uri.provider, "actor": actor})
	if s.config.FetchTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.config.FetchTimeout)
		defer cancel()
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(gavel.URL, "/")+"/api/credentials/fetch", bytes.NewReader(body))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "build the request to gavel %s: %v", uri.gavel, err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))
	response, err := s.config.HTTP.Do(request)
	if errors.Is(err, context.DeadlineExceeded) {
		return nil, status.Errorf(codes.DeadlineExceeded, "gavel %s didn't answer within %s", uri.gavel, s.config.FetchTimeout)
	}
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "ask gavel %s: %v", uri.gavel, err)
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(response.Body, 64<<10))
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "read gavel %s's answer: %v", uri.gavel, err)
	}
	var answer struct {
		Value string `json:"value"`
		Error string `json:"error"`
	}
	_ = json.Unmarshal(payload, &answer)
	switch code := response.StatusCode; {
	case code == http.StatusOK && answer.Value != "":
		return []byte(answer.Value), nil
	case code == http.StatusOK:
		return nil, status.Errorf(codes.Unavailable, "gavel %s answered without a key", uri.gavel)
	case code == http.StatusUnauthorized:
		return nil, status.Errorf(codes.PermissionDenied, "gavel %s refused this provider's token, so its tokenFile is wrong: %s", uri.gavel, answer.Error)
	case code == http.StatusForbidden:
		return nil, status.Errorf(codes.PermissionDenied, "gavel %s: %s", uri.gavel, answer.Error)
	case code == http.StatusNotFound:
		return nil, status.Errorf(codes.NotFound, "gavel %s: %s", uri.gavel, answer.Error)
	default:
		return nil, status.Errorf(codes.Unavailable, "gavel %s answered %d: %s", uri.gavel, code, answer.Error)
	}
}

func (s *Server) cached(key cacheKey) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.cache[key]
	if !ok || time.Now().After(entry.expires) {
		return nil, false
	}
	return entry.value, true
}

// store keeps value for the TTL and drops expired entries, so the cache holds only live runs.
func (s *Server) store(key cacheKey, value []byte) {
	if s.config.CacheTTL <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, entry := range s.cache {
		if now.After(entry.expires) {
			delete(s.cache, k)
		}
	}
	s.cache[key] = cached{value: value, expires: now.Add(s.config.CacheTTL)}
}
