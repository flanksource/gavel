package overlayplain

import "testing"

func TestPlain(t *testing.T) {
	if "plain" == "" {
		t.Fatal("unexpected empty name")
	}
}
