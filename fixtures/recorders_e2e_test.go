package fixtures_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flanksource/clicky/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/record"
	_ "github.com/flanksource/gavel/fixtures/types"
)

// curlPath resolves the client the recorder fixtures drive. It is required
// rather than skipped: a silently skipped recorder test is exactly the failure
// mode `requireEntries` exists to prevent, and curl ships on every platform CI
// runs on.
func curlPath(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("curl")
	require.NoError(t, err, "curl is required to exercise the http recorder")
	return path
}

// recordingServer answers two paths so a fixture can produce more than one
// entry and assert on the difference between them.
func recordingServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Set-Cookie", "session=super-secret")
		fmt.Fprint(w, `{"status":"ok"}`)
	})
	mux.HandleFunc("/missing", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// runFixture writes body as a markdown fixture in its own directory and runs it
// through the full runner, returning the leaf result.
func runFixture(t *testing.T, body string) fixtures.FixtureResult {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "record.fixture.md")
	require.NoError(t, os.WriteFile(path, []byte(strings.TrimSpace(body)+"\n"), 0644))

	runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: []string{path}, WorkDir: dir})
	require.NoError(t, err)

	tree, _ := runner.Run()
	require.NotNil(t, tree)

	var leaf *fixtures.FixtureNode
	tree.Walk(func(node *fixtures.FixtureNode) {
		if node.Test != nil && node.Results != nil {
			leaf = node
		}
	})
	require.NotNil(t, leaf, "expected a fixture result")
	return *leaf.Results
}

// The full path: `record: http` starts a proxy for the file, the child inherits
// it through HTTP_PROXY, its traffic lands in a HAR under .gavel/recordings,
// and the fixture's CEL expression can assert on what was recorded.
func TestRecordHTTPProducesHARAndCELVars(t *testing.T) {
	curl := curlPath(t)
	server := recordingServer(t)

	result := runFixture(t, fmt.Sprintf(`
---
record: http
---

# Recorded

## fetches two urls

`+"```bash"+`
%s -sS -o /dev/null %s/ok -o /dev/null %s/missing
`+"```"+`
`, curl, server.URL, server.URL))

	require.Equal(t, "", result.Error, "fixture should pass")
	require.Len(t, result.Recordings, 1)

	recording := result.Recordings[0]
	assert.Equal(t, record.KindHTTP, recording.Kind)
	assert.Equal(t, "har-1.2", recording.Format)
	assert.Equal(t, 2, recording.Count, "one entry per request")
	assert.Empty(t, recording.Error)

	har, err := os.ReadFile(recording.Path)
	require.NoError(t, err, "recording path must resolve")
	assert.Contains(t, string(har), "/ok")
	assert.Contains(t, string(har), "/missing")
	// The response's Set-Cookie is blanked on disk, not merely hidden from CEL.
	assert.NotContains(t, string(har), "super-secret")
	assert.Contains(t, string(har), "[redacted]")

	assert.Contains(t, filepath.ToSlash(recording.Path), ".gavel/recordings/",
		"recordings live in a subdirectory so snapshots' run-*.json glob is untouched")
}

// A CEL expression over the `http` root is the point of the whole feature, so
// it is asserted through a fixture that would fail if the root were missing.
func TestRecordHTTPCELRootIsAssertable(t *testing.T) {
	curl := curlPath(t)
	server := recordingServer(t)

	result := runFixture(t, fmt.Sprintf(`
---
record: http
cel: |
  http.entries == 2 &&
  http.errors == 1 &&
  http.methods["GET"] == 2 &&
  http.statuses["404"] == 1 &&
  http.requests.filter(r, r.path == "/ok").size() == 1
---

# Recorded

## asserts on the recording

`+"```bash"+`
%s -sS -o /dev/null %s/ok -o /dev/null %s/missing
`+"```"+`
`, curl, server.URL, server.URL))

	assert.Equal(t, "", result.Error, "CEL over the http root should hold")
}

// The silent failure `requireEntries` exists for: a child that never reaches
// the proxy records nothing, and without the guard its fixture passes on an
// empty HAR.
func TestRequireEntriesFailsAnEmptyRecording(t *testing.T) {
	result := runFixture(t, `
---
record:
  http: {requireEntries: 1}
---

# Recorded

## makes no calls

`+"```bash"+`
echo no traffic
`+"```"+`
`)

	require.NotEmpty(t, result.Error, "an empty recording must be a red fixture")
	assert.Contains(t, result.Error, "captured 0 of the 1 required entries")
	assert.Equal(t, task.StatusFAIL, result.Status)
}

// `record: ansi` implies a PTY, writes a cast and exposes the `cast` root — and
// the fixture's own stdout still has to be the stream the child wrote, since the
// capture now produces both.
func TestRecordANSIProducesACastAndCELVars(t *testing.T) {
	result := runFixture(t, `
---
record: ansi
cel: |
  cast.exit_code == 0 &&
  cast.events > 0 &&
  cast.width == 120 &&
  cast.final.contains("rendered") &&
  !cast.has_duplicates
---

# Recorded

## renders under a pty

`+"```bash"+`
printf 'rendered\n'
`+"```"+`
`)

	require.Equal(t, "", result.Error, "the cast root should be assertable")
	assert.Contains(t, result.Stdout, "rendered", "the raw stream is still the fixture's stdout")

	require.Len(t, result.Recordings, 1)
	recording := result.Recordings[0]
	assert.Equal(t, record.KindANSI, recording.Kind)
	assert.Equal(t, "asciinema-v2", recording.Format)
	assert.Positive(t, recording.Count, "one entry per output chunk")

	cast, err := os.ReadFile(recording.Path)
	require.NoError(t, err, "recording path must resolve")
	header, events, found := strings.Cut(string(cast), "\n")
	require.True(t, found, "a cast is a header line followed by events")
	assert.Contains(t, header, `"version":2`)
	assert.Contains(t, events, `"o"`)
}

// A child that connects to the original DSN bypasses the proxy and is recorded
// as nothing, so what the fixture inherits is the whole contract. The protocol
// itself is covered in fixtures/record; this is the wiring that points a child
// at it.
func TestRecordSQLPointsTheChildAtTheProxy(t *testing.T) {
	result := runFixture(t, `
---
record:
  sql: {dsn: "postgres://gavel@127.0.0.1:5432/fixtures"}
cel: sql.statements == 0 && sql.errors == 0
---

# Recorded

## reports the connection it inherited

`+"```bash"+`
echo "PGHOST=$PGHOST PGPORT=$PGPORT PGSSLMODE=$PGSSLMODE PGDATABASE=$PGDATABASE DATABASE_URL=$DATABASE_URL"
`+"```"+`
`)

	require.Equal(t, "", result.Error, "the sql root should be assertable on an empty capture")

	fields := map[string]string{}
	for _, pair := range strings.Fields(strings.TrimSpace(result.Stdout)) {
		key, value, _ := strings.Cut(pair, "=")
		fields[key] = value
	}
	assert.Equal(t, "127.0.0.1", fields["PGHOST"])
	assert.Equal(t, "disable", fields["PGSSLMODE"], "the proxy refuses TLS, so the child must not ask for it")
	assert.Equal(t, "fixtures", fields["PGDATABASE"])
	assert.NotEqual(t, "5432", fields["PGPORT"], "the child talks to the proxy, not the upstream")
	assert.Contains(t, fields["DATABASE_URL"], ":"+fields["PGPORT"]+"/fixtures",
		"DATABASE_URL is rewritten too, or a child reading it would bypass the proxy")

	require.Len(t, result.Recordings, 1)
	assert.Equal(t, record.KindSQL, result.Recordings[0].Kind)
	assert.Equal(t, "jsonl", result.Recordings[0].Format)
}

// Recorder env is applied only for keys the fixture left unset, so a fixture
// that deliberately points somewhere else keeps doing so.
func TestFixtureEnvOutranksTheRecorder(t *testing.T) {
	result := runFixture(t, `
---
record:
  sql: {dsn: "postgres://gavel@127.0.0.1:5432/fixtures"}
env:
  PGHOST: elsewhere.example
---

# Recorded

## keeps its own connection

`+"```bash"+`
echo "PGHOST=$PGHOST"
`+"```"+`
`)

	require.Equal(t, "", result.Error)
	assert.Contains(t, result.Stdout, "PGHOST=elsewhere.example")
}

// `clients` records gavel's own outbound HTTP, which a child process cannot
// produce — so what a fixture asserts is gavel's traffic during its window, and
// a bash child correctly contributes nothing. The transport itself is covered in
// fixtures/record; this is the wiring that turns it into a recording.
func TestRecordClientsProducesAHARForGavelsOwnTraffic(t *testing.T) {
	curl := curlPath(t)
	server := recordingServer(t)

	result := runFixture(t, fmt.Sprintf(`
---
record: clients
cel: clients.entries == 0
---

# Recorded

## makes calls gavel did not make

`+"```bash"+`
%s -sS -o /dev/null %s/ok
`+"```"+`
`, curl, server.URL))

	require.Equal(t, "", result.Error, "the clients root should be assertable on an empty capture")

	require.Len(t, result.Recordings, 1)
	assert.Equal(t, record.KindClients, result.Recordings[0].Kind,
		"a client recording and a proxy recording are the same document about different traffic")
	assert.Equal(t, "har-1.2", result.Recordings[0].Format)
}

// The in-process recorders share one process-global sink, so two files asking
// for one would each get the other's traffic. Naming the conflict beats writing
// two wrong artifacts.
func TestTwoFilesCannotBothRecordGavelItself(t *testing.T) {
	dir := t.TempDir()
	var paths []string
	for _, name := range []string{"a", "b"} {
		body := `---
record: clients
---

# Recorded

## makes no calls

` + "```bash\necho " + name + "\n```\n"
		path := filepath.Join(dir, name+".fixture.md")
		require.NoError(t, os.WriteFile(path, []byte(body), 0644))
		paths = append(paths, path)
	}

	runner, err := fixtures.NewRunner(fixtures.RunnerOptions{Paths: paths, WorkDir: dir})
	require.NoError(t, err)

	_, runErr := runner.Run()
	require.Error(t, runErr)
	assert.Contains(t, runErr.Error(), "watches gavel itself")
}

// Whether a given runtime honours the trust variables is best-effort — that is
// what `requireEntries` is for — but gavel's half of the contract is not: under
// mitm the certificate must exist on disk and the child must be pointed at it.
func TestMITMPointsTheChildAtTheGeneratedCA(t *testing.T) {
	result := runFixture(t, `
---
record:
  http: {mode: mitm}
---

# Recorded

# reports the trust it inherited

`+"```bash"+`
echo "SSL_CERT_FILE=$SSL_CERT_FILE NODE_EXTRA_CA_CERTS=$NODE_EXTRA_CA_CERTS"
`+"```"+`
`)

	require.Equal(t, "", result.Error)

	fields := map[string]string{}
	for _, pair := range strings.Fields(strings.TrimSpace(result.Stdout)) {
		key, value, _ := strings.Cut(pair, "=")
		fields[key] = value
	}
	require.NotEmpty(t, fields["SSL_CERT_FILE"], "the child must inherit a trust anchor")
	assert.Equal(t, fields["SSL_CERT_FILE"], fields["NODE_EXTRA_CA_CERTS"],
		"every runtime is pointed at the same generated CA")

	pem, err := os.ReadFile(fields["SSL_CERT_FILE"])
	require.NoError(t, err, "the trust variables must point at a file that exists")
	assert.Contains(t, string(pem), "BEGIN CERTIFICATE")
	assert.NotContains(t, string(pem), "PRIVATE KEY", "the CA key never leaves the process")
}

// Without `record:` nothing starts and nothing is written — the guarantee that
// makes the feature free for the runs that do not use it.
func TestWithoutRecordNoRecordingsAreProduced(t *testing.T) {
	curl := curlPath(t)
	server := recordingServer(t)

	result := runFixture(t, fmt.Sprintf(`
# Unrecorded

## fetches a url

`+"```bash"+`
%s -sS -o /dev/null %s/ok
`+"```"+`
`, curl, server.URL))

	require.Equal(t, "", result.Error)
	assert.Empty(t, result.Recordings)
}
