package serve

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
)

// HookStep is a named shell command rendered as a pre/post block in the
// post-receive hook.
type HookStep struct {
	Name string
	Run  string
}

// HookSpec is the resolved set of commands the post-receive hook should run
// for an incoming push. The default (zero-value) main command is
// `gavel test --lint`.
type HookSpec struct {
	Pre  []HookStep
	Cmd  string // overrides the default `gavel test --lint`
	Post []HookStep
}

// HookWriterFunc is the pluggable signature used by Server.HookWriter so
// tests can swap the post-receive hook generator.
type HookWriterFunc func(bareRepo, gavelPath string) error

// NewHookWriter returns a HookWriterFunc that renders the post-receive hook
// from the given spec. The default (zero-value) spec preserves today's
// `gavel test --lint` behavior.
func NewHookWriter(spec HookSpec) HookWriterFunc {
	return func(bareRepo, gavelPath string) error {
		return writePostReceiveHookFromSpec(bareRepo, gavelPath, spec)
	}
}

// writePostReceiveHook preserves the original zero-config entry point used
// by Server when no custom HookWriter is supplied.
func writePostReceiveHook(bareRepo, gavelPath string) error {
	return writePostReceiveHookFromSpec(bareRepo, gavelPath, HookSpec{})
}

func writePostReceiveHookFromSpec(bareRepo, gavelPath string, spec HookSpec) error {
	hooksDir := filepath.Join(bareRepo, "hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		return err
	}

	hookPath := filepath.Join(hooksDir, "post-receive")
	script := renderHookScript(bareRepo, gavelPath, spec)

	logger.V(2).Infof("Writing post-receive hook to %s", hookPath)
	logger.V(3).Infof("Hook script:\n%s", script)
	return os.WriteFile(hookPath, []byte(script), 0o755)
}

func renderHookScript(bareRepo, gavelPath string, spec HookSpec) string {
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

	mainCmd := resolveMainCommand(spec.Cmd, gavelPath)

	var buf bytes.Buffer
	fmt.Fprintln(&buf, `#!/bin/bash`)
	fmt.Fprintln(&buf, `set -euo pipefail`)
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

	writeHookSteps(&buf, "pre", spec.Pre)

	fmt.Fprintln(&buf, `  echo "============================"`)
	fmt.Fprintf(&buf, "  echo \" %s\"\n", shellEscapeEcho(mainCmd))
	fmt.Fprintln(&buf, `  echo "============================"`)
	fmt.Fprintln(&buf, `  echo ""`)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, `  set +e`)
	fmt.Fprintf(&buf, "  (cd \"$WORKDIR\" && %s) 2>&1\n", mainCmd)
	fmt.Fprintln(&buf, `  MAIN_EXIT=$?`)
	fmt.Fprintln(&buf, `  set -e`)
	fmt.Fprintln(&buf)

	writeHookSteps(&buf, "post", spec.Post)

	fmt.Fprintln(&buf, `  rm -rf "$WORKDIR"`)
	fmt.Fprintln(&buf, `  trap - EXIT`)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, `  if [ $MAIN_EXIT -ne 0 ]; then`)
	fmt.Fprintln(&buf, `    echo ""`)
	fmt.Fprintln(&buf, `    echo "============================"`)
	fmt.Fprintln(&buf, `    echo " FAILED"`)
	fmt.Fprintln(&buf, `    echo "============================"`)
	fmt.Fprintln(&buf, `    exit $MAIN_EXIT`)
	fmt.Fprintln(&buf, `  fi`)
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, `  echo ""`)
	fmt.Fprintln(&buf, `  echo "============================"`)
	fmt.Fprintln(&buf, `  echo " PASSED"`)
	fmt.Fprintln(&buf, `  echo "============================"`)
	fmt.Fprintln(&buf, `done`)

	return buf.String()
}

// resolveMainCommand returns the bash snippet for the main command.
//
//   - Empty Cmd: the default `gavel test --lint` invocation (backwards compatible).
//   - Cmd starting with `gavel ` is rewritten to use the actual binary path, and
//     `--cwd "$WORKDIR"` / `--no-progress` are appended when absent.
//   - Any other Cmd runs literally (e.g. `make ci`); the outer wrapper already
//     `cd`s into $WORKDIR.
func resolveMainCommand(cmd, gavelPath string) string {
	if cmd == "" {
		return fmt.Sprintf(`%s test --lint --no-progress --cwd "$WORKDIR"`, gavelPath)
	}
	if strings.HasPrefix(cmd, "gavel ") {
		cmd = gavelPath + strings.TrimPrefix(cmd, "gavel")
	}
	if strings.HasPrefix(cmd, gavelPath+" ") || cmd == gavelPath {
		if !strings.Contains(cmd, "--cwd") {
			cmd += ` --cwd "$WORKDIR"`
		}
		if !strings.Contains(cmd, "--no-progress") {
			cmd += " --no-progress"
		}
	}
	return cmd
}

// writeHookSteps renders a series of steps with a banner per step. Steps run
// in a subshell rooted at $WORKDIR. A failing pre-step aborts the push
// (inherited from `set -e`); a failing post-step is logged but does not mask
// the main exit code.
func writeHookSteps(buf *bytes.Buffer, phase string, steps []HookStep) {
	if len(steps) == 0 {
		return
	}
	for _, step := range steps {
		if step.Run == "" {
			continue
		}
		label := step.Name
		if label == "" {
			label = phase
		}
		fmt.Fprintln(buf, `  echo "----------------------------"`)
		fmt.Fprintf(buf, "  echo \" %s: %s\"\n", phase, shellEscapeEcho(label))
		fmt.Fprintln(buf, `  echo "----------------------------"`)
		if phase == "post" {
			// Post-steps never mask main failure.
			fmt.Fprintln(buf, `  set +e`)
			fmt.Fprintf(buf, "  (cd \"$WORKDIR\" && %s) 2>&1\n", step.Run)
			fmt.Fprintln(buf, `  set -e`)
		} else {
			fmt.Fprintf(buf, "  (cd \"$WORKDIR\" && %s) 2>&1\n", step.Run)
		}
		fmt.Fprintln(buf)
	}
}

// shellEscapeEcho escapes a string for use inside a double-quoted bash
// `echo` argument. Only `"`, `\`, `$`, and backtick need to be handled.
func shellEscapeEcho(s string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		`$`, `\$`,
		"`", "\\`",
	)
	return r.Replace(s)
}
