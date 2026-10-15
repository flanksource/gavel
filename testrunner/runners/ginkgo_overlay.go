package runners

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//go:embed ginkgo_reporter_source.txt
var ginkgoReporterSource embed.FS

// BuildGinkgoOverlay injects a Ginkgo reporter into selected test packages.
// Stable paths and contents let go test reuse its own result cache.
func BuildGinkgoOverlay(workDir string, packages []string, tags []string) (string, error) {
	if len(packages) == 0 {
		return "", nil
	}
	root, err := filepath.Abs(workDir)
	if err != nil {
		return "", err
	}
	content, err := ginkgoReporterSource.ReadFile("ginkgo_reporter_source.txt")
	if err != nil {
		return "", err
	}
	sorted := append([]string(nil), packages...)
	sort.Strings(sorted)
	digest := sha256.Sum256([]byte(strings.Join(sorted, "\n")))
	overlayDir := filepath.Join(root, ".tmp", "gavel-ginkgo-overlay")
	if err := os.MkdirAll(overlayDir, 0o755); err != nil {
		return "", fmt.Errorf("create Ginkgo overlay directory: %w", err)
	}
	files := make(map[string]string, len(sorted))
	for _, pkg := range sorted {
		pkgDir := filepath.Join(root, pkg)
		packageName, err := ginkgoPackageName(pkgDir, tags)
		if err != nil {
			return "", fmt.Errorf("package %s: %w", pkg, err)
		}
		virtual := filepath.Join(pkgDir, "gavel_metadata_test.go")
		if _, err := os.Stat(virtual); err == nil {
			return "", fmt.Errorf("Ginkgo overlay target already exists: %s", virtual)
		} else if !os.IsNotExist(err) {
			return "", err
		}
		pkgHash := sha256.Sum256([]byte(pkgDir))
		backing := filepath.Join(overlayDir, hex.EncodeToString(pkgHash[:8])+".go")
		if err := writeOverlayFile(backing, append([]byte("package "+packageName+"\n\n"), content...)); err != nil {
			return "", fmt.Errorf("write Ginkgo reporter for %s: %w", pkg, err)
		}
		files[virtual] = backing
	}
	manifest := filepath.Join(overlayDir, hex.EncodeToString(digest[:8])+".json")
	data, err := json.Marshal(struct {
		Replace map[string]string `json:"Replace"`
	}{Replace: files})
	if err != nil {
		return "", err
	}
	if err := writeOverlayFile(manifest, data); err != nil {
		return "", fmt.Errorf("write Ginkgo overlay manifest: %w", err)
	}
	return manifest, nil
}

func ginkgoPackageName(dir string, tags []string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var packageName string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		matched, err := matchesBuildConstraints(dir, entry.Name(), tags)
		if err != nil {
			return "", err
		}
		if !matched {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			return "", fmt.Errorf("parse %s: %w", entry.Name(), err)
		}
		for _, imp := range file.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return "", err
			}
			if path == "github.com/onsi/ginkgo" {
				return "", fmt.Errorf("Ginkgo v1 is unsupported; upgrade %s to github.com/onsi/ginkgo/v2", entry.Name())
			}
			if path == "github.com/onsi/ginkgo/v2" {
				name := file.Name.Name
				// pkg and pkg_test link into one test binary sharing one Ginkgo suite; the reporter goes in pkg_test.
				switch {
				case packageName == "" || packageName == name || name == packageName+"_test":
					packageName = name
				case packageName != name+"_test":
					return "", fmt.Errorf("multiple Ginkgo test packages in %s: %s and %s", dir, packageName, name)
				}
			}
		}
	}
	if packageName == "" {
		return "", fmt.Errorf("no active Ginkgo v2 test file in %s", dir)
	}
	return packageName, nil
}

func writeOverlayFile(path string, content []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".gavel-overlay-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(content); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
