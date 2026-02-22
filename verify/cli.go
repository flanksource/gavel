package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flanksource/clicky"
	"github.com/flanksource/commons/logger"
)

var adapters = map[string]Adapter{
	"claude": Claude{},
	"gemini": Gemini{},
	"codex":  Codex{},
}

func ResolveAdapter(model string) (Adapter, string) {
	if a, ok := adapters[model]; ok {
		return a, model
	}
	prefix := model
	if idx := strings.IndexByte(model, '-'); idx > 0 {
		prefix = model[:idx]
	}
	if a, ok := adapters[prefix]; ok {
		return a, model
	}
	return adapters["claude"], model
}

func Execute(adapter Adapter, prompt, model, schemaFile, workDir string, debug bool) (string, error) {
	args := adapter.BuildVerifyArgs(prompt, model, schemaFile, debug)
	name := adapter.Name()

	logger.V(1).Infof("exec: %s %s", name, strings.Join(args, " "))

	proc := clicky.Exec(name, args...).
		WithCwd(workDir).
		WithTimeout(5*time.Minute).
		Stream(os.Stderr, os.Stderr)

	if debug {
		proc = proc.Debug()
	}

	result := proc.Run().Result()

	logger.V(1).Infof("stdout: %s", result.Stdout)
	if result.Stderr != "" {
		logger.V(1).Infof("stderr: %s", result.Stderr)
	}

	if result.Error != nil || result.ExitCode != 0 {
		msg := extractErrorMessage(result.Stdout)
		if msg == "" {
			msg = result.Stderr
		}
		if containsModelError(msg) {
			if hint := formatModelHint(adapter); hint != "" {
				msg = msg + "\n" + hint
			}
		}
		if result.Error != nil {
			return "", fmt.Errorf("%s failed: %w\n%s", name, result.Error, msg)
		}
		return "", fmt.Errorf("%s exited with code %d\n%s", name, result.ExitCode, msg)
	}
	return result.Stdout, nil
}

func extractErrorMessage(stdout string) string {
	stdout = strings.TrimSpace(stdout)
	if !strings.HasPrefix(stdout, "{") {
		return ""
	}
	var envelope struct {
		Result  string `json:"result"`
		IsError bool   `json:"is_error"`
	}
	if json.Unmarshal([]byte(stdout), &envelope) == nil && envelope.Result != "" {
		return envelope.Result
	}
	return ""
}
