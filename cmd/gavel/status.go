package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky"
	clickyai "github.com/flanksource/clicky/ai"
	"github.com/flanksource/clicky/api"
	gavelai "github.com/flanksource/gavel/ai"
	"github.com/flanksource/gavel/internal/prompting"
	"github.com/flanksource/gavel/status"
	"github.com/flanksource/repomap"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type StatusOptions struct {
	WorkDir   string   `flag:"work-dir" help:"Working directory"`
	NoRepomap bool     `flag:"no-repomap" help:"Skip repomap enrichment (faster)"`
	AI        bool     `flag:"ai" help:"Add a one-line AI summary for each changed file"`
	Args      []string `json:"-" args:"true"`
}

func (o StatusOptions) Help() api.Textable {
	const (
		heading    = "font-bold text-purple-600"
		flagStyle  = "text-cyan-600 font-bold"
		muted      = "text-muted"
		staged     = "text-green-500 font-bold"
		modified   = "text-yellow-500 font-bold"
		untracked  = "text-purple-500 font-bold"
		conflicted = "text-red-600 font-bold underline"
		deleted    = "text-red-500 font-bold"
		renamed    = "text-blue-500 font-bold"
	)

	legend := func(sym, style, label string) api.Text {
		return clicky.Text("  ").
			Append(sym, style).
			Append("  ", "").
			Append(label, muted)
	}

	t := clicky.Text("gavel status", heading).Space().
		Append("— list changed files grouped by scope", muted).NewLine().NewLine().
		Append("Shows one row per file using ", "").
		Append("Starship git_status", "italic").
		Append(" symbols, plus line deltas, the working-tree age, inferred scopes, and repomap findings (Kubernetes refs, architecture violations).", "").
		NewLine().
		Append("By default the listing is scoped to the current working directory. Pass a ", "").
		Append("[folder]", flagStyle).
		Append(" to scope to a different subdirectory of the repo, or ", "").
		Append(".", flagStyle).
		Append(" from the repo root for the whole repo.", "").
		NewLine().
		Append("Use ", "").
		Append("--ai", flagStyle).
		Append(" to add a one-line LLM summary of each file change.", muted).
		NewLine().NewLine().
		Append("SYMBOLS", heading).NewLine().
		Add(legend("+", staged, "staged")).NewLine().
		Add(legend("!", modified, "modified (unstaged)")).NewLine().
		Add(legend("?", untracked, "untracked")).NewLine().
		Add(legend("=", conflicted, "conflicted")).NewLine().
		Add(legend("✘", deleted, "deleted")).NewLine().
		Add(legend("»", renamed, "renamed / copied")).NewLine().NewLine().
		Append("FLAGS", heading).NewLine().
		Append("  --work-dir", flagStyle).Append("   path to a git repo (default: cwd)", muted).NewLine().
		Append("  --no-repomap", flagStyle).Append(" skip repomap enrichment (faster)", muted).NewLine().
		Append("  --ai", flagStyle).Append("         add one-line AI summaries per changed file", muted).NewLine().
		Append("  --ai-model", flagStyle).Append("   override the AI model used with --ai", muted).NewLine().NewLine().
		Append("EXAMPLES", heading).NewLine().
		Append("  gavel status", flagStyle).Append("                 changes under cwd", muted).NewLine().
		Append("  gavel status .", flagStyle).Append("               whole repo (run from git root)", muted).NewLine().
		Append("  gavel status cmd/gavel", flagStyle).Append("       changes under cmd/gavel", muted).NewLine().
		Append("  gavel status --no-repomap", flagStyle).Append("    skip repomap lookup", muted).NewLine().
		Append("  gavel status --ai", flagStyle).Append("             include AI one-line file summaries", muted).NewLine()

	return t
}

func init() {
	statusCmd := clicky.AddNamedCommand("status", rootCmd, StatusOptions{}, runStatus)
	statusCmd.Use = "status [folder]"
	statusCmd.Args = cobra.MaximumNArgs(1)
	clickyai.BindFlags(statusCmd.Flags())
}

func runStatus(opts StatusOptions) (any, error) {
	workDir, folderFilter, err := resolveStatusWorkDir(opts.WorkDir, opts.Args)
	if err != nil {
		return nil, err
	}

	if !opts.AI {
		return status.Gather(workDir, status.Options{
			NoRepomap:    opts.NoRepomap,
			FolderFilter: folderFilter,
		})
	}

	agent, err := gavelai.NewAgent(clickyai.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("create AI agent for status: %w", err)
	}
	defer agent.Close()

	ctx := context.Background()
	gatherOpts := status.Options{
		NoRepomap:    opts.NoRepomap,
		FolderFilter: folderFilter,
		Agent:        agent,
		Context:      ctx,
		AIMaxWorkers: clickyai.DefaultConfig().MaxConcurrent,
	}

	result, err := status.GatherBase(workDir, gatherOpts)
	if err != nil {
		return nil, err
	}

	prompting.Prepare()
	result.PrepareAISummaries()
	updates := status.StreamAISummaries(ctx, workDir, agent, result.Files, gatherOpts.AIMaxWorkers)
	if err := renderStatusOutput(os.Stdout, result, updates, isTerminalWriter(os.Stdout)); err != nil {
		return nil, err
	}

	return nil, nil
}

// resolveStatusWorkDir resolves the git root to scan and the optional
// folder-relative-to-root filter. The first return is always the git root so
// status enrichment (numstat, repomap, snapshot) sees the full repo. The
// second return is a slash-separated path relative to the git root, or "" if
// no scoping is requested.
//
//   - workDirFlag is the value of --work-dir (may be empty).
//   - args is the positional argument list (0 or 1 entry per cobra config).
//
// When no positional arg is given the filter defaults to the current working
// directory, so `gavel status` from a subdirectory shows only that subtree.
// Pass `.` from the repo root to see the whole repo.
func resolveStatusWorkDir(workDirFlag string, args []string) (string, string, error) {
	workDir := workDirFlag
	if workDir == "" {
		wd, err := getWorkingDir()
		if err != nil {
			return "", "", err
		}
		workDir = wd
	}

	root := repomap.FindGitRoot(workDir)
	if root == "" {
		root = workDir
	}

	target := workDir
	if len(args) == 1 && args[0] != "" {
		target = args[0]
		if !filepath.IsAbs(target) {
			cwd, err := getWorkingDir()
			if err != nil {
				return "", "", err
			}
			target = filepath.Join(cwd, target)
		}
	}

	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", "", fmt.Errorf("resolve status path: %w", err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", "", fmt.Errorf("resolve git root: %w", err)
	}

	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil {
		return "", "", fmt.Errorf("resolve status path relative to git root: %w", err)
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		rel = ""
	}
	if strings.HasPrefix(rel, "../") || rel == ".." {
		return "", "", fmt.Errorf("path %q is outside git repository %q", absTarget, absRoot)
	}

	return absRoot, rel, nil
}

func renderStatusOutput(w io.Writer, result *status.Result, updates <-chan status.AISummaryUpdate, interactive bool) error {
	if !interactive {
		for update := range updates {
			result.ApplyAISummaryUpdate(update)
		}
		_, err := io.WriteString(w, formatStatusResult(result))
		return err
	}

	renderState := statusRenderState{}
	if err := renderState.write(w, formatStatusResult(result)); err != nil {
		return err
	}
	for update := range updates {
		result.ApplyAISummaryUpdate(update)
		if err := renderState.write(w, formatStatusResult(result)); err != nil {
			return err
		}
	}
	return nil
}

type statusRenderState struct {
	lines int
}

func (s *statusRenderState) write(w io.Writer, rendered string) error {
	if s.lines > 0 {
		if _, err := fmt.Fprintf(w, "\x1b[%dA\x1b[J", s.lines); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, rendered); err != nil {
		return err
	}
	s.lines = countRenderedLines(rendered)
	return nil
}

func formatStatusResult(result *status.Result) string {
	rendered := result.Pretty().ANSI()
	if !strings.HasSuffix(rendered, "\n") {
		rendered += "\n"
	}
	return rendered
}

func countRenderedLines(rendered string) int {
	if rendered == "" {
		return 0
	}
	lines := strings.Count(rendered, "\n")
	if strings.HasSuffix(rendered, "\n") {
		return lines
	}
	return lines + 1
}

func isTerminalWriter(w io.Writer) bool {
	file, ok := w.(*os.File)
	return ok && term.IsTerminal(int(file.Fd()))
}
