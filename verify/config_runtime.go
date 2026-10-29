package verify

import (
	"fmt"
	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/runconfig"
	"path/filepath"
	"slices"
)

type TestFrameworkConfig struct {
	Enabled *bool                  `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Options runconfig.Test         `yaml:"options,omitempty" json:"options,omitempty"`
	Args    []string               `yaml:"args,omitempty" json:"args,omitempty"`
	Runtime processruntime.Options `yaml:"runtime,omitempty" json:"runtime,omitempty"`
}

type RuntimeLayer struct {
	Global      processruntime.Options
	Lint        processruntime.Options
	Test        processruntime.Options
	Linters     map[string]LintLinterConfig
	Frameworks  map[string]TestFrameworkConfig
	LintOptions runconfig.Lint
	TestOptions runconfig.Test
}

func (c GavelConfig) RuntimeFor(category, name string, caller processruntime.Options) (processruntime.Options, []string, error) {
	if category != "lint" && category != "test" {
		return processruntime.Options{}, nil, fmt.Errorf("unknown runtime category %q", category)
	}
	layers := c.RuntimeLayers
	if len(layers) == 0 {
		layers = []RuntimeLayer{c.runtimeLayer()}
	}
	var runtime processruntime.Options
	var args []string
	for _, layer := range layers {
		runtime = runtime.Merge(layer.Global)
		if category == "lint" {
			runtime = runtime.Merge(layer.Lint).Merge(layer.Linters[name].Runtime)
			if layer.Linters[name].Args != nil {
				args = slices.Clone(layer.Linters[name].Args)
			}
		} else {
			runtime = runtime.Merge(layer.Test).Merge(layer.Frameworks[frameworkKey(name)].Runtime)
			if layer.Frameworks[frameworkKey(name)].Args != nil {
				args = slices.Clone(layer.Frameworks[frameworkKey(name)].Args)
			}
		}
	}
	runtime = runtime.Merge(caller)
	return runtime, args, runtime.Validate()
}

func (c GavelConfig) runtimeLayer() RuntimeLayer {
	frameworks := make(map[string]TestFrameworkConfig, len(c.Test.Frameworks))
	for name, config := range c.Test.Frameworks {
		frameworks[frameworkKey(name)] = config
	}
	lint, test := c.Lint.Runtime, c.Test.Runtime
	if c.Test.LintTimeout != "" && lint.Timeout == nil {
		lint.Timeout = &c.Test.LintTimeout
	}
	if c.Test.TestTimeout != "" && test.Timeout == nil {
		test.Timeout = &c.Test.TestTimeout
	}
	return RuntimeLayer{Global: c.Runtime, Lint: lint, Test: test, Linters: c.Lint.Linters, Frameworks: frameworks, LintOptions: c.Lint.Options, TestOptions: c.Test.Options}
}

func frameworkKey(name string) string {
	switch name {
	case "go", "go test", "gotest", "go-test", "go_test":
		return "go"
	}
	return name
}

func (c GavelConfig) validateRuntime() error {
	if err := runconfig.Validate(c.Lint.Options, false); err != nil {
		return fmt.Errorf("lint.options: %w", err)
	}
	if err := runconfig.Validate(c.Test.Options, false); err != nil {
		return fmt.Errorf("test.options: %w", err)
	}
	if n := c.Commit.Lint.AIFixMaxIterations; n != nil && *n < 1 {
		return fmt.Errorf("commit.lint.aiFixMaxIterations must be positive")
	}
	for _, options := range []processruntime.Options{c.Runtime, c.Lint.Runtime, c.Test.Runtime} {
		if err := options.Validate(); err != nil {
			return err
		}
	}
	for name, config := range c.Lint.Linters {
		if err := runconfig.Validate(config.Options, true); err != nil {
			return fmt.Errorf("lint.linters.%s.options: %w", name, err)
		}
		builtin := slices.Contains(BuiltinLinterNames, name)
		switch {
		case builtin && len(config.Command) > 0:
			return fmt.Errorf("lint.linters.%s declares a command, but %s is a built-in linter; name the command linter differently", name, name)
		case !builtin && len(config.Command) == 0:
			return fmt.Errorf("unknown linter %q in lint.linters; declare its command to add an external linter", name)
		case len(config.Includes) > 0 && len(config.Command) == 0:
			return fmt.Errorf("lint.linters.%s.includes applies only to a linter declared by command", name)
		}
		if len(config.Config) > 0 && name != "uir" {
			return fmt.Errorf("lint.linters.%s.config: native config documents are supported only for uir", name)
		}
		if err := config.Runtime.Validate(); err != nil {
			return fmt.Errorf("lint.linters.%s: %w", name, err)
		}
	}
	frameworks := make(map[string]bool)
	for name, config := range c.Test.Frameworks {
		if err := runconfig.Validate(config.Options, true); err != nil {
			return fmt.Errorf("test.frameworks.%s.options: %w", name, err)
		}
		key := frameworkKey(name)
		if !slices.Contains([]string{"go", "ginkgo", "jest", "vitest", "playwright", "fixture"}, key) {
			return fmt.Errorf("unknown framework %q in test.frameworks", name)
		}
		if frameworks[key] {
			return fmt.Errorf("duplicate framework %q in test.frameworks", key)
		}
		frameworks[key] = true
		if err := config.Runtime.Validate(); err != nil {
			return fmt.Errorf("test.frameworks.%s: %w", name, err)
		}
	}
	return nil
}

var BuiltinLinterNames = []string{"golangci-lint", "ruff", "eslint", "oxlint", "react-doctor", "pyright", "tsc", "markdownlint", "vale", "jscpd", "betterleaks", "uir", "json"}

func RuntimeConfigDirectory(executionDir, invocationDir string) string {
	if invocationDir == "" {
		return executionDir
	}
	relative, err := filepath.Rel(executionDir, invocationDir)
	if err != nil || !filepath.IsLocal(relative) {
		return executionDir
	}
	return invocationDir
}
