package testui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/utils"
	"github.com/flanksource/gavel/verify"
)

type IgnoreRequest struct {
	Source  string `json:"source,omitempty"`
	Rule    string `json:"rule,omitempty"`
	File    string `json:"file,omitempty"`
	WorkDir string `json:"work_dir,omitempty"`
}

type IgnoreResponse struct {
	RuleCount int `json:"rule_count"`
	Filtered  int `json:"filtered"`
}

func (s *Server) handleLintIgnore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req IgnoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Rule == "" && req.Source == "" && req.File == "" {
		http.Error(w, "at least one of rule, source, or file is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	root := s.resolveGitRootLocked(req.WorkDir)
	s.mu.Unlock()
	if root == "" {
		if req.WorkDir != "" {
			http.Error(w, "work_dir is not part of the current lint results", http.StatusBadRequest)
			return
		}
		http.Error(w, "git root not configured", http.StatusInternalServerError)
		return
	}

	repoCfg, err := verify.LoadSingleGavelConfig(filepath.Join(root, ".gavel.yaml"))
	if err != nil && !os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("load .gavel.yaml: %v", err), http.StatusInternalServerError)
		return
	}

	newRule := verify.LintIgnoreRule{Source: req.Source, Rule: req.Rule, File: req.File}
	if !slices.Contains(repoCfg.Lint.Ignore, newRule) {
		repoCfg.Lint.Ignore = append(repoCfg.Lint.Ignore, newRule)
		if err := verify.SaveGavelConfig(root, repoCfg); err != nil {
			http.Error(w, fmt.Sprintf("save .gavel.yaml: %v", err), http.StatusInternalServerError)
			return
		}
	}

	s.mu.Lock()
	filtered := linters.FilterIgnoredViolations(s.lint, []verify.LintIgnoreRule{newRule})
	s.mu.Unlock()
	s.notify()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(IgnoreResponse{RuleCount: len(repoCfg.Lint.Ignore), Filtered: filtered})
}

func (s *Server) resolveGitRootLocked(workDir string) string {
	if workDir == "" {
		if root := cleanExistingRoot(s.gitRoot); root != "" {
			return root
		}
		if s.git != nil {
			if root := cleanExistingRoot(s.git.Root); root != "" {
				return root
			}
		}
		for _, result := range s.lint {
			if result != nil {
				if root := cleanExistingRoot(result.WorkDir); root != "" {
					return root
				}
			}
		}
		return ""
	}

	requested, err := filepath.Abs(workDir)
	if err != nil {
		return ""
	}
	requested = filepath.Clean(requested)

	for _, result := range s.lint {
		if result == nil || result.WorkDir == "" {
			continue
		}
		trustedWorkDir, err := filepath.Abs(result.WorkDir)
		if err != nil || requested != filepath.Clean(trustedWorkDir) {
			continue
		}
		return cleanExistingRoot(result.WorkDir)
	}

	for _, configured := range []string{s.gitRoot, snapshotGitRoot(s.git)} {
		if root := cleanExistingRoot(configured); root != "" && utils.IsWithin(requested, root) {
			return root
		}
	}

	return ""
}

func snapshotGitRoot(git *SnapshotGit) string {
	if git == nil {
		return ""
	}
	return git.Root
}
