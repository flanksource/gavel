package launch

import (
	"errors"
	"fmt"
)

var ErrPreparationExpired = errors.New("workspace preparation expired; prepare the workspace again")

type InputError struct{ Err error }

func (e *InputError) Error() string { return fmt.Sprintf("invalid launch input: %v", e.Err) }
func (e *InputError) Unwrap() error { return e.Err }
