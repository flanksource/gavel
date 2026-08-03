package record

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

func init() { Implemented[KindSQL] = true }

// Postgres startup codes that arrive in place of a StartupMessage. The proxy
// declines both: it forwards a plain stream to the upstream, so a client that
// negotiated TLS with us would be talking into a tunnel we cannot re-establish.
//
// CancelRequest (80877102) is deliberately absent: it arrives on its own
// connection carrying no SQL and is forwarded like any other startup packet.
const (
	sslRequestCode    = 80877103
	gssEncRequestCode = 80877104
)

// maxStatements bounds what one recorder keeps in memory. Reaching it marks the
// artifact truncated rather than growing without limit — a migration or a seed
// script can issue tens of thousands of statements.
const maxStatements = 20000

// dsnEnvKeys are the variables the recorder falls back to when `dsn:` is unset,
// in preference order.
var dsnEnvKeys = []string{"GAVEL_DB_DSN", "DATABASE_URL"}

// SQLRecorder is a postgres wire proxy the fixture's child process is pointed
// at. It forwards bytes verbatim in both directions and decodes a copy, so
// extended-query, COPY and pipelined traffic pass through unchanged: a
// re-serialising proxy corrupts those in ways that present as flaky tests.
//
// It records what a *child process* does. gavel's own queries go through
// GormLogger instead — a proxy in this process would see nothing of the child,
// and the logger sees nothing of gavel.
type SQLRecorder struct {
	opts     SQLOptions
	upstream string   // host:port the child's traffic is forwarded to
	dsn      *url.URL // the resolved upstream DSN, rewritten for the child
	listener net.Listener

	mu         sync.Mutex
	statements []Statement
	truncated  bool
	conns      []net.Conn
	closed     bool
}

// StartSQL resolves the upstream DSN and starts the proxy on a loopback port.
//
// The upstream must accept unencrypted connections: the proxy answers the
// child's SSLRequest with a refusal and dials the upstream in the clear. That
// suits an embedded or local postgres, which is what fixtures run against; a
// managed cloud database that requires TLS cannot be recorded this way.
func StartSQL(opts SQLOptions) (*SQLRecorder, error) {
	if opts.Mode == "" {
		opts.Mode = SQLProxy
	}
	if opts.Mode != SQLProxy {
		return nil, fmt.Errorf("record: sql mode %q does not start a proxy", opts.Mode)
	}

	dsn, err := resolveDSN(opts.DSN)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("record: sql proxy listen: %w", err)
	}

	recorder := &SQLRecorder{opts: opts, upstream: upstreamAddr(dsn), dsn: dsn, listener: listener}
	go recorder.serve()
	return recorder, nil
}

// resolveDSN expands `$VAR` references and falls back to the environment when
// `dsn:` is unset. It fails loudly rather than recording nothing: a proxy with
// no upstream is the silent-empty-capture failure this package exists to
// prevent.
func resolveDSN(declared string) (*url.URL, error) {
	raw := strings.TrimSpace(os.ExpandEnv(declared))
	if raw == "" {
		for _, key := range dsnEnvKeys {
			if value := strings.TrimSpace(os.Getenv(key)); value != "" {
				raw = value
				break
			}
		}
	}
	if raw == "" {
		return nil, fmt.Errorf("record: sql needs an upstream — set `dsn:` or one of %s", strings.Join(dsnEnvKeys, ", "))
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("record: sql dsn %q: %w", raw, err)
	}
	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return nil, fmt.Errorf("record: sql dsn must be a postgres:// URL, got %q", parsed.Scheme)
	}
	return parsed, nil
}

func upstreamAddr(dsn *url.URL) string {
	host, port := dsn.Hostname(), dsn.Port()
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "5432"
	}
	return net.JoinHostPort(host, port)
}

// Addr is the proxy's host:port.
func (r *SQLRecorder) Addr() string { return r.listener.Addr().String() }

// Upstream is the host:port traffic is forwarded to.
func (r *SQLRecorder) Upstream() string { return r.upstream }

// ChildEnv is the environment a child process needs so its queries are
// observed: the same DSN with the host and port swapped for the proxy's, plus
// the libpq variables a client that builds its own connection string reads.
//
// sslmode is forced off because the proxy refuses TLS; leaving a `require` in
// the DSN would make the child fail to connect rather than fail to be recorded.
func (r *SQLRecorder) ChildEnv() map[string]string {
	rewritten := *r.dsn
	rewritten.Host = r.Addr()

	query := rewritten.Query()
	query.Set("sslmode", "disable")
	rewritten.RawQuery = query.Encode()

	host, port, _ := net.SplitHostPort(r.Addr())
	env := map[string]string{
		"GAVEL_DB_DSN": rewritten.String(),
		// Rewritten too: it is where the DSN most often came from, and a child
		// reading the original would bypass the proxy and record nothing.
		"DATABASE_URL": rewritten.String(),
		"PGHOST":       host,
		"PGPORT":       port,
		"PGSSLMODE":    "disable",
	}
	if database := strings.TrimPrefix(rewritten.Path, "/"); database != "" {
		env["PGDATABASE"] = database
	}
	return env
}

func (r *SQLRecorder) serve() {
	for {
		client, err := r.listener.Accept()
		if err != nil {
			return // Close closed the listener
		}
		r.track(client)
		go r.handle(client)
	}
}

// handle proxies one client connection, sniffing a copy of each direction.
func (r *SQLRecorder) handle(client net.Conn) {
	defer client.Close()

	server, err := net.Dial("tcp", r.upstream)
	if err != nil {
		r.add(Statement{started: time.Now(), SQL: "", Op: "CONNECT", Rows: -1,
			Error: fmt.Sprintf("dial %s: %v", r.upstream, err)})
		return
	}
	defer server.Close()
	r.track(server)

	if err := r.negotiateStartup(client, server); err != nil {
		return
	}

	session := newPGSession(r, r.opts.Params)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = io.Copy(io.MultiWriter(server, &pgSniffer{onMessage: session.frontend}), client)
		// Half-close so the server sees the client's EOF and answers it,
		// instead of both sides waiting on the other.
		if half, ok := server.(*net.TCPConn); ok {
			_ = half.CloseWrite()
		}
	}()

	_, _ = io.Copy(io.MultiWriter(client, &pgSniffer{onMessage: session.backend}), server)
	<-done
	session.flush()
}

// negotiateStartup forwards the client's StartupMessage, refusing any TLS or
// GSS negotiation that precedes it. A refusal is what the protocol expects
// ('N'), so a client configured for `sslmode=prefer` falls back cleanly.
func (r *SQLRecorder) negotiateStartup(client, server net.Conn) error {
	for {
		header := make([]byte, 8)
		if _, err := io.ReadFull(client, header); err != nil {
			return err
		}
		length := int(binary.BigEndian.Uint32(header[:4]))
		if length < 8 || length > maxWireMessage {
			return fmt.Errorf("record: sql startup packet length %d out of range", length)
		}

		switch binary.BigEndian.Uint32(header[4:8]) {
		case sslRequestCode, gssEncRequestCode:
			if _, err := client.Write([]byte{'N'}); err != nil {
				return err
			}
			continue
		}

		packet := make([]byte, length)
		copy(packet, header)
		if _, err := io.ReadFull(client, packet[8:]); err != nil {
			return err
		}
		_, err := server.Write(packet)
		return err
	}
}

func (r *SQLRecorder) add(statement Statement) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.statements) >= maxStatements {
		r.truncated = true
		return
	}
	r.statements = append(r.statements, statement)
}

func (r *SQLRecorder) track(conn net.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conns = append(r.conns, conn)
}

// Statements returns everything recorded so far, oldest first. Connections are
// sniffed concurrently, so the sort is what makes the artifact reproducible.
func (r *SQLRecorder) Statements() []Statement {
	r.mu.Lock()
	statements := append([]Statement(nil), r.statements...)
	r.mu.Unlock()

	sort.SliceStable(statements, func(i, j int) bool { return statements[i].started.Before(statements[j].started) })
	return statements
}

// StatementsBetween narrows to one fixture's window — a heuristic under the
// default file scope, where the file's tests share a recorder and overlap.
func (r *SQLRecorder) StatementsBetween(from, to time.Time) []Statement {
	return StatementsBetween(r.Statements(), from, to)
}

// Save writes statements as a JSONL artifact and returns the reference that
// goes on the fixture's result.
func (r *SQLRecorder) Save(store *Store, label string, statements []Statement) (Result, error) {
	result, err := SaveStatements(store, label, statements)

	r.mu.Lock()
	result.Truncated = r.truncated
	r.mu.Unlock()

	return result, err
}

// Close stops accepting and drops every connection. Unlike the HTTP recorder
// there is no drain window: a postgres connection is long-lived by design, so
// waiting for one to end on its own would wait for the child to exit, which has
// already happened by the time this runs.
func (r *SQLRecorder) Close() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	conns := append([]net.Conn(nil), r.conns...)
	r.mu.Unlock()

	err := r.listener.Close()
	for _, conn := range conns {
		_ = conn.Close()
	}
	return err
}
