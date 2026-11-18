package editor

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/gavel/editor/preview"
)

type previewHost struct {
	origin string
	cache  *preview.Cache
	close  func() error
}

func (s *Server) startPreviewHost(ctx context.Context, opts PreviewOptions, selector string) (*previewHost, error) {
	cache := &preview.Cache{Root: filepath.Join(opts.CacheRoot, "page-runtime"), Local: s.PreviewLocal}
	if selector != "" {
		runtime, err := preview.OpenNamedRuntime(ctx, selector, opts.Directory)
		if err != nil {
			return nil, err
		}
		cache.Pinned = runtime
	} else if len(cache.Local) > 0 {
		frozen := map[string]string{}
		for name := range cache.Local {
			root, err := cache.Package(ctx, name, "latest")
			if err != nil {
				return nil, err
			}
			frozen[name] = root
		}
		cache.Local = frozen
		cache.LocalFrozen = true
	}
	host, err := s.servePreview(opts.HostOrigin, cache)
	if cache.Pinned != nil {
		if err != nil {
			return nil, errors.Join(err, cache.Pinned.Close())
		}
		stop := host.close
		host.close = func() error { return errors.Join(stop(), cache.Pinned.Close()) }
	}
	return host, err
}

func (h *previewHost) profile(hostOrigin string) map[string]any {
	managed := map[string]string{"@flanksource/facet": "0.1.74", "@flanksource/clicky-ui": "^0.3.57"}
	policy := map[string]any{"dependencies": "any"}
	if runtime := h.cache.Pinned; runtime != nil {
		for name := range managed {
			managed[name] = runtime.Versions[name]
		}
		allowed := make([]string, 0, len(runtime.Packages))
		for name := range runtime.Packages {
			allowed = append(allowed, name)
		}
		sort.Strings(allowed)
		policy = map[string]any{"dependencies": "strict", "allow": allowed}
	}
	profile := map[string]any{
		"runtimeDomain": "127.0.0.1", "embedders": []string{hostOrigin},
		"packageOrigin": h.origin + "/packages", "registry": h.origin + "/registry",
		"managed": managed, "policy": policy, "minReleaseAgeDays": 0,
		"network": map[string]any{"allow": []string{"https://fonts.googleapis.com/*", "https://fonts.gstatic.com/*"}},
		"csp":     map[string]any{"wasm": true, "add": map[string][]string{"style-src": {"https://fonts.googleapis.com"}, "font-src": {"'self'", "data:", "https://fonts.gstatic.com"}}},
		"links":   "deny", "sources": []string{"host"}, "readinessTimeoutMs": 30000,
	}
	if runtime := h.cache.Pinned; runtime != nil && len(runtime.BrowserCatalog) > 0 {
		profile["runtimeExports"] = true
	}
	return profile
}

func (s *Server) servePreview(hostOrigin string, cache *preview.Cache) (*previewHost, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	host := &previewHost{origin: "http://" + listener.Addr().String(), cache: cache}
	profile := host.profile(hostOrigin)
	mux := http.NewServeMux()
	if runtime := cache.Pinned; runtime != nil && len(runtime.BrowserCatalog) > 0 {
		mux.HandleFunc("GET /runtime-exports.json", func(w http.ResponseWriter, r *http.Request) {
			respond(w, runtime.BrowserCatalog, nil)
		})
	}
	mux.HandleFunc("GET /profile.json", func(w http.ResponseWriter, r *http.Request) {
		respond(w, profile, nil)
	})
	mux.HandleFunc("GET /sources/{id}/{path...}", s.Store.previewSource)
	mux.HandleFunc("GET /registry/{name...}", func(w http.ResponseWriter, r *http.Request) {
		metadata, err := cache.Metadata(r.Context(), preview.MetadataRequest{Name: r.PathValue("name")})
		respond(w, metadata, err)
	})
	mux.HandleFunc("GET /packages/{path...}", host.previewPackage)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if strings.HasPrefix(name, "frame/") {
			name = "bootstrap.html"
		}
		if !fs.ValidPath(name) || strings.Contains(name, "/") {
			http.NotFound(w, r)
			return
		}
		if name == "sw.js" {
			w.Header().Set("Service-Worker-Allowed", "/frame/")
		}
		runtime := s.PreviewAssets
		var err error
		if runtime == nil {
			runtime, err = fs.Sub(s.Assets, "page-runtime")
		}
		if err != nil {
			http.Error(w, "embedded browser runtime missing: "+err.Error(), 500)
			return
		}
		r.URL.Path = "/" + name
		http.FileServerFS(runtime).ServeHTTP(w, r)
	})
	server := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != listener.Addr().String() {
			http.Error(w, "invalid runtime host", 403)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", hostOrigin)
		w.Header().Set("Cache-Control", "no-store")
		mux.ServeHTTP(w, r)
	})}
	go func() { _ = server.Serve(listener) }()
	host.close = func() error {
		// The listener is already bound even if Serve has yet to register it.
		err := server.Close()
		if closeErr := listener.Close(); !errors.Is(closeErr, net.ErrClosed) {
			err = errors.Join(err, closeErr)
		}
		return err
	}
	return host, nil
}

func (h *previewHost) previewPackage(w http.ResponseWriter, r *http.Request) {
	value := r.PathValue("path")
	separator := strings.Index(value, "/")
	if strings.HasPrefix(value, "@") {
		if separator < 0 {
			http.Error(w, "invalid package path", 400)
			return
		}
		next := strings.Index(value[separator+1:], "/")
		if next < 0 {
			http.Error(w, "invalid package path", 400)
			return
		}
		separator += next + 1
	}
	if separator < 0 {
		http.Error(w, "invalid package path", 400)
		return
	}
	identity, name := value[:separator], value[separator+1:]
	versionAt := strings.LastIndex(identity, "@")
	if versionAt <= 0 {
		http.Error(w, "package requires exact version", 400)
		return
	}
	root, err := h.cache.Package(r.Context(), identity[:versionAt], identity[versionAt+1:])
	if err != nil {
		respond(w, nil, err)
		return
	}
	dir, err := os.OpenRoot(root)
	if err != nil {
		respond(w, nil, err)
		return
	}
	defer dir.Close()
	packageFS := dir.FS()
	if r.URL.Query().Has("meta") {
		files := []map[string]string{}
		err = fs.WalkDir(packageFS, ".", func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			data, err := fs.ReadFile(packageFS, path)
			if err != nil {
				return err
			}
			hash := sha256.Sum256(data)
			files = append(files, map[string]string{"path": "/" + path, "integrity": "sha256-" + base64.StdEncoding.EncodeToString(hash[:])})
			return nil
		})
		respond(w, map[string]any{"files": files}, err)
		return
	}
	if !fs.ValidPath(name) {
		http.Error(w, "invalid package file", 400)
		return
	}
	http.ServeFileFS(w, r, packageFS, name)
}

func (s *Server) previewCatalog(w http.ResponseWriter, r *http.Request) {
	sessions := s.previews
	if sessions == nil {
		respond(w, nil, fmt.Errorf("browser runtime is not started"))
		return
	}
	cache := sessions.defaultHost.cache
	if sessions.options.Runtime != "" {
		respond(w, map[string]string{"pinnedRuntime": cache.Pinned.Label()}, nil)
		return
	}
	if len(sessions.installed) > 0 {
		respond(w, struct {
			Installed []string `json:"installed"`
		}{sessions.installed}, nil)
		return
	}
	metadata, err := cache.Metadata(r.Context(), preview.MetadataRequest{Name: "@flanksource/facet", Refresh: r.URL.Query().Has("refresh")})
	if err != nil {
		respond(w, nil, err)
		return
	}
	cached, local := []string{}, []string{}
	for version, manifest := range metadata.Versions {
		if _, ok := cache.Local[manifest.Name]; ok {
			local = append(local, version)
			continue
		}
		hit, err := cache.Cached(manifest)
		if err != nil {
			respond(w, nil, err)
			return
		}
		if hit {
			cached = append(cached, version)
		}
	}
	respond(w, struct {
		preview.Metadata
		Cached []string `json:"cached"`
		Local  []string `json:"local"`
	}{metadata, cached, local}, nil)
}

func (h *previewHost) resolve(ctx context.Context, deps map[string]string, selected string) error {
	if runtime := h.cache.Pinned; runtime != nil {
		for name := range deps {
			if version, ok := runtime.Versions[name]; ok {
				deps[name] = version
			} else {
				delete(deps, name)
			}
		}
		for _, name := range []string{"@flanksource/facet", "@flanksource/clicky-ui"} {
			deps[name] = runtime.Versions[name]
		}
		return nil
	}
	if selected == "" {
		selected = deps["@flanksource/facet"]
	}
	if selected == "" {
		selected = "0.1.74"
	}
	facet, err := h.cache.Resolve(ctx, "@flanksource/facet", selected)
	if err != nil {
		return err
	}
	range_ := facet.PeerDependencies["@flanksource/clicky-ui"]
	if range_ == "" {
		return fmt.Errorf("Facet %s has no Clicky UI peer dependency", facet.Version)
	}
	ui, err := h.cache.Resolve(ctx, "@flanksource/clicky-ui", range_)
	if err != nil {
		return err
	}
	for _, library := range []preview.Manifest{facet, ui} {
		root, err := h.cache.Package(ctx, library.Name, library.Version)
		if err != nil {
			return err
		}
		for _, name := range []string{"importmap.json", "barrels.json"} {
			data, err := os.ReadFile(filepath.Join(root, "dist", "runtime", name))
			if err != nil {
				return fmt.Errorf("%s@%s does not ship runtime %s: %w", library.Name, library.Version, name, err)
			}
			if !json.Valid(data) {
				return fmt.Errorf("%s@%s has invalid runtime %s", library.Name, library.Version, name)
			}
		}
		deps[library.Name] = library.Version
	}
	return nil
}
