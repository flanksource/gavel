package testrunner

import (
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/testrunner/runners"
	"github.com/samber/lo"
)

// Registry manages test runners and result parsers.
type Registry struct {
	workDir string
	runners map[Framework]runners.Runner
	parsers map[string]parsers.ResultParser
}

// NewRegistry creates a new registry.
func NewRegistry(workDir string) *Registry {
	return &Registry{
		workDir: workDir,
		runners: make(map[Framework]runners.Runner),
		parsers: make(map[string]parsers.ResultParser),
	}
}

// DefaultRegistry creates a registry with GoTest and Ginkgo runners pre-registered.
func DefaultRegistry(workDir string) *Registry {
	reg := NewRegistry(workDir)

	// Register parsers
	goTestParser := parsers.NewGoTestJSON(workDir)
	reg.RegisterParser(goTestParser)

	// Register runners
	goTestRunner := runners.NewGoTest(workDir)
	ginkgoRunner := runners.NewGinkgo(workDir)

	reg.Register(goTestRunner)
	reg.Register(ginkgoRunner)

	return reg
}

// Register adds a test runner to the registry.
func (r *Registry) Register(runner runners.Runner) {
	if runner == nil {
		return
	}
	r.runners[runner.Name()] = runner
}

// RegisterParser adds a result parser to the registry.
func (r *Registry) RegisterParser(parser parsers.ResultParser) {
	if parser == nil {
		return
	}
	r.parsers[parser.Name()] = parser
}

// Get retrieves a runner by framework.
func (r *Registry) Get(framework Framework) (runners.Runner, bool) {
	runner, ok := r.runners[framework]
	return runner, ok
}

// GetParser retrieves a parser by name.
func (r *Registry) GetParser(name string) (parsers.ResultParser, bool) {
	parser, ok := r.parsers[name]
	return parser, ok
}

// DetectAll returns all frameworks that are detected in the working directory.
func (r *Registry) DetectAll() ([]Framework, error) {
	var detected []Framework

	for framework, runner := range r.runners {
		found, err := runner.Detect(r.workDir)
		if err != nil {
			return nil, fmt.Errorf("failed to detect %s: %w", framework, err)
		}
		if found {
			detected = append(detected, framework)
		}
	}

	return detected, nil
}

func (r Registry) Pretty() api.Text {
	t := clicky.Text("")
	t = t.Append("runners: ", "text-muted").Append(clicky.CompactList(lo.Keys(r.runners)))
	if len(r.parsers) > 0 {
		t = t.Space().Append("parsers: ", "text-muted").Append(clicky.CompactList(lo.Keys(r.parsers)))
	}
	return t
}
