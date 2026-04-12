package serve

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flanksource/commons/logger"
)

func writePostReceiveHook(bareRepo, gavelPath string) error {
	hooksDir := filepath.Join(bareRepo, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return err
	}

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

	hookPath := filepath.Join(hooksDir, "post-receive")
	script := fmt.Sprintf(`#!/bin/bash
set -euo pipefail

while read oldrev newrev refname; do
  WORKDIR=$(mktemp -d)
  trap "rm -rf $WORKDIR" EXIT

  unset GIT_DIR
  git clone --no-checkout "%s" "$WORKDIR" 2>/dev/null
  git -C "$WORKDIR" checkout "$newrev" 2>/dev/null
%s
  echo "============================"
  echo " gavel test --lint"
  echo "============================"
  echo ""

  %s test --lint --no-progress --cwd "$WORKDIR" 2>&1
  EXIT=$?

  rm -rf "$WORKDIR"
  trap - EXIT

  if [ $EXIT -ne 0 ]; then
    echo ""
    echo "============================"
    echo " FAILED"
    echo "============================"
    exit 1
  fi

  echo ""
  echo "============================"
  echo " PASSED"
  echo "============================"
done
`, bareRepo, debugBlock, gavelPath)

	logger.V(2).Infof("Writing post-receive hook to %s", hookPath)
	logger.V(3).Infof("Hook script:\n%s", script)
	return os.WriteFile(hookPath, []byte(script), 0o755)
}
