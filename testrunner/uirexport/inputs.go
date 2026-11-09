package uirexport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/flanksource/commons/logger"
)

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

// inputsDigest covers every module file UIR's Go index does not, plus build-selected source from
// mutable local dependencies. A file is an input when it is
//
//   - embedded by a package of the module: `go list` reports it as EmbedFiles, TestEmbedFiles or
//     XTestEmbedFiles under the run's build tags, whatever its name, even when git ignores it; or
//   - tracked or untracked and non-ignored below the module, unless it is regular Go source UIR
//     indexes; or
//   - build-selected source absent from Git's visible inventory, even when ignored; or
//   - a visible file or build-selected source of a local replacement or workspace dependency,
//     since its bytes are not fixed by a module version.
//
// Nested modules are separate execution roots, except when Go selects one as a mutable dependency.
// Immutable module-cache dependencies are represented by UIR's dependency-set hash. The export file
// itself is excluded because Gavel creates it after execution; generated cache directories are
// already ignored by CaptureProvenance before this inventory is read.
//
// The digest is "sha256:" and the SHA-256 of one "<sha256 of content>  <path>\n" line per input,
// sorted by stable module-relative or dependency-qualified path, so a rename changes it too.
func inputsDigest(m *module, tags []string, output string) (string, int, error) {
	inputs := map[string]string{}
	embedded, err := embeddedFiles(m, tags)
	if err != nil {
		return "", 0, err
	}
	for rel := range embedded {
		inputs[rel] = filepath.Join(m.root, filepath.FromSlash(rel))
	}

	listed, err := git(m.root, nil, "ls-files", "-z", "--cached", "--others", "--exclude-standard", "--", ".")
	if err != nil {
		return "", 0, fmt.Errorf("list module test inputs: %w", err)
	}
	visible := map[string]bool{}
	for _, name := range strings.Split(listed, "\x00") {
		if name == "" {
			continue
		}
		rel := filepath.ToSlash(name)
		visible[rel] = true
		path := filepath.Join(m.root, filepath.FromSlash(rel))
		if m.relative(m.root, path) == "" {
			continue
		}
		info, err := os.Lstat(path)
		if err != nil {
			return "", 0, fmt.Errorf("inspect test input %s: %w", rel, err)
		}
		if info.IsDir() {
			continue
		}
		if !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
			return "", 0, fmt.Errorf("test input %s is not a regular file or symlink", rel)
		}
		if info.Mode()&os.ModeSymlink != 0 || !indexedSource(rel) {
			inputs[rel] = path
		}
	}

	dependencies, err := mutableDependencyInputs(m, tags, visible)
	if err != nil {
		return "", 0, err
	}
	for label, path := range dependencies {
		inputs[label] = path
	}

	if output != "" {
		if abs, err := filepath.Abs(output); err == nil {
			if rel, err := filepath.Rel(m.root, abs); err == nil {
				rel = filepath.ToSlash(rel)
				if embedded[rel] {
					return "", 0, fmt.Errorf("UIR export %s is also an embedded test input", rel)
				}
				delete(inputs, rel)
			}
		}
	}
	paths := make([]string, 0, len(inputs))
	for path := range inputs {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	digest := sha256.New()
	for _, label := range paths {
		file, err := os.Open(inputs[label])
		if err != nil {
			return "", 0, fmt.Errorf("read test input %s: %w", label, err)
		}
		content := sha256.New()
		_, err = io.Copy(content, file)
		file.Close()
		if err != nil {
			return "", 0, fmt.Errorf("read test input %s: %w", label, err)
		}
		fmt.Fprintf(digest, "%s  %s\n", hex.EncodeToString(content.Sum(nil)), label)
	}
	return "sha256:" + hex.EncodeToString(digest.Sum(nil)), len(paths), nil
}

// mutableDependencyInputs returns build-selected source from local replacements and workspace
// modules outside the invocation module. Versioned module-cache and standard-library packages are
// immutable under their dependency/toolchain identities and need no second content digest here.
func mutableDependencyInputs(m *module, tags []string, visible map[string]bool) (map[string]string, error) {
	fields := "Dir,ImportPath,Module,GoFiles,CgoFiles,TestGoFiles,XTestGoFiles,CFiles,CXXFiles,MFiles,HFiles,FFiles,SFiles,SwigFiles,SwigCXXFiles,SysoFiles,EmbedFiles"
	args := []string{"list", "-deps", "-test", "-json=" + fields}
	if len(tags) > 0 {
		args = append(args, "-tags="+strings.Join(tags, ","))
	}
	command := exec.Command("go", append(args, "./...")...)
	command.Dir = m.root
	output, err := command.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			err = fmt.Errorf("%w: %s", err, exit.Stderr)
		}
		return nil, fmt.Errorf("enumerate mutable dependency inputs: %w", err)
	}
	type selectedModule struct {
		Path, Version, Dir string
		Main               bool
		Replace            *selectedModule
	}
	type listedPackage struct {
		Dir, ImportPath                                                                                                                                string
		Module                                                                                                                                         *selectedModule
		GoFiles, CgoFiles, TestGoFiles, XTestGoFiles, CFiles, CXXFiles, MFiles, HFiles, FFiles, SFiles, SwigFiles, SwigCXXFiles, SysoFiles, EmbedFiles []string
	}
	inputs := map[string]string{}
	modules := map[string]string{}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for decoder.More() {
		var pkg listedPackage
		if err := decoder.Decode(&pkg); err != nil {
			return nil, fmt.Errorf("decode mutable dependency inputs: %w", err)
		}
		if pkg.Module == nil || pkg.Dir == "" {
			continue
		}
		if m.rootOf(pkg.Dir) == m.root {
			files := slices.Concat(pkg.GoFiles, pkg.CgoFiles, pkg.TestGoFiles, pkg.XTestGoFiles, pkg.CFiles, pkg.CXXFiles, pkg.MFiles, pkg.HFiles, pkg.FFiles,
				pkg.SFiles, pkg.SwigFiles, pkg.SwigCXXFiles, pkg.SysoFiles, pkg.EmbedFiles)
			for _, name := range files {
				// go list -test also reports its generated absolute _testmain.go; it is tool output,
				// not a source input owned by the module.
				if filepath.IsAbs(name) {
					continue
				}
				path := filepath.Join(pkg.Dir, filepath.FromSlash(name))
				rel := m.relative(m.root, path)
				if rel == "" {
					return nil, fmt.Errorf("selected module input %s/%s is outside the module", pkg.ImportPath, name)
				}
				if indexedSource(rel) && visible[rel] {
					continue
				}
				inputs[rel] = path
			}
			continue
		}
		effective := pkg.Module
		if pkg.Module.Replace != nil {
			effective = pkg.Module.Replace
		}
		if !effective.Main && effective.Version != "" {
			continue
		}
		if effective.Dir == "" {
			return nil, fmt.Errorf("mutable dependency %s has no source directory", pkg.Module.Path)
		}
		if existing := modules[pkg.Module.Path]; existing != "" && existing != effective.Dir {
			return nil, fmt.Errorf("mutable dependency %s has multiple source directories", pkg.Module.Path)
		}
		modules[pkg.Module.Path] = effective.Dir
		files := slices.Concat(pkg.GoFiles, pkg.CgoFiles, pkg.CFiles, pkg.CXXFiles, pkg.MFiles, pkg.HFiles,
			pkg.FFiles, pkg.SFiles, pkg.SwigFiles, pkg.SwigCXXFiles, pkg.SysoFiles, pkg.EmbedFiles)
		for _, name := range files {
			if filepath.IsAbs(name) {
				continue
			}
			path := filepath.Join(pkg.Dir, filepath.FromSlash(name))
			info, err := os.Lstat(path)
			if err != nil {
				return nil, fmt.Errorf("inspect mutable dependency input %s/%s: %w", pkg.ImportPath, name, err)
			}
			if !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
				return nil, fmt.Errorf("mutable dependency input %s/%s is not a regular file or symlink", pkg.ImportPath, name)
			}
			label := "dependency/" + pkg.Module.Path + "/" + pkg.ImportPath + "/" + filepath.ToSlash(name)
			inputs[label] = path
		}
	}
	for modulePath, root := range modules {
		visible, err := gitVisibleModuleInputs(modulePath, root)
		if err != nil {
			return nil, err
		}
		for label, path := range visible {
			inputs[label] = path
		}
	}
	return inputs, nil
}

// gitVisibleModuleInputs enumerates a mutable dependency's tracked and untracked, non-ignored files.
// A dependency outside Git cannot establish complete input identity, so the caller disables skipping.
func gitVisibleModuleInputs(modulePath, root string) (map[string]string, error) {
	top, err := git(root, nil, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("locate the Git checkout of mutable dependency %s: %w", modulePath, err)
	}
	prefix, err := filepath.Rel(top, root)
	if err != nil || !filepath.IsLocal(prefix) {
		return nil, fmt.Errorf("locate mutable dependency %s within its Git checkout", modulePath)
	}
	prefix = filepath.ToSlash(prefix)
	if prefix == "." {
		prefix = ""
	} else {
		prefix += "/"
	}
	listed, err := git(root, nil, "ls-files", "-z", "--full-name", "--cached", "--others", "--exclude-standard", "--", ".")
	if err != nil {
		return nil, fmt.Errorf("list mutable dependency %s inputs: %w", modulePath, err)
	}
	inputs := map[string]string{}
	for _, name := range strings.Split(listed, "\x00") {
		rel, ok := strings.CutPrefix(filepath.ToSlash(name), prefix)
		if !ok || rel == "" {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(rel))
		info, err := os.Lstat(path)
		if err != nil {
			return nil, fmt.Errorf("inspect mutable dependency input %s/%s: %w", modulePath, rel, err)
		}
		if info.IsDir() {
			continue
		}
		if !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
			return nil, fmt.Errorf("mutable dependency input %s/%s is not a regular file or symlink", modulePath, rel)
		}
		inputs["dependency/"+modulePath+"/"+rel] = path
	}
	return inputs, nil
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
