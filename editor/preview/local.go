package preview

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/safefs"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func (c *Cache) snapshot(local string, manifest Manifest) (string, error) {
	data, err := os.ReadFile(filepath.Join(local, "dist/runtime/importmap.json"))
	if err != nil {
		return "", fmt.Errorf("dev %s: build its page runtime first: %w", manifest.Name, err)
	}
	var fragment struct {
		Integrity map[string]string `json:"integrity"`
	}
	if err := json.Unmarshal(data, &fragment); err != nil {
		return "", err
	}
	for name, expected := range fragment.Integrity {
		file, ok := strings.CutPrefix(name, manifest.Name+"@"+manifest.Version+"/")
		if !ok {
			continue
		}
		if !fs.ValidPath(file) {
			return "", fmt.Errorf("invalid dev runtime path %s", file)
		}
		content, err := os.ReadFile(filepath.Join(local, file))
		if err != nil {
			return "", err
		}
		hash := sha512.Sum384(content)
		if expected != "sha384-"+base64.StdEncoding.EncodeToString(hash[:]) {
			return "", fmt.Errorf("dev %s: stale runtime file %s; rebuild its page runtime", manifest.Name, file)
		}
	}
	root := filepath.Join(c.Root, "local", fmt.Sprintf("%x", sha256.Sum256(data)))
	if _, err := os.Stat(filepath.Join(root, "package.json")); err == nil {
		return root, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(root), 0700); err != nil {
		return "", err
	}
	staging, err := os.MkdirTemp(filepath.Dir(root), ".generation-")
	if err != nil {
		return "", err
	}
	defer func() {
		if err := safefs.SafeRemove(staging, safefs.RemoveContext{AllowedRoot: c.Root}); err != nil {
			logger.Warnf("remove runtime staging %s: %v", staging, err)
		}
	}()
	files := []string{"dist/runtime/importmap.json", "dist/runtime/barrels.json"}
	for name := range fragment.Integrity {
		if file, ok := strings.CutPrefix(name, manifest.Name+"@"+manifest.Version+"/"); ok {
			files = append(files, file)
		}
	}
	files = append(files, "package.json")
	for _, name := range files {
		if !fs.ValidPath(name) {
			return "", fmt.Errorf("invalid dev runtime path %s", name)
		}
		data, err := os.ReadFile(filepath.Join(local, name))
		if err != nil {
			return "", err
		}
		if expected, ok := fragment.Integrity[manifest.Name+"@"+manifest.Version+"/"+name]; ok {
			hash := sha512.Sum384(data)
			if expected != "sha384-"+base64.StdEncoding.EncodeToString(hash[:]) {
				return "", fmt.Errorf("dev %s: stale runtime file %s; rebuild its page runtime", manifest.Name, name)
			}
		}
		target := filepath.Join(staging, name)
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return "", err
		}
		if err := os.WriteFile(target, data, 0600); err != nil {
			return "", err
		}
	}
	if err := os.Rename(staging, root); err != nil {
		if os.IsExist(err) {
			return root, nil
		}
		return "", err
	}
	return root, nil
}
