package verify

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
)

type CLITool struct {
	Binary    string
	BuildArgs func(prompt, model, schemaFile string, debug bool) []string
}

var cliTools = map[string]CLITool{
	"claude": {
		Binary: "claude",
		BuildArgs: func(prompt, model, schemaFile string, debug bool) []string {
			args := []string{"-p", prompt, "--output-format", "json"}
			if model != "" && model != "claude" {
				args = append(args, "--model", model)
			}
			if schemaFile != "" {
				if data, err := os.ReadFile(schemaFile); err == nil {
					args = append(args, "--json-schema", string(data))
				}
			}
			if debug {
				args = append(args, "--verbose")
			}
			return args
		},
	},
	"gemini": {
		Binary: "gemini",
		BuildArgs: func(prompt, model, _ string, debug bool) []string {
			args := []string{"-p", prompt, "--output-format", "json"}
			if model != "" && model != "gemini" {
				args = append(args, "-m", model)
			}
			if debug {
				args = append(args, "--debug")
			}
			return args
		},
	},
	"codex": {
		Binary: "codex",
		BuildArgs: func(prompt, model, schemaFile string, debug bool) []string {
			args := []string{"exec", "--json"}
			if model != "" && model != "codex" {
				args = append(args, "-m", model)
			}
			if schemaFile != "" {
				args = append(args, "--output-schema", schemaFile)
			}
			args = append(args, "--", prompt)
			return args
		},
	},
}

func ResolveCLI(model string) (CLITool, string) {
	if tool, ok := cliTools[model]; ok {
		return tool, model
	}
	prefix := model
	if idx := strings.IndexByte(model, '-'); idx > 0 {
		prefix = model[:idx]
	}
	if tool, ok := cliTools[prefix]; ok {
		return tool, model
	}
	return cliTools["claude"], model
}

func Execute(tool CLITool, prompt, model, schemaFile, workDir string, debug bool) (string, error) {
	args := tool.BuildArgs(prompt, model, schemaFile, debug)

	logger.V(1).Infof("exec: %s %s", tool.Binary, strings.Join(args, " "))

	proc := clicky.Exec(tool.Binary, args...).
		WithCwd(workDir).
		WithTimeout(5 * time.Minute).
		Stream(os.Stderr, os.Stderr)

	if debug {
		proc = proc.Debug()
	}

	result := proc.Run().Result()

	logger.V(1).Infof("stdout: %s", result.Stdout)
	if result.Stderr != "" {
		logger.V(1).Infof("stderr: %s", result.Stderr)
	}

	if result.Error != nil {
		return "", fmt.Errorf("%s failed: %w\nstderr: %s", tool.Binary, result.Error, result.Stderr)
	}

	if result.ExitCode != 0 {
		return "", fmt.Errorf("%s exited with code %d\nstderr: %s", tool.Binary, result.ExitCode, result.Stderr)
	}

	return result.Stdout, nil
}
