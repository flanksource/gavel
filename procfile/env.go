package procfile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// LoadDotEnv reads a foreman-style .env file: KEY=value per line, `#` comments,
// an optional leading `export `, and optional matching single/double quotes
// around the value. A missing file returns an empty map and no error so callers
// can treat .env as optional. Malformed lines are a loud error.
func LoadDotEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	env := map[string]string{}
	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		raw = strings.TrimPrefix(raw, "export ")
		key, val, ok := strings.Cut(raw, "=")
		if !ok {
			return nil, fmt.Errorf("%s line %d: expected KEY=value, got %q", path, lineNo, raw)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("%s line %d: empty key", path, lineNo)
		}
		env[key] = unquote(strings.TrimSpace(val))
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return env, nil
}

func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// MergeEnv layers the given maps left-to-right (later wins) into a fresh map.
// It does NOT include the parent process environment — the result is the
// overlay handed to clicky.Exec via WithEnv, which prepends os.Environ() itself.
// The process's shell expands $VAR references against this environment at
// runtime. nil layers are skipped.
func MergeEnv(layers ...map[string]string) map[string]string {
	env := map[string]string{}
	for _, layer := range layers {
		for k, v := range layer {
			env[k] = v
		}
	}
	return env
}
