// Package sshkeys stores each user's SSH public keys, which identify them to the hosted
// push server. Only strong key types are accepted, and a key belongs to one user.
package sshkeys

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	gossh "golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

// minRSABits is the smallest RSA key accepted.
const minRSABits = 3072

var (
	// ErrUnsupportedKey rejects text that isn't a strong SSH public key.
	ErrUnsupportedKey = errors.New("unsupported SSH key")
	// ErrDuplicate rejects a key some user already added.
	ErrDuplicate = errors.New("this SSH key is already registered")
	// ErrNotFound means the user has no key with that ID.
	ErrNotFound = errors.New("SSH key not found")
)

// Key is a registered key, without its key material.
type Key struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Fingerprint string     `json:"fingerprint"`
	CreatedAt   time.Time  `json:"createdAt"`
	LastUsedAt  *time.Time `json:"lastUsedAt"`
}

// Parse reads one authorized_keys line, accepting ssh-ed25519, ecdsa-sha2-*, and ssh-rsa
// of at least 3072 bits.
func Parse(line string) (gossh.PublicKey, error) {
	key, _, _, _, err := gossh.ParseAuthorizedKey([]byte(strings.TrimSpace(line)))
	if err != nil {
		return nil, fmt.Errorf("%w: not an SSH public key (paste one line of a .pub file)", ErrUnsupportedKey)
	}
	switch kind := key.Type(); {
	case kind == gossh.KeyAlgoED25519, strings.HasPrefix(kind, "ecdsa-sha2-"):
		return key, nil
	case kind == gossh.KeyAlgoRSA:
		public, ok := key.(gossh.CryptoPublicKey).CryptoPublicKey().(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("%w: unreadable ssh-rsa key", ErrUnsupportedKey)
		}
		if bits := public.N.BitLen(); bits < minRSABits {
			return nil, fmt.Errorf("%w: ssh-rsa keys must be at least %d bits, this one is %d", ErrUnsupportedKey, minRSABits, bits)
		}
		return key, nil
	default:
		return nil, fmt.Errorf("%w: %s keys aren't accepted; use ssh-ed25519", ErrUnsupportedKey, kind)
	}
}

// Store reads and writes user_ssh_keys.
type Store struct {
	DB *gorm.DB
}

// Add registers line as one of user's keys under name.
func (s Store) Add(ctx context.Context, user uuid.UUID, name, line string) (Key, error) {
	key, err := Parse(line)
	if err != nil {
		return Key{}, err
	}
	var added []Key
	err = s.DB.WithContext(ctx).Raw(`
		INSERT INTO user_ssh_keys (user_id, name, fingerprint, public_key) VALUES (?, ?, ?, ?)
		ON CONFLICT (fingerprint) DO NOTHING
		RETURNING id, name, fingerprint, created_at, last_used_at`,
		user, strings.TrimSpace(name), gossh.FingerprintSHA256(key), strings.TrimSpace(string(gossh.MarshalAuthorizedKey(key)))).
		Scan(&added).Error
	if err != nil {
		return Key{}, fmt.Errorf("save the SSH key: %w", err)
	}
	if len(added) == 0 {
		return Key{}, fmt.Errorf("%w (%s)", ErrDuplicate, gossh.FingerprintSHA256(key))
	}
	return added[0], nil
}

// List returns user's keys, oldest first.
func (s Store) List(ctx context.Context, user uuid.UUID) ([]Key, error) {
	keys := []Key{}
	if err := s.DB.WithContext(ctx).Raw(`
		SELECT id, name, fingerprint, created_at, last_used_at FROM user_ssh_keys WHERE user_id = ? ORDER BY created_at, id`, user).
		Scan(&keys).Error; err != nil {
		return nil, fmt.Errorf("read the SSH keys: %w", err)
	}
	return keys, nil
}

// Delete removes one of user's keys; another user's key is ErrNotFound.
func (s Store) Delete(ctx context.Context, user, id uuid.UUID) error {
	result := s.DB.WithContext(ctx).Exec(`DELETE FROM user_ssh_keys WHERE id = ? AND user_id = ?`, id, user)
	if result.Error != nil {
		return fmt.Errorf("delete the SSH key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// UserForKey returns the user who registered key, recording that the key was used.
func (s Store) UserForKey(ctx context.Context, key gossh.PublicKey) (uuid.UUID, bool, error) {
	var owners []struct{ UserID uuid.UUID }
	if err := s.DB.WithContext(ctx).Raw(`
		UPDATE user_ssh_keys SET last_used_at = now() WHERE fingerprint = ? RETURNING user_id`,
		gossh.FingerprintSHA256(key)).Scan(&owners).Error; err != nil {
		return uuid.Nil, false, fmt.Errorf("look up the SSH key: %w", err)
	}
	if len(owners) == 0 {
		return uuid.Nil, false, nil
	}
	return owners[0].UserID, true, nil
}
