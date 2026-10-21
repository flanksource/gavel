// Package shims installs PATH shims that run heavy linters through gavel with
// fast-only defaults and CPU, memory and time limits.
package shims

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Limits are the throttles derived from the shim's --cpu-limit and --mem-limit.
type Limits struct {
	// Slots is the parallelism budget: one per 100% of CPU.
	Slots int
	// MemBytes is the hard memory limit; 0 disables the soft memory caps.
	MemBytes uint64
}

// Profile describes how one linter is made cheaper: the arguments injected into
// its invocation and the environment that caps its parallelism and memory.
type Profile struct {
	Tool   string
	Inject func(args []string, limits Limits) ([]string, error)
	Env    func(limits Limits, environ map[string]string) map[string]string
}

// Profiles lists every tool install-shims can shim, keyed by executable name.
var Profiles = map[string]Profile{
	"golangci-lint": {Tool: "golangci-lint", Inject: injectGolangci, Env: goEnv},
	"eslint":        {Tool: "eslint", Inject: injectESLint, Env: nodeEnv},
	"tsc":           {Tool: "tsc", Inject: func(args []string, _ Limits) ([]string, error) { return args, nil }, Env: nodeEnv},
}

// ToolNames returns the shimmable tools in a stable order.
func ToolNames() []string {
	names := make([]string, 0, len(Profiles))
	for name := range Profiles {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Slots converts a CPU percentage (200 = two cores) into a parallelism budget.
// A non-positive percentage disables the CPU limit and returns 0 (uncapped).
func Slots(cpuPercent float64) int {
	if cpuPercent <= 0 {
		return 0
	}
	return max(1, int(math.Ceil(cpuPercent/100)))
}

func injectGolangci(args []string, limits Limits) ([]string, error) {
	run := slices.IndexFunc(args, func(arg string) bool { return !strings.HasPrefix(arg, "-") })
	if run < 0 || args[run] != "run" {
		return args, nil
	}
	var extra []string
	if !hasFlag(args, "--fast-only") && !hasFlag(args, "--enable-only") {
		extra = append(extra, "--fast-only")
	}
	if limits.Slots > 0 && !hasFlag(args, "--concurrency") && !hasFlag(args, "-j") {
		extra = append(extra, fmt.Sprintf("--concurrency=%d", limits.Slots))
	}
	// golangci-lint serializes runs on a per-user lock; each shimmed run is
	// already capped, so queueing behind other runs would only burn the timeout.
	if !hasFlag(args, "--allow-parallel-runners") && !hasFlag(args, "--allow-serial-runners") {
		extra = append(extra, "--allow-parallel-runners")
	}
	return slices.Concat(args[:run+1], extra, args[run+1:]), nil
}

func injectESLint(args []string, _ Limits) ([]string, error) {
	if hasFlag(args, "--cache") || hasFlag(args, "--no-cache") || hasFlag(args, "--cache-location") {
		return args, nil
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("eslint cache location: %w", err)
	}
	return append([]string{"--cache", "--cache-location", filepath.Join(cache, "gavel", "eslint") + string(filepath.Separator)}, args...), nil
}

func goEnv(limits Limits, environ map[string]string) map[string]string {
	env := slotEnv("GOMAXPROCS", limits)
	if limits.MemBytes > 0 && environ["GOMEMLIMIT"] == "" {
		env["GOMEMLIMIT"] = fmt.Sprintf("%dB", limits.MemBytes/10*9)
	}
	return env
}

func nodeEnv(limits Limits, environ map[string]string) map[string]string {
	env := slotEnv("UV_THREADPOOL_SIZE", limits)
	existing := environ["NODE_OPTIONS"]
	if limits.MemBytes == 0 || strings.Contains(existing, "--max-old-space-size") {
		return env
	}
	heap := fmt.Sprintf("--max-old-space-size=%d", limits.MemBytes/4*3/(1<<20))
	env["NODE_OPTIONS"] = strings.TrimSpace(existing + " " + heap)
	return env
}

func slotEnv(key string, limits Limits) map[string]string {
	if limits.Slots == 0 {
		return map[string]string{}
	}
	return map[string]string{key: fmt.Sprint(limits.Slots)}
}

func hasFlag(args []string, flag string) bool {
	return slices.ContainsFunc(args, func(arg string) bool {
		return arg == flag || strings.HasPrefix(arg, flag+"=")
	})
}
