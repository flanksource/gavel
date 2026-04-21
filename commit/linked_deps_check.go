package commit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/cmd/gavel/choose"
	"github.com/flanksource/gavel/utils"
	"golang.org/x/mod/modfile"
)

var ErrLinkedDepsCancelled = errors.New("commit cancelled: staged manifest references a path outside the git root")

type LinkedDepKind string

const (
	LinkedDepKindGoModReplace   LinkedDepKind = "go.mod replace"
	LinkedDepKindGoWorkUse      LinkedDepKind = "go.work use"
	LinkedDepKindPkgJSONFile    LinkedDepKind = "package.json file:"
	LinkedDepKindPkgJSONLink    LinkedDepKind = "package.json link:"
	LinkedDepKindPkgJSONPortal  LinkedDepKind = "package.json portal:"
	LinkedDepKindPkgJSONRelPath LinkedDepKind = "package.json relative path"
)

type LinkedDepViolation struct {
	File     string        // path relative to git root, e.g. "services/api/go.mod"
	Kind     LinkedDepKind // what produced the violation
	Name     string        // module or dep name
	Target   string        // raw right-hand side as written by the user
	Resolved string        // absolutized target, for display
}

type LinkedDepDecision int

const (
	LinkedDepDecisionCancel LinkedDepDecision = iota
	LinkedDepDecisionUnstage
)

type LinkedDepDecider func(ctx context.Context, v LinkedDepViolation) (LinkedDepDecision, error)

type LinkedDepsParams struct {
	WorkDir     string
	GitRoot     string
	StagedFiles []string
	Changes     []stagedChange
	Mode        string
	Decider     LinkedDepDecider
}

var (
	interactiveLinkedDepDecider = runChooseLinkedDepDecider
)

// EvaluateLinkedDeps inspects the staged blob of each manifest in stagedFiles
// and returns every local reference whose resolved filesystem target escapes
// gitRoot. stagedFiles MUST be repo-relative paths as reported by
// `git diff --cached --name-only`. deletedFiles lists paths staged for deletion;
// those manifests are skipped because deletion cannot introduce a bad
// reference.
func EvaluateLinkedDeps(workDir, gitRoot string, stagedFiles, deletedFiles []string) ([]LinkedDepViolation, error) {
	if len(stagedFiles) == 0 {
		return nil, nil
	}
	deleted := make(map[string]struct{}, len(deletedFiles))
	for _, d := range deletedFiles {
		deleted[d] = struct{}{}
	}

	var violations []LinkedDepViolation
	for _, rel := range stagedFiles {
		if _, isDel := deleted[rel]; isDel {
			continue
		}
		base := filepath.Base(rel)
		switch base {
		case "go.mod":
			vs, err := evaluateGoMod(workDir, gitRoot, rel)
			if err != nil {
				return nil, err
			}
			violations = append(violations, vs...)
		case "go.work":
			vs, err := evaluateGoWork(workDir, gitRoot, rel)
			if err != nil {
				return nil, err
			}
			violations = append(violations, vs...)
		case "package.json":
			vs, err := evaluatePackageJSON(workDir, gitRoot, rel)
			if err != nil {
				return nil, err
			}
			violations = append(violations, vs...)
		}
	}
	return violations, nil
}

// readStagedBlob returns the contents of path as staged in the index.
func readStagedBlob(workDir, path string) ([]byte, error) {
	cmd := exec.Command("git", "show", ":"+path)
	cmd.Dir = workDir
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git show :%s: %w: %s", path, err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func evaluateGoMod(workDir, gitRoot, rel string) ([]LinkedDepViolation, error) {
	blob, err := readStagedBlob(workDir, rel)
	if err != nil {
		return nil, err
	}
	f, err := modfile.Parse(rel, blob, nil)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", rel, err)
	}
	manifestDir := filepath.Join(gitRoot, filepath.Dir(rel))
	var out []LinkedDepViolation
	for _, r := range f.Replace {
		if !isLocalReplaceTarget(r.New.Path) {
			continue
		}
		resolved := resolveFrom(manifestDir, r.New.Path)
		if utils.IsWithin(resolved, gitRoot) {
			continue
		}
		out = append(out, LinkedDepViolation{
			File:     rel,
			Kind:     LinkedDepKindGoModReplace,
			Name:     r.Old.Path,
			Target:   r.New.Path,
			Resolved: resolved,
		})
	}
	return out, nil
}

func evaluateGoWork(workDir, gitRoot, rel string) ([]LinkedDepViolation, error) {
	blob, err := readStagedBlob(workDir, rel)
	if err != nil {
		return nil, err
	}
	f, err := modfile.ParseWork(rel, blob, nil)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", rel, err)
	}
	manifestDir := filepath.Join(gitRoot, filepath.Dir(rel))
	var out []LinkedDepViolation
	for _, u := range f.Use {
		resolved := resolveFrom(manifestDir, u.Path)
		if utils.IsWithin(resolved, gitRoot) {
			continue
		}
		out = append(out, LinkedDepViolation{
			File:     rel,
			Kind:     LinkedDepKindGoWorkUse,
			Name:     u.Path,
			Target:   u.Path,
			Resolved: resolved,
		})
	}
	for _, r := range f.Replace {
		if !isLocalReplaceTarget(r.New.Path) {
			continue
		}
		resolved := resolveFrom(manifestDir, r.New.Path)
		if utils.IsWithin(resolved, gitRoot) {
			continue
		}
		out = append(out, LinkedDepViolation{
			File:     rel,
			Kind:     LinkedDepKindGoModReplace,
			Name:     r.Old.Path,
			Target:   r.New.Path,
			Resolved: resolved,
		})
	}
	return out, nil
}

// isLocalReplaceTarget reports whether a go.mod replace right-hand side is a
// filesystem path rather than a module@version pair.
func isLocalReplaceTarget(target string) bool {
	if target == "" {
		return false
	}
	return strings.HasPrefix(target, ".") || strings.HasPrefix(target, "/") || filepath.IsAbs(target)
}

type pkgJSON struct {
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
}

func evaluatePackageJSON(workDir, gitRoot, rel string) ([]LinkedDepViolation, error) {
	blob, err := readStagedBlob(workDir, rel)
	if err != nil {
		return nil, err
	}
	var pkg pkgJSON
	if err := json.Unmarshal(blob, &pkg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", rel, err)
	}
	manifestDir := filepath.Join(gitRoot, filepath.Dir(rel))

	var out []LinkedDepViolation
	for _, deps := range []map[string]string{
		pkg.Dependencies, pkg.DevDependencies,
		pkg.PeerDependencies, pkg.OptionalDependencies,
	} {
		for name, raw := range deps {
			kind, path, ok := classifyPkgDep(raw)
			if !ok {
				continue
			}
			resolved := resolveFrom(manifestDir, path)
			if utils.IsWithin(resolved, gitRoot) {
				continue
			}
			out = append(out, LinkedDepViolation{
				File:     rel,
				Kind:     kind,
				Name:     name,
				Target:   raw,
				Resolved: resolved,
			})
		}
	}
	return out, nil
}

// classifyPkgDep returns the kind of local reference and the path portion if
// raw is a local filesystem reference; the boolean is false for anything else
// (semver, git+ssh, workspace:*, etc.).
func classifyPkgDep(raw string) (LinkedDepKind, string, bool) {
	switch {
	case strings.HasPrefix(raw, "file:"):
		return LinkedDepKindPkgJSONFile, strings.TrimPrefix(raw, "file:"), true
	case strings.HasPrefix(raw, "link:"):
		return LinkedDepKindPkgJSONLink, strings.TrimPrefix(raw, "link:"), true
	case strings.HasPrefix(raw, "portal:"):
		return LinkedDepKindPkgJSONPortal, strings.TrimPrefix(raw, "portal:"), true
	case strings.HasPrefix(raw, "./") || strings.HasPrefix(raw, "../"):
		return LinkedDepKindPkgJSONRelPath, raw, true
	case filepath.IsAbs(raw):
		return LinkedDepKindPkgJSONRelPath, raw, true
	}
	return "", "", false
}

func resolveFrom(manifestDir, target string) string {
	if filepath.IsAbs(target) {
		return filepath.Clean(target)
	}
	return filepath.Clean(filepath.Join(manifestDir, target))
}

// RunLinkedDepsCheck evaluates staged manifests, prompts for each violation,
// and applies the user's choice. The Mode field follows the same idiom as the
// gitignore check: "prompt" (default), "fail", or "skip"; non-TTY prompts
// without an injected Decider escalate to "fail".
func RunLinkedDepsCheck(ctx context.Context, p LinkedDepsParams) (CheckOutcome, error) {
	violations, err := EvaluateLinkedDeps(p.WorkDir, rootOrWorkDir(p), p.StagedFiles, deletedPaths(p.Changes))
	if err != nil {
		return CheckOutcome{}, err
	}
	if len(violations) == 0 {
		return CheckOutcome{}, nil
	}

	mode := p.Mode
	if mode == "" {
		mode = IgnoreCheckModePrompt
	}
	if mode == IgnoreCheckModePrompt && p.Decider == nil && !stdinIsTerminal() {
		logger.Warnf("linked-deps check: stdin is not a terminal; escalating to --linked-deps-check=fail")
		mode = IgnoreCheckModeFail
	}

	switch mode {
	case IgnoreCheckModeSkip:
		for _, v := range violations {
			logger.Warnf("linked-deps check skipped: %s %q in %s points to %s (outside git root)",
				v.Kind, v.Name, v.File, v.Resolved)
		}
		return CheckOutcome{}, nil
	case IgnoreCheckModeFail:
		return CheckOutcome{}, formatLinkedDepsError(violations)
	case IgnoreCheckModePrompt:
	default:
		return CheckOutcome{}, fmt.Errorf("unknown linked-deps-check mode: %q", mode)
	}

	decider := p.Decider
	if decider == nil {
		decider = interactiveLinkedDepDecider
	}

	unstage := make(map[string]struct{})
	for _, v := range violations {
		d, err := decider(ctx, v)
		if err != nil {
			return CheckOutcome{}, fmt.Errorf("linked-deps prompt: %w", err)
		}
		switch d {
		case LinkedDepDecisionCancel:
			return CheckOutcome{Cancelled: true}, nil
		case LinkedDepDecisionUnstage:
			unstage[v.File] = struct{}{}
		}
	}

	if len(unstage) == 0 {
		return CheckOutcome{}, nil
	}

	files := make([]string, 0, len(unstage))
	for f := range unstage {
		files = append(files, f)
	}
	if err := resetFiles(p.WorkDir, files); err != nil {
		return CheckOutcome{}, fmt.Errorf("unstage: %w", err)
	}
	return CheckOutcome{Unstaged: files}, nil
}

func rootOrWorkDir(p LinkedDepsParams) string {
	if p.GitRoot != "" {
		return p.GitRoot
	}
	return p.WorkDir
}

func deletedPaths(changes []stagedChange) []string {
	var out []string
	for _, c := range changes {
		if c.Status == "deleted" {
			out = append(out, c.Path)
		}
	}
	return out
}

func formatLinkedDepsError(violations []LinkedDepViolation) error {
	lines := []string{"staged manifests reference paths outside the git root:"}
	for _, v := range violations {
		lines = append(lines, fmt.Sprintf("  - %s: %s %q -> %s (raw: %s)",
			v.File, v.Kind, v.Name, v.Resolved, v.Target))
	}
	return errors.New(strings.Join(lines, "\n"))
}

func runChooseLinkedDepDecider(_ context.Context, v LinkedDepViolation) (LinkedDepDecision, error) {
	header := fmt.Sprintf("%s in %s: %q -> %s (outside git root)",
		v.Kind, v.File, v.Name, v.Resolved)
	items := []string{
		fmt.Sprintf("Unstage %s (drop the edit from this commit)", v.File),
		"Cancel commit",
	}
	idx, err := choose.Run(items, choose.WithHeader(header), choose.WithLimit(1))
	if err != nil {
		return LinkedDepDecisionCancel, err
	}
	if len(idx) == 0 {
		return LinkedDepDecisionCancel, nil
	}
	if idx[0] == 0 {
		return LinkedDepDecisionUnstage, nil
	}
	return LinkedDepDecisionCancel, nil
}

// applyLinkedDepsCheck runs the check and, if any file was unstaged, re-reads
// the staged source so the caller sees the updated file list. Returns
// ErrLinkedDepsCancelled when the user cancels.
func applyLinkedDepsCheck(ctx context.Context, opts Options, source stagedSource) (stagedSource, error) {
	outcome, err := RunLinkedDepsCheck(ctx, LinkedDepsParams{
		WorkDir:     opts.WorkDir,
		GitRoot:     opts.WorkDir,
		StagedFiles: source.Files,
		Changes:     source.Changes,
		Mode:        opts.LinkedDepsCheck,
	})
	if err != nil {
		return source, err
	}
	if outcome.Cancelled {
		return source, ErrLinkedDepsCancelled
	}
	if len(outcome.Unstaged) == 0 {
		return source, nil
	}
	refreshed, err := readStagedSource(opts.WorkDir)
	if err != nil {
		return source, fmt.Errorf("re-read staged source after linked-deps unstage: %w", err)
	}
	return refreshed, nil
}
