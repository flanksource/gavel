package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flanksource/clicky"
	commonsdb "github.com/flanksource/commons-db/db"
	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github"
	"github.com/flanksource/gavel/service"
	"github.com/jackc/pgx/v5"
)

type SystemInstallOptions struct {
	DSN             string `flag:"dsn" help:"Postgres DSN for the github cache (mutually exclusive with --embedded; overrides the embedded default)"`
	Embedded        bool   `flag:"embedded" help:"Launch an embedded postgres managed by the pr-ui daemon (default when --dsn is not supplied)"`
	Token           string `flag:"token" help:"GitHub token to persist for the daemon (defaults to GITHUB_TOKEN, GH_TOKEN, or gh auth token)"`
	SkipVerifyToken bool   `flag:"skip-verify-token" help:"Skip the live GitHub API probe of the token (for GHE or unusual token scopes)"`
	BinaryPath      string `flag:"binary" help:"Path to the gavel binary (defaults to the current executable)"`
	DryRun          bool   `flag:"dry-run" help:"Print actions and rendered service file without writing anything"`
	Force           bool   `flag:"force" help:"Overwrite an existing service file"`
}

func (SystemInstallOptions) Help() string {
	return `Install a user-level background service that keeps gavel pr list --all --ui
running across logins.

Database backend (defaults to --embedded when no flag is given):
  --embedded            Launch embedded postgres from the pr-ui daemon.
                        Binaries are downloaded on first run; install
                        verifies start/stop immediately so you find out
                        up front if the binaries can't be fetched.
                        This is the default.
  --dsn=postgres://...  Use an external postgres you manage. The DSN is
                        validated with a real connection at install time
                        so bad configs fail here instead of at runtime.

GitHub authentication:
  A token is required for the daemon to poll pull requests. launchd / systemd
  --user services do NOT inherit your shell env, so this command persists a
  token to ~/.config/gavel/auth.json (0600 perms). Discovery order:
    1. --token=... (explicit)
    2. GITHUB_TOKEN env var
    3. GH_TOKEN env var
    4. gh auth token (if the gh CLI is installed and authenticated)

  Before writing the token, install probes GitHub to confirm it works:
    - GET /rate_limit  → catches expired / revoked tokens
    - GET /user/orgs   → catches tokens missing the read:org scope
  Install fails fast if either call errors — so you find out before the
  daemon is launched. Use --skip-verify-token for GHE or unusual scopes.
  If no token is found the install still proceeds with a warning.

On macOS this writes ~/Library/LaunchAgents/com.flanksource.gavel-pr-ui.plist
and loads it via launchctl. On Linux it writes
~/.config/systemd/user/gavel-pr-ui.service and starts it via systemctl --user;
run 'loginctl enable-linger $USER' if you want it to survive logout.

No root required. Use --dry-run to preview.`
}

func init() {
	clicky.AddNamedCommand("install", systemCmd, SystemInstallOptions{}, runSystemInstall)
}

func runSystemInstall(opts SystemInstallOptions) (any, error) {
	cfg, err := resolveInstallDBConfig(opts)
	if err != nil {
		return nil, err
	}

	if err := verifyDBConfig(cfg); err != nil {
		return nil, err
	}

	if !opts.DryRun {
		if err := service.SaveDBConfig(cfg); err != nil {
			return nil, fmt.Errorf("save db config: %w", err)
		}
		path, _ := service.DBConfigPath()
		logger.Infof("Wrote db config to %s (mode=%s)", path, cfg.Mode)
	} else {
		logger.Infof("[dry-run] would write db config mode=%s", cfg.Mode)
	}

	if err := persistGitHubToken(opts); err != nil {
		return nil, err
	}

	return nil, service.Install(service.InstallOptions{
		BinaryPath: opts.BinaryPath,
		DryRun:     opts.DryRun,
		Force:      opts.Force,
	})
}

// persistGitHubToken discovers a GitHub token (explicit --token, env vars,
// or `gh auth token`), verifies it works against the live GitHub API, then
// writes it to ~/.config/gavel/auth.json so the launchd/systemd daemon —
// which doesn't inherit the user's shell env — can authenticate.
//
// Verification runs BEFORE the write so an invalid token is rejected with
// an actionable error rather than silently saved and rediscovered later as
// a "no GitHub token" poller warning. The daemon's background probe would
// eventually notice, but surfacing the problem at install time is cheaper.
//
// A missing token is not fatal: the install proceeds with a warning so
// the user can run the daemon and come back with a token later.
func persistGitHubToken(opts SystemInstallOptions) error {
	token, source := opts.Token, "--token"
	if token == "" {
		discovered, src, err := service.DiscoverGitHubToken()
		if err != nil {
			return fmt.Errorf("discover github token: %w", err)
		}
		token, source = discovered, string(src)
	}

	if token == "" {
		logger.Warnf("No GitHub token found (checked --token, GITHUB_TOKEN, GH_TOKEN, `gh auth token`); daemon will log auth errors until you re-run `gavel system install --token=...`")
		return nil
	}

	if err := verifyGitHubToken(token, opts.SkipVerifyToken); err != nil {
		return err
	}

	if opts.DryRun {
		path, _ := service.AuthConfigPath()
		logger.Infof("[dry-run] would write GitHub token to %s (source=%s)", path, source)
		return nil
	}

	if err := service.SaveAuthConfig(service.AuthConfig{Token: token}); err != nil {
		return fmt.Errorf("save auth config: %w", err)
	}
	path, _ := service.AuthConfigPath()
	logger.Infof("Wrote GitHub token to %s (source=%s)", path, source)
	return nil
}

// verifyGitHubToken hits GitHub with the discovered token to confirm it's
// valid before we persist it. Two probes:
//
//  1. GET /rate_limit — cheap, authenticated, tells us if the token is
//     accepted at all. Distinguishes expired/revoked (401) from network
//     failures, scopes problems (403), and rate-limit exhaustion.
//  2. GET /user/orgs — confirms the token can list org memberships, which
//     is what the header's org chooser and `--all --org=...` searches
//     depend on. Tokens without the read:org scope succeed at /rate_limit
//     but return empty here; we don't fail on empty (user may genuinely
//     have no orgs) but we DO fail on API error (scope missing surfaces
//     as 403 or an unexpected response).
//
// --skip-verify-token short-circuits the whole check — useful for GHE
// (api.github.com hardcoded in ProbeToken) or fine-grained tokens with
// unusual scopes.
func verifyGitHubToken(token string, skip bool) error {
	if skip {
		logger.Warnf("Skipping token verification (--skip-verify-token); daemon may log auth errors if the token is invalid")
		return nil
	}

	ghOpts := github.Options{Token: token}

	probe := github.ProbeToken(ghOpts)
	switch probe.State {
	case github.AuthStateOK:
		logger.Infof("Verified GitHub token: %s", probe.Message)
	case github.AuthStateInvalid:
		return fmt.Errorf("GitHub rejected the token: %s", probe.Message)
	case github.AuthStateUnreachable:
		return fmt.Errorf("GitHub API unreachable during token verification: %s (use --skip-verify-token to bypass)", probe.Message)
	case github.AuthStateRateLimited:
		// Rate-limited means the token IS valid — just no budget to
		// confirm org listing right now. Warn and proceed so the install
		// isn't blocked waiting for a rate-limit reset.
		logger.Warnf("Token is rate-limited: %s. Skipping /user/orgs check.", probe.Message)
		return nil
	default:
		return fmt.Errorf("unexpected token probe state %q: %s", probe.State, probe.Message)
	}

	if _, err := github.FetchUserOrgs(ghOpts); err != nil {
		return fmt.Errorf("token cannot list orgs/repos — likely missing the read:org scope (use --skip-verify-token to bypass): %w", err)
	}
	return nil
}

// resolveInstallDBConfig converts install CLI flags into a DBConfig after
// enforcing the --dsn / --embedded mutex. Keeps the handler body focused on
// I/O.
func resolveInstallDBConfig(opts SystemInstallOptions) (service.DBConfig, error) {
	switch {
	case opts.DSN != "" && opts.Embedded:
		return service.DBConfig{}, errors.New("--dsn and --embedded are mutually exclusive")
	case opts.DSN != "":
		return service.DBConfig{Mode: service.DBModeDSN, DSN: opts.DSN}, nil
	default:
		// No flags (or explicit --embedded) → embedded postgres managed by
		// the daemon. This is the zero-config default so `gavel system
		// install` Just Works on a fresh machine.
		return service.DBConfig{Mode: service.DBModeEmbedded}, nil
	}
}

// verifyDBConfig exercises the selected backend at install time so
// misconfigurations fail here instead of inside the background daemon.
func verifyDBConfig(cfg service.DBConfig) error {
	switch cfg.Mode {
	case service.DBModeDSN:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		conn, err := pgx.Connect(ctx, cfg.DSN)
		if err != nil {
			return fmt.Errorf("verify DSN: %w", err)
		}
		defer conn.Close(context.Background()) //nolint:errcheck
		if err := conn.Ping(ctx); err != nil {
			return fmt.Errorf("ping postgres: %w", err)
		}
		logger.Infof("Verified connection to %s", cfg.DSN)
		return nil
	case service.DBModeEmbedded:
		dataDir, err := service.EmbeddedDataDir()
		if err != nil {
			return err
		}
		logger.Infof("Verifying embedded postgres at %s (may download binaries on first run)", dataDir)
		dsn, stop, err := commonsdb.StartEmbedded(commonsdb.EmbeddedConfig{
			DataDir:  dataDir,
			Database: "gavel",
		})
		if err != nil {
			return fmt.Errorf("embedded postgres smoke test: %w", err)
		}
		logger.Infof("Embedded postgres ready at %s; shutting down test instance", dsn)
		if err := stop(); err != nil {
			return fmt.Errorf("stop embedded postgres: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unknown db mode %q", cfg.Mode)
	}
}
