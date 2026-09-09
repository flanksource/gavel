package record

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
)

func init() { Implemented[KindHTTP] = true }

// httpTimeFormat is HAR's startedDateTime: RFC3339 with millisecond precision,
// which is what every viewer expects to be able to sort on.
const httpTimeFormat = "2006-01-02T15:04:05.000Z07:00"

// drainTimeout bounds Close. A fixture killed by the run's task timeout leaves
// its child's connections open; the recorder waits this long for them to finish
// on their own and then force-closes, because one hung child must not hang the
// whole run.
const drainTimeout = 2 * time.Second

// HTTPRecorder is an HTTP proxy the fixture's child process is pointed at,
// writing what it sees as a HAR document.
//
// In connect mode — the default — it records one entry per CONNECT tunnel: the
// host, the duration and the byte counts, with the payload left encrypted.
// Plain (non-TLS) requests are recorded in full regardless of mode, since they
// pass through the proxy in the clear whether or not anyone asked to see them.
type HTTPRecorder struct {
	opts     HTTPOptions
	listener net.Listener
	server   *http.Server
	// upstream chains to the developer's own proxy when HTTPS_PROXY is set in
	// gavel's environment. goproxy would do this for us via ConnectDial, but
	// ConnectDial takes precedence over the per-request Dialer that does the
	// byte counting, so the chaining is re-done here instead.
	upstream func(network, addr string) (net.Conn, error)
	// ca and caPath are set only in mitm mode: the ephemeral authority the
	// per-host certificates are signed with, and where its certificate was
	// materialised for children to trust.
	ca     *CA
	caPath string

	mu      sync.Mutex
	entries []Entry
	conns   map[net.Conn]struct{}
	closed  bool
}

// StartHTTP starts a recording proxy on a loopback port.
//
// The listener is created here and kept, rather than going through the runner's
// freePort() helper: that closes its listener before returning the port, and
// the resulting window between choosing and binding is a real race once several
// recorders start at once.
func StartHTTP(opts HTTPOptions) (*HTTPRecorder, error) {
	if opts.Mode == "" {
		opts.Mode = HTTPConnect
	}
	var ca *CA
	if opts.Mode == HTTPMITM {
		var err error
		if ca, err = newCA(); err != nil {
			return nil, err
		}
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("record: http proxy listen: %w", err)
	}

	recorder := &HTTPRecorder{opts: opts, listener: listener, ca: ca, conns: map[net.Conn]struct{}{}}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Logger = silentLogger{}
	// ConnectDial is cleared so ctx.Dialer wins; see the upstream field.
	recorder.upstream = proxy.ConnectDial
	proxy.ConnectDial = nil
	proxy.OnRequest().HandleConnectFunc(recorder.handleConnect)
	proxy.OnRequest().DoFunc(recorder.onRequest)
	proxy.OnResponse().DoFunc(recorder.onResponse)

	recorder.server = &http.Server{Handler: proxy, ReadHeaderTimeout: 30 * time.Second}
	go func() { _ = recorder.server.Serve(listener) }()

	return recorder, nil
}

// Addr is the proxy's host:port.
func (r *HTTPRecorder) Addr() string { return r.listener.Addr().String() }

// ProxyEnv is the environment a child process needs to route through the
// recorder. Both cases are set because curl reads the lowercase names and Node
// reads the uppercase ones.
//
// noProxy names hosts the child must reach directly: the proxy's own address,
// plus any daemon port the fixture started — otherwise a `curl
// localhost:{{.port}}` fixture fills its own HAR with self-traffic.
//
// Loopback as a whole is deliberately *not* excluded. A fixture pointing at a
// server it started on 127.0.0.1 is the most common thing worth recording, and
// blanket-excluding loopback would make the recorder blind to it.
func (r *HTTPRecorder) ProxyEnv(noProxy []string) map[string]string {
	url := "http://" + r.Addr()
	joined := strings.Join(append([]string{r.Addr()}, noProxy...), ",")
	env := map[string]string{
		"HTTP_PROXY":  url,
		"http_proxy":  url,
		"HTTPS_PROXY": url,
		"https_proxy": url,
		"NO_PROXY":    joined,
		"no_proxy":    joined,
	}
	if r.caPath != "" {
		for key, value := range trustEnv(r.caPath) {
			env[key] = value
		}
	}
	return env
}

// WriteCA materialises the generated CA certificate so children can be pointed
// at it, and must be called before ProxyEnv for the trust variables to appear.
// A no-op outside mitm mode.
//
// Only the certificate is written — the private key stays in this process, so a
// leftover file in .gavel/recordings cannot be used to intercept anyone.
func (r *HTTPRecorder) WriteCA(store *Store, label string) error {
	if r.ca == nil {
		return nil
	}
	file, path, err := store.CreateSidecar(label, KindHTTP, "ca.pem")
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(r.ca.certPEM); err != nil {
		return fmt.Errorf("write ca %s: %w", path, err)
	}
	r.caPath = path
	return nil
}

// CAPath is where the generated certificate was written, empty outside mitm
// mode or before WriteCA.
func (r *HTTPRecorder) CAPath() string { return r.caPath }

// records reports whether host is in scope. An empty `hosts:` records
// everything; otherwise the globs decide.
func (r *HTTPRecorder) records(host string) bool {
	if len(r.opts.Hosts) == 0 {
		return true
	}
	host = hostOnly(host)
	for _, pattern := range r.opts.Hosts {
		if ok, _ := path.Match(pattern, host); ok {
			return true
		}
	}
	return false
}

// handleConnect accepts every tunnel and, for in-scope hosts, swaps in a dialer
// whose connection counts its own bytes. Counting this way rather than
// hijacking the tunnel leaves goproxy's half-close handling — and its careful
// TCP teardown — exactly as it is.
func (r *HTTPRecorder) handleConnect(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	if !r.records(host) {
		return goproxy.OkConnect, host
	}
	// A decrypted tunnel is recorded as the exchanges inside it, by the same
	// request/response handlers plain HTTP goes through, so there is nothing for
	// the byte counter below to add. Out-of-scope hosts already returned above
	// and stay encrypted — `hosts:` is what limits what mitm decrypts.
	if r.ca != nil {
		return &goproxy.ConnectAction{
			Action:    goproxy.ConnectMitm,
			TLSConfig: goproxy.TLSConfigFromCA(&r.ca.cert),
		}, host
	}
	started := time.Now()
	ctx.Dialer = func(_ context.Context, network, addr string) (net.Conn, error) {
		conn, err := r.dial(network, addr)
		if err != nil {
			r.add(connectEntry(host, started, time.Since(started), 0, 0, err))
			return nil, err
		}
		return r.track(newCountingConn(conn, r, host, started)), nil
	}
	return goproxy.OkConnect, host
}

func (r *HTTPRecorder) dial(network, addr string) (net.Conn, error) {
	if r.upstream != nil {
		return r.upstream(network, addr)
	}
	return net.Dial(network, addr)
}

// onRequest stamps the start time and, when bodies are wanted, captures the
// request body before the transport consumes it. goproxy hands the same
// ProxyCtx to the response handler, so UserData carries the state across.
func (r *HTTPRecorder) onRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	state := exchangeState{started: time.Now()}
	if r.opts.Bodies > 0 && req.Body != nil && r.records(req.URL.Host) {
		state.post = capturePostData(req, r.opts.Bodies.Or(defaultBodies))
	}
	ctx.UserData = state
	return req, nil
}

func (r *HTTPRecorder) onResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	if ctx.Req == nil || !r.records(ctx.Req.URL.Host) {
		return resp
	}
	state, ok := ctx.UserData.(exchangeState)
	if !ok {
		state = exchangeState{started: time.Now()}
	}
	r.add(exchangeEntry(r.policy(), ctx.Req, resp, state, ctx.Error))
	return resp
}

func (r *HTTPRecorder) policy() capturePolicy {
	return capturePolicy{Bodies: r.opts.Bodies, Redact: r.opts.Redact}
}

func (r *HTTPRecorder) add(entry Entry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, entry)
}

func (r *HTTPRecorder) track(conn net.Conn) net.Conn {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conns[conn] = struct{}{}
	return conn
}

func (r *HTTPRecorder) untrack(conn net.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conns, conn)
}

// Entries returns everything recorded so far, oldest first. Concurrent fixtures
// append out of order, so the sort is what makes the artifact reproducible.
func (r *HTTPRecorder) Entries() []Entry {
	r.mu.Lock()
	entries := append([]Entry(nil), r.entries...)
	r.mu.Unlock()

	sort.SliceStable(entries, func(i, j int) bool { return entries[i].started.Before(entries[j].started) })
	return entries
}

// EntriesBetween narrows to the entries that started inside a window. Under the
// default file scope one proxy is shared by every test in the file, and those
// tests run concurrently, so attributing an entry to a single test is a
// heuristic — `scope: test` is the exact escape hatch.
func (r *HTTPRecorder) EntriesBetween(from, to time.Time) []Entry {
	var window []Entry
	for _, entry := range r.Entries() {
		if entry.started.Before(from) || entry.started.After(to) {
			continue
		}
		window = append(window, entry)
	}
	return window
}

// Save writes entries as a HAR artifact under store and returns the reference
// that goes on the fixture's result. The payload stays on disk: only the counts
// travel with the result.
func (r *HTTPRecorder) Save(store *Store, label string, entries []Entry) (Result, error) {
	return SaveHAR(store, label, KindHTTP, entries)
}

// SaveHAR writes entries as a HAR artifact. The kind is a parameter because a
// proxy recording and a client recording are the same document about different
// traffic, and only the kind tells them apart on the result.
func SaveHAR(store *Store, label string, kind Kind, entries []Entry) (Result, error) {
	file, result, err := store.Create(label, kind)
	if err != nil {
		return Result{}, err
	}
	defer file.Close()

	if err := WriteHAR(file, entries); err != nil {
		return result, fmt.Errorf("write har %s: %w", result.Path, err)
	}
	if info, err := file.Stat(); err == nil {
		result.Bytes = info.Size()
	}

	result.Count = len(entries)
	for _, entry := range entries {
		if entry.Gavel != nil && entry.Gavel.Error != "" {
			result.Errors++
		}
		result.DurationMs += int64(entry.Time)
	}
	return result, nil
}

// Close stops accepting, waits out the drain window for in-flight tunnels and
// then force-closes whatever is left.
func (r *HTTPRecorder) Close() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil
	}
	r.closed = true
	r.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), drainTimeout)
	defer cancel()
	// Shutdown does not track hijacked connections, which is every CONNECT
	// tunnel, so the force-close below is the real teardown rather than a
	// fallback.
	err := r.server.Shutdown(ctx)

	r.mu.Lock()
	conns := make([]net.Conn, 0, len(r.conns))
	for conn := range r.conns {
		conns = append(conns, conn)
	}
	r.mu.Unlock()
	for _, conn := range conns {
		_ = conn.Close()
	}

	if err != nil && err != context.DeadlineExceeded {
		return err
	}
	return nil
}

// hostOnly strips a :port suffix, leaving IPv6 literals intact.
func hostOnly(host string) string {
	if stripped, _, err := net.SplitHostPort(host); err == nil {
		return stripped
	}
	return host
}

// silentLogger drops goproxy's chatter, which by default goes to stderr and
// would interleave with the fixture's own output. Transport failures are not
// lost with it: they are recorded on the entry as _gavel.error.
type silentLogger struct{}

func (silentLogger) Printf(string, ...any) {}
