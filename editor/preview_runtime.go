package editor

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/gavel/editor/preview"
)

func (s *Server) StartPreview(ctx context.Context, hostOrigin, cacheRoot string) (func() error, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	s.PreviewOrigin = "http://" + listener.Addr().String()
	s.PreviewCache = &preview.Cache{Root: filepath.Join(cacheRoot, "page-runtime"), Local: s.PreviewLocal}
	if len(s.PreviewLocal) > 0 {
		frozen := map[string]string{}
		for name := range s.PreviewLocal {
			root, err := s.PreviewCache.Package(ctx, name, "latest")
			if err != nil {
				_ = listener.Close()
				return nil, err
			}
			frozen[name] = root
		}
		s.PreviewCache.Local = frozen
		s.PreviewCache.LocalFrozen = true
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /profile.json", func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]any{
			"runtimeDomain": "127.0.0.1", "embedders": []string{hostOrigin},
			"packageOrigin": s.PreviewOrigin + "/packages", "registry": s.PreviewOrigin + "/registry",
			"managed": map[string]string{"@flanksource/facet": "0.1.74", "@flanksource/clicky-ui": "^0.3.57"},
			"policy":  map[string]string{"dependencies": "any"}, "minReleaseAgeDays": 0,
			"network": map[string]any{"allow": []string{"https://fonts.googleapis.com/*", "https://fonts.gstatic.com/*"}},
			"csp":     map[string]any{"wasm": true, "add": map[string][]string{"style-src": {"https://fonts.googleapis.com"}, "font-src": {"'self'", "data:", "https://fonts.gstatic.com"}}},
			"links":   "deny", "sources": []string{"host"}, "readinessTimeoutMs": 30000,
		}, nil)
	})
	mux.HandleFunc("GET /sources/{id}/{path...}", s.Store.previewSource)
	mux.HandleFunc("GET /registry/{name...}", func(w http.ResponseWriter, r *http.Request) {
		metadata, err := s.PreviewCache.Metadata(r.Context(), preview.MetadataRequest{Name: r.PathValue("name")})
		respond(w, metadata, err)
	})
	mux.HandleFunc("GET /packages/{path...}", s.previewPackage)
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
	go func() { <-ctx.Done(); _ = server.Close() }()
	go func() { _ = server.Serve(listener) }()
	return server.Close, nil
}

func (s *Server) previewPackage(w http.ResponseWriter, r *http.Request) {
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
	root, err := s.PreviewCache.Package(r.Context(), identity[:versionAt], identity[versionAt+1:])
	if err != nil {
		respond(w, nil, err)
		return
	}
	if r.URL.Query().Has("meta") {
		files := []map[string]string{}
		err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			hash := sha256.Sum256(data)
			files = append(files, map[string]string{"path": "/" + filepath.ToSlash(rel), "integrity": "sha256-" + base64.StdEncoding.EncodeToString(hash[:])})
			return nil
		})
		respond(w, map[string]any{"files": files}, err)
		return
	}
	if !fs.ValidPath(name) {
		http.Error(w, "invalid package file", 400)
		return
	}
	http.ServeFileFS(w, r, os.DirFS(root), name)
}

func (s *Server) previewCatalog(w http.ResponseWriter, r *http.Request) {
	if s.PreviewCache == nil {
		respond(w, nil, fmt.Errorf("browser runtime is not started"))
		return
	}
	metadata, err := s.PreviewCache.Metadata(r.Context(), preview.MetadataRequest{Name: "@flanksource/facet", Refresh: r.URL.Query().Has("refresh")})
	if err != nil {
		respond(w, nil, err)
		return
	}
	cached, local := []string{}, []string{}
	for version, manifest := range metadata.Versions {
		if _, ok := s.PreviewLocal[manifest.Name]; ok {
			local = append(local, version)
			continue
		}
		hit, err := s.PreviewCache.Cached(manifest)
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

func (s *Server) resolvePreview(ctx context.Context, deps map[string]string, selected string) error {
	if s.PreviewCache == nil {
		return fmt.Errorf("browser runtime is not started")
	}
	if selected == "" {
		selected = deps["@flanksource/facet"]
	}
	if selected == "" {
		selected = "0.1.74"
	}
	facet, err := s.PreviewCache.Resolve(ctx, "@flanksource/facet", selected)
	if err != nil {
		return err
	}
	range_ := facet.PeerDependencies["@flanksource/clicky-ui"]
	if range_ == "" {
		return fmt.Errorf("Facet %s has no Clicky UI peer dependency", facet.Version)
	}
	ui, err := s.PreviewCache.Resolve(ctx, "@flanksource/clicky-ui", range_)
	if err != nil {
		return err
	}
	for _, library := range []preview.Manifest{facet, ui} {
		root, err := s.PreviewCache.Package(ctx, library.Name, library.Version)
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
