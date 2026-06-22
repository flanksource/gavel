package cmux

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/clicky"
)

const defaultCommandTimeout = 30 * time.Second

// Runner executes the cmux CLI. Tests inject this to assert command arguments
// without requiring a live cmux.app instance.
type Runner func(ctx context.Context, cwd, binary string, timeout time.Duration, args ...string) (stdout string, err error)

type Client struct {
	Binary  string
	Timeout time.Duration
	Runner  Runner
}

type NewWorkspaceOpts struct {
	Cwd     string
	Name    string
	Command string
	Focus   bool
}

type WorkspaceRef struct {
	WorkspaceID string
	SurfaceID   string
	Raw         string
}

func NewClient(binary string) *Client {
	return &Client{Binary: binary}
}

func (c *Client) Available(ctx context.Context) error {
	if _, err := c.run(ctx, "", "ping"); err != nil {
		return fmt.Errorf("cmux is not available or not running: %w", err)
	}
	return nil
}

func (c *Client) NewWorkspace(ctx context.Context, opts NewWorkspaceOpts) (WorkspaceRef, error) {
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		name = filepath.Base(filepath.Clean(opts.Cwd))
	}
	if name == "." || name == string(filepath.Separator) {
		name = "gavel-todos"
	}

	args := []string{
		"new-workspace",
		"--cwd", opts.Cwd,
		"--name", name,
		"--focus", strconv.FormatBool(opts.Focus),
	}
	if opts.Command != "" {
		args = append(args, "--command", opts.Command)
	}
	args = append(args, "--id-format", "both")

	stdout, err := c.run(ctx, opts.Cwd, args...)
	if err != nil {
		return WorkspaceRef{}, err
	}
	ref := ParseWorkspaceRef(stdout)
	if ref.String() == "" {
		return WorkspaceRef{}, fmt.Errorf("cmux new-workspace did not return a workspace reference: %q", strings.TrimSpace(stdout))
	}
	return ref, nil
}

func (c *Client) Send(ctx context.Context, workspaceRef, text string) error {
	if strings.TrimSpace(workspaceRef) == "" {
		return fmt.Errorf("workspace reference is required")
	}
	_, err := c.run(ctx, "", "send", "--workspace", workspaceRef, "--", text)
	return err
}

func (r WorkspaceRef) String() string {
	if r.WorkspaceID != "" {
		return r.WorkspaceID
	}
	if r.Raw != "" {
		return r.Raw
	}
	return r.SurfaceID
}

func ParseWorkspaceRef(stdout string) WorkspaceRef {
	raw := strings.TrimSpace(stdout)
	if raw == "" {
		return WorkspaceRef{}
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err == nil {
		ref := WorkspaceRef{
			WorkspaceID: firstString(m, "workspace_id", "workspaceId", "workspace", "id"),
			SurfaceID:   firstString(m, "surface_id", "surfaceId", "surface"),
			Raw:         raw,
		}
		if ref.String() != "" {
			return ref
		}
	}

	ref := WorkspaceRef{
		WorkspaceID: matchRef(raw, `(?i)workspace(?:[_ -]?id)?["'=:\s]+([A-Za-z0-9._:/-]+)`),
		SurfaceID:   matchRef(raw, `(?i)surface(?:[_ -]?id)?["'=:\s]+([A-Za-z0-9._:/-]+)`),
		Raw:         raw,
	}
	if ref.String() != "" && ref.WorkspaceID != raw {
		return ref
	}

	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			ref.Raw = line
			return ref
		}
	}
	return ref
}

func (c *Client) run(ctx context.Context, cwd string, args ...string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	binary := c.Binary
	if binary == "" {
		binary = "cmux"
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = defaultCommandTimeout
	}
	runner := c.Runner
	if runner == nil {
		runner = defaultRunner
	}
	return runner(ctx, cwd, binary, timeout, args...)
}

func defaultRunner(_ context.Context, cwd, binary string, timeout time.Duration, args ...string) (string, error) {
	proc := clicky.Exec(binary, args...).WithTimeout(timeout)
	if cwd != "" {
		proc = proc.WithCwd(cwd)
	}
	result := proc.Run().Result()
	if result.Error != nil || result.ExitCode != 0 {
		msg := strings.TrimSpace(result.Stderr)
		if msg == "" {
			msg = strings.TrimSpace(result.Stdout)
		}
		if msg == "" && result.Error != nil {
			msg = result.Error.Error()
		}
		return result.Stdout, fmt.Errorf("%s %s failed (exit %d): %s", binary, strings.Join(args, " "), result.ExitCode, msg)
	}
	return result.Stdout, nil
}

func matchRef(s, pattern string) string {
	m := regexp.MustCompile(pattern).FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return strings.Trim(m[1], `"'`)
}
