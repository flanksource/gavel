package launch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
	"sync"

	"github.com/flanksource/gavel/lint"
	promptlint "github.com/flanksource/gavel/lint/prompt"
	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/processruntime"
	"github.com/flanksource/gavel/resultstore"
	"github.com/flanksource/gavel/runconfig"
	"github.com/flanksource/gavel/testrunner"
	"github.com/flanksource/gavel/workspace"
	"gopkg.in/yaml.v3"
)

type NativeExecutor struct{ Schemas SchemaProvider }

type nativeExecution struct {
	tests   *testrunner.RunOptions
	linters []nativeLinter
}

type nativeLinter struct {
	id     string
	opts   lint.Options
	prompt *promptlint.Prepared
}

func decodeNative(input any, output any) error {
	data, err := yaml.Marshal(input)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, output)
}

func (executor NativeExecutor) Prepare(ctx context.Context, ws PreparedWorkspace, config Configuration) (Execution, error) {
	dirs, err := workspace.Resolve(ws.Directory)
	if err != nil {
		return nil, err
	}
	execution := &nativeExecution{}
	env := processruntime.Options{Env: map[string]string{}}
	for _, item := range ws.Env {
		key, value, ok := strings.Cut(item, "=")
		if !ok {
			return nil, fmt.Errorf("invalid prepared environment entry %q", item)
		}
		env.Env[key] = value
	}
	if config.Test != nil {
		opts := testrunner.RunOptions{}
		options, err := executor.defaults("test", config.Test)
		if err != nil {
			return nil, err
		}
		if err := decodeNative(options, &opts); err != nil {
			return nil, err
		}
		opts.OptionFields, err = suppliedFields(config.Test)
		if err != nil {
			return nil, err
		}
		opts.OptionFields["work-dir"] = true
		opts.WorkDir, opts.Dirs, opts.Context = ws.Directory, dirs, ctx
		opts.Runtime = opts.Runtime.Merge(env)
		if err := opts.Configure(); err != nil {
			return nil, err
		}
		execution.tests = &opts
	}
	if config.Lint == nil {
		return execution, nil
	}
	entries := config.Lint.Linters
	if len(entries) == 0 {
		entries = []LinterConfig{{ID: "configured", Options: map[string]any{}}}
	}
	for _, entry := range entries {
		job := nativeLinter{id: entry.ID}
		if entry.Linter == "prompt" {
			scope, empty, err := (lint.Options{Context: ctx, WorkDir: ws.Directory, Files: config.Lint.Files, Changed: config.Lint.Changed, Since: config.Lint.Since}).WithChanges()
			if err != nil {
				return nil, err
			}
			if empty {
				continue
			}
			raw, err := json.Marshal(entry.Options)
			if err != nil {
				return nil, err
			}
			var options PromptConfig
			if err := json.Unmarshal(raw, &options); err != nil {
				return nil, err
			}
			job.prompt, err = promptlint.Prepare(ctx, promptlint.PrepareOptions{
				Directory: ws.Directory, Reference: options.Prompt, Version: options.PromptVersion,
				Variables: options.Variables, Presets: options.Presets, Spec: options.Spec,
				Paths: scope.Files, Env: ws.Env,
			})
			if err != nil {
				return nil, err
			}
		} else {
			options, err := executor.defaults("lint", entry.Options)
			if err != nil {
				return nil, err
			}
			if _, exists := options["linters"]; exists {
				return nil, errors.New("linter options cannot override the selected engine")
			}
			if len(config.Lint.Files) > 0 {
				options["files"] = config.Lint.Files
			}
			if config.Lint.Changed {
				options["changed"] = true
			}
			if config.Lint.Since != "" {
				options["since"] = config.Lint.Since
			}
			if err := decodeNative(options, &job.opts); err != nil {
				return nil, err
			}
			job.opts.OptionFields, err = suppliedFields(entry.Options)
			if err != nil {
				return nil, err
			}
			job.opts.OptionFields["work-dir"] = true
			if len(config.Lint.Files) > 0 {
				job.opts.OptionFields["files"] = true
			}
			if config.Lint.Changed {
				job.opts.OptionFields["changed"] = true
			}
			if config.Lint.Since != "" {
				job.opts.OptionFields["since"] = true
			}
			if entry.Linter != "" {
				job.opts.OptionFields["linters"] = true
			}
			job.opts.WorkDir, job.opts.Dirs, job.opts.Context = ws.Directory, dirs, ctx
			job.opts.Runtime = job.opts.Runtime.Merge(env)
			if entry.Linter != "" {
				job.opts.Linters = []string{entry.Linter}
			}
			if err := lint.Validate(job.opts); err != nil {
				return nil, err
			}
		}
		execution.linters = append(execution.linters, job)
	}
	return execution, nil
}

func suppliedFields(options map[string]any) (map[string]bool, error) {
	data, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}
	return runconfig.JSONFields(data)
}

func (executor NativeExecutor) defaults(kind string, input map[string]any) (map[string]any, error) {
	if executor.Schemas == nil {
		return nil, errors.New("native executor requires the native command schema provider")
	}
	schema, err := executor.Schemas.Schema(kind)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}
	var normalized map[string]any
	if err := json.Unmarshal(data, &normalized); err != nil {
		return nil, err
	}
	properties, ok := normalized["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s command schema has no properties", kind)
	}
	options := map[string]any{}
	for key, value := range properties {
		if property, ok := value.(map[string]any); ok {
			if initial, exists := property["default"]; exists {
				options[key] = initial
			}
		}
	}
	maps.Copy(options, input)
	return options, nil
}

func (e *nativeExecution) Run(ctx context.Context, publish Publish) error {
	var failures []error
	for _, job := range e.linters {
		if err := ctx.Err(); err != nil {
			return errors.Join(append(failures, err)...)
		}
		var results []*linters.LinterResult
		var err error
		if job.prompt != nil {
			publisher := &lintPublisher{id: job.id, publish: publish}
			results = job.prompt.Run(ctx, publisher)
			err = publisher.failure()
		} else {
			job.opts.Context = ctx
			results, err = lint.Run(ctx, job.opts)
		}
		for _, result := range results {
			result.ConfigurationID = job.id
		}
		if publishErr := publish(resultstore.Run{Lint: results}); publishErr != nil {
			return errors.Join(err, publishErr)
		}
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", job.id, err))
		}
	}
	if e.tests != nil && ctx.Err() == nil {
		failures = append(failures, runTests(ctx, *e.tests, publish))
	}
	return errors.Join(failures...)
}

type lintPublisher struct {
	id      string
	publish Publish
	mu      sync.Mutex
	err     error
}

func (p *lintPublisher) UpsertLintResult(result *linters.LinterResult) {
	p.mu.Lock()
	defer p.mu.Unlock()
	result.ConfigurationID = p.id
	if err := p.publish(resultstore.Run{Lint: []*linters.LinterResult{result}}); err != nil {
		p.err = errors.Join(p.err, err)
	}
}

func (p *lintPublisher) RemoveLintResult(linter, workDir string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.publish(resultstore.Run{RemoveLint: &linters.LinterResult{ConfigurationID: p.id, Linter: linter, WorkDir: workDir}}); err != nil {
		p.err = errors.Join(p.err, err)
	}
}

func (p *lintPublisher) failure() error { p.mu.Lock(); defer p.mu.Unlock(); return p.err }
