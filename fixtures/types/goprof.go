package types

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/flanksource/commons-db/tracing/goprof"
	"github.com/flanksource/commons/logger"
	"github.com/google/pprof/profile"

	"github.com/flanksource/gavel/fixtures"
)

// preparedGoProfiles are where a command writes its `goProfiles:` under
// --profile: the arguments and environment that tell it, and each entry by
// name.
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
	for _, name := range sortedNames(outputs) {
		if err := prepared.add(name, outputs[name], destinations); err != nil {
			return nil, fmt.Errorf("goProfiles.%s: %w", name, err)
		}
	}
	return prepared, nil
}

func (p *preparedGoProfiles) add(name string, output fixtures.GoProfileOutput, destinations map[string]string) error {
	path, err := goProfilePath(p.Root, output)
	if err != nil {
		return err
	}
	if prior, exists := destinations[path]; exists {
		return fmt.Errorf("destination is already used by goProfiles.%s", prior)
	}
	destinations[path] = name
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create parent: %w", err)
	}
	if output.Directory != "" {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
	}
	if output.Env != "" {
		if _, exists := p.Env[output.Env]; exists {
			return fmt.Errorf("environment variable %s is already assigned", output.Env)
		}
		p.Env[output.Env] = path
	}
	for _, arg := range output.Args {
		if !strings.Contains(arg, "{{.path}}") {
			return fmt.Errorf("argument %q must contain {{.path}}", arg)
		}
		p.Args = append(p.Args, strings.ReplaceAll(arg, "{{.path}}", path))
	}
	p.Paths[name] = output
	return nil
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

// ingestGoProfiles reads every profile the command pid wrote into the
// fixture's goprof trace, one window per file. An entry the command did not
// write is logged, not an error.
func ingestGoProfiles(ctx context.Context, traces *fixtures.FixtureTraces, prepared *preparedGoProfiles, pid int) error {
	var failures []error
	for _, name := range sortedNames(prepared.Paths) {
		files, err := goProfileFiles(prepared.Root, name, prepared.Paths[name])
		if err != nil {
			failures = append(failures, fmt.Errorf("goProfiles.%s: %w", name, err))
			continue
		}
		if len(files) == 0 {
			logger.Infof("goProfiles.%s: the command wrote no profile", name)
		}
		for _, file := range files {
			file.PID = pid
			failures = append(failures, traces.IngestGoProfile(ctx, file))
		}
	}
	return errors.Join(failures...)
}

// goProfileFiles are the profiles entry name's output holds: its file, or
// every file under its directory, each with the profile it is.
func goProfileFiles(root, name string, output fixtures.GoProfileOutput) ([]fixtures.GoProfileFile, error) {
	path, err := goProfilePath(root, output)
	if err != nil {
		return nil, err
	}
	var paths []string
	if output.File != "" {
		paths = []string{path}
	} else if paths, err = filesUnder(path); err != nil {
		return nil, err
	}
	var files []fixtures.GoProfileFile
	for _, file := range paths {
		if _, err := os.Lstat(file); errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err := insideRoot(root, file); err != nil {
			return nil, err
		}
		kind, err := profileOf(name, file)
		if err != nil {
			return nil, err
		}
		files = append(files, fixtures.GoProfileFile{Name: name, Path: file, Profile: kind})
	}
	return files, nil
}

// insideRoot refuses a profile that resolves, through a symlink, outside the
// fixture's profile directory, or is not a regular file.
func insideRoot(root, path string) error {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	if rel, err := filepath.Rel(resolvedRoot, resolved); err != nil || !filepath.IsLocal(rel) {
		return fmt.Errorf("profile %s is outside the fixture profile directory", path)
	}
	if info, err := os.Stat(resolved); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("profile %s is not a regular file", path)
	}
	return nil
}

func filesUnder(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(file string, entry fs.DirEntry, err error) error {
		if err == nil && !entry.IsDir() {
			files = append(files, file)
		}
		return err
	})
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return files, err
}

// profileOf is which goprof profile the file at path is: the entry's name
// when it is one, else what its sample types tell.
func profileOf(name, path string) (string, error) {
	if slices.Contains(goprof.Profiles, name) {
		return name, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	parsed, err := profile.Parse(file)
	if err != nil {
		return "", fmt.Errorf("%s is not a pprof profile: %w", path, err)
	}
	for _, sampleType := range parsed.SampleType {
		switch sampleType.Type {
		case "cpu":
			return "cpu", nil
		case "inuse_space", "alloc_space":
			return "heap", nil
		case "goroutine":
			return "goroutine", nil
		}
	}
	return "", fmt.Errorf("cannot tell which profile %s is: name the entry one of %v", path, goprof.Profiles)
}

func sortedNames[V any](entries map[string]V) []string {
	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
