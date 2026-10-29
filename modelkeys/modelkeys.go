// Package modelkeys stores each organization's model API keys, encrypted by commons-db's secret
// keeper, and describes how a sandbox job's requests carry them: the host, header and env of each kind.
package modelkeys

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	dbcontext "github.com/flanksource/commons-db/context"
	"github.com/flanksource/commons-db/secret"
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

// KeeperConnection is the secret keeper used when GAVEL_SECRET_KEEPER is unset: a local key
// derived from GAVEL_SECRET_KEY, so it can't be mistaken for another value derived from it.
func KeeperConnection(gavelSecret []byte) string {
	mac := hmac.New(sha256.New, gavelSecret)
	mac.Write([]byte("gavel-model-keys"))
	return "local_kms://" + base64.URLEncoding.EncodeToString(mac.Sum(nil))
}

// Store reads and writes org_model_keys, encrypting with commons-db's secret keeper, which
// secret.KMSConnection configures.
type Store struct {
	DB *gorm.DB
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
	ciphertext, err := secret.Encrypt(dbcontext.NewContext(ctx), secret.Sensitive(owner(org, k.Provider)+value))
	if err != nil {
		return fmt.Errorf("encrypt organization %s's %s key: %w", org, k.Provider, err)
	}
	if err := s.DB.WithContext(ctx).Exec(`
		INSERT INTO org_model_keys (org_id, provider, kind, ciphertext, hint, updated_by, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, now())
		ON CONFLICT (org_id, provider) DO UPDATE SET kind = excluded.kind, ciphertext = excluded.ciphertext,
			hint = excluded.hint, updated_by = excluded.updated_by, updated_at = now()`,
		org, k.Provider, k.Name, []byte(ciphertext), value[len(value)-4:], user).Error; err != nil {
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
	var rows []struct{ Ciphertext []byte }
	if err := s.DB.WithContext(ctx).Raw(`SELECT ciphertext FROM org_model_keys WHERE org_id = ? AND provider = ?`, org, provider).
		Scan(&rows).Error; err != nil {
		return "", fmt.Errorf("read organization %s's %s key: %w", org, provider, err)
	}
	if len(rows) == 0 {
		return "", fmt.Errorf("%w: organization %s has no %s key", ErrNotFound, org, provider)
	}
	decrypted, err := secret.Decrypt(dbcontext.NewContext(ctx), secret.Ciphertext(rows[0].Ciphertext))
	if err != nil {
		return "", fmt.Errorf("can't decrypt organization %s's %s key; the secret keeper (GAVEL_SECRET_KEEPER, or GAVEL_SECRET_KEY by default) changed since it was set, so an organization admin must set it again: %w", org, provider, err)
	}
	value, ok := strings.CutPrefix(decrypted.PlainText(), owner(org, provider))
	decrypted.Clear()
	if !ok {
		return "", fmt.Errorf("organization %s's %s key was encrypted for another organization or provider; an organization admin must set it again", org, provider)
	}
	return value, nil
}

// owner is encrypted in front of a key and checked on decrypting, so a ciphertext copied into
// another org's row or another provider's isn't used there.
func owner(org uuid.UUID, provider string) string {
	return org.String() + "/" + provider + "\x00"
}
