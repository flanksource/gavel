package preview

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/safefs"
)

type Manifest struct {
	Name             string            `json:"name"`
	Version          string            `json:"version"`
	Type             string            `json:"type,omitempty"`
	Main             string            `json:"main,omitempty"`
	Module           string            `json:"module,omitempty"`
	Exports          json.RawMessage   `json:"exports,omitempty"`
	SideEffects      json.RawMessage   `json:"sideEffects,omitempty"`
	Dependencies     map[string]string `json:"dependencies,omitempty"`
	PeerDependencies map[string]string `json:"peerDependencies"`
	Dist             struct {
		Tarball   string `json:"tarball"`
		Integrity string `json:"integrity"`
	} `json:"dist"`
}
type Metadata struct {
	Versions map[string]Manifest `json:"versions"`
	Time     map[string]string   `json:"time"`
	Tags     map[string]string   `json:"dist-tags"`
}
type Cache struct {
	Root        string
	Local       map[string]string
	LocalFrozen bool
	Pinned      *NamedRuntime
	mu          sync.Mutex
}

var packageName = regexp.MustCompile(`^(?:@[a-z0-9][a-z0-9._-]*/)?[a-z0-9][a-z0-9._-]*$`)

func download(ctx context.Context, target string) ([]byte, error) {
	u, err := url.Parse(target)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Host != "registry.npmjs.org" {
		return nil, fmt.Errorf("invalid npm URL %s", target)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("npm %s: HTTP %d", target, response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, (256<<20)+1))
	if len(data) > 256<<20 {
		return nil, fmt.Errorf("npm download exceeds size limit")
	}
	return data, err
}

type MetadataRequest struct {
	Name    string
	Refresh bool
}

func (c *Cache) Metadata(ctx context.Context, request MetadataRequest) (Metadata, error) {
	name := request.Name
	if c.Pinned != nil {
		return c.pinnedMetadata(name)
	}
	if root, ok := c.Local[name]; ok {
		data, err := os.ReadFile(filepath.Join(root, "package.json"))
		if err != nil {
			return Metadata{}, err
		}
		var manifest Manifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return Metadata{}, err
		}
		return Metadata{Versions: map[string]Manifest{manifest.Version: manifest}, Time: map[string]string{manifest.Version: "2020-01-01T00:00:00Z"}, Tags: map[string]string{"latest": manifest.Version}}, nil
	}
	if !packageName.MatchString(name) {
		return Metadata{}, fmt.Errorf("invalid package name %q", name)
	}
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(name)))
	path := filepath.Join(c.Root, "metadata", key+".json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) || request.Refresh {
		data, err = download(ctx, "https://registry.npmjs.org/"+url.PathEscape(name))
		if err != nil {
			return Metadata{}, err
		}
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return Metadata{}, err
		}
		temp, err := os.CreateTemp(filepath.Dir(path), ".metadata-")
		if err != nil {
			return Metadata{}, err
		}
		defer os.Remove(temp.Name())
		_, err = temp.Write(data)
		closeErr := temp.Close()
		if err != nil {
			return Metadata{}, err
		}
		if closeErr != nil {
			return Metadata{}, closeErr
		}
		if err = os.Rename(temp.Name(), path); err != nil {
			return Metadata{}, err
		}
	} else if err != nil {
		return Metadata{}, err
	}
	var result Metadata
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *Cache) Resolve(ctx context.Context, name, constraint string) (Manifest, error) {
	metadata, err := c.Metadata(ctx, MetadataRequest{Name: name})
	if err != nil {
		return Manifest{}, err
	}
	if tagged, ok := metadata.Tags[constraint]; ok {
		constraint = tagged
	}
	range_, err := semver.NewConstraint(constraint)
	if err != nil {
		return Manifest{}, err
	}
	var chosen *semver.Version
	for version := range metadata.Versions {
		parsed, err := semver.NewVersion(version)
		if err == nil && range_.Check(parsed) && (chosen == nil || parsed.GreaterThan(chosen)) {
			chosen = parsed
		}
	}
	if chosen == nil {
		return Manifest{}, fmt.Errorf("%s: no version matches %s", name, constraint)
	}
	return metadata.Versions[chosen.String()], nil
}

func (c *Cache) Package(ctx context.Context, name, version string) (string, error) {
	if c.Pinned != nil {
		if root := c.Pinned.Packages[name][version]; root != "" {
			return root, nil
		}
		return "", fmt.Errorf("package %s@%s is absent from Facet runtime %s; add the dependency and run facet runtime build", name, version, c.Pinned.Label())
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	manifest, err := c.Resolve(ctx, name, version)
	if err != nil {
		return "", err
	}
	if local, ok := c.Local[name]; ok {
		if c.LocalFrozen {
			return local, nil
		}
		return c.snapshot(local, manifest)
	}
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(name+"@"+manifest.Version+"/"+manifest.Dist.Integrity)))
	root := filepath.Join(c.Root, "packages", key)
	if _, err := os.Stat(filepath.Join(root, "package.json")); err == nil {
		return root, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	data, err := download(ctx, manifest.Dist.Tarball)
	if err != nil {
		return "", err
	}
	if !validIntegrity(data, manifest.Dist.Integrity) {
		return "", fmt.Errorf("%s@%s: tarball integrity mismatch", name, version)
	}
	if err := os.MkdirAll(filepath.Dir(root), 0700); err != nil {
		return "", err
	}
	staging, err := os.MkdirTemp(filepath.Dir(root), ".package-")
	if err != nil {
		return "", err
	}
	defer func() {
		if err := safefs.SafeRemove(staging, safefs.RemoveContext{AllowedRoot: c.Root}); err != nil {
			logger.Warnf("remove package staging %s: %v", staging, err)
		}
	}()
	if err := extract(data, staging); err != nil {
		return "", err
	}
	if err := os.Rename(staging, root); err != nil {
		if os.IsExist(err) {
			return root, nil
		}
		return "", err
	}
	return root, nil
}

func validIntegrity(data []byte, expected string) bool {
	hash := sha512.Sum512(data)
	return expected == "sha512-"+base64.StdEncoding.EncodeToString(hash[:])
}

func (c *Cache) Cached(manifest Manifest) (bool, error) {
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(manifest.Name+"@"+manifest.Version+"/"+manifest.Dist.Integrity)))
	_, err := os.Stat(filepath.Join(c.Root, "packages", key, "package.json"))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func extract(data []byte, root string) error {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gz.Close()
	archive := tar.NewReader(gz)
	var total int64
	for {
		header, err := archive.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if header.Name == "package/" && header.Typeflag == tar.TypeDir {
			continue
		}
		name, ok := strings.CutPrefix(header.Name, "package/")
		if !ok || !fs.ValidPath(strings.TrimSuffix(name, "/")) || strings.Contains(name, "\\") {
			return fmt.Errorf("invalid npm archive path %s", header.Name)
		}
		target := filepath.Join(root, filepath.FromSlash(name))
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(target, 0700); err != nil {
				return err
			}
			continue
		}
		if header.Typeflag != tar.TypeReg {
			return fmt.Errorf("npm archive contains unsupported entry %s", header.Name)
		}
		total += header.Size
		if total > 1<<30 {
			return fmt.Errorf("npm archive exceeds extraction limit")
		}
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return err
		}
		file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(file, archive)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
}
