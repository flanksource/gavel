// Azure DevOps build and timeline states as the dashboard spells a check's status and conclusion,
// for the CI commons-db reads with each provider's own states.
package azuredevops

import (
	"fmt"
	"strings"
)

// ExecutionState is an Azure DevOps build or timeline record's state and result as the dashboard
// spells a check's status and conclusion: QUEUED, IN_PROGRESS or COMPLETED, and SUCCESS, FAILURE...
func ExecutionState(state, result string) (string, string, error) {
	switch state {
	case "pending", "notStarted", "postponed":
		return "QUEUED", "", nil
	case "inProgress", "cancelling":
		return "IN_PROGRESS", "", nil
	case "completed":
		conclusions := map[string]string{"succeeded": "SUCCESS", "failed": "FAILURE", "canceled": "CANCELLED", "skipped": "SKIPPED", "abandoned": "CANCELLED", "partiallySucceeded": "PARTIAL_SUCCESS", "succeededWithIssues": "PARTIAL_SUCCESS"}
		if conclusion, ok := conclusions[result]; ok {
			return "COMPLETED", conclusion, nil
		}
	}
	return "", "", fmt.Errorf("unexpected Azure execution state %q/result %q", strings.TrimSpace(state), strings.TrimSpace(result))
}
