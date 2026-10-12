package record

import (
	"io"
	"sync"
	"unicode/utf8"
)

// streamCapture tees an event stream's bytes as the child reads them. Such a
// stream never ends while its page is open, so unlike a finite body it cannot be
// read ahead and replayed: the entry is materialized from whatever has passed
// at the moment the recording is read, and every byte counts toward Size even
// past the `bodies:` cap that bounds the kept text.
type streamCapture struct {
	mu     sync.Mutex
	limit  int
	kept   []byte
	total  int64
	closed bool
}

func (c *streamCapture) record(chunk []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.total += int64(len(chunk))
	if room := c.limit - len(c.kept); room > 0 {
		c.kept = append(c.kept, chunk[:min(room, len(chunk))]...)
	}
}

func (c *streamCapture) finish() {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
}

func (c *streamCapture) apply(content *Content) {
	c.mu.Lock()
	defer c.mu.Unlock()
	content.Size = c.total
	kept := c.kept
	if c.total > int64(len(kept)) || !c.closed {
		kept = trimPartialRune(kept)
	}
	content.Text, content.Encoding = encodeBody(kept)
	switch {
	case c.total > int64(len(c.kept)):
		content.Comment = "truncated by record.http.bodies"
	case !c.closed:
		content.Comment = "event stream still open when recorded"
	default:
		content.Comment = ""
	}
}

// trimPartialRune drops a UTF-8 rune split by the cap or a chunk boundary, so a
// cut-off prefix still encodes as text instead of falling back to base64.
func trimPartialRune(b []byte) []byte {
	for i := len(b) - 1; i >= 0 && i >= len(b)-utf8.UTFMax; i-- {
		if utf8.RuneStart(b[i]) {
			if !utf8.FullRune(b[i:]) {
				return b[:i]
			}
			break
		}
	}
	return b
}

type teeStream struct {
	io.ReadCloser
	capture *streamCapture
}

func (t teeStream) Read(p []byte) (int, error) {
	n, err := t.ReadCloser.Read(p)
	t.capture.record(p[:n])
	if err != nil {
		t.capture.finish()
	}
	return n, err
}

func (t teeStream) Close() error {
	t.capture.finish()
	return t.ReadCloser.Close()
}

// materializeStreams fills each streamed entry's content from its capture, on
// copies the caller owns.
func materializeStreams(entries []Entry) {
	for i := range entries {
		if entries[i].stream != nil {
			entries[i].stream.apply(&entries[i].Response.Content)
		}
	}
}
