package types

import (
	"fmt"
	"strings"
)

type PRRunEnvelope struct {
	ResultEnvelope
	AddressedThreadIDs []string `json:"addressedThreadIds,omitempty" jsonschema_description:"Native PR review thread IDs whose feedback this completed turn addressed; omit other threads"`
}

func (e *PRRunEnvelope) Validate() error {
	if err := e.ResultEnvelope.Validate(); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, id := range e.AddressedThreadIDs {
		if id == "" || id != strings.TrimSpace(id) || seen[id] {
			return fmt.Errorf("invalid or repeated addressed thread ID %q", id)
		}
		seen[id] = true
	}
	if e.EndStatus != EndCompleted && len(e.AddressedThreadIDs) > 0 {
		return fmt.Errorf("only a completed turn can address PR threads")
	}
	return nil
}
