package commit

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/mod/semver"
)

// ErrNoTaggedVersions is returned when `go list -m -versions <module>` reports
// no semver tags for the module. Callers fall back to re-prompting rather
// than synthesizing a pseudo-version.
var ErrNoTaggedVersions = errors.New("module has no tagged versions")

// lookupLatestGoVersion resolves the highest semver tag for a Go module via
// `go list -m -versions <module>`. Swappable so tests can inject a fake.
var lookupLatestGoVersion = realLatestGoVersion

// goVersionsCommand runs `go list -m -versions <module>` and returns stdout.
// Swappable so tests can supply canned output without an actual go toolchain
// or network.
var goVersionsCommand = func(ctx context.Context, module string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-versions", module)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("`go` not on PATH: %w", err)
		}
		return nil, fmt.Errorf("go list -m -versions %s: %w: %s", module, err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

// runGoModTidy runs `go mod tidy` in modDir. Swappable for tests.
var runGoModTidy = func(modDir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = modDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func realLatestGoVersion(ctx context.Context, module string) (string, error) {
	out, err := goVersionsCommand(ctx, module)
	if err != nil {
		return "", err
	}
	versions := parseGoVersionsOutput(string(out))
	if len(versions) == 0 {
		return "", ErrNoTaggedVersions
	}
	return pickLatestVersion(versions), nil
}

// parseGoVersionsOutput parses `go list -m -versions` stdout. The first
// whitespace-separated token is the module path; remaining tokens are the
// version list. Tokens that are not valid semver are dropped.
func parseGoVersionsOutput(stdout string) []string {
	fields := strings.Fields(stdout)
	if len(fields) <= 1 {
		return nil
	}
	out := make([]string, 0, len(fields)-1)
	for _, f := range fields[1:] {
		if semver.IsValid(f) {
			out = append(out, f)
		}
	}
	return out
}

// pickLatestVersion returns the highest version, preferring non-prerelease
// tags when any exist. Assumes the input slice has been filtered through
// semver.IsValid.
func pickLatestVersion(versions []string) string {
	var stable, pre []string
	for _, v := range versions {
		if semver.Prerelease(v) == "" {
			stable = append(stable, v)
		} else {
			pre = append(pre, v)
		}
	}
	pool := stable
	if len(pool) == 0 {
		pool = pre
	}
	best := pool[0]
	for _, v := range pool[1:] {
		if semver.Compare(v, best) > 0 {
			best = v
		}
	}
	return best
}
