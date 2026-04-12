package serve

import (
	"fmt"
	"os"
	"path/filepath"
)

func writePreReceiveHook(bareRepo, gavelPath string) error {
	hooksDir := filepath.Join(bareRepo, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return err
	}

	hookPath := filepath.Join(hooksDir, "pre-receive")
	script := fmt.Sprintf(`#!/bin/bash
set -euo pipefail

while read oldrev newrev refname; do
  WORKDIR=$(mktemp -d)
  trap "rm -rf $WORKDIR" EXIT

  export GIT_DIR="%s"
  git --work-tree="$WORKDIR" checkout "$newrev" -- . 2>/dev/null

  echo "============================"
  echo " gavel test --lint"
  echo "============================"
  echo ""

  %s test --lint --cwd "$WORKDIR" 2>&1
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
`, bareRepo, gavelPath)

	return os.WriteFile(hookPath, []byte(script), 0o755)
}
