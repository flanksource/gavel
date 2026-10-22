package fixtures

import (
	"encoding/json"
	"fmt"
)

func (result *FixtureResult) UnmarshalJSON(data []byte) error {
	type resultJSON FixtureResult
	var decoded resultJSON
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if stored, ok := decoded.Metadata["attempts"]; ok {
		raw, err := json.Marshal(stored)
		if err != nil {
			return fmt.Errorf("encode stored attempts: %w", err)
		}
		var attempts []AttemptReport
		if err := json.Unmarshal(raw, &attempts); err != nil {
			return fmt.Errorf("decode stored attempts: %w", err)
		}
		decoded.Metadata["attempts"] = attempts
	}
	*result = FixtureResult(decoded)
	return nil
}
