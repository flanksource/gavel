package commit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/flanksource/gavel/status"
)

func readTreePatch(ctx context.Context, root string, file status.FileStatus) (string, error) {
	noIndex := file.State == status.StateUntracked
	if !noIndex {
		cmd := exec.CommandContext(ctx, "git", "rev-parse", "--verify", "--quiet", "HEAD")
		cmd.Dir = root
		if err := cmd.Run(); err != nil {
			if !gitExited(err, 1) {
				return "", fmt.Errorf("resolve HEAD for %q: %w", file.Path, err)
			}
			if file.WorkKind == status.KindDeleted {
				return "", nil
			}
			noIndex = true
		}
	}

	args := []string{"--no-optional-locks", "--literal-pathspecs", "diff", "--patch",
		"--no-color", "--no-ext-diff", "--no-textconv", "--unified=3", "--src-prefix=a/", "--dst-prefix=b/"}
	if noIndex {
		args = append(args, "--no-index", "--", os.DevNull, file.Path)
	} else {
		args = append(args, "--find-renames", "HEAD", "--", file.Path)
		if file.PreviousPath != "" {
			args = append(args, file.PreviousPath)
		}
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	// --no-index implies --exit-code. A successful comparison with a patch exits 1.
	if err != nil && !(noIndex && gitExited(err, 1) && len(out) > 0) {
		return "", fmt.Errorf("git diff %q: %w: %s", file.Path, err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}

func gitExited(err error, code int) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr) && exitErr.ExitCode() == code
}
