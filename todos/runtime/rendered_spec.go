package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/captain/pkg/api"
)

// renderedSpec stores only the resolved spec dispatched to Captain.
func renderedSpec(spec api.Spec) (map[string]any, error) {
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("marshal rendered spec: %w", err)
	}
	var rendered map[string]any
	if err := json.Unmarshal(data, &rendered); err != nil {
		return nil, fmt.Errorf("decode rendered spec: %w", err)
	}
	return rendered, nil
}
