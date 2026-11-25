package proposal

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// ErrConflict is a proposal whose patch no longer applies to the working tree or index.
var ErrConflict = errors.New("proposal does not apply")

// Stat lists the files base..commit changes with their line counts.
func Stat(ctx context.Context, dir, base, commit string) ([]FileStat, error) {
	if err := checkCommits(base, commit); err != nil {
		return nil, err
	}
	out, err := call{dir: dir}.run(ctx, "diff", "--numstat", "-z", "--no-renames", base, commit, "--")
	if err != nil {
		return nil, fmt.Errorf("proposal: stat %s..%s: %w", base, commit, err)
	}
	files := []FileStat{}
	for _, record := range strings.Split(strings.TrimSuffix(out, "\x00"), "\x00") {
		if record == "" {
			continue
		}
		file, err := parseNumstat(record)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func parseNumstat(record string) (FileStat, error) {
	fields := strings.SplitN(record, "\t", 3)
	if len(fields) != 3 {
		return FileStat{}, fmt.Errorf("proposal: malformed numstat record %q", record)
	}
	if fields[0] == "-" && fields[1] == "-" {
		return FileStat{Path: fields[2]}, nil
	}
	additions, err := strconv.Atoi(fields[0])
	if err != nil {
		return FileStat{}, fmt.Errorf("proposal: malformed numstat record %q: %w", record, err)
	}
	deletions, err := strconv.Atoi(fields[1])
	if err != nil {
		return FileStat{}, fmt.Errorf("proposal: malformed numstat record %q: %w", record, err)
	}
	return FileStat{Path: fields[2], Additions: additions, Deletions: deletions}, nil
}

// Apply writes base..commit into the working tree, and the index when stage.
// git apply is atomic: a patch that does not apply cleanly changes nothing and
// is ErrConflict.
func Apply(ctx context.Context, dir, base, commit string, stage bool) error {
	if err := checkCommits(base, commit); err != nil {
		return err
	}
	top, err := topLevel(ctx, dir)
	if err != nil {
		return err
	}
	patch, err := call{dir: top}.run(ctx, "diff", "--binary", "--no-color", "--no-ext-diff", "--no-renames", base, commit, "--")
	if err != nil {
		return fmt.Errorf("proposal: diff %s..%s: %w", base, commit, err)
	}
	args := []string{"apply"}
	if stage {
		args = append(args, "--index")
	}
	_, err = call{dir: top, stdin: []byte(patch)}.run(ctx, append(args, "-")...)
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf("%w: %w", ErrConflict, err)
	}
	return err
}

var fullSHA = regexp.MustCompile(`^([0-9a-f]{40}|[0-9a-f]{64})$`)

// checkCommits accepts only the full object names Create returns.
func checkCommits(shas ...string) error {
	for _, sha := range shas {
		if !fullSHA.MatchString(sha) {
			return fmt.Errorf("proposal: invalid commit hash %q: want a full object name", sha)
		}
	}
	return nil
}
