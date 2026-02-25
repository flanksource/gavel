package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePathRef(t *testing.T) {
	tests := []struct {
		input    string
		expected PathRef
	}{
		{"pkg/auth.go", PathRef{File: "pkg/auth.go", Line: 0, EndLine: 0}},
		{"pkg/auth.go:42", PathRef{File: "pkg/auth.go", Line: 42, EndLine: 0}},
		{"pkg/auth.go:10-50", PathRef{File: "pkg/auth.go", Line: 10, EndLine: 50}},
		{"main.go:1", PathRef{File: "main.go", Line: 1, EndLine: 0}},
		{"pkg/auth.go:invalid", PathRef{File: "pkg/auth.go", Line: 0, EndLine: 0}},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, ParsePathRef(tc.input))
		})
	}
}

func TestPathRef_String(t *testing.T) {
	tests := []struct {
		ref      PathRef
		expected string
	}{
		{PathRef{File: "pkg/auth.go"}, "pkg/auth.go"},
		{PathRef{File: "pkg/auth.go", Line: 42}, "pkg/auth.go:42"},
		{PathRef{File: "pkg/auth.go", Line: 10, EndLine: 50}, "pkg/auth.go:10-50"},
	}
	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.ref.String())
		})
	}
}

func TestParsePathRef_RoundTrip(t *testing.T) {
	inputs := []string{"pkg/auth.go", "pkg/auth.go:42", "pkg/auth.go:10-50"}
	for _, input := range inputs {
		assert.Equal(t, input, ParsePathRef(input).String())
	}
}

func TestPathRef_IsWholeFile(t *testing.T) {
	assert.True(t, PathRef{File: "main.go"}.IsWholeFile())
	assert.False(t, PathRef{File: "main.go", Line: 1}.IsWholeFile())
}

func TestTODOFrontmatter_PathRefs(t *testing.T) {
	fm := TODOFrontmatter{
		Path: StringOrSlice{"pkg/auth.go:42", "pkg/db.go", "cmd/main.go:10-20"},
	}
	refs := fm.PathRefs()
	assert.Equal(t, []PathRef{
		{File: "pkg/auth.go", Line: 42},
		{File: "pkg/db.go"},
		{File: "cmd/main.go", Line: 10, EndLine: 20},
	}, refs)
}
