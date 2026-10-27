package lint

import "fmt"

func Validate(opts Options) error {
	if opts.WorkDir == "" {
		return fmt.Errorf("lint workspace is required")
	}
	names, _, err := resolveRequestedLinters(buildLinterRegistry(opts.WorkDir), opts.Linters)
	if err != nil {
		return err
	}
	if err := checkPassThrough(names, opts.PassThroughArgs); err != nil {
		return err
	}
	if _, err := ParseRuleExclusions(opts.Exclude); err != nil {
		return err
	}
	_, err = parseTimeout(opts.Timeout)
	return err
}
