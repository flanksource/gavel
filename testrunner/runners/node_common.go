package runners

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/gavel/utils"
)

var nodeConfigExts = []string{"js", "ts", "mjs", "cjs"}

// pkgJSON captures the subset of package.json we care about.
type pkgJSON struct {
	Jest            json.RawMessage   `json:"jest,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
	Workspaces      json.RawMessage   `json:"workspaces,omitempty"`
}

// readPackageJSON reads <dir>/package.json. Missing or malformed returns
// (nil, false) — callers treat that as "not a node package".
func readPackageJSON(dir string) (*pkgJSON, bool) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil, false
	}
	var pkg pkgJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, false
	}
	return &pkg, true
}

// hasConfigFile reports whether any <basename>.<ext> exists in dir.
func hasConfigFile(dir string, basenames, exts []string) bool {
	for _, base := range basenames {
		for _, ext := range exts {
			if _, err := os.Stat(filepath.Join(dir, base+"."+ext)); err == nil {
				return true
			}
		}
	}
	return false
}

// hasNpmDep reports whether pkg declares name in dependencies or devDependencies.
func hasNpmDep(pkg *pkgJSON, name string) bool {
	if pkg == nil {
		return false
	}
	if _, ok := pkg.Dependencies[name]; ok {
		return true
	}
	_, ok := pkg.DevDependencies[name]
	return ok
}

// detectPackageManager returns the command + prefix args used to invoke a
// node tool from dir. It walks up looking for lockfiles and falls back to npx.
//   - pnpm-lock.yaml    → ("pnpm", ["exec"])
//   - yarn.lock         → ("yarn", nil)
//   - package-lock.json → ("npm", ["exec", "--"])
//   - bun.lockb         → ("bun", ["x"])
//   - none              → ("npx", nil)
func detectPackageManager(dir string) (string, []string) {
	cur, _ := filepath.Abs(dir)
	for {
		if _, err := os.Stat(filepath.Join(cur, "pnpm-lock.yaml")); err == nil {
			return "pnpm", []string{"exec"}
		}
		if _, err := os.Stat(filepath.Join(cur, "yarn.lock")); err == nil {
			return "yarn", nil
		}
		if _, err := os.Stat(filepath.Join(cur, "bun.lockb")); err == nil {
			return "bun", []string{"x"}
		}
		if _, err := os.Stat(filepath.Join(cur, "package-lock.json")); err == nil {
			return "npm", []string{"exec", "--"}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "npx", nil
		}
		cur = parent
	}
}

// walkNodePackages finds every directory under root that (a) contains a
// package.json and (b) satisfies detect(dir, pkgJSON). Returns gavel-style
// relative paths ("./apps/web"). Skips node_modules/ defensively and honors
// .gitignore via utils.WalkGitIgnoredBounded.
func walkNodePackages(root string, detect func(string, *pkgJSON) bool) ([]string, error) {
	var out []string
	seen := map[string]bool{}

	err := utils.WalkGitIgnoredBounded(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "node_modules" {
				return fs.SkipDir
			}
			return nil
		}
		if d.Name() != "package.json" {
			return nil
		}
		pkgDir := filepath.Dir(path)
		if seen[pkgDir] {
			return nil
		}
		seen[pkgDir] = true
		pkg, _ := readPackageJSON(pkgDir)
		if detect(pkgDir, pkg) {
			rel, relErr := filepath.Rel(root, pkgDir)
			if relErr != nil {
				rel = pkgDir
			}
			out = append(out, "./"+filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// normalizeNodeFilePath converts an absolute path from a Node tool's JSON
// reporter to a path relative to workDir. Leaves already-relative paths alone.
func normalizeNodeFilePath(workDir, filePath string) string {
	if filePath == "" {
		return ""
	}
	if !filepath.IsAbs(filePath) && !strings.HasPrefix(filePath, "..") {
		return filePath
	}
	if rel, err := filepath.Rel(workDir, filePath); err == nil {
		return rel
	}
	return filePath
}
