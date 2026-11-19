// Tests the status a todo request's error answers with: a run its organization's budgets refuse
// says so with 429, as a project action's does.
package ui

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/flanksource/gavel/sandbox"
	"github.com/stretchr/testify/assert"
)

func TestATodoRunItsBudgetsRefuseAnswersTooManyRequests(t *testing.T) {
	refused := fmt.Errorf("step plan: run 7 was not admitted: start sandbox run 8: %w: this organization has used 2 tokens of its 1 tokens budget", sandbox.ErrBudgetUsedUp)
	assert.Equal(t, http.StatusTooManyRequests, todoErrorStatus(http.StatusBadRequest, refused))
}
