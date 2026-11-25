package query

import (
	"fmt"
	"strings"
	"sync"

	cdbquery "github.com/flanksource/commons-db/query"
	"github.com/flanksource/commons-db/query/sqlactivity"

	"github.com/flanksource/gavel/fixtures"
)

// liveQuery is the step's query record while it runs — the connection it runs
// on, the query as authored and its params, then the statement the provider
// issued with its arguments, then each batch end — recorded on the step under
// QueryMetadataKey as each arrives (StepContext.SetDetail), which publishes it
// live and keeps it as the result's record whether the run succeeds or not.
// Server messages (PRINT, RAISERROR) are the step's stdout and stderr.
type liveQuery struct {
	step   *fixtures.StepContext
	mu     sync.Mutex
	record map[string]any
	// err is the first record that failed; Execute fails the step with it,
	// since the provider's callbacks cannot return one.
	err error
}

func (r *run) startLive(step *fixtures.StepContext) *liveQuery {
	record := map[string]any{"template": r.profile.Query, "params": r.params, "connection": r.connectionOutput()}
	if r.profile.QueryFile != "" {
		record["queryFile"] = r.profile.QueryFile
	}
	live := &liveQuery{step: step, record: record}
	live.update(func(map[string]any) {})
	return live
}

// update applies change to the record and records it; SetDetail copies it,
// so the next change does not reach a published snapshot.
func (l *liveQuery) update(change func(record map[string]any)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	change(l.record)
	if err := l.step.SetDetail(fixtures.QueryMetadataKey, l.record); err != nil && l.err == nil {
		l.err = err
	}
}

// issued records the statement the provider is about to issue and the
// arguments it bound (commons-db DiagnosticOptions.OnRequest).
func (l *liveQuery) issued(request cdbquery.ProviderDiagnosticRequest) {
	l.update(func(record map[string]any) {
		record["sql"] = request.Query
		if len(request.Arguments) > 0 {
			record["arguments"] = request.Arguments
		}
	})
}

// batch takes a SQL program's batch event (commons-db WithBatchObserver): a
// server message is a line of the step's output — the ANSI warnings SQL Server
// prefixes with "Warning:" on stderr, PRINT and RAISERROR text on stdout; a
// batch's end, with the rows it changed where the driver reports them, goes on
// the record under `batches`.
func (l *liveQuery) batch(event cdbquery.BatchEvent) {
	if !event.Done {
		output := l.step.Stdout()
		if strings.HasPrefix(event.Message, "Warning:") {
			output = l.step.Stderr()
		}
		fmt.Fprintln(output, event.Message)
		return
	}
	l.update(func(record map[string]any) {
		entry := map[string]any{"batch": event.Batch, "run": event.Run}
		if event.RowsAffected != nil {
			entry["rowsAffected"] = *event.RowsAffected
		}
		batches, _ := record["batches"].([]map[string]any)
		record["batches"] = append(batches, entry)
	})
}

// diagnoseLimit is how many AutoDiagnose samples the record keeps, the
// latest; older ones are counted under diagnoseDropped.
const diagnoseLimit = 100

// session records the server session the statement runs on (commons-db
// DiagnosticOptions.OnSession); a walk's later pages replace it.
func (l *liveQuery) session(session cdbquery.ServerSession) {
	l.update(func(record map[string]any) { record["session"] = session })
}

// sample appends an AutoDiagnose sample of the running statement's session
// (commons-db DiagnosticOptions.OnSample), so the live view streams its CPU,
// IO, waits and blocking.
func (l *liveQuery) sample(sample sqlactivity.Sample) {
	l.update(func(record map[string]any) {
		samples, _ := record["diagnose"].([]sqlactivity.Sample)
		samples = append(samples, sample)
		if len(samples) > diagnoseLimit {
			dropped, _ := record["diagnoseDropped"].(int)
			record["diagnoseDropped"] = dropped + len(samples) - diagnoseLimit
			samples = samples[len(samples)-diagnoseLimit:]
		}
		record["diagnose"] = samples
	})
}

// sampleError records why AutoDiagnose stopped sampling (commons-db
// DiagnosticOptions.OnSampleError) and prints it on the step's stderr. It is
// the diagnostics' failure, not the query's, so the step goes on.
func (l *liveQuery) sampleError(err error) {
	fmt.Fprintf(l.step.Stderr(), "autoDiagnose: %v\n", err)
	l.update(func(record map[string]any) { record["diagnoseError"] = err.Error() })
}

// final settles the record once the run has finished, failed or not: its SQL
// and arguments as the diagnostics recorded them (the templated query when the
// provider recorded no statement of its own).
func (l *liveQuery) final(diagnostics *cdbquery.ProviderDiagnostics) (map[string]any, error) {
	l.update(func(record map[string]any) {
		sql := diagnostics.Request.Query
		if sql == "" {
			sql = diagnostics.Request.Rendered
		}
		if sql != "" {
			record["sql"] = sql
		}
		if len(diagnostics.Request.Arguments) > 0 {
			record["arguments"] = diagnostics.Request.Arguments
		}
	})
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.record, l.err
}
