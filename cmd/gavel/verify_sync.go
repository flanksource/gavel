package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/todos/types"
	"github.com/flanksource/gavel/verify"
	"github.com/samber/lo"
)

type SyncOptions struct {
	TodosDir       string
	ScoreThreshold int
	RepoPath       string
}

type SyncResult struct {
	Created   []string
	Updated   []string
	Completed []string
}

func (r SyncResult) Pretty() api.Text {
	text := clicky.Text("Todo Sync", "font-bold")

	if len(r.Created) > 0 {
		text = text.NewLine().Append(fmt.Sprintf("  Created (%d)", len(r.Created)), "text-green-600")
		for _, f := range r.Created {
			text = text.NewLine().Append("    ", "").Add(icons.Check.WithStyle("text-green-600")).Append(" "+f, "")
		}
	}
	if len(r.Updated) > 0 {
		text = text.NewLine().Append(fmt.Sprintf("  Updated (%d)", len(r.Updated)), "text-yellow-600")
		for _, f := range r.Updated {
			text = text.NewLine().Append("    ", "").Add(icons.Edit).Append(" "+f, "")
		}
	}
	if len(r.Completed) > 0 {
		text = text.NewLine().Append(fmt.Sprintf("  Completed (%d)", len(r.Completed)), "text-blue-600")
		for _, f := range r.Completed {
			text = text.NewLine().Append("    ", "").Add(icons.Check).Append(" "+f, "")
		}
	}
	if len(r.Created)+len(r.Updated)+len(r.Completed) == 0 {
		text = text.Append(" — no changes", "text-gray-500")
	}
	return text
}

func SyncTodos(result *verify.VerifyResult, opts SyncOptions) (*SyncResult, error) {
	if opts.ScoreThreshold <= 0 {
		opts.ScoreThreshold = 80
	}

	if err := os.MkdirAll(opts.TodosDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create todos dir: %w", err)
	}

	lang := detectRepoLanguage(opts.RepoPath)
	existing := loadExistingTodos(opts.TodosDir)
	syncResult := &SyncResult{}

	// Sync failing checks
	for id, cr := range result.Checks {
		filename := id + ".md"
		filePath := filepath.Join(opts.TodosDir, filename)
		category := checkCategory(id)

		if !cr.Pass {
			paths := evidencePaths(cr.Evidence)
			if _, ok := existing[filename]; ok {
				if err := updateVerifyTodo(filePath, cr.Evidence); err != nil {
					return nil, fmt.Errorf("updating %s: %w", filename, err)
				}
				syncResult.Updated = append(syncResult.Updated, filename)
			} else {
				todo := buildCheckTodo(id, category, paths, lang, opts.ScoreThreshold)
				todo.Implementation = formatEvidenceMarkdown(cr.Evidence)
				if err := todos.WriteTODOFile(filePath, todo); err != nil {
					return nil, fmt.Errorf("creating %s: %w", filename, err)
				}
				syncResult.Created = append(syncResult.Created, filename)
			}
		} else if _, ok := existing[filename]; ok {
			if err := markCompleted(filePath); err != nil {
				return nil, fmt.Errorf("completing %s: %w", filename, err)
			}
			syncResult.Completed = append(syncResult.Completed, filename)
		}
	}

	// Sync low-scoring ratings
	for dim, rr := range result.Ratings {
		filename := "rating-" + dim + ".md"
		filePath := filepath.Join(opts.TodosDir, filename)

		if rr.Score < opts.ScoreThreshold {
			paths := evidencePaths(rr.Findings)
			if _, ok := existing[filename]; ok {
				if err := updateVerifyTodo(filePath, rr.Findings); err != nil {
					return nil, fmt.Errorf("updating %s: %w", filename, err)
				}
				syncResult.Updated = append(syncResult.Updated, filename)
			} else {
				todo := buildRatingTodo(dim, rr.Score, paths, lang, opts.ScoreThreshold)
				todo.Implementation = formatEvidenceMarkdown(rr.Findings)
				if err := todos.WriteTODOFile(filePath, todo); err != nil {
					return nil, fmt.Errorf("creating %s: %w", filename, err)
				}
				syncResult.Created = append(syncResult.Created, filename)
			}
		} else if _, ok := existing[filename]; ok {
			if err := markCompleted(filePath); err != nil {
				return nil, fmt.Errorf("completing %s: %w", filename, err)
			}
			syncResult.Completed = append(syncResult.Completed, filename)
		}
	}

	return syncResult, nil
}

func buildCheckTodo(id, category string, paths []string, lang types.Language, threshold int) *types.TODO {
	return &types.TODO{
		TODOFrontmatter: types.TODOFrontmatter{
			Title:    humanize(id),
			Priority: types.PriorityHigh,
			Status:   types.StatusPending,
			Language: lang,
			Path:     types.StringOrSlice(paths),
			Verify: &types.TODOVerifyConfig{
				Categories:     []string{category},
				ScoreThreshold: threshold,
			},
		},
	}
}

func buildRatingTodo(dim string, score int, paths []string, lang types.Language, threshold int) *types.TODO {
	return &types.TODO{
		TODOFrontmatter: types.TODOFrontmatter{
			Title:    fmt.Sprintf("Rating: %s (score %d)", humanize(dim), score),
			Priority: types.PriorityMedium,
			Status:   types.StatusPending,
			Language: lang,
			Path:     types.StringOrSlice(paths),
			Verify: &types.TODOVerifyConfig{
				ScoreThreshold: threshold,
			},
		},
	}
}

func loadExistingTodos(dir string) map[string]bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	result := make(map[string]bool)
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			result[e.Name()] = true
		}
	}
	return result
}

func updateVerifyTodo(filePath string, evidence []verify.Evidence) error {
	fm, err := todos.ReadTODOState(filePath)
	if err != nil {
		return err
	}
	if fm.Status == types.StatusCompleted {
		// Reopen
		return todos.UpdateTODOState(&types.TODO{FilePath: filePath, TODOFrontmatter: *fm}, todos.StateUpdate{
			Status: lo.ToPtr(types.StatusPending),
		})
	}
	return nil
}

func markCompleted(filePath string) error {
	fm, err := todos.ReadTODOState(filePath)
	if err != nil {
		return err
	}
	if fm.Status == types.StatusCompleted {
		return nil
	}
	return todos.UpdateTODOState(&types.TODO{FilePath: filePath, TODOFrontmatter: *fm}, todos.StateUpdate{
		Status: lo.ToPtr(types.StatusCompleted),
	})
}

func evidencePaths(evidence []verify.Evidence) []string {
	seen := map[string]bool{}
	var paths []string
	for _, e := range evidence {
		if e.File != "" && !seen[e.File] {
			seen[e.File] = true
			paths = append(paths, e.File)
		}
	}
	return paths
}

func checkCategory(id string) string {
	for _, c := range verify.AllChecks {
		if c.ID == id {
			return c.Category
		}
	}
	return ""
}

func humanize(s string) string {
	parts := strings.Split(s, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

func detectRepoLanguage(repoPath string) types.Language {
	indicators := []struct {
		file string
		lang types.Language
	}{
		{"go.mod", types.LanguageGo},
		{"package.json", types.LanguageTypeScript},
		{"pyproject.toml", types.LanguagePython},
		{"requirements.txt", types.LanguagePython},
	}
	for _, ind := range indicators {
		if _, err := os.Stat(filepath.Join(repoPath, ind.file)); err == nil {
			return ind.lang
		}
	}
	return ""
}

func formatEvidenceMarkdown(evidence []verify.Evidence) string {
	if len(evidence) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## Findings\n\n")
	for _, e := range evidence {
		if e.Line > 0 {
			sb.WriteString(fmt.Sprintf("- `%s:%d` — %s\n", e.File, e.Line, e.Message))
		} else if e.File != "" {
			sb.WriteString(fmt.Sprintf("- `%s` — %s\n", e.File, e.Message))
		} else {
			sb.WriteString(fmt.Sprintf("- %s\n", e.Message))
		}
	}
	return sb.String()
}
