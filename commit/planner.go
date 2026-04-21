package commit

import (
	"fmt"
	"sort"
	"strings"

	"github.com/flanksource/commons/logger"
)

type commitGroup struct {
	Label   string
	Changes []stagedChange
}

func (g commitGroup) Files() []string {
	files := make([]string, 0, len(g.Changes))
	for _, change := range g.Changes {
		files = append(files, change.Path)
	}
	return files
}

func (g commitGroup) GitPaths() []string {
	seen := make(map[string]struct{}, len(g.Changes)*2)
	var paths []string
	for _, change := range g.Changes {
		for _, path := range change.GitPaths() {
			if _, ok := seen[path]; ok {
				continue
			}
			seen[path] = struct{}{}
			paths = append(paths, path)
		}
	}
	return paths
}

func (g commitGroup) diff() string {
	var patches []string
	for _, change := range g.Changes {
		patches = append(patches, strings.TrimRight(change.Patch, "\n"))
	}
	if len(patches) == 0 {
		return ""
	}
	return strings.Join(patches, "\n") + "\n"
}

func (g commitGroup) labelOrDefault() string {
	if g.Label != "" {
		return g.Label
	}
	if len(g.Changes) == 1 {
		return g.Changes[0].Path
	}
	return fmt.Sprintf("%d files", len(g.Changes))
}

const rootGroupLabel = "root"

// groupChangesByDir buckets changes by top-level directory, then recursively
// subdivides any bucket that exceeds maxFiles or maxLines. Values <= 0 disable
// the corresponding budget. Output is sorted by label for deterministic order.
func groupChangesByDir(changes []stagedChange, maxFiles, maxLines int) []commitGroup {
	if len(changes) == 0 {
		return nil
	}

	buckets := bucketByDepth(changes, 1)
	keys := make([]string, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var out []commitGroup
	for _, key := range keys {
		out = append(out, splitBucket(buckets[key], key, 1, maxFiles, maxLines)...)
	}
	return out
}

func splitBucket(changes []stagedChange, label string, depth, maxFiles, maxLines int) []commitGroup {
	if len(changes) == 0 {
		return nil
	}

	totalLines := countBudgetedLines(changes)

	fitsFiles := maxFiles <= 0 || len(changes) <= maxFiles
	fitsLines := maxLines <= 0 || totalLines <= maxLines
	if fitsFiles && fitsLines {
		return []commitGroup{{Label: label, Changes: changes}}
	}

	buckets := bucketByDepth(changes, depth+1)
	hasSubdir := false
	for _, bucket := range buckets {
		if len(bucket) > 1 {
			hasSubdir = true
			break
		}
	}
	if len(buckets) <= 1 || !hasSubdir {
		logger.Warnf("commit group %q exceeds budget (files=%d, lines=%d) and cannot be subdivided further",
			label, len(changes), totalLines)
		return []commitGroup{{Label: label, Changes: changes}}
	}

	keys := make([]string, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var out []commitGroup
	for _, key := range keys {
		out = append(out, splitBucket(buckets[key], key, depth+1, maxFiles, maxLines)...)
	}
	return out
}

// countBudgetedLines sums Adds+Dels for the --max-lines budget, skipping
// newly-inserted files so adding a large generated/vendored/first-party file
// does not push an otherwise-small group over the limit.
func countBudgetedLines(changes []stagedChange) int {
	total := 0
	for _, c := range changes {
		if c.Status == "inserted" {
			continue
		}
		total += c.Adds + c.Dels
	}
	return total
}

func bucketByDepth(changes []stagedChange, depth int) map[string][]stagedChange {
	buckets := make(map[string][]stagedChange)
	for _, change := range changes {
		key := pathPrefix(change.Path, depth)
		buckets[key] = append(buckets[key], change)
	}
	return buckets
}

// pathPrefix joins the first `depth` path segments. Top-level files collapse
// into rootGroupLabel so they share a single bucket; paths shorter than depth
// return the full path so distinct leaves never collide.
func pathPrefix(path string, depth int) string {
	segments := strings.Split(path, "/")
	if len(segments) == 1 && depth == 1 {
		return rootGroupLabel
	}
	if len(segments) <= depth {
		return path
	}
	return strings.Join(segments[:depth], "/")
}
