package uirexport

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

// Environment returns the effective Go build and test identity for a run.
func Environment(moduleRoot, command string) (map[string]string, error) {
	goEnv := exec.Command("go", "env", "-json", "GOOS", "GOARCH", "GOVERSION", "GOFLAGS", "CGO_ENABLED", "GOEXPERIMENT")
	goEnv.Dir = moduleRoot
	output, err := goEnv.Output()
	if err != nil {
		return nil, fmt.Errorf("go env: %w", err)
	}
	var values map[string]string
	if err := json.Unmarshal(output, &values); err != nil {
		return nil, fmt.Errorf("decode go env: %w", err)
	}
	env := map[string]string{
		"GOOS": values["GOOS"], "GOARCH": values["GOARCH"], "go": values["GOVERSION"],
		"CGO_ENABLED": values["CGO_ENABLED"], "GOEXPERIMENT": canonicalList(values["GOEXPERIMENT"]),
		"race": "false", "tags": "", "count": "",
	}
	args := strings.Fields(values["GOFLAGS"])
	if before, _, found := strings.Cut(command, " -args "); found {
		command = before
	}
	args = append(args, strings.Fields(command)...)
	flags := map[string]string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			continue
		}
		name, value, hasValue := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		if !hasValue && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			i++
			value = args[i]
			hasValue = true
		}
		if !hasValue {
			value = "true"
		}
		flags[name] = value
		switch name {
		case "race":
			race, _ := strconv.ParseBool(value)
			env["race"] = strconv.FormatBool(race)
		case "tags", "count":
			if name == "tags" {
				value = canonicalList(value)
			}
			env[name] = value
		}
	}
	keys := make([]string, 0, len(flags))
	for name := range flags {
		keys = append(keys, name)
	}
	slices.Sort(keys)
	canonical := make([]string, 0, len(keys))
	for _, name := range keys {
		value := flags[name]
		if name == "tags" {
			value = canonicalList(value)
		}
		canonical = append(canonical, name+"="+value)
	}
	env["GOFLAGS"] = strings.Join(canonical, "\x00")
	return env, nil
}

func canonicalList(value string) string {
	items := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' })
	slices.Sort(items)
	return strings.Join(slices.Compact(items), ",")
}
