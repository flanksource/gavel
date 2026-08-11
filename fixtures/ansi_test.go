package fixtures

import "testing"

func TestHasAltScreen(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"enter alt screen", "before\x1b[?1049hafter", true},
		{"leave alt screen", "report\x1b[?1049l\x1b[?25h", true},
		{"plain text", "no escapes here", false},
		{"colour only", "\x1b[38;2;1;2;3mgreen\x1b[0m", false},
		{"cursor show is not alt screen", "\x1b[?25h", false},
		{"cursor hide is not alt screen", "\x1b[?25l", false},
		// The digits matter: ?1049 is the alt screen, ?104 is not a prefix of it.
		{"unrelated private mode", "\x1b[?104h", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hasAltScreen(c.in); got != c.want {
				t.Errorf("hasAltScreen(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}
