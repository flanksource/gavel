package serve

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/commons/logger"
)

// writePostReceiveHook generates the bash post-receive hook script for a
// bare repo and writes it to disk. The script is intentionally minimal:
// clone/checkout the push into a worktree, then either exec the pushed
// repo's `ssh.cmd` (read from $WORKDIR/.gavel.yaml via yq) or fall back to
// `gavel test --lint`. Pre/post hooks are NOT rendered here — `gavel test`
// itself runs them when CI=1 is in the environment (which this script
// exports). See cmd/gavel/test.go:runTests for the --skip-hooks logic.
//
// The server is now config-agnostic: it does NOT read .gavel.yaml at
// startup and does NOT bake per-project commands into the hook. The
// pushed repo's own config drives the push pipeline.
func writePostReceiveHook(bareRepo, gavelPath string) error {
	hooksDir := filepath.Join(bareRepo, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return err
	}

	hookPath := filepath.Join(hooksDir, "post-receive")
	script := renderHookScript(bareRepo, gavelPath)

	logger.V(2).Infof("Writing post-receive hook to %s", hookPath)
	logger.V(3).Infof("Hook script:\n%s", script)
	return os.WriteFile(hookPath, []byte(script), 0o755)
}

func renderHookScript(bareRepo, gavelPath string) string {
	var debugBlock string
	if logger.V(1).Enabled() {
		debugBlock = `
  echo "[debug] worktree contents:"
  find "$WORKDIR" -not -path '*/.git/*' -not -path '*/.git' | sort | head -50
  echo ""
  echo "[debug] last commit:"
  git -C "$WORKDIR" log -1 --oneline
  echo ""
`
	}

	var buf bytes.Buffer
	fmt.Fprintln(&buf, `#!/bin/bash`)
	fmt.Fprintln(&buf, `set -euo pipefail`)
	fmt.Fprintln(&buf)
	// Fail loud if yq is missing instead of silently running the default
	// main command. yq is used to extract ssh.cmd from .gavel.yaml.
	fmt.Fprintln(&buf, `command -v yq >/dev/null || { echo "error: yq not found on $PATH (needed to parse .gavel.yaml)" >&2; exit 1; }`)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, `while read oldrev newrev refname; do`)
	fmt.Fprintln(&buf, `  WORKDIR=$(mktemp -d)`)
	fmt.Fprintln(&buf, `  trap "rm -rf $WORKDIR" EXIT`)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, `  unset GIT_DIR`)
	fmt.Fprintf(&buf, "  git clone --no-checkout %q \"$WORKDIR\" 2>/dev/null\n", bareRepo)
	fmt.Fprintln(&buf, `  git -C "$WORKDIR" checkout "$newrev" 2>/dev/null`)
	if debugBlock != "" {
		buf.WriteString(debugBlock)
	}
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, `  SSH_CMD=""`)
	fmt.Fprintln(&buf, `  if [ -f "$WORKDIR/.gavel.yaml" ]; then`)
	fmt.Fprintln(&buf, `    SSH_CMD=$(yq -r '.ssh.cmd // ""' "$WORKDIR/.gavel.yaml")`)
	fmt.Fprintln(&buf, `  fi`)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, `  export CI=1`)
	fmt.Fprintln(&buf, `  set +e`)
	fmt.Fprintln(&buf, `  if [ -n "$SSH_CMD" ]; then`)
	fmt.Fprintln(&buf, `    echo "============================"`)
	fmt.Fprintln(&buf, `    echo " ssh.cmd: $SSH_CMD"`)
	fmt.Fprintln(&buf, `    echo "============================"`)
	fmt.Fprintln(&buf, `    (cd "$WORKDIR" && eval "$SSH_CMD") 2>&1`)
	fmt.Fprintln(&buf, `  else`)
	fmt.Fprintln(&buf, `    echo "============================"`)
	fmt.Fprintf(&buf, "    echo \" %s test --lint --ui --addr 0.0.0.0 --cwd $WORKDIR\"\n", gavelPath)
	fmt.Fprintln(&buf, `    echo "============================"`)
	fmt.Fprintf(&buf, "    %s test --lint --ui --addr 0.0.0.0 --no-progress --cwd \"$WORKDIR\" 2>&1\n", gavelPath)
	fmt.Fprintln(&buf, `  fi`)
	fmt.Fprintln(&buf, `  EXIT=$?`)
	fmt.Fprintln(&buf, `  set -e`)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, `  rm -rf "$WORKDIR"`)
	fmt.Fprintln(&buf, `  trap - EXIT`)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, `  if [ $EXIT -ne 0 ]; then`)
	fmt.Fprintln(&buf, `    echo ""`)
	fmt.Fprintln(&buf, `    echo "============================"`)
	fmt.Fprintln(&buf, `    echo " FAILED"`)
	fmt.Fprintln(&buf, `    echo "============================"`)
	fmt.Fprintln(&buf, `    exit $EXIT`)
	fmt.Fprintln(&buf, `  fi`)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, `  echo ""`)
	fmt.Fprintln(&buf, `  echo "============================"`)
	fmt.Fprintln(&buf, `  echo " PASSED"`)
	fmt.Fprintln(&buf, `  echo "============================"`)
	fmt.Fprintln(&buf, `done`)

	return buf.String()
}
