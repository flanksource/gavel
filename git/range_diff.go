package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// DiffStat is the change footprint of a run's work: the commits it made, and
// the files and lines they changed. Files counts each touched path once;
// Adds/Dels exclude binary files (git reports "-" for those).
type DiffStat struct {
	Commits int `json:"commits"`
	Files   int `json:"files"`
	Adds    int `json:"adds"`
	Dels    int `json:"dels"`
}

// RangeDiffStat is the footprint of from..to in the repository at path: the
// commits reachable from to but not from (`git rev-list --count`), and the net
// file/line change between the two trees (`git diff --numstat`). Both shas are
// validated before git runs; a sha git cannot resolve is an error.
func RangeDiffStat(path, from, to string) (DiffStat, error) {
	if !IsValidCommitHash(from) {
		return DiffStat{}, fmt.Errorf("invalid commit hash %q", from)
	}
	if !IsValidCommitHash(to) {
		return DiffStat{}, fmt.Errorf("invalid commit hash %q", to)
	}
	count, err := gitOutput(path, "rev-list", "--count", from+".."+to)
	if err != nil {
		return DiffStat{}, err
	}
	commits, err := strconv.Atoi(strings.TrimSpace(string(count)))
	if err != nil {
		return DiffStat{}, fmt.Errorf("git rev-list --count %s..%s in %s: unexpected output %q", from, to, path, count)
	}
	numstat, err := gitOutput(path, "diff", "--numstat", from, to, "--")
	if err != nil {
		return DiffStat{}, err
	}
	stat := DiffStat{Commits: commits}
	scanner := bufio.NewScanner(bytes.NewReader(numstat))
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		adds, dels, _, isBinary, ok := parseNumstatLine(scanner.Text())
		if !ok {
			continue
		}
		stat.Files++
		if !isBinary {
			stat.Adds += adds
			stat.Dels += dels
		}
	}
	if err := scanner.Err(); err != nil {
		return DiffStat{}, fmt.Errorf("read git diff --numstat %s %s in %s: %w", from, to, path, err)
	}
	return stat, nil
}

func gitOutput(path string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		var stderr []byte
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = exitErr.Stderr
		}
		return nil, fmt.Errorf("git %s in %s: %w\n%s", strings.Join(args, " "), path, err, stderr)
	}
	return out, nil
}
