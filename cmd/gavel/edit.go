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
	"github.com/flanksource/gavel/internal/database"
	"github.com/flanksource/gavel/internal/vitedev"
	jsonlint "github.com/flanksource/gavel/linters/json"
	"github.com/flanksource/gavel/sdk/status"
	"github.com/flanksource/gavel/verify"
)

type EditOptions struct {
	Paths          []string `args:"true"`
	Port           int      `flag:"port" help:"Local HTTP port (0 chooses an available port)"`
	NoOpen         bool     `flag:"no-open" help:"Print the URL without opening a browser"`
	Schema         string   `flag:"schema" help:"JSON Schema path or HTTP(S) URL overriding document declarations"`
	Ext            []string `flag:"ext" help:"File extensions folders list, e.g. .md or .go, instead of the built-in documents and images; repeat for each (overrides .gavel.yaml edit.extensions)"`
	IncludeIgnored []string `flag:"include-ignored" help:"Folder-relative glob or directory whose files are listed even though .gitignore excludes them; repeat for each (overrides .gavel.yaml edit.includeIgnored)"`
	Dev            bool     `flag:"dev" help:"Serve the page from the editor/ui Vite dev server, with clicky-ui from the sibling checkout's source when present (requires a source checkout)"`
	// Session defaults to the agent session gavel edit runs inside.
	Session string `flag:"session" help:"Agent session id new comments are assigned and delivered to (default: the Claude/Codex session gavel edit runs in, as captain detects it; \"none\" routes through the files' comment agents)"`
}

// editSession is the agent session new comments go to: the flag's session,
// none for "none", else the live session captain finds this process inside
// (inside an agent that is not live, an error).
func editSession(ctx context.Context, flag string) (session string, label string, err error) {
	switch flag {
	case "none":
		return "", "", nil
	case "":
		record, err := agents.Captain{}.Current(ctx)
		if err != nil || record == nil {
			return "", "", err
		}
		label = record.Source + " session " + record.ID
		if record.Title != "" {
			label += " (" + record.Title + ")"
		}
		return record.ID, label, nil
	default:
		return flag, "session " + flag, nil
	}
}

func init() {
	cmd := clicky.AddNamedCommandWithContext("edit", rootCmd, EditOptions{}, runEdit)
	cmd.Short = "Edit Markdown, MDX, TSX, JSX, JSON and YAML files in a local browser"
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
	editorServer := &editor.Server{ListingFlags: verify.EditConfig{Extensions: opts.Ext, IncludeIgnored: opts.IncludeIgnored}}
	listingDir := dir
	if len(opts.Paths) > 0 {
		listingDir = opts.Paths[0]
		if info, err := os.Stat(listingDir); err == nil && !info.IsDir() {
			listingDir = filepath.Dir(listingDir)
		}
	}
	listing, err := editorServer.ListingFor(listingDir)
	if err != nil {
		return nil, err
	}
	store, err := editor.NewStore(editor.StoreOptions{Paths: opts.Paths, Listing: listing})
	if err != nil {
		return nil, err
	}
	store.EnableAccess()
	if _, err := database.Shared(ctx, database.WithMigrations()); err != nil {
		return nil, fmt.Errorf("initialize editor database: %w", err)
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
	session, sessionLabel, err := editSession(ctx, opts.Session)
	if err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("resolving the agent session for new comments (pass --session=none to skip): %w", err)
	}
	editorServer.Store, editorServer.Schema, editorServer.Assets, editorServer.Context = store, opts.Schema, editorui.Assets(), ctx
	editorServer.Agents, editorServer.Session, editorServer.Version = agents.Captain{}, session, version
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
		sourceDir, err := vitedev.ResolveDir(editorDev, "")
		if err != nil {
			_ = listener.Close()
			return nil, err
		}
		siblings := filepath.Clean(filepath.Join(sourceDir, "../../.."))
		editorServer.PreviewLocal = map[string]string{"@flanksource/facet": filepath.Join(siblings, "facet"), "@flanksource/clicky-ui": filepath.Join(siblings, "clicky-ui/packages/ui")}
		editorServer.PreviewAssets = os.DirFS(filepath.Join(siblings, "clicky-ui/packages/page-runtime/dist/static"))
	}
	closePreview, err := editorServer.StartPreview(ctx, "http://"+listener.Addr().String(), cache)
	if err != nil {
		_ = listener.Close()
		return nil, err
	}
	defer closePreview()
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
	if sessionLabel != "" {
		fmt.Println("Comments go to " + sessionLabel)
	}
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
