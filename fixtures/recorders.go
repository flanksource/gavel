package fixtures

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/flanksource/commons/logger"

	"github.com/flanksource/gavel/fixtures/record"
)

// fixtureEnv is the per-file runtime environment a fixture executes in: the
// prepared `setup:` and, keyed the same way, the recorders watching it. Both
// are file-level, so they travel together rather than as separate parameters
// that could disagree about which file they came from.
type fixtureEnv struct {
	file  string
	setup *PreparedSetup
}

// recorderSet holds the recorders prepared for one markdown file. Under the
// default `scope: file` every test in the file shares them, so a test's own
// artifact is a time-slice of the file's — see EntriesBetween.
//
// ANSI is the exception: a cast is produced by the fixture's own PTY, so there
// is nothing to start here and nothing shared — only the options the fixture
// captures with.
type recorderSet struct {
	http     *record.HTTPRecorder
	httpOpts record.HTTPOptions
	ansi     *record.ANSIOptions
	// sql is the wire proxy a child process is pointed at; sqlLog is the
	// in-process alternative that records gavel's own queries. They are
	// alternatives, never both: one watches the child, the other watches gavel.
	sql    *record.SQLRecorder
	sqlLog *record.StatementLog
	// clients records gavel's own outbound HTTP, which no proxy can see: the
	// child never makes those calls.
	clients *record.ClientLog
}

// RecorderContext is what a fixture type needs to run under the recorders:
// the environment a child process must inherit to be observed. It is passed on
// RunOptions rather than on FixtureTest because FixtureTest is serialized into
// the result, and a proxy address is runtime state, not a declaration.
type RecorderContext struct {
	// ProxyEnv routes the child's HTTP traffic through the recorder. Applied
	// only for keys the fixture has not set itself.
	ProxyEnv map[string]string

	// ANSI is non-nil when the fixture must run under a PTY and hand its
	// capture back through Harvest. It implies `terminal: pty` — there is no
	// ANSI to record from a pipe.
	ANSI *record.ANSIOptions

	// Harvest writes the artifacts for one fixture's slice of the recording.
	// The fixture type calls it rather than the runner because the CEL roots it
	// returns have to exist before the fixture's own expression is evaluated,
	// and only the fixture type knows when its child finished.
	Harvest func(HarvestRequest) Harvest
}

// HarvestRequest is one fixture's claim on the recording: which test, the
// window its child ran in, and anything the fixture captured itself.
type HarvestRequest struct {
	Label      string
	Start, End time.Time

	// Capture is the fixture's PTY recording, set when RecorderContext.ANSI
	// asked for one.
	Capture *Capture
}

// Harvest is what one fixture's recordings turned into: the artifact references
// that travel with the result, and the CEL roots the fixture can assert on.
type Harvest struct {
	Recordings []Recording
	CELVars    map[string]any

	// Err fails the fixture. Only `requireEntries` sets it: a recorder that
	// captured nothing is indistinguishable from a fixture that made no calls,
	// and that ambiguity is exactly what the fixture asked to rule out.
	Err error
}

// effectiveRecord is the spec a fixture runs with: its own declaration if it
// made one, and the run-wide --record only if it did not. An explicit
// `record: none` parses to an empty non-nil Spec precisely so it lands here as
// a declaration and outranks the flag.
func (r *Runner) effectiveRecord(fixture FixtureTest) *record.Spec {
	if spec := fixture.ExecBase().Record; spec != nil {
		return spec
	}
	return r.options.Record
}

// prepareRecorders starts one recorder per markdown file that asks for one.
// Runs after the daemon so the daemon's port can be excluded from the proxy —
// otherwise a `curl localhost:{{.port}}` fixture fills its own HAR with
// self-traffic.
func (r *Runner) prepareRecorders() error {
	byFile := map[string]*record.Spec{}
	r.tree.Walk(func(node *FixtureNode) {
		if node.Test == nil || node.Origin == nil {
			return
		}
		spec := r.effectiveRecord(*node.Test)
		if spec.IsEmpty() {
			return
		}
		if _, exists := byFile[node.Origin.File]; !exists {
			byFile[node.Origin.File] = spec
		}
	})
	if len(byFile) == 0 {
		return nil
	}

	// WorkDir is otherwise resolved lazily, per command, deep inside
	// executeFixture — too late for a store that must exist before the first
	// fixture starts.
	if r.options.WorkDir == "" {
		r.options.WorkDir, _ = os.Getwd()
	}

	r.recorders = map[string]*recorderSet{}
	r.store = record.NewStore(r.options.WorkDir, time.Now())

	// The recorders that watch gavel itself — `sql: {mode: inprocess}` and
	// `clients` — share one process-global sink each. A second file declaring
	// one would capture the first file's traffic into its own artifact and leave
	// the first empty, so the conflict is named rather than recorded wrongly.
	owners := map[record.Kind]string{}
	claim := func(kind record.Kind, file string) error {
		if owner, taken := owners[kind]; taken {
			return fmt.Errorf("%s: record: %s watches gavel itself, so only one file can record it per run (already declared by %s)",
				file, kind, owner)
		}
		owners[kind] = file
		return nil
	}

	for file, spec := range byFile {
		set := &recorderSet{ansi: spec.ANSI}
		if opts := spec.HTTP; spec.Enabled(record.KindHTTP) {
			if opts.Scope == record.ScopeTest {
				return fmt.Errorf("%s: record: http scope %q is not implemented yet", file, record.ScopeTest)
			}
			recorder, err := record.StartHTTP(*opts)
			if err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
			// Before the first ProxyEnv: the trust variables only appear once the
			// certificate exists on disk for the child to read.
			if err := recorder.WriteCA(r.store, filepath.Base(file)); err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
			logger.V(2).Infof("recording http for %s via %s", file, recorder.Addr())
			set.http, set.httpOpts = recorder, *opts
		}
		if opts := spec.SQL; spec.Enabled(record.KindSQL) {
			if opts.Scope == record.ScopeTest {
				return fmt.Errorf("%s: record: sql scope %q is not implemented yet", file, record.ScopeTest)
			}
			if opts.Mode == record.SQLInProcess {
				if err := claim(record.KindSQL, file); err != nil {
					return err
				}
			}
			if err := startSQL(set, *opts); err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
		}
		if opts := spec.Clients; spec.Enabled(record.KindClients) {
			if err := claim(record.KindClients, file); err != nil {
				return err
			}
			set.clients = record.StartClients(*opts)
		}
		r.recorders[file] = set
	}
	return nil
}

// startSQL attaches whichever of the two SQL recorders the mode asks for. They
// watch different processes — the proxy watches the fixture's child, the
// in-process log watches gavel itself — so the mode is a choice about whose
// queries are interesting, not an implementation detail.
func startSQL(set *recorderSet, opts record.SQLOptions) error {
	if opts.Mode == record.SQLInProcess {
		set.sqlLog = record.StartInProcess()
		return nil
	}
	recorder, err := record.StartSQL(opts)
	if err != nil {
		return err
	}
	logger.V(2).Infof("recording sql via %s -> %s", recorder.Addr(), recorder.Upstream())
	set.sql = recorder
	return nil
}

// recorderContext returns what a fixture in file needs to be observed, or nil
// when nothing is recording it.
func (r *Runner) recorderContext(file string) *RecorderContext {
	set := r.recorders[file]
	if set == nil {
		return nil
	}
	ctx := &RecorderContext{
		ANSI: set.ansi,
		Harvest: func(req HarvestRequest) Harvest {
			return r.harvest(set, req)
		},
	}
	if set.http != nil {
		var noProxy []string
		if r.daemonPort > 0 {
			noProxy = append(noProxy, "127.0.0.1:"+strconv.Itoa(r.daemonPort), "localhost:"+strconv.Itoa(r.daemonPort))
		}
		ctx.ProxyEnv = set.http.ProxyEnv(noProxy)
	}
	if set.sql != nil {
		if ctx.ProxyEnv == nil {
			ctx.ProxyEnv = map[string]string{}
		}
		for key, value := range set.sql.ChildEnv() {
			ctx.ProxyEnv[key] = value
		}
	}
	return ctx
}

// harvest writes one fixture's artifacts and summarises them for CEL. A
// recorder that fails to write is reported on the recording itself rather than
// failing the fixture: the fixture's own assertions already ran, and losing the
// evidence is not the same as failing the test.
func (r *Runner) harvest(set *recorderSet, req HarvestRequest) Harvest {
	harvest := Harvest{CELVars: map[string]any{}}

	if req.Capture != nil {
		cast := req.Capture.Cast()
		result, err := record.SaveCast(r.store, req.Label, cast)
		if err != nil {
			result.Kind = record.KindANSI
			result.Error = err.Error()
		}
		harvest.Recordings = append(harvest.Recordings, result)
		harvest.CELVars["cast"] = record.CastCELVars(cast, result.Path)
	}

	if set.http != nil {
		entries := set.http.EntriesBetween(req.Start, req.End)
		result, err := set.http.Save(r.store, req.Label, entries)
		if err != nil {
			result.Kind = record.KindHTTP
			result.Error = err.Error()
		}
		harvest.Recordings = append(harvest.Recordings, result)
		harvest.CELVars["http"] = record.HTTPCELVars(entries, result.Path)

		if want := set.httpOpts.RequireEntries; want > 0 && len(entries) < want {
			harvest.Err = fmt.Errorf("record: http captured %d of the %d required entries — the child either made no calls or could not reach the proxy",
				len(entries), want)
		}
	}

	if statements, save := set.sqlHarvester(); save != nil {
		window := statements(req.Start, req.End)
		result, err := save(r.store, req.Label, window)
		if err != nil {
			result.Kind = record.KindSQL
			result.Error = err.Error()
		}
		harvest.Recordings = append(harvest.Recordings, result)
		harvest.CELVars["sql"] = record.SQLCELVars(window, result.Path)
	}

	if set.clients != nil {
		entries := set.clients.Between(req.Start, req.End)
		result, err := set.clients.Save(r.store, req.Label, entries)
		if err != nil {
			result.Kind = record.KindClients
			result.Error = err.Error()
		}
		harvest.Recordings = append(harvest.Recordings, result)
		harvest.CELVars["clients"] = record.HTTPCELVars(entries, result.Path)
	}
	return harvest
}

// sqlHarvester resolves the two SQL modes to one pair of funcs, so harvest does
// not branch on which recorder is attached.
func (s *recorderSet) sqlHarvester() (func(from, to time.Time) []record.Statement,
	func(*record.Store, string, []record.Statement) (record.Result, error)) {
	switch {
	case s.sql != nil:
		return s.sql.StatementsBetween, s.sql.Save
	case s.sqlLog != nil:
		return s.sqlLog.Between, s.sqlLog.Save
	}
	return nil, nil
}

// closeRecorders stops every recorder. Registered after the daemon's defer so
// LIFO closes the proxies first: a daemon shutting down can still emit
// requests, and they belong in the recording.
func (r *Runner) closeRecorders() {
	for file, set := range r.recorders {
		if set.http != nil {
			if err := set.http.Close(); err != nil {
				logger.Warnf("%s: closing http recorder: %v", file, err)
			}
		}
		if set.sql != nil {
			if err := set.sql.Close(); err != nil {
				logger.Warnf("%s: closing sql recorder: %v", file, err)
			}
		}
		if set.sqlLog != nil {
			record.StopInProcess()
		}
		if set.clients != nil {
			record.StopClients()
		}
	}
}
