package record

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// halfCloser is goproxy's `halfClosable`. goproxy type-asserts the dialed
// connection against it and takes a materially better teardown path when it
// matches — its own comment documents the race in the fallback — so the
// counting wrapper must preserve the capability rather than hide it.
type halfCloser interface {
	net.Conn
	CloseWrite() error
	CloseRead() error
}

// countingConn wraps the proxy-to-target side of a CONNECT tunnel and counts
// the bytes crossing it. Reads carry the server's response (bytes in) and
// writes carry the client's request (bytes out).
type countingConn struct {
	net.Conn
	recorder *HTTPRecorder
	host     string
	started  time.Time
	in       atomic.Int64
	out      atomic.Int64
	once     sync.Once
	// self is the value the recorder tracks, which is the outer half-closable
	// wrapper when there is one. Untracking `c` instead would leave the real
	// key in the map and defeat the force-close on shutdown.
	self net.Conn
}

// newCountingConn returns a wrapper that keeps whatever capabilities the inner
// connection had. Claiming CloseWrite/CloseRead unconditionally would push
// goproxy down its fast path with a connection that cannot honour it.
func newCountingConn(inner net.Conn, recorder *HTTPRecorder, host string, started time.Time) net.Conn {
	counting := &countingConn{Conn: inner, recorder: recorder, host: host, started: started}
	counting.self = counting
	if _, ok := inner.(halfCloser); ok {
		wrapped := &countingHalfCloser{countingConn: counting}
		counting.self = wrapped
		return wrapped
	}
	return counting
}

func (c *countingConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	c.in.Add(int64(n))
	return n, err
}

func (c *countingConn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	c.out.Add(int64(n))
	return n, err
}

func (c *countingConn) Close() error {
	c.finish()
	return c.Conn.Close()
}

// finish records the tunnel exactly once. goproxy closes both ends of a tunnel
// and can reach Close twice on the same connection.
func (c *countingConn) finish() {
	c.once.Do(func() {
		c.recorder.untrack(c.self)
		c.recorder.add(connectEntry(c.host, c.started, time.Since(c.started), c.in.Load(), c.out.Load(), nil))
	})
}

// countingHalfCloser forwards the half-close methods to the inner connection.
type countingHalfCloser struct {
	*countingConn
}

func (c *countingHalfCloser) CloseWrite() error { return c.inner().CloseWrite() }
func (c *countingHalfCloser) CloseRead() error  { return c.inner().CloseRead() }

func (c *countingHalfCloser) inner() halfCloser { return c.countingConn.Conn.(halfCloser) }
