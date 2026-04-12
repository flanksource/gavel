package main

import (
	"testing"

	"github.com/flanksource/gavel/linters"
	"github.com/flanksource/gavel/models"
	"github.com/stretchr/testify/assert"
)

func TestCollectViolationTypes(t *testing.T) {
	msg1 := "error return value not checked"
	msg2 := "unused variable"

	results := []*linters.LinterResult{
		{
			Violations: []models.Violation{
				{Source: "golangci-lint", Rule: &models.Rule{Method: "errcheck"}, File: "a.go", Message: &msg1},
				{Source: "golangci-lint", Rule: &models.Rule{Method: "errcheck"}, File: "b.go", Message: &msg1},
				{Source: "golangci-lint", Rule: &models.Rule{Method: "unused"}, File: "a.go", Message: &msg2},
			},
		},
		{
			Violations: []models.Violation{
				{Source: "eslint", Rule: &models.Rule{Method: "no-unused-vars"}, File: "x.ts"},
			},
		},
		nil,
	}

	types := collectViolationTypes(results)

	assert.Len(t, types, 3)

	// Sorted by count descending
	assert.Equal(t, "errcheck", types[0].Rule)
	assert.Equal(t, "golangci-lint", types[0].Source)
	assert.Equal(t, 2, types[0].Count)
	assert.ElementsMatch(t, []string{"a.go", "b.go"}, types[0].Files)
	assert.Equal(t, msg1, types[0].Example)

	assert.Equal(t, 1, types[1].Count)
	assert.Equal(t, 1, types[2].Count)
}

func TestCollectViolationTypes_Empty(t *testing.T) {
	types := collectViolationTypes(nil)
	assert.Empty(t, types)
}

func TestCollectViolationTypes_NilRule(t *testing.T) {
	msg := "some error"
	results := []*linters.LinterResult{
		{
			Violations: []models.Violation{
				{Source: "custom", File: "f.go", Message: &msg},
				{Source: "custom", File: "g.go"},
			},
		},
	}

	types := collectViolationTypes(results)
	assert.Len(t, types, 1)
	assert.Equal(t, "", types[0].Rule)
	assert.Equal(t, "custom", types[0].Source)
	assert.Equal(t, 2, types[0].Count)
}
