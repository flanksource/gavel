package lint

import "fmt"

func Validate(opts Options) error {
	if opts.WorkDir == "" {
		return fmt.Errorf("lint workspace is required")
	}
	registry, err := loadLinterRegistry(opts.WorkDir, opts.ConfigDir)
	if err != nil {
		return err
	}
	names, _, err := resolveRequestedLinters(registry, opts.Linters)
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
