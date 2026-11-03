package uir

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/verify"
	"gopkg.in/yaml.v3"
)

func (u *UIR) writeFragment(ctx context.Context) (string, error) {
	config, err := verify.LoadGavelConfig(verify.RuntimeConfigDirectory(u.WorkDir, u.ConfigDir))
	if err != nil {
		return "", err
	}
	native := maps.Clone(config.Lint.Linters[u.Name()].Config)
	if len(native) == 0 {
		return "", nil
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Minute)
	}
	native["_expires"] = deadline.UTC().Format(time.RFC3339Nano)
	data, err := yaml.Marshal(native)
	if err != nil {
		return "", fmt.Errorf("encode UIR configuration: %w", err)
	}
	directory := filepath.Join(u.WorkDir, ".uir.d")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", err
	}
	file, err := os.CreateTemp(directory, "gavel-*.part")
	if err != nil {
		return "", err
	}
	if _, err := file.Write(data); err != nil {
		return "", errors.Join(err, file.Close(), removeFragment(file.Name()))
	}
	if err := file.Close(); err != nil {
		return "", errors.Join(err, removeFragment(file.Name()))
	}
	path := strings.TrimSuffix(file.Name(), ".part") + ".yaml"
	if err := os.Rename(file.Name(), path); err != nil {
		return "", errors.Join(err, removeFragment(file.Name()))
	}
	return path, nil
}

func removeFragment(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove generated UIR fragment %s: %w", path, err)
	}
	return nil
}

func (u *UIR) HasDirectConfig(workDir string) bool {
	if _, err := os.Stat(filepath.Join(workDir, ".uir.yaml")); !os.IsNotExist(err) {
		return true
	}
	entries, err := os.ReadDir(filepath.Join(workDir, ".uir.d"))
	if err != nil {
		return !os.IsNotExist(err)
	}
	for _, entry := range entries {
		extension := filepath.Ext(entry.Name())
		if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") && (extension == ".yaml" || extension == ".yml") {
			return true
		}
	}
	return false
}

func (u *UIR) HasDefaultActivation(workDir string) bool {
	config, err := verify.LoadGavelConfig(verify.RuntimeConfigDirectory(workDir, u.ConfigDir))
	if err != nil || len(config.Lint.Linters[u.Name()].Config) > 0 {
		return true
	}
	if home, err := os.UserHomeDir(); err == nil && u.HasDirectConfig(home) {
		return true
	}
	for current := workDir; ; current = filepath.Dir(current) {
		if u.HasDirectConfig(current) {
			return true
		}
		if filepath.Dir(current) == current {
			return false
		}
	}
}
