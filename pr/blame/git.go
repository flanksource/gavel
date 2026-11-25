package blame

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/bluekeyes/go-gitdiff/gitdiff"
	"github.com/hmarr/codeowners"
)

type repository struct {
	ctx context.Context
	dir string
}

func (r repository) git(args ...string) (string, error) {
	cmd := exec.CommandContext(r.ctx, "git", args...)
	cmd.Dir = r.dir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

func (r *repository) comparison(opts Options) (*Result, []*gitdiff.File, error) {
	root, err := r.git("rev-parse", "--show-toplevel")
	if err != nil {
		return nil, nil, err
	}
	r.dir = strings.TrimSpace(root)
	shallow, err := r.git("rev-parse", "--is-shallow-repository")
	if err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(shallow) == "true" {
		return nil, nil, fmt.Errorf("reviewer history is incomplete in a shallow repository; fetch full history first")
	}
	result := &Result{HistoryLimit: opts.HistoryLimit, Files: []string{}, Reviewers: []Reviewer{}}
	for _, revision := range []struct {
		ref    string
		target *string
	}{{opts.Base, &result.Base}, {opts.Head, &result.Head}} {
		resolved, err := r.git("rev-parse", "--verify", "--end-of-options", revision.ref+"^{commit}")
		if err != nil {
			return nil, nil, err
		}
		*revision.target = strings.TrimSpace(resolved)
	}
	mergeBase, err := r.git("merge-base", result.Base, result.Head)
	if err != nil {
		return nil, nil, err
	}
	result.MergeBase = strings.TrimSpace(mergeBase)
	patch, err := r.git("diff", "--no-ext-diff", "--no-textconv", "--no-color", "--find-renames", "--unified=0", "--inter-hunk-context=0", "--src-prefix=a/", "--dst-prefix=b/", result.MergeBase, result.Head, "--")
	if err != nil {
		return nil, nil, err
	}
	files, _, err := gitdiff.Parse(strings.NewReader(patch))
	return result, files, err
}

func (r repository) owners(base string) (codeowners.Ruleset, string, error) {
	for _, path := range []string{".github/CODEOWNERS", "CODEOWNERS", "docs/CODEOWNERS"} {
		entry, err := r.git("ls-tree", base, "--", path)
		if err != nil {
			return nil, "", err
		}
		if entry == "" {
			continue
		}
		if !strings.HasPrefix(entry, "100644 blob ") && !strings.HasPrefix(entry, "100755 blob ") {
			return nil, "", fmt.Errorf("%s must be a regular CODEOWNERS file", path)
		}
		content, err := r.git("show", base+":"+path)
		if err != nil {
			return nil, "", err
		}
		if len(content) >= 3<<20 {
			return nil, "", fmt.Errorf("%s exceeds GitHub's 3 MB CODEOWNERS limit", path)
		}
		rules, err := codeowners.ParseFile(strings.NewReader(content))
		if err != nil {
			return nil, "", fmt.Errorf("parse %s: %w", path, err)
		}
		return rules, path, nil
	}
	return nil, "", nil
}

type author struct {
	Identity
	commit      string
	committedAt int64
}

func (r repository) history(base, path string, opts Options) ([]author, error) {
	args := []string{"log", "--follow", fmt.Sprintf("-n%d", opts.HistoryLimit), "--format=%H%x00%an%x00%ae%x00%ct"}
	if !opts.Since.IsZero() {
		args = append(args, "--since-as-filter="+opts.Since.UTC().Format(time.RFC3339))
	}
	out, err := r.git(append(args, base, "--", ":(literal)"+path)...)
	if err != nil {
		return nil, err
	}
	var authors []author
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\x00")
		if len(parts) != 4 || parts[0] == "" || parts[2] == "" {
			return nil, fmt.Errorf("invalid git history record for %s: %q", path, line)
		}
		committedAt, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid commit date for %s: %w", path, err)
		}
		authors = append(authors, author{Identity: Identity{Name: parts[1], Email: parts[2]}, commit: parts[0], committedAt: committedAt})
	}
	return authors, nil
}

func (r repository) blame(base, path string) ([]author, error) {
	out, err := r.git("blame", "--line-porcelain", base, "--", path)
	if err != nil {
		return nil, err
	}
	var lines []author
	var current author
	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "author "):
			current.Name = strings.TrimPrefix(line, "author ")
		case strings.HasPrefix(line, "author-mail "):
			current.Email = strings.Trim(strings.TrimPrefix(line, "author-mail "), "<>")
		case strings.HasPrefix(line, "committer-time "):
			current.committedAt, err = strconv.ParseInt(strings.TrimPrefix(line, "committer-time "), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid blame commit date for %s: %w", path, err)
			}
		case strings.HasPrefix(line, "\t"):
			if current.commit == "" || current.Email == "" {
				return nil, fmt.Errorf("incomplete git blame identity for %s", path)
			}
			lines = append(lines, current)
		case len(strings.Fields(line)) == 4:
			parts := strings.Fields(line)
			if len(parts[0]) == 40 || len(parts[0]) == 64 {
				current = author{commit: parts[0]}
			}
		}
	}
	return lines, nil
}
