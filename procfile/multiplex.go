package procfile

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

// ansiColors is the rotating palette used to colour per-process log prefixes in
// foreground (`gavel proc run`) mode, mirroring foreman's coloured output.
var ansiColors = []string{
	"\033[36m", // cyan
	"\033[33m", // yellow
	"\033[32m", // green
	"\033[35m", // magenta
	"\033[34m", // blue
	"\033[31m", // red
}

const ansiReset = "\033[0m"

// prefixWriter prepends "name | " (optionally coloured) to every complete line
// it forwards to w. Partial lines are buffered until their newline arrives so a
// prefix is never inserted mid-line. A shared mutex serialises writes from all
// processes to the same underlying writer so concurrent output stays
// line-atomic instead of interleaving character-by-character.
type prefixWriter struct {
	w      io.Writer
	prefix string
	mu     *sync.Mutex
	buf    []byte
}

// newPrefixWriter builds a prefix writer for process name. width pads the name
// so prefixes from different processes line up. colorIdx selects a palette
// entry; a negative index disables colour.
func newPrefixWriter(w io.Writer, mu *sync.Mutex, name string, width, colorIdx int) *prefixWriter {
	label := fmt.Sprintf("%-*s | ", width, name)
	if colorIdx >= 0 {
		label = ansiColors[colorIdx%len(ansiColors)] + label + ansiReset
	}
	return &prefixWriter{w: w, prefix: label, mu: mu}
}

func (p *prefixWriter) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.buf = append(p.buf, b...)
	for {
		i := bytes.IndexByte(p.buf, '\n')
		if i < 0 {
			break
		}
		line := p.buf[:i]
		p.buf = p.buf[i+1:]
		if _, err := fmt.Fprintf(p.w, "%s%s\n", p.prefix, line); err != nil {
			return len(b), err
		}
	}
	return len(b), nil
}
