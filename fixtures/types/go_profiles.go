package types

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flanksource/gavel/fixtures"
	"github.com/google/pprof/profile"
)

type preparedGoProfiles struct {
	Root  string
	Args  []string
	Env   map[string]string
	Paths map[string]fixtures.GoProfileOutput
}

func prepareGoProfiles(workDir string, outputs map[string]fixtures.GoProfileOutput) (*preparedGoProfiles, error) {
	absoluteWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve fixture profile working directory: %w", err)
	}
	base := filepath.Join(absoluteWorkDir, ".gavel", "profiles", "fixtures")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return nil, fmt.Errorf("create fixture profile store: %w", err)
	}
	root, err := os.MkdirTemp(base, "run-")
	if err != nil {
		return nil, fmt.Errorf("allocate fixture profile directory: %w", err)
	}
	prepared := &preparedGoProfiles{Root: root, Env: map[string]string{}, Paths: map[string]fixtures.GoProfileOutput{}}
	destinations := make(map[string]string)
	names := make([]string, 0, len(outputs))
	for name := range outputs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		output := outputs[name]
		path, err := goProfilePath(root, output)
		if err != nil {
			return nil, fmt.Errorf("goProfiles.%s: %w", name, err)
		}
		if prior, exists := destinations[path]; exists {
			return nil, fmt.Errorf("goProfiles.%s: destination is already used by goProfiles.%s", name, prior)
		}
		destinations[path] = name
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, fmt.Errorf("goProfiles.%s: create parent: %w", name, err)
		}
		if output.Directory != "" {
			if err := os.MkdirAll(path, 0o755); err != nil {
				return nil, fmt.Errorf("goProfiles.%s: create directory: %w", name, err)
			}
		}
		if output.Env != "" {
			if _, exists := prepared.Env[output.Env]; exists {
				return nil, fmt.Errorf("goProfiles.%s: environment variable %s is already assigned", name, output.Env)
			}
			prepared.Env[output.Env] = path
		}
		for _, arg := range output.Args {
			if !strings.Contains(arg, "{{.path}}") {
				return nil, fmt.Errorf("goProfiles.%s: argument %q must contain {{.path}}", name, arg)
			}
			prepared.Args = append(prepared.Args, strings.ReplaceAll(arg, "{{.path}}", path))
		}
		prepared.Paths[name] = output
	}
	return prepared, nil
}

func goProfilePath(root string, output fixtures.GoProfileOutput) (string, error) {
	if (output.File == "") == (output.Directory == "") {
		return "", fmt.Errorf("specify exactly one of file or directory")
	}
	name := output.File
	if output.Directory != "" {
		name = output.Directory
	}
	if !filepath.IsLocal(name) || name == "." {
		return "", fmt.Errorf("profile path %q must be relative to the fixture artifact directory", name)
	}
	return filepath.Join(root, name), nil
}

func collectGoProfiles(prepared *preparedGoProfiles) ([]fixtures.GoProfileArtifact, error) {
	names := make([]string, 0, len(prepared.Paths))
	for name := range prepared.Paths {
		names = append(names, name)
	}
	sort.Strings(names)
	var artifacts []fixtures.GoProfileArtifact
	var failures []error
	for _, name := range names {
		output := prepared.Paths[name]
		path, err := goProfilePath(prepared.Root, output)
		if err != nil {
			return artifacts, err
		}
		if output.File != "" {
			artifact, err := inspectGoProfile(prepared.Root, name, path)
			artifacts = append(artifacts, artifact)
			failures = append(failures, err)
			continue
		}
		if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
			artifacts = append(artifacts, fixtures.GoProfileArtifact{Name: name, Status: "not_emitted"})
			continue
		} else if err != nil {
			failures = append(failures, fmt.Errorf("goProfiles.%s: inspect directory: %w", name, err))
			continue
		}
		found := false
		err = filepath.WalkDir(path, func(file string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			found = true
			rel, err := filepath.Rel(path, file)
			if err != nil {
				return err
			}
			artifact, inspectErr := inspectGoProfile(prepared.Root, name+"/"+filepath.ToSlash(rel), file)
			artifacts = append(artifacts, artifact)
			failures = append(failures, inspectErr)
			return nil
		})
		if err != nil {
			failures = append(failures, fmt.Errorf("goProfiles.%s: inspect directory: %w", name, err))
		}
		if !found && err == nil {
			artifacts = append(artifacts, fixtures.GoProfileArtifact{Name: name, Status: "not_emitted"})
		}
	}
	return artifacts, errors.Join(failures...)
}

func inspectGoProfile(root, name, path string) (fixtures.GoProfileArtifact, error) {
	artifact := fixtures.GoProfileArtifact{Name: name, Path: path, Status: "not_emitted"}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		artifact.Path = ""
		return artifact, nil
	}
	if err != nil {
		return invalidGoProfile(artifact, err)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return invalidGoProfile(artifact, err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return invalidGoProfile(artifact, err)
	}
	relToRoot, err := filepath.Rel(resolvedRoot, resolved)
	if err != nil || !filepath.IsLocal(relToRoot) {
		return invalidGoProfile(artifact, fmt.Errorf("outside fixture profile directory"))
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return invalidGoProfile(artifact, fmt.Errorf("expected a nonempty regular pprof file"))
	}
	file, err := os.Open(path)
	if err != nil {
		return invalidGoProfile(artifact, err)
	}
	defer file.Close()
	parsed, err := profile.Parse(file)
	if err != nil {
		return invalidGoProfile(artifact, err)
	}
	rel, err := filepath.Rel(filepath.Dir(root), path)
	if err != nil {
		return invalidGoProfile(artifact, err)
	}
	artifact.ID = filepath.ToSlash(rel)
	artifact.Bytes = info.Size()
	artifact.Status = "captured"
	for _, kind := range parsed.SampleType {
		artifact.SampleTypes = append(artifact.SampleTypes, kind.Type+"/"+kind.Unit)
	}
	return artifact, nil
}

func invalidGoProfile(artifact fixtures.GoProfileArtifact, cause error) (fixtures.GoProfileArtifact, error) {
	artifact.Status = "invalid"
	artifact.Error = cause.Error()
	return artifact, fmt.Errorf("go profile %s at %s: %w", artifact.Name, artifact.Path, cause)
}
