package testrunner

import "testing"

func TestShellQuote(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"plain", "plain"},
		{"with-dash", "with-dash"},
		{"./pkg/a", "./pkg/a"},
		{"", "''"},
		{"has space", "'has space'"},
		{"has'quote", `'has'\''quote'`},
		{"$VAR", "'$VAR'"},
		{"a|b", "'a|b'"},
		{"glob*.md", "'glob*.md'"},
	}
	for _, c := range cases {
		got := shellQuote(c.in)
		if got != c.want {
			t.Errorf("shellQuote(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestShellJoin(t *testing.T) {
	got := shellJoin([]string{"go", "test", "-json", "./pkg with space"})
	want := "go test -json './pkg with space'"
	if got != want {
		t.Errorf("shellJoin = %q, want %q", got, want)
	}
}
