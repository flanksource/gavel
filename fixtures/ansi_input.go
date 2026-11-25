package fixtures

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

// PTYInput is one exchange of a scripted terminal session: wait until Expect
// appears in the output (escape sequences stripped, after the previous item's
// match), then type Send followed by Keys.
type PTYInput struct {
	Expect string   `yaml:"expect" json:"expect"`
	Send   string   `yaml:"send,omitempty" json:"send,omitempty"`
	Keys   []string `yaml:"keys,omitempty" json:"keys,omitempty"`
}

// ptyKeys are the named keys a script can press, as the bytes a terminal
// sends for them.
var ptyKeys = map[string]string{
	"enter":     "\r",
	"tab":       "\t",
	"space":     " ",
	"esc":       "\x1b",
	"backspace": "\x7f",
	"up":        "\x1b[A",
	"down":      "\x1b[B",
	"right":     "\x1b[C",
	"left":      "\x1b[D",
	"ctrl-c":    "\x03",
	"ctrl-d":    "\x04",
}

// terminalQueryReplies answer the queries TUI libraries send to learn about
// the terminal: a real terminal answers them, and without an answer the
// program waits on its own timeout before every prompt.
var terminalQueryReplies = map[string]string{
	"\x1b[6n":         "\x1b[1;1R",
	"\x1b]11;?\x1b\\": "\x1b]11;rgb:0000/0000/0000\x1b\\",
	"\x1b]11;?\a":     "\x1b]11;rgb:0000/0000/0000\a",
}

// ValidatePTYInput reports the first item that cannot be typed: one without
// an expectation, or one pressing a key that has no name.
func ValidatePTYInput(items []PTYInput) error {
	for i, item := range items {
		if item.Expect == "" {
			return fmt.Errorf("input %d: expect is required", i+1)
		}
		if _, err := item.text(); err != nil {
			return fmt.Errorf("input %d: %w", i+1, err)
		}
	}
	return nil
}

func (in PTYInput) text() (string, error) {
	var b strings.Builder
	b.WriteString(in.Send)
	for _, key := range in.Keys {
		seq, ok := ptyKeys[key]
		if !ok {
			names := make([]string, 0, len(ptyKeys))
			for name := range ptyKeys {
				names = append(names, name)
			}
			sort.Strings(names)
			return "", fmt.Errorf("unknown key %q (known: %s)", key, strings.Join(names, ", "))
		}
		b.WriteString(seq)
	}
	return b.String(), nil
}

// ptyScript types a PTYInput script into a running command. The capture's
// read loop feeds it every chunk; Run waits for each expectation in turn.
type ptyScript struct {
	items   []PTYInput
	term    io.Writer
	writeMu sync.Mutex

	mu      sync.Mutex
	cond    *sync.Cond
	raw     strings.Builder
	pending string
	closed  bool
}

func newPTYScript(items []PTYInput, term io.Writer) *ptyScript {
	s := &ptyScript{items: items, term: term}
	s.cond = sync.NewCond(&s.mu)
	return s
}

// feed records a chunk of output and answers any terminal query it completes.
func (s *ptyScript) feed(chunk []byte) error {
	s.mu.Lock()
	s.raw.Write(chunk)
	s.pending += string(chunk)
	// Replies go in the order the queries were asked: termenv reads its OSC
	// reply before the cursor position, and gives up on the first reply
	// that is not the one it expects.
	var replies []string
	for {
		at, query := -1, ""
		for q := range terminalQueryReplies {
			if i := strings.Index(s.pending, q); i >= 0 && (at < 0 || i < at) {
				at, query = i, q
			}
		}
		if at < 0 {
			break
		}
		replies = append(replies, terminalQueryReplies[query])
		s.pending = s.pending[:at] + s.pending[at+len(query):]
	}
	if len(s.pending) > 64 {
		s.pending = s.pending[len(s.pending)-64:]
	}
	s.cond.Broadcast()
	s.mu.Unlock()
	for _, reply := range replies {
		if err := s.write(reply); err != nil {
			return fmt.Errorf("answer terminal query: %w", err)
		}
	}
	return nil
}

// close wakes Run for good: the command's output has ended.
func (s *ptyScript) close() {
	s.mu.Lock()
	s.closed = true
	s.cond.Broadcast()
	s.mu.Unlock()
}

func (s *ptyScript) write(text string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, err := io.WriteString(s.term, text)
	return err
}

// Run types the script, returning the first expectation the output ended
// without, with the tail of what the command did print.
func (s *ptyScript) Run() error {
	offset := 0
	for i, item := range s.items {
		text, err := item.text()
		if err != nil {
			return fmt.Errorf("pty input %d: %w", i+1, err)
		}
		s.mu.Lock()
		for {
			seen := stripFixtureANSI(s.raw.String())
			if at := strings.Index(seen[min(offset, len(seen)):], item.Expect); at >= 0 {
				offset += at + len(item.Expect)
				break
			}
			if s.closed {
				s.mu.Unlock()
				return fmt.Errorf("pty input %d: expected %q before exit; output ended with:\n%s", i+1, item.Expect, outputTail(seen, 2000))
			}
			s.cond.Wait()
		}
		s.mu.Unlock()
		if err := s.write(text); err != nil {
			return fmt.Errorf("pty input %d: type %q: %w", i+1, item.Expect, err)
		}
	}
	return nil
}

func outputTail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "…" + s[len(s)-n:]
}
