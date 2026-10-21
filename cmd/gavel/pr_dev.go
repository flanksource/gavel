package main

import (
	"net/http"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/internal/vitedev"
)

// The UIs `--dev` serves from their Vite dev servers. Each port is pinned by
// server.strictPort in the UI's vite.config.ts, away from Vite's default 5173
// and the clicky-ui playground's 5274.
var (
	prDashboardDev = vitedev.App{Name: "PR dashboard", Dir: "pr/ui", Port: 5273, Entry: "/src/index.tsx"}
	testRunnerDev  = vitedev.App{Name: "test-runner UI", Dir: "testrunner/ui", Port: 5275, Entry: "/src/index.tsx"}
)

// devProxy is the proxy a `--dev` server puts in front of app: it finds the
// app's source (override first), reuses or starts its Vite dev server, and
// proxies to it.
func devProxy(app vitedev.App, override string) (http.Handler, error) {
	dir, err := vitedev.ResolveDir(app, override)
	if err != nil {
		return nil, err
	}
	if err := vitedev.Ensure(app, dir); err != nil {
		return nil, err
	}
	logger.Infof("dev mode: reverse-proxying the %s to %s", app.Name, app.URL())
	return vitedev.Proxy(app), nil
}
