package testrunner

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/flanksource/gavel/runconfig"
	"github.com/flanksource/gavel/testrunner/parsers"
	"github.com/flanksource/gavel/verify"
)

func (o *RunOptions) Configure() error {
	if o.OptionFields == nil {
		o.OptionFields = runconfig.Present(*o)
	}
	if o.WorkDir == "" {
		var err error
		o.WorkDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	cfg, err := verify.LoadGavelConfig(o.WorkDir)
	if err != nil {
		return err
	}
	options := cfg.TestOptionsFor("")
	options.ExtraArgs, options.Tags, options.Nodes, options.Bench, options.TestTimeout = nil, nil, nil, nil, nil
	return runconfig.Apply(o, options, o.OptionFields)
}

func (o *TestOrchestrator) configuredPackages(framework parsers.Framework, packages []string) ([]string, error) {
	var selected []string
	for _, pkg := range packages {
		dir := pkg
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(o.WorkDir, pkg)
		}
		if info, err := os.Stat(dir); err != nil {
			return nil, err
		} else if !info.IsDir() {
			dir = filepath.Dir(dir)
		}
		cfg, err := verify.LoadGavelConfig(verify.RuntimeConfigDirectory(dir, o.runtimeConfigDir))
		if err != nil {
			return nil, err
		}
		options := o.RunOptions
		if err := runconfig.Apply(&options, cfg.TestOptionsFor(string(framework)), options.OptionFields); err != nil {
			return nil, err
		}
		name := string(framework)
		if framework == parsers.GoTest {
			name = "go"
		}
		configured := cfg.Test.Frameworks[name]
		if configured.Enabled != nil && !*configured.Enabled || packageIgnored(pkg, options.Ignore) {
			o.configurationScoped = true
			continue
		}
		selected = append(selected, pkg)
	}
	return selected, nil
}

func (o *RunOptions) UnmarshalJSON(data []byte) error {
	type plain RunOptions
	if err := json.Unmarshal(data, (*plain)(o)); err != nil {
		return err
	}
	fields, err := runconfig.JSONFields(data)
	if err != nil {
		return err
	}
	if fields["starting-paths"] {
		fields["paths"] = true
	}
	if fields["frameworks"] {
		fields["framework"] = true
	}
	o.OptionFields = fields
	return nil
}
