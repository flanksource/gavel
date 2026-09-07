package streamtee

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// safeBuffer guards the sink: a Writer is fed from os/exec's copy goroutine, so
// an unguarded buffer races the assertion.
type safeBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (s *safeBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func TestWriterEmitsWholeLinesOnly(t *testing.T) {
	var sink safeBuffer
	w := New(Options{Out: &sink, Prefix: "| "})

	_, err := w.Write([]byte("abc"))
	require.NoError(t, err)
	assert.Empty(t, sink.String(), "a partial line must not be emitted")

	_, err = w.Write([]byte("def\n"))
	require.NoError(t, err)
	assert.Equal(t, "\n| abcdef\n", sink.String())
}

func TestWriterNeverReturnsAnErrorToItsCaller(t *testing.T) {
	failing := writerFunc(func([]byte) (int, error) { return 0, errors.New("sink is gone") })
	w := New(Options{Out: failing})

	n, err := w.Write([]byte("one\ntwo\n"))

	require.NoError(t, err, "io.MultiWriter would stop feeding the tail buffer on an error")
	assert.Equal(t, len("one\ntwo\n"), n)
	assert.Error(t, w.Close(), "the error is reported once, at Close")
}

func TestWriterIsLineAtomicUnderConcurrentWrites(t *testing.T) {
	var sink safeBuffer
	w := New(Options{Out: &sink, Prefix: "> "})

	const writers = 20
	var wg sync.WaitGroup
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _ = w.Write([]byte(fmt.Sprintf("line-%02d", i)))
			_, _ = w.Write([]byte("-mid"))
			_, _ = w.Write([]byte("-end\n"))
		}(i)
	}
	wg.Wait()
	require.NoError(t, w.Close())

	for _, line := range strings.Split(strings.TrimSpace(sink.String()), "\n") {
		if line == "" {
			continue
		}
		assert.True(t, strings.HasPrefix(line, "> "), "every line carries the gutter: %q", line)
	}
	// Chunks from different goroutines may interleave within a line, but no
	// line may ever be split across two output lines.
	assert.Equal(t, writers, strings.Count(sink.String(), "-end\n"))
}

func TestWriterBreaksOutOfAnInPlaceLine(t *testing.T) {
	var sink safeBuffer
	w := New(Options{Out: &sink})

	_, _ = w.Write([]byte("first\n"))
	_, _ = w.Write([]byte("second\n"))

	assert.Equal(t, "\nfirst\nsecond\n", sink.String(), "the cursor is committed once, not before every line")
}

func TestWriterCapsALongLine(t *testing.T) {
	var sink safeBuffer
	w := New(Options{Out: &sink, MaxLineBytes: 16})

	_, _ = w.Write([]byte(strings.Repeat("x", 100) + "\n"))

	assert.Equal(t, "\n"+strings.Repeat("x", 16)+truncationMarker+"\n", sink.String())
}

func TestWriterCloseFlushesATrailingPartialLine(t *testing.T) {
	var sink safeBuffer
	w := New(Options{Out: &sink, Prefix: "| "})

	_, _ = w.Write([]byte("no newline"))
	require.NoError(t, w.Close())

	assert.Equal(t, "\n| no newline\n", sink.String(), "a child that dies mid-line must not lose it")
}

type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }
