// Package modelkeys stores each organization's model API keys, encrypted, and describes how a
// sandbox job's requests carry them: the host, header and environment variable of each kind.
package modelkeys

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// maxKeyBytes bounds a key; real ones are a few hundred bytes at most.
const maxKeyBytes = 4096

var (
	// ErrInvalidKey rejects a kind or key the egress gateway couldn't inject.
	ErrInvalidKey = errors.New("invalid model key")
	// ErrNotFound means the org has no key for that provider.
	ErrNotFound = errors.New("model key not found")
)

// Kind is a kind of model key: which provider it is for, and how a request carries it. The
// egress gateway puts Prefix and the key into Header on requests to Host; the job's client reads
// a placeholder from Env, so it sends the header.
type Kind struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Label    string `json:"label"`
	Host     string `json:"-"`
	Header   string `json:"-"`
	Prefix   string `json:"-"`
	Env      string `json:"-"`
	// KeyPrefix is how every key of this kind starts.
	KeyPrefix string `json:"-"`
}

// Kinds are the keys an org can set. An org has at most one key per provider.
var Kinds = []Kind{
	{Name: "anthropic-api-key", Provider: "anthropic", Label: "Anthropic API key",
		Host: "api.anthropic.com", Header: "x-api-key", Env: "ANTHROPIC_API_KEY", KeyPrefix: "sk-ant-api"},
	{Name: "claude-code-oauth-token", Provider: "anthropic", Label: "Claude Code OAuth token",
		Host: "api.anthropic.com", Header: "Authorization", Prefix: "Bearer ", Env: "CLAUDE_CODE_OAUTH_TOKEN", KeyPrefix: "sk-ant-oat"},
}

// KindNamed returns the kind called name.
func KindNamed(name string) (Kind, bool) {
	for _, kind := range Kinds {
		if kind.Name == name {
			return kind, true
		}
	}
	return Kind{}, false
}

// Summary describes a stored key without its value.
type Summary struct {
	Provider  string     `json:"provider"`
	Kind      string     `json:"kind"`
	Hint      string     `json:"hint"`
	UpdatedAt time.Time  `json:"updatedAt"`
	UpdatedBy *uuid.UUID `json:"updatedBy"`
}

// EncryptionKey derives the key model keys are encrypted with from GAVEL_SECRET_KEY, so it can't
// be mistaken for another value derived from the secret.
func EncryptionKey(secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte("gavel-model-keys"))
	return mac.Sum(nil)
}

// Store reads and writes org_model_keys. Key is the EncryptionKey.
type Store struct {
	DB  *gorm.DB
	Key []byte
}

// Set validates value as a key of kind and stores it encrypted as org's key for the kind's
// provider, replacing any key the org had for it.
func (s Store) Set(ctx context.Context, org uuid.UUID, user *uuid.UUID, kind, value string) error {
	k, ok := KindNamed(kind)
	if !ok {
		names := make([]string, len(Kinds))
		for i, known := range Kinds {
			names[i] = known.Name
		}
		return fmt.Errorf("%w: unknown kind %q; kinds are %s", ErrInvalidKey, kind, strings.Join(names, ", "))
	}
	value = strings.TrimSpace(value)
	if err := validate(k, value); err != nil {
		return err
	}
	aead, err := s.aead()
	if err != nil {
		return err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("generate a nonce for the model key: %w", err)
	}
	ciphertext := aead.Seal(nil, nonce, []byte(value), additionalData(org, k.Provider))
	if err := s.DB.WithContext(ctx).Exec(`
		INSERT INTO org_model_keys (org_id, provider, kind, ciphertext, nonce, hint, updated_by, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, now())
		ON CONFLICT (org_id, provider) DO UPDATE SET kind = excluded.kind, ciphertext = excluded.ciphertext,
			nonce = excluded.nonce, hint = excluded.hint, updated_by = excluded.updated_by, updated_at = now()`,
		org, k.Provider, k.Name, ciphertext, nonce, value[len(value)-4:], user).Error; err != nil {
		return fmt.Errorf("save organization %s's %s key: %w", org, k.Provider, err)
	}
	return nil
}

// validate refuses a value the gateway couldn't put into a header, or one that isn't of kind.
func validate(kind Kind, value string) error {
	switch {
	case value == "":
		return fmt.Errorf("%w: the key is empty", ErrInvalidKey)
	case len(value) > maxKeyBytes:
		return fmt.Errorf("%w: the key is longer than %d bytes", ErrInvalidKey, maxKeyBytes)
	case strings.IndexFunc(value, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0:
		return fmt.Errorf("%w: the key contains whitespace or a control character", ErrInvalidKey)
	case !strings.HasPrefix(value, kind.KeyPrefix):
		return fmt.Errorf("%w: %s %s starts with %q", ErrInvalidKey, article(kind.Label), kind.Label, kind.KeyPrefix)
	}
	return nil
}

func article(label string) string {
	if strings.ContainsRune("AEIOU", rune(label[0])) {
		return "an"
	}
	return "a"
}

// Delete removes org's key for provider.
func (s Store) Delete(ctx context.Context, org uuid.UUID, provider string) error {
	result := s.DB.WithContext(ctx).Exec(`DELETE FROM org_model_keys WHERE org_id = ? AND provider = ?`, org, provider)
	if result.Error != nil {
		return fmt.Errorf("delete organization %s's %s key: %w", org, provider, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: organization %s has no %s key", ErrNotFound, org, provider)
	}
	return nil
}

// List returns org's keys by provider, without their values.
func (s Store) List(ctx context.Context, org uuid.UUID) ([]Summary, error) {
	keys := []Summary{}
	if err := s.DB.WithContext(ctx).Raw(`
		SELECT provider, kind, hint, updated_at, updated_by FROM org_model_keys WHERE org_id = ? ORDER BY provider`, org).
		Scan(&keys).Error; err != nil {
		return nil, fmt.Errorf("read organization %s's model keys: %w", org, err)
	}
	return keys, nil
}

// Value decrypts org's key for provider.
func (s Store) Value(ctx context.Context, org uuid.UUID, provider string) (string, error) {
	var rows []struct{ Ciphertext, Nonce []byte }
	if err := s.DB.WithContext(ctx).Raw(`SELECT ciphertext, nonce FROM org_model_keys WHERE org_id = ? AND provider = ?`, org, provider).
		Scan(&rows).Error; err != nil {
		return "", fmt.Errorf("read organization %s's %s key: %w", org, provider, err)
	}
	if len(rows) == 0 {
		return "", fmt.Errorf("%w: organization %s has no %s key", ErrNotFound, org, provider)
	}
	aead, err := s.aead()
	if err != nil {
		return "", err
	}
	value, err := aead.Open(nil, rows[0].Nonce, rows[0].Ciphertext, additionalData(org, provider))
	if err != nil {
		return "", fmt.Errorf("can't decrypt organization %s's %s key: GAVEL_SECRET_KEY changed since it was set, or the row was altered; an organization admin must set it again", org, provider)
	}
	return string(value), nil
}

func (s Store) aead() (cipher.AEAD, error) {
	block, err := aes.NewCipher(s.Key)
	if err != nil {
		return nil, fmt.Errorf("model key encryption needs a 32-byte key from EncryptionKey: %w", err)
	}
	return cipher.NewGCM(block)
}

// additionalData binds a ciphertext to its org and provider, so it can't be opened in another
// org's row or another provider's.
func additionalData(org uuid.UUID, provider string) []byte {
	return []byte(org.String() + "/" + provider)
}
