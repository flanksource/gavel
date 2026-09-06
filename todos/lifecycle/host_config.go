package lifecycle

import (
	"errors"

	"github.com/flanksource/captain/pkg/api"
	"github.com/flanksource/captain/pkg/captainconfig"
)

func (h *Host) savedDefaults() *captainconfig.AIDefaults {
	if h.Saved == nil {
		return nil
	}
	return &h.Saved.AI
}

func runtimeConfigurationError(err error) error {
	var saved *api.SavedDefaultsError
	if errors.As(err, &saved) {
		return &ConfigurationError{Err: err}
	}
	return err
}
