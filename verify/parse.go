package verify

import "fmt"

// validateVerifyResult is the required-field gate for parsed verify replies —
// tolerant YAML decoding accepts nearly anything, so a reply only counts as a
// verify result when it actually scored at least one check.
func validateVerifyResult(r *VerifyResult) error {
	if len(r.Checks) == 0 {
		return fmt.Errorf("verify result has no checks")
	}
	return nil
}
