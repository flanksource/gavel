package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
)

const (
	// viteDevHost/viteDevURL is where the pr/ui Vite dev server listens (pinned
	// by server.strictPort in pr/ui/vite.config.ts). The Go server
	// reverse-proxies the UI here; Vite's HMR websocket runs on its own port
	// (24778). The port is deliberately not Vite's default 5173 to avoid
	// colliding with sibling UIs.
	viteDevHost = "localhost:5273"
	viteDevURL  = "http://" + viteDevHost
	// viteEntryMarker is the entry script Vite injects into our index.html;
	// finding it confirms the server on the port is the pr/ui app and not some
	// other process that merely answers HTTP on the same port.
	viteEntryMarker = "/src/index.tsx"
	viteWaitFor     = 60 * time.Second
)

// resolveDevDir locates the pr/ui source directory for the Vite dev server. It
// tries the configured path against the cwd first, then walks up to the gavel
// repo root and uses its pr/ui. It fails loudly when no valid source dir is
// found — `--dev` only makes sense from a source checkout.
func resolveDevDir(dir string) (string, error) {
	if abs, err := filepath.Abs(dir); err == nil && isDevDir(abs) {
		return abs, nil
	}
	if root := findGavelRoot(); root != "" {
		if cand := filepath.Join(root, "pr", "ui"); isDevDir(cand) {
			return cand, nil
		}
	}
	abs, _ := filepath.Abs(dir)
	return "", fmt.Errorf("--dev: pr/ui source not found at %s (run from the gavel repo, or pass --dev-dir)", abs)
}

// isDevDir reports whether path looks like the pr/ui source (has both a
// package.json and a vite.config.ts), guarding against proxying a stray dir.
func isDevDir(path string) bool {
	for _, f := range []string{"package.json", "vite.config.ts"} {
		if _, err := os.Stat(filepath.Join(path, f)); err != nil {
			return false
		}
	}
	return true
}

// findGavelRoot walks up from the cwd to the directory whose go.mod declares the
// gavel module, returning "" if none is found.
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

// ensureVite makes the pr/ui Vite dev server reachable at viteDevURL: it reuses
// one that is already running, fails loudly if the port is held by a different
// process, otherwise spawns `pnpm dev` in devDir (tied to this process's
// lifetime via clicky.Exec) and waits for it to come up.
func ensureVite(devDir string) error {
	if portOccupied(viteDevHost) {
		if viteReachable() {
			logger.Infof("dev mode: reusing Vite dev server at %s", viteDevURL)
			return nil
		}
		return fmt.Errorf("%s is in use but is not the pr/ui Vite dev server — free the port and retry", viteDevHost)
	}

	logger.Infof("dev mode: starting Vite (pnpm dev) in %s", devDir)
	if err := clicky.Exec("pnpm", "dev").WithCwd(devDir).Stream(os.Stdout, os.Stderr).Start(); err != nil {
		return fmt.Errorf("start vite dev server: %w", err)
	}

	deadline := time.Now().Add(viteWaitFor)
	for time.Now().Before(deadline) {
		if viteReachable() {
			return nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return fmt.Errorf("vite dev server did not become reachable at %s within %s", viteDevURL, viteWaitFor)
}

// portOccupied reports whether something is already listening on host.
func portOccupied(host string) bool {
	conn, err := net.DialTimeout("tcp", host, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// viteReachable reports whether the pr/ui Vite dev server is serving its app —
// it must answer with HTML containing our entry script, so a stray HTTP server
// on the same port is not mistaken for our dev server.
func viteReachable() bool {
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get(viteDevURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return false
	}
	return strings.Contains(string(body), viteEntryMarker)
}
