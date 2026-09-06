package lifecycle

import "fmt"

// ConfigurationError identifies invalid project or catalog configuration.
type ConfigurationError struct {
	Err error
}

func (e *ConfigurationError) Error() string { return fmt.Sprintf("lifecycle configuration: %v", e.Err) }

func (e *ConfigurationError) Unwrap() error { return e.Err }
