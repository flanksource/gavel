package runners

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/flanksource/clicky/exec"
)

// GoTestCommand builds one `go test` over opts.Packages. ginkgo names the
// packages that are Ginkgo suites: they get the reporter overlay, and when
// every package is one, names, skips and files render as Ginkgo filters
// instead of go test's -run/-skip.
//
// Packages always precede the user's flags: go test ends its package list at
// the first flag it does not know, so a flag such as -ginkgo.focus before
// the packages would make it test the module root instead.
func GoTestCommand(opts TestOptions, ginkgo []string) (*exec.Process, error) {
	if len(opts.Packages) == 0 {
		return nil, fmt.Errorf("go test needs at least one package")
	}
	asGinkgo := len(ginkgo) > 0 && len(ginkgo) == len(opts.Packages)
	if len(ginkgo) > 0 && !asGinkgo && opts.selectsTests() {
		return nil, fmt.Errorf("cannot filter tests in one go test run that mixes Ginkgo suites %v with plain go test packages", ginkgo)
	}
	args := []string{"test", "-json", "-v"}
	if len(ginkgo) > 0 {
		overlay, err := BuildGinkgoOverlay(opts.Cwd, ginkgo, opts.Tags)
		if err != nil {
			return nil, err
		}
		if overlay != "" {
			args = append(args, "-overlay="+overlay)
		}
	}
	if opts.Timeout > 0 {
		args = append(args, fmt.Sprintf("-timeout=%s", opts.Timeout))
	}
	if opts.CoverProfile != "" {
		args = append(args, "-coverprofile="+opts.CoverProfile)
	}
	if len(opts.Tags) > 0 {
		args = append(args, "-tags="+strings.Join(opts.Tags, ","))
	}
	if opts.Parallel > 0 {
		args = append(args, fmt.Sprintf("-p=%d", opts.Parallel))
	}
	args = append(args, opts.Packages...)
	filters, err := goTestFilters(opts, asGinkgo)
	if err != nil {
		return nil, err
	}
	args = append(args, filters...)
	if opts.FailFast {
		args = append(args, "-failfast")
	}
	if opts.Bench != "" {
		args = append(args, "-bench="+opts.Bench)
	}
	for _, arg := range opts.ExtraArgs {
		if !asGinkgo && isGinkgoFlag(arg) {
			continue
		}
		args = append(args, arg)
	}
	if asGinkgo {
		ginkgoArgs, err := ginkgoFilters(opts)
		if err != nil {
			return nil, err
		}
		args = append(args, ginkgoArgs...)
	}
	process := exec.NewExec("go", args...).WithCwd(opts.Cwd).WithProcessGroup()
	process.SucceedOnNonZero = true
	return process, nil
}

func (o TestOptions) selectsTests() bool {
	return !o.Focus.Empty() || !o.Skip.Empty() || !o.Files.Empty() || o.LabelFilter != ""
}

func isGinkgoFlag(arg string) bool {
	return strings.HasPrefix(strings.TrimLeft(arg, "-"), "ginkgo.")
}

// goTestFilters renders -run/-skip. A Ginkgo run gets -run only to pick the
// suite bootstrap of a file selection; its names go through ginkgoFilters.
func goTestFilters(opts TestOptions, asGinkgo bool) ([]string, error) {
	if asGinkgo {
		if opts.Files.Empty() || len(opts.Packages) != 1 {
			return nil, nil
		}
		files, err := opts.PackageFiles(opts.Packages[0], false)
		if err != nil || len(files) == 0 {
			return nil, err
		}
		suite, err := ginkgoSuiteFunc(filepath.Join(opts.Cwd, opts.Packages[0]))
		if err != nil {
			return nil, err
		}
		return []string{"-run=^" + suite + "$"}, nil
	}
	if opts.LabelFilter != "" {
		return nil, UnsupportedOptionError{Framework: "go test", Option: "label filters"}
	}
	var args []string
	run, err := goRunPattern(opts)
	if err != nil {
		return nil, err
	}
	if run != "" {
		args = append(args, "-run="+run)
	}
	if !opts.Skip.Empty() {
		skip, err := opts.Skip.Pattern()
		if err != nil {
			return nil, err
		}
		args = append(args, "-skip="+skip)
	}
	return args, nil
}

// goRunPattern is the focus, or, for a selection of test files, the exact
// names of the tests those files declare that the focus also selects.
func goRunPattern(opts TestOptions) (string, error) {
	var names []string
	fileFiltered := false
	for _, pkg := range opts.Packages {
		files, err := opts.PackageFiles(pkg, false)
		if err != nil {
			return "", err
		}
		if len(files) == 0 {
			continue
		}
		if len(opts.Packages) != 1 {
			return "", fmt.Errorf("test files %v need their own go test run", files)
		}
		declared, whole, err := goTestNames(filepath.Join(opts.Cwd, pkg), files)
		if err != nil {
			return "", err
		}
		if whole {
			break
		}
		fileFiltered = true
		for _, name := range declared {
			if !opts.Focus.Empty() {
				ok, err := opts.Focus.Matches(name)
				if err != nil {
					return "", err
				}
				if !ok {
					continue
				}
			}
			names = append(names, name)
		}
		if len(names) == 0 {
			return "", fmt.Errorf("no test in %v matches focus %s", files, opts.Focus)
		}
	}
	if fileFiltered {
		return FileSelection{Exact: names}.Pattern()
	}
	if opts.Focus.Empty() {
		return "", nil
	}
	return opts.Focus.Pattern()
}

var goTestFuncRe = regexp.MustCompile(`^Test[^a-z]`)

// goTestNames lists the Test functions the files declare. A non-test source
// file selects its whole package (whole = true).
func goTestNames(dir string, files []string) (names []string, whole bool, err error) {
	for _, file := range files {
		if !strings.HasSuffix(file, "_test.go") {
			return nil, true, nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, file), nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, false, fmt.Errorf("parse %s: %w", file, err)
		}
		found := false
		for _, decl := range parsed.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name == "TestMain" || (fn.Name.Name != "Test" && !goTestFuncRe.MatchString(fn.Name.Name)) {
				continue
			}
			found = true
			if !slices.Contains(names, fn.Name.Name) {
				names = append(names, fn.Name.Name)
			}
		}
		if !found {
			return nil, false, fmt.Errorf("%s declares no Test function", file)
		}
	}
	return names, false, nil
}

var runSpecsRe = regexp.MustCompile(`(?s)func\s+(Test\w*)\s*\([^)]*\*testing\.T\)\s*\{[^}]*RunSpecs\(`)

// ginkgoSuiteFunc finds the Test function that bootstraps dir's Ginkgo suite.
func ginkgoSuiteFunc(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return "", err
		}
		if match := runSpecsRe.FindSubmatch(content); match != nil {
			return string(match[1]), nil
		}
	}
	return "", fmt.Errorf("no Test function calling RunSpecs in %s", dir)
}

// ginkgoFilters renders the Ginkgo flags registered in the go test binary.
func ginkgoFilters(opts TestOptions) ([]string, error) {
	var args []string
	if !opts.Focus.Empty() {
		focus, err := opts.Focus.Pattern()
		if err != nil {
			return nil, err
		}
		args = append(args, "-ginkgo.focus="+focus)
	}
	if !opts.Skip.Empty() {
		skip, err := opts.Skip.Pattern()
		if err != nil {
			return nil, err
		}
		args = append(args, "-ginkgo.skip="+skip)
	}
	if opts.LabelFilter != "" {
		args = append(args, "-ginkgo.label-filter="+opts.LabelFilter)
	}
	if opts.FailFast {
		args = append(args, "-ginkgo.fail-fast")
	}
	for _, pkg := range opts.Packages {
		files, err := opts.PackageFiles(pkg, false)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			args = append(args, "-ginkgo.focus-file=(^|/)"+regexp.QuoteMeta(file)+"$")
		}
	}
	return args, nil
}
