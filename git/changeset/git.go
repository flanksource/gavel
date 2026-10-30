package changeset

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

func gitOutput(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %v: %w: %s; ensure the required refs and commit history are checked out", args, err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func gitText(ctx context.Context, dir string, args ...string) (string, error) {
	out, err := gitOutput(ctx, dir, args...)
	return strings.TrimSuffix(string(out), "\n"), err
}

var commitRefPattern = regexp.MustCompile(`^[^-\x00-\x1f\x7f][^\x00-\x1f\x7f]*$`)

func commit(ctx context.Context, dir, ref string) (string, error) {
	if !commitRefPattern.MatchString(ref) {
		return "", fmt.Errorf("invalid commit reference %q", ref)
	}
	return gitText(ctx, dir, "rev-parse", "--verify", "--end-of-options", ref+"^{commit}")
}

func (s *ChangeSet) collect(ctx context.Context, args ...string) error {
	out, err := gitOutput(ctx, s.Root, args...)
	if err != nil {
		return err
	}
	fields := strings.Split(strings.TrimSuffix(string(out), "\x00"), "\x00")
	for i := 0; len(out) > 0 && i < len(fields); i++ {
		status := fields[i]
		if status == "" || i+1 >= len(fields) {
			return fmt.Errorf("invalid name-status record %q", status)
		}
		i++
		change := Change{Path: fields[i]}
		switch status[0] {
		case 'A':
			change.Status = "added"
		case 'M', 'T':
			change.Status = "modified"
		case 'D':
			change.Status = "deleted"
		case 'R', 'C':
			if i+1 >= len(fields) {
				return fmt.Errorf("missing destination for %q", change.Path)
			}
			i++
			change.PreviousPath, change.Path = change.Path, fields[i]
			change.Status = "renamed"
			if status[0] == 'C' {
				change.Status = "copied"
			}
		default:
			return fmt.Errorf("unsupported or unresolved Git status %q for %q", status, change.Path)
		}
		s.Changes = append(s.Changes, change)
	}
	return nil
}
