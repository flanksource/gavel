package runners

import (
	"fmt"
	"strconv"
	"strings"
)

// GinkgoGoTestArgs maps the Ginkgo options accepted by gavel onto flags
// registered by Ginkgo in the go test binary.
func GinkgoGoTestArgs(args []string) ([]string, error) {
	var mapped []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		for _, flag := range []string{"--focus", "--label-filter"} {
			if arg == flag {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("%s requires a value", flag)
				}
				i++
				arg = flag + "=" + args[i]
			}
			if value, ok := strings.CutPrefix(arg, flag+"="); ok {
				arg = "-ginkgo." + strings.TrimPrefix(flag, "--") + "=" + value
				break
			}
		}
		if arg == "--nodes" {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--nodes requires a value")
			}
			i++
			arg = "--nodes=" + args[i]
		}
		if value, ok := strings.CutPrefix(arg, "--nodes="); ok {
			nodes, err := strconv.Atoi(value)
			if err != nil || nodes != 1 {
				return nil, fmt.Errorf("Ginkgo --nodes=%s is unsupported by the unified go test runner", value)
			}
			continue
		}
		if value, ok := strings.CutPrefix(arg, "--tags="); ok {
			arg = "-tags=" + value
		}
		mapped = append(mapped, arg)
	}
	return mapped, nil
}
