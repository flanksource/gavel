package utils

import (
	"testing"
)

func TestIsSemver(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"1.2.3", true},
		{"v1.2.3", true},
		{"1.0.0", true},
		{"2.1.0-alpha", true},
		{"1.2.3-beta.1", true},
		{"latest", false},
		{"stable", false},
		{"abc123def456", false},
		{"", false},
		{"1.2", true},
		{"1", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := IsSemver(tt.value)
			if result != tt.expected {
				t.Errorf("IsSemver(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsSHA256(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"abc123def456789012345678901234567890abcdef1234567890abcdef123456", true},
		{"ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890", true},
		{"abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", true},
		{"abc123", false},
		{"abc123def456789012345678901234567890abcdef1234567890abcdef12345", false},
		{"abc123def456789012345678901234567890abcdef1234567890abcdef1234567", false},
		{"ghij23def456789012345678901234567890abcdef1234567890abcdef123456", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value[:min(len(tt.value), 10)], func(t *testing.T) {
			result := IsSHA256(tt.value)
			if result != tt.expected {
				t.Errorf("IsSHA256(%q...) = %v, want %v", tt.value[:min(len(tt.value), 10)], result, tt.expected)
			}
		})
	}
}

func TestIsGitSHA(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"abc123def456789012345678901234567890abcd", true},
		{"ABCDEF1234567890ABCDEF1234567890ABCDEF12", true},
		{"abcdef1234567890abcdef1234567890abcdef12", true},
		{"abc123", false},
		{"abc123def456789012345678901234567890abc", false},
		{"abc123def456789012345678901234567890abcde", false},
		{"ghij23def456789012345678901234567890ab", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.value[:min(len(tt.value), 10)], func(t *testing.T) {
			result := IsGitSHA(tt.value)
			if result != tt.expected {
				t.Errorf("IsGitSHA(%q...) = %v, want %v", tt.value[:min(len(tt.value), 10)], result, tt.expected)
			}
		})
	}
}

func TestParseImageReference(t *testing.T) {
	tests := []struct {
		image    string
		expected ImageReference
	}{
		{
			image:    "nginx:1.21",
			expected: ImageReference{Name: "nginx", Tag: "1.21"},
		},
		{
			image:    "nginx@sha256:abc123def456789012345678901234567890abcdef1234567890abcdef123456",
			expected: ImageReference{Name: "nginx", Digest: "sha256:abc123def456789012345678901234567890abcdef1234567890abcdef123456"},
		},
		{
			image:    "nginx:1.21@sha256:abc123def456789012345678901234567890abcdef1234567890abcdef123456",
			expected: ImageReference{Name: "nginx", Tag: "1.21", Digest: "sha256:abc123def456789012345678901234567890abcdef1234567890abcdef123456"},
		},
		{
			image:    "registry.io/namespace/nginx:1.21",
			expected: ImageReference{Registry: "registry.io", Name: "namespace/nginx", Tag: "1.21"},
		},
		{
			image:    "registry.io/namespace/nginx@sha256:abc123def456789012345678901234567890abcdef1234567890abcdef123456",
			expected: ImageReference{Registry: "registry.io", Name: "namespace/nginx", Digest: "sha256:abc123def456789012345678901234567890abcdef1234567890abcdef123456"},
		},
		{
			image:    "localhost:5000/myapp:v1.0",
			expected: ImageReference{Registry: "localhost:5000", Name: "myapp", Tag: "v1.0"},
		},
		{
			image:    "nginx",
			expected: ImageReference{Name: "nginx"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.image, func(t *testing.T) {
			result := ParseImageReference(tt.image)
			if result != tt.expected {
				t.Errorf("ParseImageReference(%q) = %+v, want %+v", tt.image, result, tt.expected)
			}
		})
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		fieldPath string
		pattern   string
		expected  bool
	}{
		{"spec.image", "*.image", true},
		{"metadata.image", "*.image", true},
		{"spec.containers.0.image", "*.image", false},
		{"spec.containers.0.image", "**.image", true},
		{"spec.containers.0.image", "spec.containers.*.image", true},
		{"spec.containers.1.image", "spec.containers.*.image", true},
		{"spec.template.spec.containers.0.image", "**.containers.*.image", true},
		{"spec.version", "*.version", true},
		{"spec.tag", "*.tag", true},
		{"metadata.labels.version", "**.version", true},
		{"spec.image", "*.version", false},
		{"spec.containers.0.tag", "**.tag", true},
	}

	for _, tt := range tests {
		t.Run(tt.fieldPath+" vs "+tt.pattern, func(t *testing.T) {
			result := MatchesPattern(tt.fieldPath, tt.pattern)
			if result != tt.expected {
				t.Errorf("MatchesPattern(%q, %q) = %v, want %v", tt.fieldPath, tt.pattern, result, tt.expected)
			}
		})
	}
}
