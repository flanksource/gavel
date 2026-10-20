// Package vitedev runs a gavel UI from its Vite dev server: it finds the UI's
// source directory, reuses or starts `pnpm dev` there, and builds the reverse
// proxy the Go server puts in front of it. Each UI is an App; the PR dashboard
// and the test-runner UI share everything but their directory and port.
package vitedev

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/httpx"
)

// App is one UI with a Vite dev server.
type App struct {
	// Name names the UI in logs and errors.
	Name string
	// Dir is the UI's source directory, relative to the gavel repo root.
	Dir string
	// Port is the dev server's port, pinned by server.strictPort in the UI's
	// vite.config.ts.
	Port int
	// Entry is the module script the UI's index.html loads; finding it in the
	// page confirms the server on Port is this UI and not another process.
	Entry string
}

// WaitFor bounds how long Ensure waits for a dev server it started.
var WaitFor = 60 * time.Second

// URL is where the app's dev server listens.
func (a App) URL() string { return "http://" + a.host() }

func (a App) host() string { return fmt.Sprintf("localhost:%d", a.Port) }

// ResolveDir locates the app's source directory: override against the cwd
// first, then the app's Dir under the gavel repo root found by walking up
// from the cwd. Dev mode only makes sense from a source checkout, so a missing
// directory is an error.
func ResolveDir(app App, override string) (string, error) {
	if override != "" {
		if abs, err := filepath.Abs(override); err == nil && IsSourceDir(abs) {
			return abs, nil
		}
	}
	if root := findGavelRoot(); root != "" {
		if dir := filepath.Join(root, filepath.FromSlash(app.Dir)); IsSourceDir(dir) {
			return dir, nil
		}
	}
	where := app.Dir
	if override != "" {
		where, _ = filepath.Abs(override)
	}
	return "", fmt.Errorf("--dev: %s source not found at %s (run from the gavel repo, or pass the source dir)", app.Name, where)
}

// IsSourceDir reports whether path is a Vite app's source: it has both a
// package.json and a vite.config.ts.
func IsSourceDir(path string) bool {
	for _, f := range []string{"package.json", "vite.config.ts"} {
		if _, err := os.Stat(filepath.Join(path, f)); err != nil {
			return false
		}
	}
	return true
}

// findGavelRoot walks up from the cwd to the directory whose go.mod declares
// the gavel module, returning "" if none is found.
func findGavelRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if data, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
			if strings.Contains(string(data), "module github.com/flanksource/gavel") {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// Ensure makes the app's dev server reachable at app.URL(): it reuses one
// already serving the app, fails when the port is held by anything else, and
// otherwise starts `pnpm dev` in dir — tied to this process's lifetime by
// clicky.Exec — and waits for it.
func Ensure(app App, dir string) error {
	if portOccupied(app.host()) {
		if Serving(app) {
			logger.Infof("dev mode: reusing the %s Vite dev server at %s", app.Name, app.URL())
			return nil
		}
		return fmt.Errorf("%s is in use but is not the %s Vite dev server — free the port and retry", app.host(), app.Name)
	}

	logger.Infof("dev mode: starting Vite (pnpm dev) for the %s in %s", app.Name, dir)
	if err := clicky.Exec("pnpm", "dev").WithCwd(dir).Stream(os.Stdout, os.Stderr).Start(); err != nil {
		return fmt.Errorf("start the %s vite dev server: %w", app.Name, err)
	}
	deadline := time.Now().Add(WaitFor)
	for time.Now().Before(deadline) {
		if Serving(app) {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("the %s vite dev server did not become reachable at %s within %s", app.Name, app.URL(), WaitFor)
}

func portOccupied(host string) bool {
	conn, err := net.DialTimeout("tcp", host, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// Serving reports whether app.URL() answers with a page that loads the app's
// entry, so a stray HTTP server on the port is not mistaken for it.
func Serving(app App) bool {
	resp, err := httpx.Client(time.Second).Get(app.URL())
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return false
	}
	return strings.Contains(string(body), app.Entry)
}

// Proxy is the reverse proxy to the app's dev server. A failed upstream is a
// 502 naming the app and where `pnpm dev` runs.
func Proxy(app App) http.Handler {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: "http", Host: app.host()})
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Warnf("dev proxy %s%s: %v", app.URL(), r.URL.Path, err)
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "the %s vite dev server is unreachable at %s — is `pnpm dev` running in %s?\n", app.Name, app.URL(), app.Dir)
	}
	return proxy
}
