package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DBMode labels the two supported github-cache backends.
const (
	DBModeDSN      = "dsn"
	DBModeEmbedded = "embedded"
)

// DBConfig is persisted at ~/.config/gavel/db.json by `gavel system install`
// and read at startup wherever the github cache is opened. It's the single
// source of truth for "which postgres should the cache use".
type DBConfig struct {
	Mode string `json:"mode"`          // "dsn" | "embedded"
	DSN  string `json:"dsn,omitempty"` // only when Mode == DBModeDSN
}

const dbConfigFile = "db.json"

// DBConfigPath returns the absolute path of the persisted db config.
func DBConfigPath() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, dbConfigFile), nil
}

// EmbeddedDataDir returns the directory where the embedded postgres keeps
// its cluster + binaries when Mode == DBModeEmbedded.
func EmbeddedDataDir() (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "embedded-pg"), nil
}

// LoadDBConfig reads db.json. A missing file yields a zero-value DBConfig
// (no error) so the cache opener can fall back to its own defaults cleanly.
func LoadDBConfig() (DBConfig, error) {
	path, err := DBConfigPath()
	if err != nil {
		return DBConfig{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return DBConfig{}, nil
		}
		return DBConfig{}, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg DBConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return DBConfig{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

// SaveDBConfig persists cfg to db.json with 0600 perms — the DSN can contain
// credentials, so the file is user-only.
func SaveDBConfig(cfg DBConfig) error {
	switch cfg.Mode {
	case DBModeDSN:
		if cfg.DSN == "" {
			return errors.New("mode=dsn requires a non-empty DSN")
		}
	case DBModeEmbedded:
		// nothing to validate
	default:
		return fmt.Errorf("invalid mode %q (must be %q or %q)", cfg.Mode, DBModeDSN, DBModeEmbedded)
	}
	path, err := DBConfigPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal db config: %w", err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// MaskDSN hides the password in a postgres DSN so it can be surfaced in the
// UI or CLI without leaking credentials. Handles URL form
// (postgres://u:p@h/d) and key-value form (host=... password=...).
func MaskDSN(dsn string) string {
	if strings.Contains(dsn, "://") {
		if at := strings.LastIndex(dsn, "@"); at > 0 {
			if colon := strings.Index(dsn[:at], "://"); colon >= 0 {
				schemeEnd := colon + 3
				creds := dsn[schemeEnd:at]
				if user, _, ok := strings.Cut(creds, ":"); ok {
					return dsn[:schemeEnd] + user + ":REDACTED" + dsn[at:]
				}
			}
		}
		return dsn
	}
	parts := strings.Fields(dsn)
	for i, p := range parts {
		if strings.HasPrefix(p, "password=") {
			parts[i] = "password=REDACTED"
		}
	}
	return strings.Join(parts, " ")
}
