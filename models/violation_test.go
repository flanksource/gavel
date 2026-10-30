package models

import (
	"encoding/json"
	"testing"
)

func TestViolationPrettyCountsLocationsOnlyWhenThereAreSeveral(t *testing.T) {
	single := Violation{File: "a.xml", Line: 4, Locations: []ViolationLocation{{File: "a.xml", Line: 4}}}
	if got := single.Pretty().String(); got != "a.xml:4" {
		t.Errorf("single-location Pretty = %q, want a.xml:4", got)
	}
	several := Violation{File: "a.xml", Line: 4, Locations: []ViolationLocation{{File: "a.xml", Line: 4}, {File: "b.xml", Line: 9}, {File: "b.xml"}}}
	if got := several.Pretty().String(); got != "a.xml:4 (3 locations)" {
		t.Errorf("multi-location Pretty = %q, want a.xml:4 (3 locations)", got)
	}
}

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
