package verify

import (
	"fmt"

	"github.com/flanksource/gavel/ai"
)

// parseVerifyResponse extracts a VerifyResult from an agent's raw reply via the
// generic structured parser; a decoded result without checks is rejected.
func parseVerifyResponse(raw string) (VerifyResult, error) {
	result, err := ai.ParseStructured(raw, validateVerifyResult)
	if err != nil {
		return VerifyResult{}, err
	}
	return *result, nil
}

// validateVerifyResult is the required-field gate for parsed verify replies —
// tolerant YAML decoding accepts nearly anything, so a reply only counts as a
// verify result when it actually scored at least one check.
func validateVerifyResult(r *VerifyResult) error {
	if len(r.Checks) == 0 {
		return fmt.Errorf("verify result has no checks")
	}
	return nil
}
