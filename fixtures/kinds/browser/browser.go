package browser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/gavel/fixtures"
)

const devToolsActivePort = "DevToolsActivePort"

// BrowserConfig is a `kind: browser` step: one Chrome on one profile that
// later playwright and agent-browser steps attach to over CDP. It lives until
// the fixture ends.
type BrowserConfig struct {
	Engine       string             `json:"engine,omitempty" jsonschema:"enum=chrome,description=Browser engine (only chrome)"`
	Binary       string             `json:"binary,omitempty" jsonschema:"description=Chrome executable (default: newest Playwright Chrome for Testing, then system Chrome)"`
	Profile      Profile            `json:"profile,omitempty"`
	Headless     *bool              `json:"headless,omitempty" jsonschema:"description=Run headless (default true)"`
	Args         []string           `json:"args,omitempty" jsonschema:"description=Extra Chrome command-line arguments"`
	Session      string             `json:"session,omitempty" jsonschema:"description=BROWSER_SESSION name for agent-browser (default: generated)"`
	StartTimeout *fixtures.Duration `json:"startTimeout,omitempty" jsonschema:"description=How long to wait for the DevTools endpoint (default 30s)"`
}

// Browser is the `browser` kind.
type Browser struct{}

func (Browser) Name() string            { return "browser" }
func (Browser) Schema() json.RawMessage { return schemaOf(&BrowserConfig{}) }

// DefaultWait makes `wait:` on a browser step mean "until CDP answers".
func (Browser) DefaultWait(fixtures.StepSpec) *fixtures.WaitPolicy {
	return &fixtures.WaitPolicy{Until: "cdp.ready"}
}

// chromeHandle is the running Chrome, passed back to Observe.
type chromeHandle struct {
	process *managedProcess
	cdpURL  string
}

func (Browser) Execute(ctx context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	start := time.Now()
	cfg, err := browserConfig(step)
	if err != nil {
		return fixtures.StepOutcome{}, err
	}
	binary, err := resolveChrome(defaultChromeSearch(cfg.Binary))
	if err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("browser step %q: %w", step.Spec.ID, err)
	}
	profile, err := prepareProfile(ctx, cfg.Profile)
	if err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("browser step %q: %w", step.Spec.ID, err)
	}
	args := chromeArgs(cfg, profile.Dir)
	process, err := startManagedProcess(ctx, binary, args, "", environ(step), nil)
	if err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("browser step %q: %w", step.Spec.ID, err)
	}
	handle := &chromeHandle{process: process}

	version, err := handle.await(ctx, profile.Dir, cfg.StartTimeout.Std())
	if err == nil && profile.StorageState != "" {
		err = applyStorageState(ctx, version.WebSocketDebuggerURL, profile.StorageState)
	}
	if err != nil {
		return fixtures.StepOutcome{}, errors.Join(fmt.Errorf("browser step %q: %w", step.Spec.ID, err), handle.stop(ctx))
	}

	result := newResult(step, start)
	result.Command = binary + " " + strings.Join(args, " ")
	return fixtures.StepOutcome{
		Result: finish(result),
		Outputs: map[string]any{
			"cdp":        map[string]any{"ready": true, "url": handle.cdpURL, "wsUrl": version.WebSocketDebuggerURL, "browser": version.Browser},
			"session":    cfg.Session,
			"profileDir": profile.Dir,
			"binary":     binary,
			"pid":        process.pid(),
		},
		Exports: map[string]string{
			"BROWSER_CDP_URL":     handle.cdpURL,
			"BROWSER_WS_URL":      version.WebSocketDebuggerURL,
			"BROWSER_SESSION":     cfg.Session,
			"BROWSER_PROFILE_DIR": profile.Dir,
		},
		Handle:  handle,
		Cleanup: handle.stop,
	}, nil
}

// Observe re-probes the DevTools endpoint. An unreachable endpoint is an
// observation (cdp.ready=false with the reason), not an error, so a `wait:`
// can report it.
func (Browser) Observe(ctx context.Context, _ *fixtures.StepContext, handle any) (map[string]any, error) {
	chrome, ok := handle.(*chromeHandle)
	if !ok {
		return nil, fmt.Errorf("browser observe: unexpected handle %T", handle)
	}
	version, err := probeVersion(ctx, chrome.cdpURL)
	if err != nil {
		return map[string]any{"cdp": map[string]any{"ready": false, "url": chrome.cdpURL, "error": err.Error()}}, nil
	}
	return map[string]any{"cdp": map[string]any{"ready": true, "url": chrome.cdpURL, "wsUrl": version.WebSocketDebuggerURL, "browser": version.Browser}}, nil
}

func browserConfig(step *fixtures.StepContext) (BrowserConfig, error) {
	var cfg BrowserConfig
	if len(step.Spec.Config) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(step.Spec.Config))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&cfg); err != nil {
			return cfg, fmt.Errorf("browser step %q: decode config: %w", step.Spec.ID, err)
		}
	}
	var err error
	if cfg.Engine != "" && cfg.Engine != "chrome" {
		return cfg, fmt.Errorf("browser step %q: engine %q is not supported (only chrome)", step.Spec.ID, cfg.Engine)
	}
	fields := []struct {
		name  string
		value *string
	}{{"binary", &cfg.Binary}, {"profile.dir", &cfg.Profile.Dir}, {"profile.from", &cfg.Profile.From}, {"session", &cfg.Session}}
	for _, field := range fields {
		if *field.value, err = render(step, field.name, *field.value); err != nil {
			return cfg, err
		}
	}
	for i := range cfg.Args {
		if cfg.Args[i], err = render(step, "args", cfg.Args[i]); err != nil {
			return cfg, err
		}
	}
	if cfg.Profile.Dir == "" {
		cfg.Profile.Dir = filepath.Join(step.TmpDir, "browser-profile-"+stepName(step))
	} else {
		cfg.Profile.Dir = resolvePath(step, cfg.Profile.Dir)
	}
	cfg.Profile.From = resolvePath(step, cfg.Profile.From)
	if cfg.Session == "" {
		cfg.Session = "gavel-" + stepName(step) + "-" + randomHex(4)
	}
	if cfg.StartTimeout == nil {
		timeout := fixtures.Duration(30 * time.Second)
		cfg.StartTimeout = &timeout
	}
	return cfg, nil
}

func chromeArgs(cfg BrowserConfig, profileDir string) []string {
	args := []string{
		"--user-data-dir=" + profileDir,
		// Port 0 lets Chrome pick a free port and write it to
		// DevToolsActivePort, so no port can be stolen between pick and bind.
		"--remote-debugging-port=0",
		"--no-first-run",
		"--no-default-browser-check",
		"--use-mock-keychain",
		"--password-store=basic",
	}
	if cfg.Headless == nil || *cfg.Headless {
		args = append(args, "--headless=new")
	}
	return append(append(args, cfg.Args...), "about:blank")
}

type cdpVersion struct {
	Browser              string `json:"Browser"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

// await waits for Chrome to publish its DevTools port and answer
// /json/version, failing early (with Chrome's stderr) when it exits.
func (h *chromeHandle) await(ctx context.Context, profileDir string, timeout time.Duration) (cdpVersion, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if h.process.exited() {
			return cdpVersion{}, fmt.Errorf("chrome exited (code %d) before DevTools was ready: %s", h.process.exitCode(), strings.TrimSpace(h.process.stderr.String()))
		}
		if h.cdpURL == "" {
			if port, err := readDevToolsPort(profileDir); err == nil {
				h.cdpURL = "http://127.0.0.1:" + strconv.Itoa(port)
			} else {
				lastErr = err
			}
		}
		if h.cdpURL != "" {
			version, err := probeVersion(ctx, h.cdpURL)
			if err == nil {
				return version, nil
			}
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return cdpVersion{}, ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
	return cdpVersion{}, fmt.Errorf("chrome DevTools not ready within %s: %w", timeout, lastErr)
}

func readDevToolsPort(profileDir string) (int, error) {
	raw, err := os.ReadFile(filepath.Join(profileDir, devToolsActivePort))
	if err != nil {
		return 0, err
	}
	line, _, _ := strings.Cut(string(raw), "\n")
	port, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || port <= 0 {
		return 0, fmt.Errorf("%s holds no port yet (%q)", devToolsActivePort, line)
	}
	return port, nil
}

func probeVersion(ctx context.Context, cdpURL string) (cdpVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cdpURL+"/json/version", nil)
	if err != nil {
		return cdpVersion{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return cdpVersion{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return cdpVersion{}, fmt.Errorf("GET %s/json/version: %s", cdpURL, resp.Status)
	}
	var version cdpVersion
	if err := json.NewDecoder(resp.Body).Decode(&version); err != nil {
		return cdpVersion{}, fmt.Errorf("decode %s/json/version: %w", cdpURL, err)
	}
	if version.WebSocketDebuggerURL == "" {
		return cdpVersion{}, fmt.Errorf("%s/json/version has no webSocketDebuggerUrl", cdpURL)
	}
	return version, nil
}

// stop kills Chrome's whole process group (renderers, GPU, crashpad) and
// waits for it to be reaped.
func (h *chromeHandle) stop(context.Context) error {
	if err := h.process.killTree(); err != nil {
		return fmt.Errorf("kill chrome process tree: %w", err)
	}
	return h.process.waitWithin(10 * time.Second)
}
