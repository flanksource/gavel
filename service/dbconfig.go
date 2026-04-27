package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
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

// RunningEmbeddedPostgres represents a postmaster discovered via
// postmaster.pid in the embedded data directory. Returned by
// FindRunningEmbeddedPostgres so CLI invocations can connect to a postgres
// already started by the system service instead of spinning up a duplicate.
type RunningEmbeddedPostgres struct {
	PID  int
	Port int
}

// embeddedPGUser / embeddedPGPassword / embeddedPGDatabase mirror the values
// commons-db's StartEmbedded uses (see EmbeddedConfig defaults). The github
// cache uses Database="gavel" — keep this in sync with cache.Open.
const (
	embeddedPGUser     = "postgres"
	embeddedPGPassword = "postgres"
	embeddedPGDatabase = "gavel"
)

// posmasterLinePort is the line index in postmaster.pid that holds the
// listening port; line 0 is the pid. Format is stable across postgres
// releases and matches commons-db/db/embedded.go.
const posmasterLinePort = 3

// FindRunningEmbeddedPostgres parses <EmbeddedDataDir>/data/postmaster.pid
// and verifies the postmaster is reachable. Returns (nil, nil) when:
//   - postmaster.pid is missing (no instance ever started, or it shut down
//     cleanly and removed the file)
//   - the recorded pid is not alive (stale pidfile after a crash)
//   - the recorded port is not accepting TCP connections (postgres dying
//     mid-shutdown, file not yet rewritten)
//
// Returns an error only for unexpected I/O / parse failures so callers can
// distinguish "definitely not running" from "couldn't determine".
func FindRunningEmbeddedPostgres() (*RunningEmbeddedPostgres, error) {
	dataDir, err := EmbeddedDataDir()
	if err != nil {
		return nil, err
	}
	return findRunningEmbeddedPostgresIn(dataDir)
}

func findRunningEmbeddedPostgresIn(dataDir string) (*RunningEmbeddedPostgres, error) {
	pidPath := filepath.Join(dataDir, "data", "postmaster.pid")
	raw, err := os.ReadFile(pidPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", pidPath, err)
	}
	lines := strings.Split(string(raw), "\n")
	if len(lines) <= posmasterLinePort {
		return nil, fmt.Errorf("%s has %d lines, need >%d", pidPath, len(lines), posmasterLinePort)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil || pid <= 0 {
		return nil, fmt.Errorf("%s: invalid pid %q: %w", pidPath, lines[0], err)
	}
	port, err := strconv.Atoi(strings.TrimSpace(lines[posmasterLinePort]))
	if err != nil || port <= 0 || port > 65535 {
		return nil, fmt.Errorf("%s: invalid port %q: %w", pidPath, lines[posmasterLinePort], err)
	}
	if !pidAlive(pid) {
		return nil, nil
	}
	if !tcpReachable("localhost", port) {
		return nil, nil
	}
	return &RunningEmbeddedPostgres{PID: pid, Port: port}, nil
}

// EmbeddedDSN returns the DSN string for the embedded postgres at the given
// port. Uses the same user/password/database that commons-db.StartEmbedded
// would have used, so connecting via this DSN is equivalent to the path that
// went through StartEmbedded.
func EmbeddedDSN(port int) string {
	return fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable",
		embeddedPGUser, embeddedPGPassword, port, embeddedPGDatabase)
}

// pidAlive reports whether a process with pid exists. Uses signal 0 which is
// a no-op send — succeeds when the process exists and can receive signals
// from the current user.
func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

// tcpReachable returns true if a short TCP dial to host:port succeeds. Used
// to confirm the postmaster recorded in postmaster.pid is actually serving;
// a stale file pointing at a port nobody's listening on counts as "not
// running" rather than a hard failure.
func tcpReachable(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), 200*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
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
