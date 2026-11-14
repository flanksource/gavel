package proposal

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type call struct {
	dir   string
	env   []string
	stdin []byte
}

func (c call) run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", c.dir, "--literal-pathspecs"}, args...)...)
	cmd.Env = append(append(os.Environ(), "GIT_TERMINAL_PROMPT=0"), c.env...)
	if c.stdin != nil {
		cmd.Stdin = bytes.NewReader(c.stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s in %s: %w: %s", strings.Join(args, " "), c.dir, err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func (c call) trimmed(ctx context.Context, args ...string) (string, error) {
	out, err := c.run(ctx, args...)
	return strings.TrimSpace(out), err
}

func topLevel(ctx context.Context, dir string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("proposal: repository directory is required")
	}
	return call{dir: dir}.trimmed(ctx, "rev-parse", "--show-toplevel")
}

func checkRef(ctx context.Context, dir, ref string) error {
	if !strings.HasPrefix(ref, "refs/") {
		return fmt.Errorf("proposal: invalid ref %q: must start with refs/", ref)
	}
	if _, err := (call{dir: dir}).run(ctx, "check-ref-format", ref); err != nil {
		return fmt.Errorf("proposal: invalid ref %q: %w", ref, err)
	}
	return nil
}
