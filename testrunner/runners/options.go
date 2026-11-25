package runners

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// TestOptions is the one description of a test run that every runner parses
// its arguments into and builds its command from, so focus, skips, files and
// parallelism mean the same thing whichever framework spells them.
type TestOptions struct {
	Cwd      string   `json:"cwd,omitempty"`      // execution root: git root, nearest go.mod, or the workdir
	Args     []string `json:"args,omitempty"`     // argv this was parsed from: positional paths, then -- and runner arguments
	Paths    []string `json:"paths,omitempty"`    // starting directories, Cwd-relative ("./x")
	Packages []string `json:"packages,omitempty"` // the packages BuildCommand runs, set after discovery

	Focus     FileSelection `json:"focus,omitempty"`      // test names to run
	Skip      FileSelection `json:"skip,omitempty"`       // test names to leave out
	Files     FileSelection `json:"files,omitempty"`      // test files to run, Cwd-relative
	SkipFiles FileSelection `json:"skip_files,omitempty"` // packages and files to leave out, Cwd-relative

	LabelFilter  string        `json:"label_filter,omitempty"`  // Ginkgo label filter
	Recursive    bool          `json:"recursive,omitempty"`     // discover packages below each path
	Tags         []string      `json:"tags,omitempty"`          // Go build tags
	Parallel     int           `json:"parallel,omitempty"`      // workers inside one runner process; 0 = the runner's default
	FailFast     bool          `json:"fail_fast,omitempty"`     // stop at the first failure
	Bench        string        `json:"bench,omitempty"`         // Go benchmark regexp
	Timeout      time.Duration `json:"timeout,omitempty"`       // subprocess deadline armed inside the test process
	CoverProfile string        `json:"cover_profile,omitempty"` // Go coverage profile the run writes
	ExtraArgs    []string      `json:"extra_args,omitempty"`    // runner arguments no parser recognised, passed verbatim

	consumed []bool
}

// FlagKind says which TestOptions field a flag sets.
type FlagKind int

const (
	FocusFlag FlagKind = iota
	SkipFlag
	FailFastFlag
	LabelFilterFlag
	TagsFlag
	BenchFlag
	ParallelFlag
	// NodesFlag is Ginkgo's --nodes: the unified go test runner runs one node.
	NodesFlag
)

// FlagSpec names the spellings of one runner flag. Each name is accepted with
// one or two leading dashes, as name=value or as name value.
type FlagSpec struct {
	Names []string
	Kind  FlagKind
}

// SharedFlags are recognised by every runner, so a focus written in go test,
// Ginkgo or vitest style selects tests in whichever frameworks run.
var SharedFlags = []FlagSpec{
	{Names: []string{"run", "focus", "ginkgo.focus", "t", "testNamePattern", "test-name-pattern"}, Kind: FocusFlag},
	{Names: []string{"skip", "ginkgo.skip"}, Kind: SkipFlag},
	{Names: []string{"failfast", "fail-fast", "ginkgo.fail-fast", "bail"}, Kind: FailFastFlag},
}

// ParseTestArgs parses argv relative to cwd. Arguments before "--" are
// starting paths: directories become Paths and files become Files, and a
// missing one is an error. After "--", the shared flags and the runner's own
// flags fill their fields, an existing path fills Paths or Files, and anything
// else is kept in ExtraArgs.
func ParseTestArgs(cwd string, argv []string, native ...FlagSpec) (*TestOptions, error) {
	opts := &TestOptions{Cwd: cwd, Args: slices.Clone(argv), consumed: make([]bool, len(argv))}
	specs := append(slices.Clone(SharedFlags), native...)
	runnerArgs := false
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if arg == "--" && !runnerArgs {
			runnerArgs = true
			opts.consumed[i] = true
			continue
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			added, err := opts.addPath(arg, !runnerArgs)
			if err != nil {
				return nil, err
			}
			opts.consumed[i] = added
			continue
		}
		if !runnerArgs {
			return nil, fmt.Errorf("unexpected flag %s before --", arg)
		}
		name, value, hasValue := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		spec, ok := findFlag(specs, name)
		if !ok {
			continue
		}
		opts.consumed[i] = true
		if spec.Kind == FailFastFlag {
			enabled := true
			if hasValue {
				parsed, err := strconv.ParseBool(value)
				if err != nil {
					return nil, fmt.Errorf("%s=%s: expected true or false", arg[:strings.Index(arg, "=")], value)
				}
				enabled = parsed
			}
			opts.FailFast = opts.FailFast || enabled
			continue
		}
		if !hasValue {
			if i+1 >= len(argv) || argv[i+1] == "--" {
				return nil, fmt.Errorf("%s requires a value", arg)
			}
			i++
			value = argv[i]
			opts.consumed[i] = true
		}
		if err := opts.set(spec.Kind, strings.SplitN(arg, "=", 2)[0], value); err != nil {
			return nil, err
		}
	}
	opts.ExtraArgs = opts.extraArgs()
	return opts, nil
}

func findFlag(specs []FlagSpec, name string) (FlagSpec, bool) {
	for _, spec := range specs {
		if slices.Contains(spec.Names, name) {
			return spec, true
		}
	}
	return FlagSpec{}, false
}

func (o *TestOptions) set(kind FlagKind, flag, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s requires a non-empty value", flag)
	}
	switch kind {
	case FocusFlag:
		o.Focus = o.Focus.Merge(FileSelection{Regex: []string{value}})
	case SkipFlag:
		o.Skip = o.Skip.Merge(FileSelection{Regex: []string{value}})
	case LabelFilterFlag:
		if o.LabelFilter != "" && o.LabelFilter != value {
			return fmt.Errorf("conflicting label filters %q and %q", o.LabelFilter, value)
		}
		o.LabelFilter = value
	case TagsFlag:
		for _, tag := range strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' }) {
			if !slices.Contains(o.Tags, tag) {
				o.Tags = append(o.Tags, tag)
			}
		}
	case BenchFlag:
		if o.Bench != "" && o.Bench != value {
			return fmt.Errorf("conflicting benchmark patterns %q and %q", o.Bench, value)
		}
		o.Bench = value
	case ParallelFlag:
		n, err := strconv.Atoi(value)
		if err != nil || n < 1 {
			return fmt.Errorf("%s=%s: expected a positive number", flag, value)
		}
		if o.Parallel != 0 && o.Parallel != n {
			return fmt.Errorf("conflicting parallelism %d and %d", o.Parallel, n)
		}
		o.Parallel = n
	case NodesFlag:
		if value != "1" {
			return fmt.Errorf("Ginkgo %s=%s is unsupported by the unified go test runner", flag, value)
		}
	}
	return nil
}

// addPath records an existing directory as a starting path and an existing
// file as a test file. A missing path is an error when required (gavel's own
// positional arguments) and otherwise left to the runner.
func (o *TestOptions) addPath(arg string, required bool) (bool, error) {
	full := arg
	if !filepath.IsAbs(full) {
		full = filepath.Join(o.Cwd, arg)
	}
	info, err := os.Stat(full)
	if err != nil {
		if required {
			return false, fmt.Errorf("starting path %s does not exist: %w", arg, err)
		}
		return false, nil
	}
	rel, err := filepath.Rel(o.Cwd, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
		if required {
			return false, fmt.Errorf("starting path %s is outside %s", arg, o.Cwd)
		}
		return false, nil
	}
	rel = "./" + filepath.ToSlash(rel)
	if info.IsDir() {
		if !slices.Contains(o.Paths, rel) {
			o.Paths = append(o.Paths, rel)
		}
	} else {
		o.Files = o.Files.Merge(FileSelection{Exact: []string{strings.TrimPrefix(rel, "./")}})
	}
	return true, nil
}

func (o *TestOptions) extraArgs() []string {
	var extra []string
	for i, arg := range o.Args {
		if i < len(o.consumed) && o.consumed[i] {
			continue
		}
		extra = append(extra, arg)
	}
	return extra
}

// Merge combines two descriptions of the same run. Selections and lists are
// unioned, a runner argument is extra only when neither side recognised it,
// and scalars that both sides set must agree.
func (o TestOptions) Merge(other TestOptions) (TestOptions, error) {
	out := o
	if out.Cwd == "" {
		out.Cwd = other.Cwd
	} else if other.Cwd != "" && other.Cwd != o.Cwd {
		return out, fmt.Errorf("cannot merge test options for %s and %s", o.Cwd, other.Cwd)
	}
	switch {
	case len(out.Args) == 0:
		out.Args, out.consumed, out.ExtraArgs = other.Args, other.consumed, other.ExtraArgs
	case len(other.Args) > 0 && !slices.Equal(o.Args, other.Args):
		// Different sources (the command line and a .gavel.yaml, say) add up.
		out.Args = append(slices.Clone(o.Args), other.Args...)
		out.consumed = make([]bool, 0, len(out.Args))
		for i := range o.Args {
			out.consumed = append(out.consumed, o.consumedAt(i))
		}
		for i := range other.Args {
			out.consumed = append(out.consumed, other.consumedAt(i))
		}
		out.ExtraArgs = out.extraArgs()
	case len(other.Args) > 0:
		out.consumed = make([]bool, len(o.Args))
		for i := range out.consumed {
			out.consumed[i] = o.consumedAt(i) || other.consumedAt(i)
		}
		out.ExtraArgs = out.extraArgs()
	}
	out.Paths = unionStrings(o.Paths, other.Paths)
	out.Packages = unionStrings(o.Packages, other.Packages)
	out.Tags = unionStrings(o.Tags, other.Tags)
	out.Focus, out.Skip = o.Focus.Merge(other.Focus), o.Skip.Merge(other.Skip)
	out.Files, out.SkipFiles = o.Files.Merge(other.Files), o.SkipFiles.Merge(other.SkipFiles)
	out.Recursive = o.Recursive || other.Recursive
	out.FailFast = o.FailFast || other.FailFast
	var err error
	if out.LabelFilter, err = agree("label filter", o.LabelFilter, other.LabelFilter); err != nil {
		return out, err
	}
	if out.Bench, err = agree("benchmark pattern", o.Bench, other.Bench); err != nil {
		return out, err
	}
	if out.Parallel, err = agree("parallelism", o.Parallel, other.Parallel); err != nil {
		return out, err
	}
	if out.Timeout, err = agree("timeout", o.Timeout, other.Timeout); err != nil {
		return out, err
	}
	if out.CoverProfile, err = agree("coverage profile", o.CoverProfile, other.CoverProfile); err != nil {
		return out, err
	}
	return out, nil
}

func (o TestOptions) consumedAt(i int) bool {
	return i < len(o.consumed) && o.consumed[i]
}

func agree[T comparable](name string, a, b T) (T, error) {
	var zero T
	switch {
	case a == zero:
		return b, nil
	case b == zero || a == b:
		return a, nil
	}
	return a, fmt.Errorf("conflicting %s %v and %v", name, a, b)
}

func unionStrings(a, b []string) []string {
	out := slices.Clone(a)
	for _, v := range b {
		if !slices.Contains(out, v) {
			out = append(out, v)
		}
	}
	return out
}

// UnsupportedOptionError reports an option a runner cannot express.
type UnsupportedOptionError struct {
	Framework string
	Option    string
}

func (e UnsupportedOptionError) Error() string {
	return fmt.Sprintf("%s does not support %s", e.Framework, e.Option)
}
