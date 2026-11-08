// Package query implements the `kind: query` fixture step: it runs a
// commons-db query profile against a resolved connection, exposes the rows as
// step outputs and maps them to findings — gavel lint violations — that fail
// (error) or warn (warning, info) the step.
//
// The SQL provider binds every `{{.params.<name>}}` / `$(.params.<name>)`
// reference as a driver argument. A single SELECT is read as a page; anything
// else — EXEC, DML, several statements, a GO-separated script (inline, or an
// inline profile's `queryFile` relative to the fixture) — runs as a SQL
// program, batch by batch on one session, reported under `batches`.
package query

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/api/icons"
	"github.com/flanksource/clicky/formatters"
	"github.com/flanksource/clicky/task"
	dbcontext "github.com/flanksource/commons-db/context"
	cdbquery "github.com/flanksource/commons-db/query"
	_ "github.com/flanksource/commons-db/query/providers"

	"github.com/flanksource/gavel/fixtures"
	"github.com/flanksource/gavel/fixtures/kinds/dbconn"
	"github.com/flanksource/gavel/models"
)

//go:embed schema.json
var schema []byte

func init() { fixtures.RegisterKind(Kind{}) }

// Kind is the `query` fixture step kind.
type Kind struct{}

func (Kind) Name() string { return "query" }

func (Kind) Schema() json.RawMessage { return schema }

// RawConfigKeys leaves the query text to the SQL provider, which binds its
// params references, and the findings to Execute, whose templates read .row.
func (Kind) RawConfigKeys() []string { return []string{"query", "findings"} }

// Config is a query step's kind-specific config.
type Config struct {
	Connection any             `json:"connection,omitempty"`
	Database   string          `json:"database,omitempty"`
	Query      json.RawMessage `json:"query"`
	Params     map[string]any  `json:"params,omitempty"`
	Findings   json.RawMessage `json:"findings,omitempty"`
	Limit      int             `json:"limit,omitempty"`
}

// databaseOption is the SQL provider option that switches the connection to
// another database on the same server (commons-db SQLConnection.UseDatabase).
const databaseOption = "database"

// run is the Handle Observe re-runs.
type run struct {
	profile  cdbquery.Profile
	params   map[string]any
	resolved *dbconn.Resolved
	findings *findingsSpec
}

func (Kind) Execute(ctx context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	began := time.Now()
	r, err := prepare(ctx, step)
	if err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("query step %s: %w", stepName(step), err)
	}
	diagnostics := cdbquery.NewDiagnostics(cdbquery.DiagnosticOptions{Provider: r.profile.Provider.Type, Walk: true})
	live := cdbquery.WithDiagnosticSink(cdbquery.WithBatchObserver(r.resolved.Context, streamBatches(fixtures.OutputWriter(ctx))), diagnostics)
	outputs, violations, view, err := r.execute(live, step)
	if err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("query step %s: %w", stepName(step), err)
	}
	result := findingsResult(step, violations)
	result.Metadata = map[string]any{fixtures.ViewMetadataKey: formatters.NewClickyDocument(view), fixtures.QueryMetadataKey: r.queryRecord(diagnostics)}
	result.Command, result.Duration = r.profile.Query, time.Since(began)
	if r.profile.QueryFile != "" {
		result.Command = r.profile.QueryFile
	}
	if step.Test.Step == nil && step.Test.Expected.CEL != "" && result.Status != task.StatusFAIL {
		result = applyRowCEL(result, step, outputs)
	}
	return fixtures.StepOutcome{Result: result, Outputs: outputs, Handle: r}, nil
}

// Observe re-runs the query, so `wait.until` polls the state it reads. A poll
// does not stream its batches: Execute's run already did, and a wait polling
// every interval would repeat the same lines for as long as it waits.
func (Kind) Observe(_ context.Context, step *fixtures.StepContext, handle any) (map[string]any, error) {
	r, ok := handle.(*run)
	if !ok {
		return nil, fmt.Errorf("query: Observe got a %T handle, want the one Execute returned", handle)
	}
	outputs, _, _, err := r.execute(r.resolved.Context, step)
	return outputs, err
}

// streamBatches writes a SQL program's batch events to the step's live
// output as `batch N: <message>` and `batch N done (k rows affected)` lines.
func streamBatches(output io.Writer) func(cdbquery.BatchEvent) {
	return func(event cdbquery.BatchEvent) {
		label := fmt.Sprintf("batch %d", event.Batch)
		if event.Run > 1 {
			label += fmt.Sprintf(" run %d", event.Run)
		}
		switch {
		case !event.Done:
			fmt.Fprintf(output, "%s: %s\n", label, event.Message)
		case event.RowsAffected != nil:
			fmt.Fprintf(output, "%s done (%d rows affected)\n", label, *event.RowsAffected)
		default:
			fmt.Fprintf(output, "%s done\n", label)
		}
	}
}

func prepare(ctx context.Context, step *fixtures.StepContext) (*run, error) {
	cfg, err := loadConfig(step)
	if err != nil {
		return nil, err
	}
	findings, err := parseFindings(cfg.Findings)
	if err != nil {
		return nil, err
	}
	profile, params, err := buildProfile(ctx, step, cfg)
	if err != nil {
		return nil, err
	}
	ref := cfg.Connection
	if ref == nil && profile.Provider.Connection != "" {
		ref = profile.Provider.Connection
	}
	if ref == nil {
		return nil, errors.New("no connection: set connection on the step (or a Connection column / frontmatter connection), or provider.connection on the profile")
	}
	resolved, err := dbconn.Resolve(ctx, step, ref, dbconn.Options{})
	if err != nil {
		return nil, err
	}
	profile.Provider.Connection = resolved.Ref
	if cfg.Database != "" {
		if profile.Provider.Options == nil {
			profile.Provider.Options = map[string]any{}
		}
		profile.Provider.Options[databaseOption] = cfg.Database
	}
	if profile.Provider.Type == "" {
		if profile.Provider.Type = providerForConnection[resolved.Connection.Type]; profile.Provider.Type == "" {
			return nil, fmt.Errorf("connection %q is a %s connection, which no query provider reads; set provider.type on the profile",
				resolved.Connection.Name, resolved.Connection.Type)
		}
	}
	return &run{profile: profile, params: params, resolved: resolved, findings: findings}, nil
}

// loadConfig is a YAML step's config, or a markdown row's: its Query column,
// a Connection column (else the frontmatter `connection`) and a Findings
// column, which is the CEL form.
func loadConfig(step *fixtures.StepContext) (Config, error) {
	if step.Test.Step != nil {
		cfg, err := fixtures.KindConfig[Config](step)
		if err == nil && len(cfg.Query) == 0 {
			err = errors.New("query is required")
		}
		return cfg, err
	}
	query, err := json.Marshal(step.Test.Query)
	if err != nil {
		return Config{}, err
	}
	cfg := Config{Query: query, Connection: step.Test.Expected.Properties["connection"]}
	if cfg.Connection == nil {
		cfg.Connection = step.Test.Metadata["connection"]
	}
	if findings, ok := step.Test.Expected.Properties["findings"]; ok {
		if cfg.Findings, err = json.Marshal(findings); err != nil {
			return Config{}, err
		}
	}
	return cfg, nil
}

// execute runs the profile and builds the step outputs, findings included, and
// the response as commons-db renders it (view).
func (r *run) execute(ctx dbcontext.Context, step *fixtures.StepContext) (map[string]any, []models.Violation, api.Textable, error) {
	result, err := cdbquery.Execute(ctx, r.profile, r.params)
	if err != nil {
		return nil, nil, nil, err
	}
	outputs, violations, err := r.outputs(step, result)
	if err != nil {
		return nil, nil, nil, err
	}
	return outputs, violations, responseView(r.profile, result), nil
}

// responseView is the query's response through commons-db's own clicky table
// (Result.Table over the profile's declared columns), the same rendering the
// CLI and a profile's /api/v1/profile/<key> give. A SQL program shows every
// result set of every batch, under the batch it came from.
func responseView(profile cdbquery.Profile, result *cdbquery.Result) api.Textable {
	if len(result.Batches) == 0 {
		return result.Table(profile.Columns)
	}
	view := api.Text{}
	for _, batch := range result.Batches {
		heading := fmt.Sprintf("batch %d (lines %d-%d)", batch.Index, batch.StartLine, batch.EndLine)
		if batch.Run > 1 {
			heading += fmt.Sprintf(" run %d", batch.Run)
		}
		if batch.RowsAffected != nil {
			heading += fmt.Sprintf(" · %d row(s) affected", *batch.RowsAffected)
		}
		view = view.Append(heading, "font-medium").NewLine()
		for _, message := range batch.Messages {
			view = view.Add(serverMessage(message)).NewLine()
		}
		for _, set := range batch.ResultSets {
			view = view.Add((&cdbquery.Result{Profile: result.Profile, Rows: set}).Table(nil)).NewLine()
		}
	}
	return view
}

// outputs are the step's CEL/template variables built from the result.
func (r *run) outputs(step *fixtures.StepContext, result *cdbquery.Result) (map[string]any, []models.Violation, error) {
	rows, err := fixtures.PlainValue(result.Rows)
	if err != nil {
		return nil, nil, err
	}
	list, _ := rows.([]any)
	if list == nil {
		list = []any{}
	}
	outputs := map[string]any{
		"rows": list, "rowCount": len(list), "columns": columnsOf(r.profile, result.Rows),
		"truncated":  result.Truncated,
		"connection": r.connectionOutput(),
	}
	if len(list) > 0 {
		outputs["row"] = list[0]
	}
	if err := addBatches(outputs, result.Batches); err != nil {
		return nil, nil, err
	}
	violations, err := r.findings.evaluate(step, outputs)
	if err != nil {
		return nil, nil, err
	}
	plain, err := fixtures.PlainValue(violations)
	if err != nil {
		return nil, nil, err
	}
	outputs["findings"] = plain
	return outputs, violations, nil
}

// queryRecord is what the step ran, for the UI's Template / SQL view: the query
// as authored (its `$(.params.x)` references intact), the statement commons-db
// issued with them bound, the driver arguments, and the step's params by name.
// Read once the run has finished, so the recorder is no longer written.
func (r *run) queryRecord(diagnostics *cdbquery.ProviderDiagnostics) map[string]any {
	record := map[string]any{"template": r.profile.Query, "params": r.params}
	if r.profile.QueryFile != "" {
		record["queryFile"] = r.profile.QueryFile
	}
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
	return record
}

// serverMessage is one PRINT / RAISERROR / ANSI-warning message a batch
// returned: SQL Server prefixes its ANSI warnings with "Warning:", which get the
// warning icon; every other message is informational.
func serverMessage(message string) api.Text {
	if strings.HasPrefix(message, "Warning:") {
		return api.Text{}.Add(icons.Warning).Append(" "+message, "text-amber-600")
	}
	return api.Text{}.Add(icons.Info).Append(" "+message, "text-muted-foreground")
}

// connectionOutput names the connection (and the database selected on it),
// never its secrets.
func (r *run) connectionOutput() map[string]any {
	out := map[string]any{"name": r.resolved.Connection.Name, "type": r.resolved.Connection.Type}
	if database, ok := r.profile.Provider.Options[databaseOption].(string); ok && database != "" {
		out[databaseOption] = database
	}
	return out
}

// addBatches exposes a SQL program's per-batch results as `batches`, and the
// rows its batches report changing as `rowsAffected`. A SELECT read as a page
// has no batches and gets neither.
func addBatches(outputs map[string]any, batches []cdbquery.BatchResult) error {
	if len(batches) == 0 {
		return nil
	}
	plain, err := fixtures.PlainValue(batches)
	if err != nil {
		return err
	}
	outputs["batches"] = plain
	var affected int64
	for _, batch := range batches {
		if batch.RowsAffected != nil {
			affected += *batch.RowsAffected
		}
	}
	outputs["rowsAffected"] = affected
	return nil
}

// columnsOf is the profile's declared column order, else the first row's
// keys sorted (a provider row carries no order of its own).
func columnsOf(profile cdbquery.Profile, rows []cdbquery.Row) []any {
	columns := []any{}
	for _, column := range profile.Columns {
		columns = append(columns, column.Name)
	}
	if len(columns) > 0 || len(rows) == 0 {
		return columns
	}
	keys := make([]string, 0, len(rows[0]))
	for key := range rows[0] {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		columns = append(columns, key)
	}
	return columns
}

// applyRowCEL evaluates a markdown row's CEL Validation over the outputs.
func applyRowCEL(result fixtures.FixtureResult, step *fixtures.StepContext, outputs map[string]any) fixtures.FixtureResult {
	expression := step.Test.Expected.CEL
	result.CELExpression = expression
	ok, err := evalBool(expression, step.CELVars(outputs))
	switch {
	case err != nil:
		return result.Errorf(err, "evaluate cel %s", expression)
	case !ok:
		return result.Failf("%s is false", expression)
	}
	result.CELResult = true
	return result
}

func stepName(step *fixtures.StepContext) string {
	for _, name := range []string{step.Spec.Name, step.Spec.ID, step.Test.Name} {
		if name != "" {
			return name
		}
	}
	return "query"
}

func resultStatus(errors, warnings int) task.Status {
	switch {
	case errors > 0:
		return task.StatusFAIL
	case warnings > 0:
		return task.StatusWarning
	}
	return task.StatusPASS
}
