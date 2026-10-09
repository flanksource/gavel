package azuredevops

import (
	"encoding/json"
	"fmt"
)

func (p *policyEvaluation) UnmarshalJSON(data []byte) error {
	type wirePolicy policyEvaluation
	var wire wirePolicy
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	var required struct {
		Configuration *struct {
			Enabled  *bool `json:"isEnabled"`
			Blocking *bool `json:"isBlocking"`
		} `json:"configuration"`
	}
	if err := json.Unmarshal(data, &required); err != nil {
		return err
	}
	if required.Configuration == nil || required.Configuration.Enabled == nil || required.Configuration.Blocking == nil || wire.Configuration.ID <= 0 || wire.Configuration.Type.ID == "" {
		return fmt.Errorf("azure policy configuration is missing identity, type, isEnabled or isBlocking")
	}
	if wire.Status == "" {
		return fmt.Errorf("azure policy evaluation is missing status")
	}
	*p = policyEvaluation(wire)
	return nil
}
