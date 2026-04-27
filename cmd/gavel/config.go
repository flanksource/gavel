package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
	gaveldocs "github.com/flanksource/gavel"
	"github.com/flanksource/gavel/verify"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ConfigOptions struct {
	Args []string `json:"-" args:"true"`
}

type ConfigResult struct {
	TargetPath string                     `json:"targetPath" yaml:"targetPath"`
	TargetDir  string                     `json:"targetDir" yaml:"targetDir"`
	GitRoot    string                     `json:"gitRoot,omitempty" yaml:"gitRoot,omitempty"`
	Sources    []verify.GavelConfigSource `json:"sources,omitempty" yaml:"sources,omitempty"`
	Merged     verify.GavelConfig         `json:"merged" yaml:"merged"`
}

func init() {
	configCmd := clicky.AddNamedCommand("config", rootCmd, ConfigOptions{}, runConfig)
	configCmd.Use = "config [path]"
	configCmd.Short = "Show merged .gavel.yaml and the files that contributed to it"
	configCmd.Args = cobra.MaximumNArgs(1)
	configCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr, configHelp(cmd).ANSI())
	})
}

func runConfig(opts ConfigOptions) (any, error) {
	workDir, err := getWorkingDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	target := "."
	if len(opts.Args) > 0 && opts.Args[0] != "" {
		target = opts.Args[0]
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(workDir, target)
	}

	trace, err := verify.LoadGavelConfigTrace(target)
	if err != nil {
		return nil, fmt.Errorf("failed to load config trace: %w", err)
	}

	return ConfigResult{
		TargetPath: trace.TargetPath,
		TargetDir:  trace.TargetDir,
		GitRoot:    trace.GitRoot,
		Sources:    trace.Sources,
		Merged:     trace.Merged,
	}, nil
}

func configHelp(cmd *cobra.Command) api.Text {
	const (
		heading = "font-bold text-purple-600"
		flag    = "text-cyan-600 font-bold"
		code    = "text-green-500"
		muted   = "text-muted"
		mono    = "font-mono"
	)

	t := clicky.Text("Show the merged .gavel.yaml for a path and every file that contributed to it.", "font-bold").
		NewLine().NewLine().
		Append("USAGE", heading).NewLine().
		Append("  ").Append("gavel config [path]", code).NewLine().NewLine().
		Append("RESOLUTION ORDER", heading).NewLine().
		Append("  0. ", flag).Append("built-in defaults", muted).NewLine().
		Append("  1. ", flag).Append("~/.gavel.yaml", code).Append(" when present", muted).NewLine().
		Append("  2. ", flag).Append("<git-root>/.gavel.yaml", code).Append(" when the target is inside a git repo", muted).NewLine().
		Append("  3. ", flag).Append("<target-dir>/.gavel.yaml", code).Append(" or the parent directory when the target path is a file", muted).NewLine().NewLine().
		Append("OUTPUT", heading).NewLine().
		Append("  interactive output shows merged YAML with comments for non-git-root sources", muted).NewLine().
		Append("  redirected output writes merged YAML only", muted).NewLine().
		Append("  use ", muted).Append("--json", flag).Append(" or ", muted).Append("--yaml", flag).Append(" for machine-readable merged config", muted).NewLine().NewLine().
		Append("EXAMPLES", heading).NewLine().
		Append("  ").Append("gavel config", code).Append("                         inspect config for the current directory", muted).NewLine().
		Append("  ").Append("gavel config ./pkg/api", code).Append("               inspect config for a nested directory", muted).NewLine().
		Append("  ").Append("gavel config ./cmd/gavel/main.go", code).Append("   inspect config for a specific file path", muted).NewLine().
		Append("  ").Append("gavel config --yaml", code).Append("                 emit merged config as YAML", muted).NewLine().
		Append("  ").Append("gavel config > merged.gavel.yaml", code).Append("      write merged YAML without source comments", muted).NewLine().
		Append("  ").Append("gavel --cwd ../repo config src/app.ts", code).Append("  resolve a relative path from another working tree", muted).NewLine().NewLine().
		Append("UBER EXAMPLE", heading).NewLine().
		Append("  bundled from ", muted).Append("gavel.yaml.example", code).Append(" so the file and help stay in sync", muted).NewLine().
		Append(indentBlock(gaveldocs.GavelConfigExample), mono).
		NewLine().
		Add(renderHelpFlags("FLAGS", cmd.NonInheritedFlags())).
		Add(renderHelpFlags("GLOBAL FLAGS", cmd.InheritedFlags()))

	return t
}

func renderHelpFlags(title string, flags *pflag.FlagSet) api.Text {
	const (
		heading = "font-bold text-purple-600"
		flag    = "text-cyan-600 font-bold"
		muted   = "text-muted"
	)

	t := clicky.Text(title, heading).NewLine()
	count := 0

	flags.VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		count++
		t = t.Append("  ").Append(formatFlagUsage(f), flag).Append("  ").Append(formatFlagDescription(f), muted).NewLine()
	})

	if count == 0 {
		t = t.Append("  none", muted).NewLine()
	}

	return t.NewLine()
}

func formatFlagUsage(f *pflag.Flag) string {
	parts := make([]string, 0, 2)
	if f.Shorthand != "" && f.ShorthandDeprecated == "" {
		parts = append(parts, "-"+f.Shorthand)
	}

	long := "--" + f.Name
	if f.Value.Type() != "bool" {
		long += " " + flagValueName(f)
	}
	parts = append(parts, long)

	return strings.Join(parts, ", ")
}

func flagValueName(f *pflag.Flag) string {
	switch f.Value.Type() {
	case "duration":
		return "duration"
	case "int":
		return "int"
	case "string":
		return "string"
	case "stringSlice":
		return "strings"
	default:
		return f.Value.Type()
	}
}

func formatFlagDescription(f *pflag.Flag) string {
	desc := f.Usage
	if shouldShowFlagDefault(f) {
		desc += fmt.Sprintf(" (default %s)", f.DefValue)
	}
	return desc
}

func shouldShowFlagDefault(f *pflag.Flag) bool {
	if f.DefValue == "" {
		return false
	}
	if f.Value.Type() == "bool" && f.DefValue == "false" {
		return false
	}
	return true
}

func configOriginLabel(origin string) string {
	switch origin {
	case "user-home":
		return "user home"
	case "git-root":
		return "git root"
	case "parent-directory":
		return "parent directory"
	case "target-directory":
		return "target directory"
	default:
		return origin
	}
}

func mustMarshalYAML(v any) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Sprintf("marshal error: %v\n", err)
	}

	s := strings.TrimRight(string(data), "\n")
	if s == "" {
		return "{}\n"
	}
	return s + "\n"
}

func prettyYAML(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return mustMarshalYAML(v)
	}

	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return mustMarshalYAML(v)
	}

	pruned := pruneEmpty(parsed)
	if pruned == nil {
		return "{}\n"
	}

	return mustMarshalYAML(pruned)
}

func pruneEmpty(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any)
		for k, val := range x {
			pruned := pruneEmpty(val)
			if pruned == nil {
				continue
			}
			out[k] = pruned
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case []any:
		out := make([]any, 0, len(x))
		for _, val := range x {
			pruned := pruneEmpty(val)
			if pruned == nil {
				continue
			}
			out = append(out, pruned)
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case string:
		if x == "" {
			return nil
		}
		return x
	case bool:
		if !x {
			return nil
		}
		return x
	case float64:
		if x == 0 {
			return nil
		}
		return x
	case nil:
		return nil
	default:
		return x
	}
}

func indentBlock(s string) string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return "    {}\n"
	}

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = "    " + line
	}
	return strings.Join(lines, "\n") + "\n"
}
