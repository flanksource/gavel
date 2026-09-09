package record

import (
	"encoding/binary"
	"io"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePostgres speaks just enough of the wire protocol to answer the messages a
// client sends, so the proxy can be exercised end to end without a real server.
// It is the only kind of postgres this package's tests need: the recorder never
// interprets results, only the shape of the conversation.
type fakePostgres struct {
	listener net.Listener
	// onQuery answers a simple or extended query with a command tag. Anything
	// starting "fail:" is answered with an ErrorResponse instead.
	tag string
}

func startFakePostgres(t *testing.T, tag string) *fakePostgres {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := &fakePostgres{listener: listener, tag: tag}
	go server.serve()
	t.Cleanup(func() { _ = listener.Close() })
	return server
}

func (f *fakePostgres) dsn() string {
	return "postgres://gavel@" + f.listener.Addr().String() + "/fixtures?sslmode=disable"
}

func (f *fakePostgres) serve() {
	for {
		conn, err := f.listener.Accept()
		if err != nil {
			return
		}
		go f.handle(conn)
	}
}

func (f *fakePostgres) handle(conn net.Conn) {
	defer conn.Close()

	backend := pgproto3.NewBackend(conn, conn)
	if _, err := backend.ReceiveStartupMessage(); err != nil {
		return
	}
	backend.Send(&pgproto3.AuthenticationOk{})
	backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
	if backend.Flush() != nil {
		return
	}

	for {
		msg, err := backend.Receive()
		if err != nil {
			return
		}
		switch msg := msg.(type) {
		case *pgproto3.Query:
			f.answer(backend, msg.String)
			backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
		case *pgproto3.Parse:
			backend.Send(&pgproto3.ParseComplete{})
		case *pgproto3.Bind:
			backend.Send(&pgproto3.BindComplete{})
		case *pgproto3.Execute:
			f.answer(backend, "")
		case *pgproto3.Sync:
			backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
		case *pgproto3.Terminate:
			return
		}
		if backend.Flush() != nil {
			return
		}
	}
}

func (f *fakePostgres) answer(backend *pgproto3.Backend, sql string) {
	if sql == "fail" {
		backend.Send(&pgproto3.ErrorResponse{Severity: "ERROR", Message: `relation "missing" does not exist`})
		return
	}
	backend.Send(&pgproto3.CommandComplete{CommandTag: []byte(f.tag)})
}

// pgClient is the fixture's child process, reduced to the protocol it speaks.
type pgClient struct {
	conn     net.Conn
	frontend *pgproto3.Frontend
}

func dialProxy(t *testing.T, recorder *SQLRecorder) *pgClient {
	t.Helper()

	conn, err := net.Dial("tcp", recorder.Addr())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	client := &pgClient{conn: conn, frontend: pgproto3.NewFrontend(conn, conn)}
	client.frontend.Send(&pgproto3.StartupMessage{
		ProtocolVersion: pgproto3.ProtocolVersionNumber,
		Parameters:      map[string]string{"user": "gavel", "database": "fixtures"},
	})
	require.NoError(t, client.frontend.Flush())
	client.until(t, func(msg pgproto3.BackendMessage) bool {
		_, ready := msg.(*pgproto3.ReadyForQuery)
		return ready
	})
	return client
}

// until reads until stop says so, which is how a test waits for the server's
// answer without assuming how many messages preceded it.
func (c *pgClient) until(t *testing.T, stop func(pgproto3.BackendMessage) bool) {
	t.Helper()

	require.NoError(t, c.conn.SetReadDeadline(time.Now().Add(10*time.Second)))
	for {
		msg, err := c.frontend.Receive()
		require.NoError(t, err)
		if stop(msg) {
			return
		}
	}
}

func (c *pgClient) query(t *testing.T, sql string) {
	t.Helper()

	c.frontend.Send(&pgproto3.Query{String: sql})
	require.NoError(t, c.frontend.Flush())
	c.until(t, func(msg pgproto3.BackendMessage) bool {
		_, ready := msg.(*pgproto3.ReadyForQuery)
		return ready
	})
}

func startSQLRecorder(t *testing.T, opts SQLOptions) *SQLRecorder {
	t.Helper()

	recorder, err := StartSQL(opts)
	require.NoError(t, err)
	t.Cleanup(func() { _ = recorder.Close() })
	return recorder
}

// eventually polls because the sniffer sees a copy of the stream: the client can
// observe the server's answer before the recorder has finished decoding it.
func eventually(t *testing.T, recorder *SQLRecorder, want int) []Statement {
	t.Helper()

	var statements []Statement
	require.Eventually(t, func() bool {
		statements = recorder.Statements()
		return len(statements) >= want
	}, 10*time.Second, 10*time.Millisecond, "recorded %d of %d statements", len(statements), want)
	return statements
}

func TestSQLProxyRecordsSimpleQueries(t *testing.T) {
	server := startFakePostgres(t, "SELECT 3")
	recorder := startSQLRecorder(t, SQLOptions{DSN: server.dsn()})

	client := dialProxy(t, recorder)
	client.query(t, "SELECT id FROM users")
	client.query(t, "INSERT INTO users (id) VALUES (1)")

	statements := eventually(t, recorder, 2)
	require.Len(t, statements, 2)

	assert.Equal(t, "SELECT id FROM users", statements[0].SQL)
	assert.Equal(t, "SELECT", statements[0].Op)
	assert.Equal(t, []string{"users"}, statements[0].Tables)
	assert.Equal(t, 3, statements[0].Rows, "the row count comes off the command tag")
	assert.Empty(t, statements[0].Error)

	assert.Equal(t, "INSERT", statements[1].Op)
}

func TestSQLProxyRecordsTheServersError(t *testing.T) {
	server := startFakePostgres(t, "SELECT 0")
	recorder := startSQLRecorder(t, SQLOptions{DSN: server.dsn()})

	client := dialProxy(t, recorder)
	client.query(t, "fail")

	statements := eventually(t, recorder, 1)
	require.Len(t, statements, 1)
	assert.Contains(t, statements[0].Error, `relation "missing" does not exist`)
	assert.Equal(t, -1, statements[0].Rows, "a failed statement reports no row count, not zero")
}

func TestSQLProxyResolvesTheExtendedQueryPortal(t *testing.T) {
	server := startFakePostgres(t, "SELECT 1")
	recorder := startSQLRecorder(t, SQLOptions{DSN: server.dsn(), Params: true})

	client := dialProxy(t, recorder)
	client.frontend.Send(&pgproto3.Parse{Name: "s1", Query: "SELECT * FROM users WHERE id = $1"})
	client.frontend.Send(&pgproto3.Bind{DestinationPortal: "", PreparedStatement: "s1",
		Parameters: [][]byte{[]byte("42")}})
	client.frontend.Send(&pgproto3.Execute{Portal: ""})
	client.frontend.Send(&pgproto3.Sync{})
	require.NoError(t, client.frontend.Flush())
	client.until(t, func(msg pgproto3.BackendMessage) bool {
		_, ready := msg.(*pgproto3.ReadyForQuery)
		return ready
	})

	statements := eventually(t, recorder, 1)
	require.Len(t, statements, 1)
	// Execute names only a portal, so the SQL has to be carried through
	// Parse → Bind → Execute for the statement to be identifiable at all.
	assert.Equal(t, "SELECT * FROM users WHERE id = $1", statements[0].SQL)
	assert.Equal(t, []string{"42"}, statements[0].Params)
}

func TestSQLProxyDropsBindParamsByDefault(t *testing.T) {
	server := startFakePostgres(t, "SELECT 1")
	recorder := startSQLRecorder(t, SQLOptions{DSN: server.dsn()})

	client := dialProxy(t, recorder)
	client.frontend.Send(&pgproto3.Parse{Name: "s1", Query: "SELECT * FROM secrets WHERE token = $1"})
	client.frontend.Send(&pgproto3.Bind{PreparedStatement: "s1", Parameters: [][]byte{[]byte("hunter2")}})
	client.frontend.Send(&pgproto3.Execute{})
	client.frontend.Send(&pgproto3.Sync{})
	require.NoError(t, client.frontend.Flush())
	client.until(t, func(msg pgproto3.BackendMessage) bool {
		_, ready := msg.(*pgproto3.ReadyForQuery)
		return ready
	})

	statements := eventually(t, recorder, 1)
	require.Len(t, statements, 1)
	assert.Nil(t, statements[0].Params, "bind values carry row data and are off unless asked for")
}

// A client configured for sslmode=prefer opens with an SSLRequest. The proxy has
// to refuse it and still complete the startup, or the child fails to connect
// rather than merely failing to be recorded.
func TestSQLProxyRefusesTLSAndStillConnects(t *testing.T) {
	server := startFakePostgres(t, "SELECT 1")
	recorder := startSQLRecorder(t, SQLOptions{DSN: server.dsn()})

	conn, err := net.Dial("tcp", recorder.Addr())
	require.NoError(t, err)
	defer conn.Close()

	request := make([]byte, 8)
	binary.BigEndian.PutUint32(request[:4], 8)
	binary.BigEndian.PutUint32(request[4:], sslRequestCode)
	_, err = conn.Write(request)
	require.NoError(t, err)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(10*time.Second)))
	answer := make([]byte, 1)
	_, err = io.ReadFull(conn, answer)
	require.NoError(t, err)
	assert.Equal(t, byte('N'), answer[0], "the protocol's decline, so sslmode=prefer falls back")

	frontend := pgproto3.NewFrontend(conn, conn)
	frontend.Send(&pgproto3.StartupMessage{
		ProtocolVersion: pgproto3.ProtocolVersionNumber,
		Parameters:      map[string]string{"user": "gavel", "database": "fixtures"},
	})
	require.NoError(t, frontend.Flush())

	for {
		msg, err := frontend.Receive()
		require.NoError(t, err)
		if _, ready := msg.(*pgproto3.ReadyForQuery); ready {
			break
		}
	}
}

func TestChildEnvPointsAtTheProxyWithoutTLS(t *testing.T) {
	server := startFakePostgres(t, "SELECT 1")
	recorder := startSQLRecorder(t, SQLOptions{DSN: server.dsn()})

	env := recorder.ChildEnv()

	host, port, err := net.SplitHostPort(recorder.Addr())
	require.NoError(t, err)
	assert.Equal(t, host, env["PGHOST"])
	assert.Equal(t, port, env["PGPORT"])
	assert.Equal(t, "disable", env["PGSSLMODE"])
	assert.Equal(t, "fixtures", env["PGDATABASE"])

	// DATABASE_URL is rewritten too: it is most often where the DSN came from,
	// and a child reading the original would bypass the proxy and record nothing.
	for _, key := range []string{"GAVEL_DB_DSN", "DATABASE_URL"} {
		parsed, err := url.Parse(env[key])
		require.NoError(t, err, key)
		assert.Equal(t, recorder.Addr(), parsed.Host, key)
		assert.Equal(t, "disable", parsed.Query().Get("sslmode"), key)
		assert.Equal(t, "/fixtures", parsed.Path, key)
	}
	assert.NotEqual(t, recorder.Addr(), recorder.Upstream(), "the proxy is not the upstream")
}

func TestResolveDSNFailsLoudRatherThanRecordingNothing(t *testing.T) {
	t.Setenv("GAVEL_DB_DSN", "")
	t.Setenv("DATABASE_URL", "")

	_, err := resolveDSN("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "GAVEL_DB_DSN")

	_, err = resolveDSN("mysql://localhost/db")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "postgres://")
}

func TestResolveDSNExpandsAndFallsBackToTheEnvironment(t *testing.T) {
	t.Setenv("GAVEL_DB_DSN", "")
	t.Setenv("DATABASE_URL", "postgres://from-env/db")
	t.Setenv("SOME_DSN", "postgres://expanded/db")

	fromEnv, err := resolveDSN("")
	require.NoError(t, err)
	assert.Equal(t, "from-env", fromEnv.Host)

	expanded, err := resolveDSN("$SOME_DSN")
	require.NoError(t, err)
	assert.Equal(t, "expanded", expanded.Host, "a declared dsn expands $VAR references")
}

func TestUpstreamAddrDefaultsToLocalPostgres(t *testing.T) {
	dsn, err := url.Parse("postgres:///fixtures")
	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:5432", upstreamAddr(dsn))
}

func TestStartSQLRejectsTheInProcessMode(t *testing.T) {
	// The two modes watch different processes, so silently starting a proxy for
	// `mode: inprocess` would record the wrong one's queries.
	_, err := StartSQL(SQLOptions{Mode: SQLInProcess, DSN: "postgres://localhost/db"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not start a proxy")
}
