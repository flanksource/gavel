package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	nethttp "net/http"
	"os"
	"runtime"
	"sort"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/gavel/service"
)

// componentStatus / healthStatus mirror pr/ui.ComponentStatus /
// pr/ui.StatusResponse over the /api/status JSON endpoint. Kept as local
// DTOs so cmd/gavel doesn't have to import pr/ui (and thereby drag in all
// the UI server dependencies).
type componentStatus struct {
	Severity string          `json:"severity"` // "ok" | "degraded" | "down"
	Message  string          `json:"message"`
	Detail   json.RawMessage `json:"detail,omitempty"`
}

type healthStatus struct {
	Overall   string          `json:"overall"`
	Database  componentStatus `json:"database"`
	GitHub    componentStatus `json:"github"`
	CheckedAt time.Time       `json:"checkedAt"`
}

// cacheDetail is what componentStatus.Detail carries for the database
// component — same shape as the /api/activity/cache response. We decode it
// lazily to pull out row counts + DSN fields for the CLI rendering.
type cacheDetail struct {
	Enabled      bool             `json:"enabled"`
	Driver       string           `json:"driver"`
	DSNSource    string           `json:"dsnSource"`
	DSNMasked    string           `json:"dsnMasked"`
	RetentionSec int64            `json:"retentionSec"`
	Counts       map[string]int64 `json:"counts"`
}

// healthURL builds the /api/status URL from the persisted UI port so
// --port=0 auto-scan installs still get polled at the right place. Called
// per-request (not cached) because the daemon could have been reinstalled
// on a different port between CLI invocations.
func healthURL() string {
	return fmt.Sprintf("http://localhost:%d/api/status", service.ReadUIPort())
}

type SystemStatusOptions struct {
	LogLines int `flag:"log-lines" help:"Number of trailing log lines to show (0 hides the log section)" default:"25"`
}

func (SystemStatusOptions) Help() string {
	return `Report health of the background gavel pr UI daemon.

Shows:
  - Service source (launchd / systemd --user / pidfile fallback)
  - Process state (running, stale, not running)
  - Configured database backend (mode + masked DSN or embedded data dir)
  - Live health — fetched from the daemon's /api/status endpoint so the
    CLI and the PR UI's header status dot share a single source of truth.
    Breaks out database and GitHub components separately.
  - Last --log-lines lines of the daemon log (default 25, pass 0 to hide)`
}

func init() {
	clicky.AddNamedCommand("status", systemCmd, SystemStatusOptions{}, runSystemStatus)
}

// systemStatusReport bundles the three sections (service / database / log
// tail) that `system status` renders. It implements clicky's Pretty()
// contract so the cobra framework prints it automatically with respect to
// the global --format flag (ANSI by default, JSON/YAML/MD if requested).
type systemStatusReport struct {
	Installed bool
	Service   service.Status
	DBConfig  service.DBConfig
	DBPath    string
	DataDir   string
	// Health is populated only when the daemon is running and answered
	// /api/status. HealthErr captures why the fetch didn't happen (daemon
	// down, connect refused, timeout) so we can tell users what to do next.
	Health    *healthStatus
	HealthErr string
	LogLines  []string
	LogWanted int
}

func runSystemStatus(opts SystemStatusOptions) (any, error) {
	installed, err := service.IsInstalled()
	if err != nil {
		return nil, err
	}
	st, err := service.ReadStatus()
	if err != nil {
		return nil, err
	}
	cfg, err := service.LoadDBConfig()
	if err != nil {
		return nil, err
	}
	dbPath, _ := service.DBConfigPath()

	var dataDir string
	if cfg.Mode == service.DBModeEmbedded {
		dataDir, _ = service.EmbeddedDataDir()
	}

	var logLines []string
	if opts.LogLines > 0 {
		logLines, err = service.TailLog(opts.LogLines)
		if err != nil {
			return nil, err
		}
	}

	// Only try to reach the UI if the service layer says the daemon is
	// running — otherwise we waste a 2s TCP timeout on every status call.
	var health *healthStatus
	var healthErr string
	if st.Running {
		h, err := fetchHealth()
		if err != nil {
			healthErr = err.Error()
		} else {
			health = h
		}
	}

	return systemStatusReport{
		Installed: installed,
		Service:   st,
		DBConfig:  cfg,
		DBPath:    dbPath,
		DataDir:   dataDir,
		Health:    health,
		HealthErr: healthErr,
		LogLines:  logLines,
		LogWanted: opts.LogLines,
	}, nil
}

// fetchHealth queries the running pr-ui daemon's /api/status endpoint. The
// daemon is the single source of truth for embedded-postgres state — opening
// a second connection from the CLI would collide with the daemon's managed
// instance.
func fetchHealth() (*healthStatus, error) {
	url := healthURL()
	client := &nethttp.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != nethttp.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}
	var h healthStatus
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		return nil, fmt.Errorf("decode status: %w", err)
	}
	return &h, nil
}

func (r systemStatusReport) Pretty() api.Text {
	t := clicky.Text("== service ==", "font-bold").NewLine()
	t = t.Append(kv("source", serviceSourceLabel(r.Installed))).NewLine()
	t = t.Append(kv("pidfile", r.Service.PidFile)).NewLine()
	t = t.Append(kv("logfile", r.Service.LogFile)).NewLine()
	t = t.Append(kv("status", "")).Add(serviceStatusText(r.Service)).NewLine()

	t = t.NewLine().Append("== database ==", "font-bold").NewLine()
	t = appendDBSection(t, r.DBConfig, r.DBPath, r.DataDir)

	if !r.Service.Running {
		t = t.NewLine().Append("== health ==", "font-bold").NewLine()
		t = t.Add(icons.Pending).Space().
			Append("daemon not running — start it with ", "text-muted").
			Append("gavel system start", "font-bold").NewLine()
	} else if r.Health == nil {
		t = t.NewLine().Append("== health ==", "font-bold").NewLine()
		t = t.Add(icons.Error).Space().
			Append("could not query daemon: ", "error").
			Append(r.HealthErr).NewLine()
	} else {
		t = t.NewLine().Append("== database health ==", "font-bold").NewLine()
		t = appendDBHealthSection(t, r.Health.Database)
		t = t.NewLine().Append("== github ==", "font-bold").NewLine()
		t = appendGitHubSection(t, r.Health.GitHub)
	}

	if r.LogWanted > 0 {
		t = t.NewLine().
			Append("== log (last ", "font-bold").
			Append(r.LogWanted).
			Append(" lines) ==", "font-bold").NewLine()
		t = appendLogSection(t, r.LogLines)
	}
	return t
}

// kv renders "label: " with the label muted — the caller appends the value.
func kv(label, value string) api.Text {
	t := clicky.Text(label+":", "text-muted").Space()
	if value != "" {
		t = t.Append(value)
	}
	return t
}

// serviceStatusText renders the running/stale/not-running line with the
// matching status icon so the line reads at a glance.
func serviceStatusText(st service.Status) api.Text {
	switch {
	case st.Running:
		return api.Text{}.
			Add(icons.Success).Space().
			Append("running", "text-green-600").
			Append(" (pid ").Append(st.PID).Append(")")
	case st.Stale:
		return api.Text{}.
			Add(icons.Warning).Space().
			Append("stale", "warning").
			Append(" (pidfile points to dead pid ").Append(st.PID).Append(")")
	default:
		return api.Text{}.
			Add(icons.Pending).Space().
			Append("not running", "text-muted")
	}
}

func appendDBSection(t api.Text, cfg service.DBConfig, path, dataDir string) api.Text {
	switch cfg.Mode {
	case "":
		t = t.Append(kv("config", path)).Space().Append("(missing)", "warning").NewLine()
		t = t.Append(kv("mode", "")).
			Add(icons.Warning).Space().
			Append("<not configured> — run ", "warning").
			Append("gavel system install", "font-bold").
			Append(" to set one", "warning").NewLine()
	case service.DBModeDSN:
		t = t.Append(kv("config", path)).NewLine()
		t = t.Append(kv("mode", "")).
			Add(icons.DB).Space().
			Append("dsn", "font-bold").Space().
			Append("(external postgres)", "text-muted").NewLine()
		t = t.Append(kv("dsn", service.MaskDSN(cfg.DSN))).NewLine()
	case service.DBModeEmbedded:
		t = t.Append(kv("config", path)).NewLine()
		t = t.Append(kv("mode", "")).
			Add(icons.DB).Space().
			Append("embedded", "font-bold").Space().
			Append("(postgres managed by pr-ui daemon)", "text-muted").NewLine()
		t = t.Append(kv("datadir", dataDir))
		if _, err := os.Stat(dataDir); errors.Is(err, fs.ErrNotExist) {
			t = t.Space().Append("(not yet initialized — will be created on first daemon start)", "text-muted")
		}
		t = t.NewLine()
	default:
		t = t.Append(kv("mode", cfg.Mode)).Space().Append("(unknown)", "error").NewLine()
	}
	return t
}

// appendDBHealthSection renders the database ComponentStatus from the
// daemon's /api/status response. Severity drives the icon; detail (a
// cacheDetail payload) surfaces the row counts and DSN metadata the CLI
// user already expects to see.
func appendDBHealthSection(t api.Text, c componentStatus) api.Text {
	t = t.Append(kv("status", "")).Add(severityText(c)).NewLine()
	var detail cacheDetail
	if len(c.Detail) > 0 {
		_ = json.Unmarshal(c.Detail, &detail)
	}
	if detail.Driver != "" {
		t = t.Append(kv("driver", detail.Driver)).NewLine()
	}
	if detail.DSNSource != "" {
		t = t.Append(kv("source", detail.DSNSource)).NewLine()
	}
	if detail.DSNMasked != "" {
		t = t.Append(kv("dsn", detail.DSNMasked)).NewLine()
	}
	if detail.RetentionSec > 0 {
		t = t.Append(kv("retention", (time.Duration(detail.RetentionSec) * time.Second).String())).NewLine()
	}
	if len(detail.Counts) > 0 {
		t = t.Append(kv("rows", "")).NewLine()
		keys := make([]string, 0, len(detail.Counts))
		for k := range detail.Counts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			t = t.Append("  ").
				Append(k+":", "text-muted").Space().
				Append(detail.Counts[k]).NewLine()
		}
	}
	return t
}

// appendGitHubSection renders the GitHub ComponentStatus. When the detail
// payload carries a rateLimit, show it — that's the most actionable piece
// of info for "why is the dot yellow?".
func appendGitHubSection(t api.Text, c componentStatus) api.Text {
	t = t.Append(kv("status", "")).Add(severityText(c)).NewLine()
	// The detail is a free-form map; probe for rateLimit / lastError.
	var detail map[string]json.RawMessage
	if len(c.Detail) > 0 {
		_ = json.Unmarshal(c.Detail, &detail)
	}
	if raw, ok := detail["rateLimit"]; ok {
		var rl struct {
			Limit     int   `json:"limit"`
			Remaining int   `json:"remaining"`
			Used      int   `json:"used"`
			Reset     int64 `json:"reset"`
		}
		if json.Unmarshal(raw, &rl) == nil && rl.Limit > 0 {
			t = t.Append(kv("rate limit", fmt.Sprintf("%d / %d remaining", rl.Remaining, rl.Limit))).NewLine()
			if rl.Reset > 0 {
				t = t.Append(kv("resets", time.Unix(rl.Reset, 0).Format(time.RFC3339))).NewLine()
			}
		}
	}
	if raw, ok := detail["lastError"]; ok {
		var msg string
		if json.Unmarshal(raw, &msg) == nil && msg != "" {
			t = t.Append(kv("last error", msg)).NewLine()
		}
	}
	return t
}

// severityText maps a ComponentStatus severity + message to an icon-prefixed
// coloured line. Shared between the db and github sections so they look
// consistent.
func severityText(c componentStatus) api.Text {
	switch c.Severity {
	case "ok":
		return api.Text{}.
			Add(icons.Success).Space().
			Append(c.Message, "text-green-600")
	case "degraded":
		return api.Text{}.
			Add(icons.Warning).Space().
			Append(c.Message, "warning")
	case "down":
		return api.Text{}.
			Add(icons.Error).Space().
			Append(c.Message, "error")
	default:
		return api.Text{}.
			Add(icons.Unknown).Space().
			Append(c.Message, "text-muted")
	}
}

func appendLogSection(t api.Text, lines []string) api.Text {
	if len(lines) == 0 {
		return t.Append("<no log yet>", "text-muted").NewLine()
	}
	for _, line := range lines {
		t = t.Append(line).NewLine()
	}
	return t
}

// serviceSourceLabel describes where ReadStatus sourced the state from, so
// users can tell whether they're looking at launchd/systemd state or the
// fallback pidfile check.
func serviceSourceLabel(installed bool) string {
	if !installed {
		return "pidfile (no service installed — run `gavel system install`)"
	}
	switch runtime.GOOS {
	case "darwin":
		return "launchd (service installed)"
	case "linux":
		return "systemd --user (service installed)"
	default:
		return "service (installed)"
	}
}
