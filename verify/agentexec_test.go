package verify

import (
	"testing"
)

func TestParseVerifyResponse(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"bare json", `{"checks":{"a":{"pass":true}},"ratings":{},"completeness":{"pass":true}}`},
		{"fenced json", "```json\n{\"checks\":{\"a\":{\"pass\":true}},\"ratings\":{},\"completeness\":{\"pass\":true}}\n```"},
		{"json in prose", "Here is my review:\n{\"checks\":{\"a\":{\"pass\":false}},\"ratings\":{},\"completeness\":{\"pass\":false}}\nDone."},
		{"json envelope", `{"result":"{\"checks\":{\"a\":{\"pass\":true}},\"ratings\":{},\"completeness\":{\"pass\":true}}"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseVerifyResponse(tc.raw)
			if err != nil {
				t.Fatalf("parseVerifyResponse: %v", err)
			}
			if len(got.Checks) != 1 {
				t.Errorf("Checks = %d, want 1", len(got.Checks))
			}
		})
	}
}

func TestParseVerifyResponseError(t *testing.T) {
	if _, err := parseVerifyResponse("not json, just an apology"); err == nil {
		t.Fatal("expected an error for a non-JSON reply")
	}
}
