package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveAIFix(t *testing.T) {
	cases := []struct {
		name  string
		aiFix bool
		yes   bool
		want  bool
	}{
		{"neither", false, false, false},
		{"explicit ai-fix", true, false, true},
		{"yes implies ai-fix", false, true, true},
		{"both", true, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveAIFix(LintOptions{AIFix: tc.aiFix, Yes: tc.yes})
			assert.Equal(t, tc.want, got)
		})
	}
}
