// Package streamtee turns a child process's byte stream into whole prefixed
// lines on a terminal another renderer is also drawing on.
//
// It is the shape procfile's per-process prefix writer uses — accumulate, split
// on newlines, hold the mutex across the whole emit — with three differences
// that matter to a tee:
//
//   - it never returns an error to its caller, because it is wrapped in an
//     io.MultiWriter alongside the buffer that feeds an agent's next iteration,
//     and io.MultiWriter stops at the first writer that fails;
//   - it flushes a trailing partial line on Close, so a child that dies without
//     a final newline does not lose its last line;
//   - it breaks out of an in-place line before its first output, so a renderer
//     that redraws with \r (captain's EventRenderer) cannot be interleaved with.
package streamtee

import (
	"bytes"
	"io"
	"sync"
)

// Options configures a Writer. Out is required.
type Options struct {
	Out          io.Writer // where whole lines are written
	Prefix       string    // gutter written before every line
	MaxLineBytes int       // 0 ⇒ unbounded
}

// Writer is an io.Writer that emits only whole, prefixed lines.
//
// Writes are synchronous: the caller is os/exec's copy goroutine, and letting it
// block is deliberate. A stalled terminal fills the child's pipe and blocks the
// child, at which point the verifier's own bounds — its timeout, its process
// group kill, and its WaitDelay — apply. A background goroutine with a drop
// policy would trade that bounded failure for an unbounded one, and the
// authoritative copy of the output is already in the feedback tail.
type Writer struct {
	mu     sync.Mutex
	out    io.Writer
	prefix string
	max    int
	buf    []byte
	opened bool
	err    error
}

// truncationMarker replaces the tail of a line longer than MaxLineBytes. The
// untruncated bytes still reach the feedback tail through the other arm of the
// MultiWriter; this bound only stops one pathological line from wrapping across
// hundreds of terminal rows.
const truncationMarker = " …"

func New(o Options) *Writer {
	return &Writer{out: o.Out, prefix: o.Prefix, max: o.MaxLineBytes}
}

// Write buffers p and emits every complete line it now holds. It always reports
// a full write with a nil error; the first sink error is kept for Close.
func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		w.emit(w.buf[:i])
		w.buf = w.buf[i+1:]
	}
	return len(p), nil
}

// Close flushes a trailing partial line and reports the first sink error seen.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.buf) > 0 {
		w.emit(w.buf)
		w.buf = nil
	}
	return w.err
}

// emit writes one line, breaking out of any in-place line the surrounding
// renderer left the cursor on before the very first one.
//
// The break is a bare newline, never \r or an erase sequence: a renderer that
// redraws in place has text on that line it does not redraw, and erasing would
// delete it rather than commit it.
func (w *Writer) emit(line []byte) {
	if w.out == nil {
		return
	}
	if !w.opened {
		w.opened = true
		w.write("\n")
	}
	w.write(w.prefix)
	w.write(string(w.truncate(line)))
	w.write("\n")
}

func (w *Writer) truncate(line []byte) []byte {
	if w.max <= 0 || len(line) <= w.max {
		return line
	}
	return append(append([]byte{}, line[:w.max]...), truncationMarker...)
}

func (w *Writer) write(s string) {
	if s == "" {
		return
	}
	if _, err := io.WriteString(w.out, s); err != nil && w.err == nil {
		w.err = err
	}
}
