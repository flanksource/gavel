package git

import "testing"

func TestParseStringArrayResult(t *testing.T) {
	tests := []struct {
		name      string
		schema    []string
		raw       string
		fieldName string
		want      []string
	}{
		{
			name:      "prefers populated schema output",
			schema:    []string{"Removed v1 API"},
			raw:       `["ignored"]`,
			fieldName: "functionalityRemoved",
			want:      []string{"Removed v1 API"},
		},
		{
			name:      "parses bare json array",
			raw:       `["Removed v1 API","Clients must migrate"]`,
			fieldName: "functionalityRemoved",
			want:      []string{"Removed v1 API", "Clients must migrate"},
		},
		{
			name:      "parses fenced json array",
			raw:       "```json\n[\"Removed v1 API\", \"Clients must migrate\"]\n```",
			fieldName: "functionalityRemoved",
			want:      []string{"Removed v1 API", "Clients must migrate"},
		},
		{
			name:      "parses wrapped functionalityRemoved object",
			raw:       `{"functionalityRemoved":["Removed v1 API","Clients must migrate"]}`,
			fieldName: "functionalityRemoved",
			want:      []string{"Removed v1 API", "Clients must migrate"},
		},
		{
			name:      "parses wrapped compatibilityIssues object",
			raw:       `{"compatibilityIssues":["Clients must migrate"]}`,
			fieldName: "compatibilityIssues",
			want:      []string{"Clients must migrate"},
		},
		{
			name:      "parses fenced wrapped json object",
			raw:       "```json\n{\"functionalityRemoved\":[\"Removed v1 API\", \"  \", \"\"]}\n```",
			fieldName: "functionalityRemoved",
			want:      []string{"Removed v1 API"},
		},
		{
			name:      "drops empty items from bare array",
			raw:       `["Removed v1 API","  ",""]`,
			fieldName: "functionalityRemoved",
			want:      []string{"Removed v1 API"},
		},
		{
			name:      "ignores unrelated object key",
			raw:       `{"compatibilityIssues":["Clients must migrate"]}`,
			fieldName: "functionalityRemoved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseStringArrayResult(tt.schema, tt.raw, tt.fieldName)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v want %v", got, tt.want)
				}
			}
		})
	}
}
