package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/gavel/comments/agents"
	commentapi "github.com/flanksource/gavel/comments/clicky"
	"github.com/flanksource/gavel/editor"
	editorui "github.com/flanksource/gavel/editor/ui"
	jsonlint "github.com/flanksource/gavel/linters/json"
	"github.com/flanksource/gavel/sdk/status"
)

type EditOptions struct {
	Paths  []string `args:"true"`
	Port   int      `flag:"port" help:"Local HTTP port (0 chooses an available port)"`
	NoOpen bool     `flag:"no-open" help:"Print the URL without opening a browser"`
	Schema string   `flag:"schema" help:"JSON Schema path or HTTP(S) URL overriding document declarations"`
	Dev    bool     `flag:"dev" help:"Serve the page from the editor/ui Vite dev server, with clicky-ui from the sibling checkout's source when present (requires a source checkout)"`
}

func init() {
	cmd := clicky.AddNamedCommandWithContext("edit", rootCmd, EditOptions{}, runEdit)
	cmd.Short = "Edit Markdown, MDX, JSON and YAML files, and preview images, in a local browser"
}
func runEdit(ctx context.Context, opts EditOptions) (any, error) {
	defer keepTasksOffTerminal()()
	dir, err := getWorkingDir()
	if err != nil {
		return nil, err
	}
	for i, path := range opts.Paths {
		if !filepath.IsAbs(path) {
			opts.Paths[i] = filepath.Join(dir, path)
		}
	}
	store, err := editor.NewStore(opts.Paths)
	if err != nil {
		return nil, err
	}
	if opts.Schema != "" {
		dir, err := getWorkingDir()
		if err != nil {
			return nil, err
		}
		opts.Schema, err = jsonlint.ResolveSchema(opts.Schema, dir)
		if err != nil {
			return nil, err
		}
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", opts.Port))
	if err != nil {
		return nil, err
	}
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	editorServer := &editor.Server{Store: store, Schema: opts.Schema, Assets: editorui.Assets(), Context: ctx, Agents: agents.Captain{}, Version: version}
	if err := editorServer.Reload(); err != nil {
		_ = listener.Close()
		return nil, err
	}
	cache, err := status.CacheDir()
	if err != nil {
		_ = listener.Close()
		return nil, err
	}
	if err := editorServer.WatchRuns(cache); err != nil {
		_ = listener.Close()
		return nil, err
	}
	if opts.Dev {
		if editorServer.DevProxy, err = devProxy(editorDev, ""); err != nil {
			_ = listener.Close()
			return nil, err
		}
	}
	you, err := commentapi.LocalUser()
	if err != nil {
		_ = listener.Close()
		return nil, err
	}
	editorServer.Comments = commentapi.LocalActor(commentapi.Handler(editorServer.OpenComments), you)
	server := &http.Server{Handler: editorServer.Handler(), ReadHeaderTimeout: 5 * time.Second}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	url := "http://" + listener.Addr().String()
	fmt.Println("Editor at " + url)
	if !opts.NoOpen {
		if err := openEditorBrowser(url); err != nil {
			cancel()
			_ = server.Close()
			return nil, err
		}
	}
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil, nil
		}
		return nil, err
	case <-ctx.Done():
	}
	cancel()
	shutdown, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	return nil, server.Shutdown(shutdown)
}
func openEditorBrowser(url string) error {
	command, args := "xdg-open", []string{url}
	switch runtime.GOOS {
	case "darwin":
		command = "open"
	case "windows":
		command = "cmd"
		args = []string{"/c", "start", "", url}
	}
	return exec.Command(command, args...).Run()
}
