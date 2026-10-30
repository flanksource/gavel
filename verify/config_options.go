package verify

import "github.com/flanksource/gavel/runconfig"

func (c GavelConfig) CommitFixSpec() PromptSpec { return Merge(c.Lint.Fix, c.Commit.Fix) }

func (c GavelConfig) LintOptionsFor(name string) runconfig.Lint {
	var options runconfig.Lint
	for _, layer := range c.configRuntimeLayers() {
		options = Merge(Merge(options, layer.LintOptions), layer.Linters[name].Options)
	}
	return options
}

func (c GavelConfig) TestOptionsFor(name string) runconfig.Test {
	var options runconfig.Test
	for _, layer := range c.configRuntimeLayers() {
		options = Merge(Merge(options, layer.TestOptions), layer.Frameworks[frameworkKey(name)].Options)
	}
	return options
}

func (c GavelConfig) configRuntimeLayers() []RuntimeLayer {
	if len(c.RuntimeLayers) == 0 {
		return []RuntimeLayer{c.runtimeLayer()}
	}
	return c.RuntimeLayers
}

func (c *GavelConfig) resolveOptionPaths(base string) {
	runconfig.ResolvePaths(&c.Lint.Options, base)
	runconfig.ResolvePaths(&c.Test.Options, base)
	for name, config := range c.Lint.Linters {
		runconfig.ResolvePaths(&config.Options, base)
		c.Lint.Linters[name] = config
	}
	for name, config := range c.Test.Frameworks {
		runconfig.ResolvePaths(&config.Options, base)
		c.Test.Frameworks[name] = config
	}
}
