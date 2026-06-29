package verify

import (
	"testing"
)

func TestGeminiParseResponse(t *testing.T) {
	adapter := Gemini{}
	validJSON := `{"checks":{"tests-added":{"pass":true}},"ratings":{"security":{"score":90}},"completeness":{"pass":true,"summary":"ok"}}`

	tests := []struct {
		name       string
		input      string
		wantChecks int
		wantErr    bool
	}{
		{
			name:       "direct JSON",
			input:      validJSON,
			wantChecks: 1,
		},
		{
			name:       "markdown fences",
			input:      "```json\n" + validJSON + "\n```",
			wantChecks: 1,
		},
		{
			name:    "invalid input",
			input:   "not json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.ParseResponse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Checks) != tt.wantChecks {
				t.Errorf("got %d checks, want %d", len(result.Checks), tt.wantChecks)
			}
		})
	}
}
