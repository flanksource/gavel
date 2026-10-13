package lint

import (
	"fmt"
	"path/filepath"
	"time"
)

func (o *Options) captureConfigDir() error {
	if o.ConfigDir == "" {
		o.ConfigDir = o.WorkDir
	}
	if o.ConfigDir == "" {
		o.ConfigDir = "."
	}
	var err error
	o.ConfigDir, err = filepath.Abs(o.ConfigDir)
	return err
}

func parseTimeout(value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}
	timeout, err := time.ParseDuration(value)
	if err != nil || timeout <= 0 {
		return 0, fmt.Errorf("invalid lint timeout %q: expected a positive duration", value)
	}
	return timeout, nil
}
