package lint

import (
	"encoding/json"
	"fmt"
	"os"

	jsonlint "github.com/flanksource/gavel/linters/json"
	"github.com/flanksource/gavel/runconfig"
	"github.com/flanksource/gavel/verify"
)

func (o *Options) Configure() error {
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
	dir := o.ConfigDir
	if dir == "" {
		dir = o.WorkDir
	}
	cfg, err := verify.LoadGavelConfig(dir)
	if err != nil {
		return err
	}
	if err := runconfig.Apply(o, cfg.LintOptionsFor(""), o.OptionFields); err != nil {
		return err
	}
	if o.Schema != "" {
		o.Schema, err = jsonlint.ResolveSchema(o.Schema, o.WorkDir)
	}
	return err
}

func (o Options) forLinter(name, dir string) (Options, error) {
	cfg, err := verify.LoadGavelConfig(dir)
	if err != nil {
		return o, err
	}
	if err := runconfig.Apply(&o, cfg.LintOptionsFor(name), o.OptionFields); err != nil {
		return o, fmt.Errorf("%s %s: %w", dir, name, err)
	}
	return o, nil
}

func (o *Options) UnmarshalJSON(data []byte) error {
	type plain Options
	if err := json.Unmarshal(data, (*plain)(o)); err != nil {
		return err
	}
	fields, err := runconfig.JSONFields(data)
	if err != nil {
		return err
	}
	o.OptionFields = fields
	return nil
}
