package ai

import (
	"strings"
	"testing"
)

// reviewReply is a stand-in agent output type; requireVerdict mirrors the kind
// of required-field validation callers supply (e.g. verify's non-empty checks).
type reviewReply struct {
	Verdict string   `json:"verdict"`
	Notes   []string `json:"notes,omitempty"`
}

func requireVerdict(r *reviewReply) error {
	if r.Verdict == "" {
		return errEmptyVerdict
	}
	return nil
}

var errEmptyVerdict = &validationError{"verdict is empty"}

type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }

func TestParseStructured(t *testing.T) {
	body := `{"verdict":"pass","notes":["clean"]}`
	cases := []struct {
		name string
		raw  string
	}{
		{"bare json", body},
		{"fenced json", "```json\n" + body + "\n```"},
		{"plain fences", "```\n" + body + "\n```"},
		{"json in prose", "Here is my verdict:\n" + body + "\nDone."},
		{"json envelope result key", `{"result":"{\"verdict\":\"pass\",\"notes\":[\"clean\"]}"}`},
		{"json envelope text key", `{"text":"{\"verdict\":\"pass\",\"notes\":[\"clean\"]}"}`},
		{"bare yaml", "verdict: pass\nnotes:\n  - clean"},
		{"yaml document block", "---\nverdict: pass\nnotes:\n  - clean\n---"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseStructured(tc.raw, requireVerdict)
			if err != nil {
				t.Fatalf("ParseStructured: %v", err)
			}
			if got.Verdict != "pass" {
				t.Errorf("Verdict = %q, want %q", got.Verdict, "pass")
			}
		})
	}
}

func TestParseStructuredErrors(t *testing.T) {
	t.Run("junk is an error", func(t *testing.T) {
		if _, err := ParseStructured("not json, just an apology", requireVerdict); err == nil {
			t.Fatal("expected an error for a non-JSON reply")
		}
	})

	// The old verify guard (len(Checks) > 0) silently classified a decoded but
	// invalid payload as "not parsed"; the validate func must surface WHY.
	t.Run("decoded but invalid payload names the validation failure", func(t *testing.T) {
		_, err := ParseStructured(`{"notes":["decodes fine, no verdict"]}`, requireVerdict)
		if err == nil {
			t.Fatal("expected an error for a payload that fails validation")
		}
		if !strings.Contains(err.Error(), "verdict is empty") {
			t.Errorf("error %q does not name the validation failure", err)
		}
	})

	t.Run("error carries a preview of the raw reply", func(t *testing.T) {
		_, err := ParseStructured("garbage reply", requireVerdict)
		if err == nil || !strings.Contains(err.Error(), "garbage reply") {
			t.Errorf("error %v does not preview the raw reply", err)
		}
	})
}

func TestSchemaInstruction(t *testing.T) {
	schema := `{"type":"object"}`
	got := SchemaInstruction(schema)
	if !strings.Contains(got, schema) {
		t.Errorf("instruction does not embed the schema: %q", got)
	}
	if !strings.Contains(got, "ONLY a single JSON object") {
		t.Errorf("instruction lost the bare-JSON contract: %q", got)
	}
}

func TestStripMarkdownFences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"json fences", "```json\n{\"a\":1}\n```", `{"a":1}`},
		{"yaml fences", "```yaml\na: 1\n```", "a: 1"},
		{"plain fences", "```\nsome text\n```", "some text"},
		{"no fences", `{"a":1}`, `{"a":1}`},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripMarkdownFences(tt.input); got != tt.want {
				t.Errorf("stripMarkdownFences() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractYAMLBlock(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"with yaml block", "---\nchecks:\n  a: true\n---", "checks:\n  a: true"},
		{"no separators", "just text", ""},
		{"single separator", "---\nfoo", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractYAMLBlock(tt.input); got != tt.want {
				t.Errorf("extractYAMLBlock() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractJSONFromText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"embedded object", `text {"a":1} more`, `{"a":1}`},
		{"nested braces", `prefix {"a":{"b":2}} suffix`, `{"a":{"b":2}}`},
		{"no JSON", "just text", ""},
		{"unmatched brace", "prefix { no close", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractJSONFromText(tt.input); got != tt.want {
				t.Errorf("extractJSONFromText() = %q, want %q", got, tt.want)
			}
		})
	}
}
