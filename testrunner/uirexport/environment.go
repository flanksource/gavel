package uirexport

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

// environment is what makes the run comparable with another: GOOS, GOARCH and the go toolchain from
// `go env` in the module root, and race, tags and count from GOFLAGS overridden by the go test command
// the run executed. Every key is always present, so an unset flag reads the same in every run.
func environment(moduleRoot, command string) (map[string]string, error) {
	goEnv := exec.Command("go", "env", "-json", "GOOS", "GOARCH", "GOVERSION", "GOFLAGS")
	goEnv.Dir = moduleRoot
	output, err := goEnv.Output()
	if err != nil {
		return nil, fmt.Errorf("go env: %w", err)
	}
	var values map[string]string
	if err := json.Unmarshal(output, &values); err != nil {
		return nil, fmt.Errorf("decode go env: %w", err)
	}
	env := map[string]string{"GOOS": values["GOOS"], "GOARCH": values["GOARCH"], "go": values["GOVERSION"], "race": "false", "tags": "", "count": ""}
	args := strings.Fields(values["GOFLAGS"])
	if before, _, found := strings.Cut(command, " -args "); found {
		command = before
	}
	args = append(args, strings.Fields(command)...)
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			continue
		}
		name, value, hasValue := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		switch name {
		case "race":
			race := true
			if hasValue {
				race, _ = strconv.ParseBool(value)
			}
			env["race"] = strconv.FormatBool(race)
		case "tags", "count":
			if !hasValue && i+1 < len(args) {
				value = args[i+1]
			}
			if name == "tags" {
				tags := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' })
				slices.Sort(tags)
				value = strings.Join(slices.Compact(tags), ",")
			}
			env[name] = value
		}
	}
	return env, nil
}
