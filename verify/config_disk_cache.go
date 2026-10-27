package verify

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/workspace"
	"github.com/ghodss/yaml"
	"github.com/gofrs/flock"
)

const configCacheVersion = 1

type configDiskLayer struct {
	Exists bool            `json:"exists"`
	Hash   string          `json:"hash,omitempty"`
	Value  json.RawMessage `json:"value,omitempty"`
}

type configDiskCache struct {
	Version int                        `json:"version"`
	Root    string                     `json:"root"`
	Layers  map[string]configDiskLayer `json:"directories"`
}

func loadConfigLayers(dir string, layers []gavelConfigLayer) ([]GavelConfig, error) {
	dirs, err := workspace.Resolve(dir)
	if err != nil {
		return nil, err
	}
	path := dirs.State("config-cache.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create config cache: %w", err)
	}
	lock := flock.New(path + ".lock")
	if err := lock.Lock(); err != nil {
		return nil, fmt.Errorf("lock config cache: %w", err)
	}
	defer func() {
		if err := lock.Unlock(); err != nil {
			logger.Errorf("unlock config cache: %v", err)
		}
	}()
	cache, err := readConfigDiskCache(path, dirs.Root())
	if err != nil {
		return nil, err
	}
	configs, changed, err := cache.resolve(layers)
	if err != nil {
		return nil, err
	}
	if changed {
		if err := writeConfigDiskCache(path, cache); err != nil {
			return nil, fmt.Errorf("write config cache %s: %w", path, err)
		}
	}
	return configs, nil
}

func readConfigDiskCache(path, root string) (configDiskCache, error) {
	empty := configDiskCache{Version: configCacheVersion, Root: root, Layers: map[string]configDiskLayer{}}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return empty, nil
	}
	if err != nil {
		return empty, fmt.Errorf("read config cache %s: %w", path, err)
	}
	var cache configDiskCache
	if err := json.Unmarshal(data, &cache); err != nil {
		logger.Warnf("Rebuilding invalid config cache %s: %v", path, err)
		return empty, nil
	}
	if cache.Version != configCacheVersion || cache.Root != root || cache.Layers == nil {
		logger.Warnf("Rebuilding incompatible config cache %s", path)
		return empty, nil
	}
	return cache, nil
}

func (c *configDiskCache) resolve(layers []gavelConfigLayer) ([]GavelConfig, bool, error) {
	configs := make([]GavelConfig, 0, len(layers))
	changed := false
	for _, layer := range layers {
		key := filepath.Dir(layer.path)
		entry := configDiskLayer{Exists: layer.exists}
		if layer.exists {
			hash := sha256.Sum256(layer.data)
			entry.Hash = hex.EncodeToString(hash[:])
			if previous := c.Layers[key]; previous.Exists && previous.Hash == entry.Hash && json.Valid(previous.Value) {
				entry.Value = previous.Value
			} else {
				if _, err := parseGavelConfig(layer.path, layer.data); err != nil {
					return nil, false, err
				}
				encoded, err := yaml.YAMLToJSON(layer.data)
				if err != nil {
					return nil, false, fmt.Errorf("parse %s: %w", layer.path, err)
				}
				entry.Value = encoded
			}
			cfg, err := decodeGavelConfig(layer.path, entry.Value)
			if err != nil {
				return nil, false, err
			}
			configs = append(configs, cfg)
		}
		previous, exists := c.Layers[key]
		changed = changed || !exists || previous.Exists != entry.Exists || previous.Hash != entry.Hash || !bytes.Equal(previous.Value, entry.Value)
		c.Layers[key] = entry
	}
	return configs, changed, nil
}

func writeConfigDiskCache(path string, cache configDiskCache) (returnErr error) {
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".config-cache-*")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil && !os.IsNotExist(err) {
			returnErr = fmt.Errorf("remove config cache temporary file: %w", err)
		}
	}()
	_, writeErr := file.Write(data)
	if err := file.Close(); err != nil {
		return err
	}
	if writeErr != nil {
		return writeErr
	}
	return os.Rename(file.Name(), path)
}
