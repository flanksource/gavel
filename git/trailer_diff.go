package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// DiffStat aggregates the change footprint of all the commits that carry a
// particular git trailer value (e.g. a todo's Gavel-Issue-Id). Files counts
// each touched path once across the commit set; Adds/Dels exclude binary files
// (git reports "-" for those) the same way `git log --numstat` does.
type DiffStat struct {
	Commits int `json:"commits"`
	Files   int `json:"files"`
	Adds    int `json:"adds"`
	Dels    int `json:"dels"`
}

// diffStatHeaderSep and diffStatFieldSep are the bytes git emits in the
// per-commit header line: \x00<hash>\x1f<trailer values>. \x00 cannot appear in
// git's own output, so it unambiguously marks a header against numstat rows. The
// format string asks for these via %x.. directives (a literal NUL in argv is
// rejected by exec); git substitutes the bytes at output time.
const (
	diffStatHeaderSep = "\x00"
	diffStatFieldSep  = "\x1f"
	diffStatValueSep  = "\x1d"
)

// TrailerDiffStats returns the aggregated diff footprint per trailer value for
// every commit (across all refs) whose message carries the given trailer key.
// One `git log --numstat` pass, pre-filtered to commits mentioning the trailer,
// yields stats for every todo in the workspace at once. The map is keyed by the
// trailer value (the todo's id); commits whose trailer git does not actually
// parse (a coincidental body line) contribute nothing.
func TrailerDiffStats(path, key string) (map[string]DiffStat, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return map[string]DiffStat{}, nil
	}

	// Pre-filter to commits whose message contains the "Key: " trailer line, then
	// let git's own trailer parser supply the value so a coincidental body mention
	// is excluded — same confirmation CommitsWithTrailer relies on.
	format := "--format=%x00%H%x1f%(trailers:key=" + key + ",valueonly,separator=%x1d)"
	cmd := exec.Command("git", "log", "--all", "--no-merges", "--numstat",
		"--fixed-strings", "--grep="+key+": ", format)
	cmd.Dir = path
	out, err := cmd.CombinedOutput()
	if err != nil {
		if isNoCommitsError(out) {
			return map[string]DiffStat{}, nil
		}
		return nil, fmt.Errorf("git log --numstat --grep %q: %w\nOutput: %s", key, err, string(out))
	}
	return parseTrailerDiffStats(out), nil
}

// trailerAgg accumulates one trailer value's stats while streaming the log,
// tracking touched files in a set so a path edited by several commits counts
// once.
type trailerAgg struct {
	commits int
	adds    int
	dels    int
	files   map[string]struct{}
}

// parseTrailerDiffStats folds the `git log --numstat` stream into per-trailer
// stats. Each header line (\x00-prefixed) sets the active trailer values; the
// numstat rows that follow are attributed to them until the next header.
func parseTrailerDiffStats(out []byte) map[string]DiffStat {
	aggs := make(map[string]*trailerAgg)
	var active []string

	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, diffStatHeaderSep) {
			active = activeTrailerValues(line, aggs)
			continue
		}
		if len(active) == 0 {
			continue
		}
		adds, dels, file, isBinary, ok := parseNumstatLine(line)
		if !ok {
			continue
		}
		for _, value := range active {
			acc := aggs[value]
			if !isBinary {
				acc.adds += adds
				acc.dels += dels
			}
			if file != "" {
				acc.files[file] = struct{}{}
			}
		}
	}

	stats := make(map[string]DiffStat, len(aggs))
	for value, acc := range aggs {
		stats[value] = DiffStat{
			Commits: acc.commits,
			Files:   len(acc.files),
			Adds:    acc.adds,
			Dels:    acc.dels,
		}
	}
	return stats
}

// activeTrailerValues parses a header line into the distinct, non-empty trailer
// values it carries, registering/incrementing each one's commit count and
// returning the set the following numstat rows belong to.
func activeTrailerValues(line string, aggs map[string]*trailerAgg) []string {
	_, rest, _ := strings.Cut(strings.TrimPrefix(line, diffStatHeaderSep), diffStatFieldSep)
	var active []string
	seen := make(map[string]struct{})
	for _, raw := range strings.Split(rest, diffStatValueSep) {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if _, dup := seen[value]; dup {
			continue
		}
		seen[value] = struct{}{}
		acc := aggs[value]
		if acc == nil {
			acc = &trailerAgg{files: make(map[string]struct{})}
			aggs[value] = acc
		}
		acc.commits++
		active = append(active, value)
	}
	return active
}
