package preview

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

const facetPackage = "@flanksource/facet"
const clickyPackage = "@flanksource/clicky-ui"

var runtimeName = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)
var runtimeVersion = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9.+_-]*$`)
var runtimeID = regexp.MustCompile(`^[a-f0-9]{8}-(?:[a-f0-9]{4}-){3}[a-f0-9]{12}$`)

type NamedRuntime struct {
	Name           string
	Version        string
	Build          BuildInfo
	Packages       map[string]map[string]string
	Versions       map[string]string
	BrowserCatalog json.RawMessage
	lease          string
}

type BuildInfo struct {
	BuiltAt *time.Time `json:"builtAt,omitempty"`
	Git     *BuildGit  `json:"git,omitempty"`
}

type BuildGit struct {
	Commit string `json:"commit,omitempty"`
	Dirty  bool   `json:"dirty"`
}

func (r *NamedRuntime) Label() string { return r.Name + "@" + r.Version }

func (r *NamedRuntime) Close() error {
	if r.lease == "" {
		return nil
	}
	return os.Remove(r.lease)
}

func runtimeStore(cwd string) (string, error) {
	cache := os.Getenv("FACET_CACHE_DIR")
	if cache == "" {
		xdg := os.Getenv("XDG_CACHE_HOME")
		if xdg == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			xdg = filepath.Join(home, ".cache")
		}
		cache = filepath.Join(xdg, "facet")
	}
	if !filepath.IsAbs(cache) {
		cache = filepath.Join(cwd, cache)
	}
	store := filepath.Join(cache, "named-runtimes")
	canonical, err := filepath.EvalSymlinks(store)
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	return canonical, err
}

func ListNamedRuntimes(cwd string) ([]string, error) {
	store, err := runtimeStore(cwd)
	if err != nil {
		return nil, err
	}
	dirs, err := os.ReadDir(store)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	type installedBuild struct {
		selector string
		builtAt  time.Time
	}
	var builds []installedBuild
	for _, dir := range dirs {
		if !dir.IsDir() || !runtimeName.MatchString(dir.Name()) {
			continue
		}
		entries, err := readRuntimeIndex(filepath.Join(store, dir.Name()))
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			builds = append(builds, installedBuild{selector: dir.Name() + "@" + entry.Version, builtAt: entry.BuiltAt})
		}
	}
	slices.SortFunc(builds, func(a, b installedBuild) int {
		if order := b.builtAt.Compare(a.builtAt); order != 0 {
			return order
		}
		return strings.Compare(a.selector, b.selector)
	})
	selectors := make([]string, 0, len(builds))
	for _, build := range builds {
		selectors = append(selectors, build.selector)
	}
	return selectors, nil
}

func OpenNamedRuntime(ctx context.Context, selector, cwd string) (*NamedRuntime, error) {
	store, err := runtimeStore(cwd)
	if err != nil {
		return nil, err
	}
	if selector == "." || selector == ".." || strings.ContainsAny(selector, `/\`) || filepath.IsAbs(selector) {
		return openRuntimeFolder(ctx, store, cwd, selector)
	}
	name, version, exact := strings.Cut(selector, "@")
	if !runtimeName.MatchString(name) || (exact && !runtimeVersion.MatchString(version)) {
		return nil, fmt.Errorf("invalid Facet runtime name or name@version %q", selector)
	}
	dir := filepath.Join(store, name)
	if err := checkRuntimeDirectory(dir); errors.Is(err, os.ErrNotExist) {
		return nil, runtimeMissing(selector)
	} else if err != nil {
		return nil, err
	}
	unlock, err := lockRuntime(ctx, dir)
	if err != nil {
		return nil, err
	}
	defer unlock()
	return openRuntimeVersion(dir, name, version, selector)
}

func runtimeMissing(selector string) error {
	return fmt.Errorf("Facet runtime %s is not installed; run facet runtime build first", selector)
}

type runtimeIndexEntry struct {
	Version string
	ID      string
	BuiltAt time.Time
}

func readRuntimeIndex(dir string) ([]runtimeIndexEntry, error) {
	path := filepath.Join(dir, "index.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []runtimeIndexEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("read Facet runtime index %s: %w", path, err)
	}
	if entries == nil {
		return nil, fmt.Errorf("invalid Facet runtime index %s", path)
	}
	for _, entry := range entries {
		if !runtimeVersion.MatchString(entry.Version) || !runtimeID.MatchString(entry.ID) || entry.BuiltAt.IsZero() {
			return nil, fmt.Errorf("invalid Facet runtime index %s", path)
		}
	}
	return entries, nil
}

func openRuntimeVersion(dir, name, version, selector string) (*NamedRuntime, error) {
	entries, err := readRuntimeIndex(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if version == "" || version == entry.Version {
			return loadRuntime(filepath.Join(dir, entry.Version, entry.ID), []string{name, entry.Version, entry.ID})
		}
	}
	return nil, runtimeMissing(selector)
}

func openRuntimeFolder(ctx context.Context, store, cwd, selector string) (*NamedRuntime, error) {
	root := selector
	if !filepath.IsAbs(root) {
		root = filepath.Join(cwd, root)
	}
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("open Facet runtime folder %s: %w", selector, err)
	}
	rel, err := filepath.Rel(store, root)
	if err != nil {
		return loadRuntime(root, nil)
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) != 3 || !runtimeName.MatchString(parts[0]) || !runtimeVersion.MatchString(parts[1]) || !runtimeID.MatchString(parts[2]) {
		return loadRuntime(root, nil)
	}
	unlock, err := lockRuntime(ctx, filepath.Join(store, parts[0]))
	if err != nil {
		return nil, err
	}
	defer unlock()
	return loadRuntime(root, parts)
}

func checkRuntimeDirectory(path string) error {
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	if canonical != path {
		return fmt.Errorf("invalid Facet runtime directory %s", path)
	}
	return nil
}

func lockRuntime(ctx context.Context, dir string) (func(), error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	path := filepath.Join(dir, ".lock")
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err == nil {
			return func() { _ = file.Close(); _ = os.Remove(path) }, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for Facet runtime lock %s: %w", path, ctx.Err())
		case <-ticker.C:
		}
	}
}

func loadRuntime(dir string, managed []string) (*NamedRuntime, error) {
	if err := checkRuntimeDirectory(dir); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	data, err := root.ReadFile("facet-runtime.json")
	if err != nil {
		return nil, fmt.Errorf("read built Facet runtime in %s: %w", dir, err)
	}
	var manifest struct {
		BuildInfo
		Schema          int
		Name, Version   string
		ID              string
		BrowserPackages map[string]struct {
			Name, Version, Path string
			Direct              bool
		}
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("decode Facet runtime in %s: %w", dir, err)
	}
	if manifest.Schema != 1 || !runtimeName.MatchString(manifest.Name) || !runtimeVersion.MatchString(manifest.Version) || !runtimeID.MatchString(manifest.ID) {
		return nil, fmt.Errorf("invalid Facet runtime manifest in %s", dir)
	}
	runtime := &NamedRuntime{Name: manifest.Name, Version: manifest.Version, Build: manifest.BuildInfo, Packages: map[string]map[string]string{}, Versions: map[string]string{}}
	for key, entry := range manifest.BrowserPackages {
		if key != entry.Name+"@"+entry.Version || !packageName.MatchString(entry.Name) || !fs.ValidPath(entry.Path) || entry.Path == "." || strings.Contains(entry.Path, `\`) {
			return nil, fmt.Errorf("invalid browser package %s in Facet runtime %s", key, runtime.Label())
		}
		if err := validateRuntimePackage(root.FS(), entry.Name, entry.Version, entry.Path); err != nil {
			return nil, fmt.Errorf("Facet runtime %s: %w", runtime.Label(), err)
		}
		path := filepath.Join(dir, filepath.FromSlash(entry.Path))
		if err := checkRuntimeDirectory(path); err != nil {
			return nil, err
		}
		if runtime.Packages[entry.Name] == nil {
			runtime.Packages[entry.Name] = map[string]string{}
		}
		runtime.Packages[entry.Name][entry.Version] = path
		if entry.Direct {
			if _, exists := runtime.Versions[entry.Name]; exists {
				return nil, fmt.Errorf("Facet runtime %s has more than one direct version of %s", runtime.Label(), entry.Name)
			}
			runtime.Versions[entry.Name] = entry.Version
		}
	}
	if runtime.Versions[facetPackage] == "" || runtime.Versions[clickyPackage] == "" {
		return nil, fmt.Errorf("Facet runtime %s is missing Facet or Clicky UI browser packages; rebuild it", runtime.Label())
	}
	if err := runtime.loadBrowserCatalog(root); err != nil {
		return nil, err
	}
	if len(managed) == 0 {
		return runtime, nil
	}
	if managed[0] != manifest.Name || managed[1] != manifest.Version || managed[2] != manifest.ID {
		return nil, fmt.Errorf("Facet runtime manifest does not match its cache entry in %s", dir)
	}
	return runtime, runtime.createLease(root)
}

func (r *NamedRuntime) loadBrowserCatalog(root *os.Root) error {
	data, err := root.ReadFile("runtime-layers.json")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var layers struct {
		BrowserCatalog json.RawMessage
	}
	if err := json.Unmarshal(data, &layers); err != nil {
		return fmt.Errorf("decode Facet runtime %s layers: %w", r.Label(), err)
	}
	var catalog struct {
		Versions map[string]string
	}
	if err := json.Unmarshal(layers.BrowserCatalog, &catalog); err != nil || catalog.Versions == nil {
		return fmt.Errorf("Facet runtime %s has an invalid browser export catalog", r.Label())
	}
	for name, version := range catalog.Versions {
		if _, ok := r.Packages[name][version]; !ok {
			return fmt.Errorf("Facet runtime %s browser exports reference missing package %s@%s", r.Label(), name, version)
		}
		r.Versions[name] = version
	}
	r.BrowserCatalog = layers.BrowserCatalog
	return nil
}

func validateRuntimePackage(root fs.FS, name, version, directory string) error {
	data, err := fs.ReadFile(root, directory+"/package.json")
	if err != nil {
		return err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	if version == "" || manifest.Name != name || manifest.Version != version {
		return fmt.Errorf("browser package %s@%s does not match its manifest", name, version)
	}
	if name != facetPackage && name != clickyPackage {
		return nil
	}
	for _, file := range []string{"importmap.json", "barrels.json"} {
		data, err := fs.ReadFile(root, directory+"/dist/runtime/"+file)
		if err != nil {
			return err
		}
		if !json.Valid(data) {
			return fmt.Errorf("browser package %s@%s has invalid %s", name, version, file)
		}
	}
	return nil
}

func (r *NamedRuntime) createLease(root *os.Root) error {
	if err := root.MkdirAll(".leases", 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Join(root.Name(), ".leases"), fmt.Sprintf("%d-", os.Getpid()))
	if err != nil {
		return err
	}
	r.lease = file.Name()
	return file.Close()
}

func (c *Cache) pinnedMetadata(name string) (Metadata, error) {
	packages := c.Pinned.Packages[name]
	if len(packages) == 0 {
		return Metadata{}, fmt.Errorf("package %s is absent from Facet runtime %s; add the dependency and run facet runtime build", name, c.Pinned.Label())
	}
	result := Metadata{Versions: map[string]Manifest{}, Time: map[string]string{}, Tags: map[string]string{}}
	for version, path := range packages {
		data, err := os.ReadFile(filepath.Join(path, "package.json"))
		if err != nil {
			return Metadata{}, err
		}
		var manifest Manifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return Metadata{}, err
		}
		result.Versions[version] = manifest
		result.Time[version] = "2020-01-01T00:00:00Z"
	}
	if version := c.Pinned.Versions[name]; version != "" {
		result.Tags["latest"] = version
	}
	return result, nil
}
