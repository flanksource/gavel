package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"golang.org/x/term"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/kinds/dbconn"
	"github.com/flanksource/gavel/fixtures/kinds/form"
)

var (
	fixturesParams     []string
	fixturesParamsFile string
)

// fixtureRunParams are the values supplied up front for a fixture's `params:`
// form: the --params file, with each --param k=v over it. --param values stay
// strings; the form types them by its schema.
func fixtureRunParams(file string, pairs []string) (map[string]any, error) {
	params := map[string]any{}
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("--params: %w", err)
		}
		if err := yaml.Unmarshal(data, &params); err != nil {
			return nil, fmt.Errorf("--params %s: want a YAML or JSON object: %w", file, err)
		}
	}
	for _, pair := range pairs {
		key, value, ok := strings.Cut(pair, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("--param %q: want key=value", pair)
		}
		params[strings.TrimSpace(key)] = value
	}
	if len(params) == 0 {
		return nil, nil
	}
	return params, nil
}

// fixtureTerminalPrompter is the Prompter a `gavel fixtures` run without --ui
// answers forms with: the terminal when stdin is one, its connection pickers
// listing the gavel connection catalog as workDir sees it, else none — a form that
// must ask then fails naming what it needed. The --ui run supplies its own
// prompter on the RunnerOptions it is handed (testrunner/ui).
func fixtureTerminalPrompter(ctx context.Context, workDir string) (fixtures.Prompter, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil, nil
	}
	by, err := form.TerminalAnswerer(ctx)
	if err != nil {
		return nil, err
	}
	return form.TTYPrompter{By: by, Lookup: dbconn.TerminalLookup(workDir, dbconn.Options{})}, nil
}

func init() {
	fixturesCmd.Flags().StringArrayVar(&fixturesParams, "param", nil, "Supply a value for the fixture's params: form as key=value (repeatable; typed by the params schema)")
	fixturesCmd.Flags().StringVar(&fixturesParamsFile, "params", "", "Supply the fixture's params: form from a YAML or JSON file; --param values override it")
}
