package models

import (
	"encoding/json"
	"testing"
)

func TestViolationJSONUsesLowercaseSeverity(t *testing.T) {
	for _, severity := range []ViolationSeverity{SeverityError, SeverityWarning, SeverityInfo} {
		raw, err := json.Marshal(Violation{File: "main.go", Line: 3, Severity: severity})
		if err != nil {
			t.Fatal(err)
		}
		want := `{"file":"main.go","line":3,"severity":"` + string(severity) + `","created_at":"0001-01-01T00:00:00Z"}`
		if string(raw) != want {
			t.Errorf("Marshal(%s) = %s, want %s", severity, raw, want)
		}
	}
}
