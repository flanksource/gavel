package uirexport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/utils"
)

// inputExtensions are the data and configuration files a test may read at run time.
var inputExtensions = []string{
	".sql", ".hcl", ".yaml", ".yml", ".json", ".toml", ".ini", ".conf", ".cfg", ".env", ".properties",
	".xml", ".csv", ".tmpl", ".tpl", ".gotmpl",
}

// InputsDigest is the inputs_digest of the Go module holding workDir; see inputsDigest.
func InputsDigest(workDir string, tags []string, output string) (string, error) {
	m, err := findModule(workDir)
	if err != nil {
		return "", err
	}
	digest, inputs, err := inputsDigest(m, tags, output)
	if err != nil {
		return "", err
	}
	logger.V(1).Infof("uir export: inputs digest %s over %d files", digest, inputs)
	return digest, nil
}

// inputsDigest digests the files of the module tests can read that uir does not index, which is only
// Go source. A file is an input when it is
//
//   - embedded by a package of the module: `go list` reports it as EmbedFiles, TestEmbedFiles or
//     XTestEmbedFiles under the run's build tags, whatever its name, even when git ignores it; or
//   - under a directory named testdata, whatever its extension, .go included; or
//   - a data or configuration file by extension (inputExtensions).
//
// The last two come from walking the module root as git sees it: ignored files, the .git directory,
// directories whose name starts with "." (gavel's own .gavel and .tmp among them), vendor and nested
// modules or checkouts are skipped. The export file itself is never an input.
//
// The digest is "sha256:" and the SHA-256 of one "<sha256 of content>  <path>\n" line per input,
// sorted by module-relative path, so a renamed input changes it as well as an edited one.
func inputsDigest(m *module, tags []string, output string) (string, int, error) {
	inputs, err := embeddedFiles(m, tags)
	if err != nil {
		return "", 0, err
	}
	err = utils.WalkGitIgnoredBounded(m.root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != m.root && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "vendor") {
				return fs.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(m.root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if slices.Contains(inputExtensions, strings.ToLower(filepath.Ext(rel))) || slices.Contains(strings.Split(rel, "/"), "testdata") {
			inputs[rel] = true
		}
		return nil
	})
	if err != nil {
		return "", 0, fmt.Errorf("walk %s for test inputs: %w", m.root, err)
	}
	if output != "" {
		if abs, err := filepath.Abs(output); err == nil {
			if rel, err := filepath.Rel(m.root, abs); err == nil {
				delete(inputs, filepath.ToSlash(rel))
			}
		}
	}
	paths := make([]string, 0, len(inputs))
	for path := range inputs {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	digest := sha256.New()
	for _, path := range paths {
		file, err := os.Open(filepath.Join(m.root, filepath.FromSlash(path)))
		if err != nil {
			return "", 0, fmt.Errorf("read test input: %w", err)
		}
		content := sha256.New()
		_, err = io.Copy(content, file)
		file.Close()
		if err != nil {
			return "", 0, fmt.Errorf("read test input %s: %w", path, err)
		}
		fmt.Fprintf(digest, "%s  %s\n", hex.EncodeToString(content.Sum(nil)), path)
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), len(paths), nil
}

// embeddedFiles lists, by module-relative path, the files the module's packages and their tests embed
// under the build tags.
func embeddedFiles(m *module, tags []string) (map[string]bool, error) {
	args := []string{"list", "-e", "-json=Dir,EmbedFiles,TestEmbedFiles,XTestEmbedFiles"}
	if len(tags) > 0 {
		args = append(args, "-tags="+strings.Join(tags, ","))
	}
	list := exec.Command("go", append(args, "./...")...)
	list.Dir = m.root
	output, err := list.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			err = fmt.Errorf("%w: %s", err, exit.Stderr)
		}
		return nil, fmt.Errorf("go list embedded files: %w", err)
	}
	files := map[string]bool{}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for decoder.More() {
		var pkg struct {
			Dir                                         string
			EmbedFiles, TestEmbedFiles, XTestEmbedFiles []string
		}
		if err := decoder.Decode(&pkg); err != nil {
			return nil, fmt.Errorf("decode go list: %w", err)
		}
		for _, name := range slices.Concat(pkg.EmbedFiles, pkg.TestEmbedFiles, pkg.XTestEmbedFiles) {
			if rel := m.relative(pkg.Dir, filepath.FromSlash(name)); rel != "" {
				files[rel] = true
			}
		}
	}
	return files, nil
}
