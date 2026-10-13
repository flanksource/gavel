package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CredentialPath is separate from service/authconfig.go's GitHub credentials.
func CredentialPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gavel", "clerk-session.json"), nil
}

// LoadSession reads the explicitly private file-backed credential store.
func LoadSession(path string) (*Session, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("not logged in; run gavel login")
	}
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return nil, errors.New("credential file must be a regular file accessible only to its owner (chmod 600)")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, errors.New("invalid credential file; run gavel login again")
	}
	if session.Token == nil || session.Token.AccessToken == "" || session.IDToken == "" {
		return nil, errors.New("incomplete credentials; run gavel login again")
	}
	return &session, nil
}

// SaveSession atomically replaces credentials, keeping tokens out of repo configuration.
// This initial store uses an owner-only file; OS keychain integration can replace it later.
func SaveSession(path string, session *Session) error {
	if session == nil || session.Token == nil || session.Token.AccessToken == "" || session.IDToken == "" {
		return errors.New("refusing to save incomplete credentials")
	}
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".clerk-session-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return fmt.Errorf("persist credentials: %w", err)
	}
	return nil
}

// Logout removes local credentials; it does not revoke the Clerk browser session.
func Logout(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
