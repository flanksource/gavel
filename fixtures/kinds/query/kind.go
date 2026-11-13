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
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/formatters"
	"github.com/flanksource/clicky/task"
	"github.com/flanksource/commons-db/connection"
	dbcontext "github.com/flanksource/commons-db/context"
	cdbmodels "github.com/flanksource/commons-db/models"
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
	// AutoDiagnose is how long a statement runs before its server session is
	// sampled for CPU, IO, waits and blocking; nil is defaultAutoDiagnose, 0
	// off, -1 from the start.
	AutoDiagnose *autoDiagnose `json:"autoDiagnose,omitempty"`
}

// defaultAutoDiagnose is when an unset autoDiagnose starts sampling.
const defaultAutoDiagnose = 10 * time.Second

// autoDiagnose is a duration string ("10s", "0s"), or -1 (number or string)
// for sampling from the start, which commons-db reads as a negative delay.
type autoDiagnose time.Duration

func (a *autoDiagnose) UnmarshalJSON(b []byte) error {
	text := strings.Trim(string(b), `"`)
	if text == "-1" {
		*a = autoDiagnose(-1)
		return nil
	}
	parsed, err := time.ParseDuration(text)
	if err != nil {
		return fmt.Errorf(`autoDiagnose must be a duration such as "10s", "0s" to turn it off, or -1 to sample from the start: %w`, err)
	}
	if parsed < 0 {
		return fmt.Errorf("autoDiagnose %s is negative: use -1 to sample from the start", text)
	}
	*a = autoDiagnose(parsed)
	return nil
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
	// target is where a SQL connection points; zero for a non-SQL connection.
	target connection.SQLTarget
	// autoDiagnose is when the server session starts being sampled; 0 never.
	autoDiagnose time.Duration
}

// Execute records the query on the step as it runs — the record under
// QueryMetadataKey, the server's messages as stdout/stderr, the query as the
// command — so a run that fails keeps the same detail one that passes does.
func (Kind) Execute(ctx context.Context, step *fixtures.StepContext) (fixtures.StepOutcome, error) {
	r, err := prepare(ctx, step)
	if err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("query step %s: %w", stepName(step), err)
	}
	command := r.profile.Query
	if r.profile.QueryFile != "" {
		command = r.profile.QueryFile
	}
	step.SetCommand(command, "")
	live := r.startLive(step)
	diagnostics := cdbquery.NewDiagnostics(cdbquery.DiagnosticOptions{Provider: r.profile.Provider.Type, Walk: true, OnRequest: live.issued,
		AutoDiagnose: r.autoDiagnose, OnSession: live.session, OnSample: live.sample, OnSampleError: live.sampleError})
	outputs, violations, response, err := r.execute(cdbquery.WithDiagnosticSink(cdbquery.WithBatchObserver(r.resolved.Context, live.batch), diagnostics), step)
	record, finalErr := live.final(diagnostics)
	if err = errors.Join(err, finalErr); err != nil {
		return fixtures.StepOutcome{}, fmt.Errorf("query step %s: %w", stepName(step), err)
	}
	if session, ok := record["session"]; ok {
		if outputs["session"], err = fixtures.PlainValue(session); err != nil {
			return fixtures.StepOutcome{}, fmt.Errorf("query step %s: %w", stepName(step), err)
		}
	}
	result := findingsResult(step, violations)
	result.Metadata = map[string]any{fixtures.ViewMetadataKey: formatters.NewClickyDocument(responseView(r.profile, response))}
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
	target, err := sqlTarget(resolved.Connection, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("connection %q: %w", resolved.Connection.Name, err)
	}
	autoDiagnose := defaultAutoDiagnose
	if cfg.AutoDiagnose != nil {
		autoDiagnose = time.Duration(*cfg.AutoDiagnose)
	}
	return &run{profile: profile, params: params, resolved: resolved, findings: findings, target: target, autoDiagnose: autoDiagnose}, nil
}

// sqlTarget is the host, port, user and database a SQL connection opens —
// database being the step's `database:` when it switches to another one.
func sqlTarget(c *cdbmodels.Connection, database string) (connection.SQLTarget, error) {
	if _, ok := providerForConnection[c.Type]; !ok {
		return connection.SQLTarget{}, nil
	}
	var sql connection.SQLConnection
	if err := sql.FromModel(*c); err != nil {
		return connection.SQLTarget{}, err
	}
	target, err := sql.Target()
	if database != "" {
		target.Database = database
	}
	return target, err
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

// execute runs the profile and builds the step outputs, findings included,
// returning the response it read.
func (r *run) execute(ctx dbcontext.Context, step *fixtures.StepContext) (map[string]any, []models.Violation, *cdbquery.Result, error) {
	result, err := cdbquery.Execute(ctx, r.profile, r.params)
	if err != nil {
		return nil, nil, nil, err
	}
	outputs, violations, err := r.outputs(step, result)
	if err != nil {
		return nil, nil, nil, err
	}
	return outputs, violations, result, nil
}

// responseView is the query's response through commons-db's own clicky table
// (Result.Table over the profile's declared columns), the same rendering the
// CLI and a profile's /api/v1/profile/<key> give. A SQL program shows every
// result set of every batch, under the batch it came from; its server
// messages are the step's stdout and stderr (liveQuery.batch), not the response.
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
		body := api.Text{}
		for _, set := range batch.ResultSets {
			body = body.Add((&cdbquery.Result{Profile: result.Profile, Rows: set}).Table(nil)).NewLine()
		}
		view = view.Add(api.Collapsed{Label: heading, Content: body}).NewLine()
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
	sets, err := resultSetsOf(list, result.Batches)
	if err != nil {
		return nil, nil, err
	}
	outputs := map[string]any{
		"rows": list, "rowCount": len(list), "columns": columnsOf(r.profile, result.Rows),
		"resultSets": sets,
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

// connectionOutput names the connection and where it points — host, port,
// user and the database queried — never its secrets.
func (r *run) connectionOutput() map[string]any {
	out := map[string]any{"name": r.resolved.Connection.Name, "type": r.resolved.Connection.Type}
	for key, value := range map[string]string{"host": r.target.Host, "port": r.target.Port, "user": r.target.User, databaseOption: r.target.Database} {
		if value != "" {
			out[key] = value
		}
	}
	return out
}

// resultSetsOf is every result set the query returned, in the order the server
// sent them across batches and runs. A SELECT read as a page has no batches:
// its rows are its one result set. A program's rows are its last set.
func resultSetsOf(rows []any, batches []cdbquery.BatchResult) ([]any, error) {
	if len(batches) == 0 {
		return []any{rows}, nil
	}
	sets := []any{}
	for _, batch := range batches {
		for _, set := range batch.ResultSets {
			plain, err := fixtures.PlainValue(set)
			if err != nil {
				return nil, err
			}
			if plain == nil {
				plain = []any{}
			}
			sets = append(sets, plain)
		}
	}
	return sets, nil
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
