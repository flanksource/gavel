package procfile

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// shellCaptureTimeout caps how long a login shell may take to print its
// environment, so a misbehaving rc file can't wedge supervisor startup.
const shellCaptureTimeout = 5 * time.Second

// cwdCoupledVars are dropped from the captured login-shell environment so a
// supervised process's inherited PWD/SHLVL don't contradict clicky's
// WithCwd(root) and the supervisor's own shell-nesting depth.
var cwdCoupledVars = []string{"PWD", "OLDPWD", "SHLVL", "_"}

// shellForHome resolves the login shell for a home directory. It is a package
// var so tests can substitute a deterministic fake shell.
var shellForHome = loginShell

// LoadUserEnv detects the home directory containing workDir, captures that
// home's login-shell environment, and returns it with the cwd-coupled vars
// removed. This is the lowest-priority layer of a supervised process's
// environment: it supplies the developer's real PATH (and the rest of their
// shell env) even when the supervisor was launched by a launchd/systemd
// service that stripped the interactive shell environment.
func LoadUserEnv(workDir string) (map[string]string, error) {
	home := homeForPath(workDir)
	if home == "" {
		return nil, fmt.Errorf("could not resolve a home directory for %q", workDir)
	}
	env, err := captureShellEnv(shellForHome(home))
	if err != nil {
		return nil, err
	}
	for _, k := range cwdCoupledVars {
		delete(env, k)
	}
	return env, nil
}

// homeForPath derives the user home directory that contains dir by matching it
// against the known home parents (/Users on macOS, /home on Linux, plus the
// parent of the running user's home) and returning <parent>/<first-segment>.
// For /Users/moshe/go/src/github.com/flanksource/clicky-ui this is
// /Users/moshe. When dir lives outside every known home parent it falls back
// to the running user's home.
func homeForPath(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	abs = filepath.Clean(abs)
	for _, parent := range homeParents() {
		if name, ok := firstSegmentUnder(abs, parent); ok {
			return filepath.Join(parent, name)
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return ""
}

// homeParents returns the directories under which user homes live. The parent
// of the running user's home keeps this portable across platforms; /Users and
// /home are added so a root-owned daemon still resolves a developer home from
// the workspace path.
func homeParents() []string {
	var parents []string
	seen := map[string]bool{}
	add := func(p string) {
		if p == "" || p == string(filepath.Separator) || seen[p] {
			return
		}
		seen[p] = true
		parents = append(parents, p)
	}
	if home, err := os.UserHomeDir(); err == nil {
		add(filepath.Dir(home))
	}
	add("/Users")
	add("/home")
	return parents
}

// firstSegmentUnder returns the first path segment of path below parent (the
// username) when path is strictly under parent. parent itself is not a match.
func firstSegmentUnder(path, parent string) (string, bool) {
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return "", false
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	if seg, _, _ := strings.Cut(rel, string(filepath.Separator)); seg != "" {
		return seg, true
	}
	return "", false
}

// loginShell returns the login shell configured for the user that owns home,
// via the platform user database (dscl on macOS, getent on Linux), falling
// back to $SHELL then /bin/sh.
func loginShell(home string) string {
	if sh := lookupUserShell(filepath.Base(home)); sh != "" {
		return sh
	}
	if sh := strings.TrimSpace(os.Getenv("SHELL")); sh != "" {
		return sh
	}
	return "/bin/sh"
}

// lookupUserShell reads name's login shell from the platform user database.
// An empty return means "unknown", and the caller falls back to $SHELL.
func lookupUserShell(name string) string {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("dscl", ".", "-read", "/Users/"+name, "UserShell").Output()
		if err != nil {
			return ""
		}
		// "UserShell: /bin/zsh"
		return lastField(string(out))
	case "linux":
		out, err := exec.Command("getent", "passwd", name).Output()
		if err != nil {
			return ""
		}
		// "name:x:1000:1000:gecos:/home/name:/bin/bash"
		if fields := strings.Split(strings.TrimSpace(string(out)), ":"); len(fields) >= 7 {
			return fields[6]
		}
	}
	return ""
}

func lastField(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1]
}

// captureShellEnv runs shell as a login+interactive shell and parses the
// environment it prints. Login+interactive applies both the profile files
// (.zprofile/.bash_profile) and the rc files (.zshrc/.bashrc), where
// developers put the PATH edits for tools like nvm/asdf/Homebrew. Output is
// captured to a buffer (never the terminal) and stdin is left at os/exec's
// /dev/null default so an rc file that reads input gets EOF instead of
// blocking. A non-zero exit is tolerated as long as some environment was
// printed, since rc files routinely end non-zero.
func captureShellEnv(shell string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), shellCaptureTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, shell, "-l", "-i", "-c", "env")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = io.Discard

	runErr := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("login shell %s did not print its environment within %s", shell, shellCaptureTimeout)
	}
	env := parseEnvOutput(&stdout)
	if len(env) == 0 && runErr != nil {
		return nil, fmt.Errorf("capture environment from %s: %w", shell, runErr)
	}
	return env, nil
}

var envKeyLine = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// parseEnvOutput parses `env` output into a map. Each line beginning with a
// shell-identifier KEY= starts a new entry; any other line is treated as the
// continuation of a previous value's embedded newline and ignored, so a stray
// multi-line value can't corrupt later keys (PATH, the value we care about,
// never contains newlines).
func parseEnvOutput(r io.Reader) map[string]string {
	env := map[string]string{}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !envKeyLine.MatchString(line) {
			continue
		}
		key, val, _ := strings.Cut(line, "=")
		env[key] = val
	}
	return env
}
