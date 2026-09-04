package todos

import (
	"strings"
	"testing"
)

func TestAbsolutizeAttachmentURLs(t *testing.T) {
	const origin = "http://gavel.example:9092"
	body := "Look here:\n\n![screen.png](" + AttachmentURLPrefix + "abc.png)\n- [log.txt](" + AttachmentURLPrefix + "def.txt)"

	got := AbsolutizeAttachmentURLs(body, origin)

	wantImg := "![screen.png](" + origin + AttachmentURLPrefix + "abc.png)"
	wantLink := "- [log.txt](" + origin + AttachmentURLPrefix + "def.txt)"
	if !strings.Contains(got, wantImg) {
		t.Errorf("image link not absolutized: %q", got)
	}
	if !strings.Contains(got, wantLink) {
		t.Errorf("file link not absolutized: %q", got)
	}

	// Idempotent: a second pass leaves the already-absolute links untouched.
	if again := AbsolutizeAttachmentURLs(got, origin); again != got {
		t.Errorf("second pass changed body:\n got %q\nwant %q", again, got)
	}

	// No origin and bodies without attachments are returned unchanged.
	if AbsolutizeAttachmentURLs(body, "") != body {
		t.Error("empty origin should not modify body")
	}
	if AbsolutizeAttachmentURLs("plain body", origin) != "plain body" {
		t.Error("body without attachments should be unchanged")
	}
}

func TestHasAttachmentURLs(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{"image target", "![s.png](" + AttachmentURLPrefix + "abc.png)", true},
		{"link target", "- [log.txt](" + AttachmentURLPrefix + "def.txt)", true},
		{"already absolute", "![s.png](http://host" + AttachmentURLPrefix + "abc.png)", false},
		{"prose mentioning the path", "attachments live under " + AttachmentURLPrefix, false},
		{"no attachments", "plain body", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasAttachmentURLs(tt.body); got != tt.want {
				t.Errorf("HasAttachmentURLs(%q) = %v, want %v", tt.body, got, tt.want)
			}
		})
	}
}
