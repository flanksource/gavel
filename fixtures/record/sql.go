package record

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// celQueryCap bounds the per-statement detail exposed to CEL, mirroring
// celRequestCap: the aggregates stay exact, the long tail lives in the artifact.
const celQueryCap = 200

// Statement is one SQL statement as it was observed — the text the server was
// asked to run, what it answered, and how long it took. It is deliberately the
// same shape whether it came from the wire proxy or from gavel's own gorm
// logger, so a fixture asserts on `sql.queries` without caring which mode
// produced it.
type Statement struct {
	// Started orders statements and decides which fixture's window one falls in.
	// Unexported for the same reason as Entry.started: it is bookkeeping, not
	// part of the artifact's schema.
	started time.Time

	SQL string `json:"sql"`
	// Op is the leading keyword, upper-cased: SELECT, INSERT, BEGIN, …
	Op string `json:"op,omitempty"`
	// Tables are the relations named in the statement, best-effort.
	Tables []string `json:"tables,omitempty"`
	// Rows is the count from the command tag; -1 when the server did not report
	// one, which is different from an honest zero.
	Rows       int      `json:"rows"`
	DurationMs int64    `json:"duration_ms"`
	Error      string   `json:"error,omitempty"`
	Params     []string `json:"params,omitempty"`
	StartedAt  string   `json:"started_at,omitempty"`
}

// Start returns when the statement was issued.
func (s Statement) Start() time.Time { return s.started }

// WriteStatements serialises statements as JSONL — one object per line — so a
// long capture streams rather than having to be parsed whole, and so a shell
// pipeline can `jq` it directly.
func WriteStatements(w io.Writer, statements []Statement) error {
	buffered := bufio.NewWriter(w)
	encoder := json.NewEncoder(buffered)
	for _, statement := range statements {
		if statement.StartedAt == "" && !statement.started.IsZero() {
			statement.StartedAt = statement.started.UTC().Format(httpTimeFormat)
		}
		if err := encoder.Encode(statement); err != nil {
			return fmt.Errorf("write sql statement: %w", err)
		}
	}
	return buffered.Flush()
}

// SaveStatements writes an artifact under store and returns the reference that
// travels on the fixture's result.
func SaveStatements(store *Store, label string, statements []Statement) (Result, error) {
	file, result, err := store.Create(label, KindSQL)
	if err != nil {
		return Result{}, err
	}
	defer file.Close()

	if err := WriteStatements(file, statements); err != nil {
		return result, fmt.Errorf("write sql %s: %w", result.Path, err)
	}
	if info, err := file.Stat(); err == nil {
		result.Bytes = info.Size()
	}

	result.Count = len(statements)
	for _, statement := range statements {
		if statement.Error != "" {
			result.Errors++
		}
		result.DurationMs += statement.DurationMs
	}
	return result, nil
}

// SQLCELVars builds the `sql` CEL root. As with the http and cast roots every
// key is present even on an empty capture, so `sql.errors == 0` is a legal
// assertion for a fixture that issued no statements.
func SQLCELVars(statements []Statement, path string) map[string]any {
	byOp := map[string]int{}
	tableSet := map[string]bool{}
	queries := make([]map[string]any, 0, min(len(statements), celQueryCap))

	var errors int
	var durationMs int64

	for _, statement := range statements {
		byOp[statement.Op]++
		for _, table := range statement.Tables {
			tableSet[table] = true
		}
		if statement.Error != "" {
			errors++
		}
		durationMs += statement.DurationMs

		if len(queries) < celQueryCap {
			queries = append(queries, map[string]any{
				"sql":         statement.SQL,
				"op":          statement.Op,
				"tables":      statement.Tables,
				"rows":        statement.Rows,
				"duration_ms": statement.DurationMs,
				"error":       statement.Error,
			})
		}
	}

	return map[string]any{
		"statements":  len(statements),
		"errors":      errors,
		"duration_ms": durationMs,
		"by_op":       byOp,
		"tables":      sortedKeys(tableSet),
		"queries":     queries,
		"path":        path,
	}
}

// StatementsBetween narrows to the statements issued inside a window. The same
// time-slice heuristic the HTTP recorder documents applies: under the default
// file scope the tests sharing a recorder run concurrently and overlap.
func StatementsBetween(statements []Statement, from, to time.Time) []Statement {
	var window []Statement
	for _, statement := range statements {
		if statement.started.Before(from) || statement.started.After(to) {
			continue
		}
		window = append(window, statement)
	}
	return window
}

// classify fills in Op and Tables from the statement text. Deliberately a
// scanner over keywords rather than a SQL parser: the goal is a summary a
// fixture can assert coarsely on (`sql.by_op["INSERT"] == 1`), and the exact
// statement is right there in the artifact when more is needed.
func classify(sql string) (op string, tables []string) {
	fields := strings.Fields(strings.TrimSpace(sql))
	if len(fields) == 0 {
		return "", nil
	}
	op = strings.ToUpper(strings.TrimLeft(fields[0], "("))

	seen := map[string]bool{}
	for i, field := range fields {
		switch strings.ToUpper(field) {
		case "FROM", "INTO", "UPDATE", "JOIN", "TABLE":
			if i+1 < len(fields) {
				if name := tableName(fields[i+1]); name != "" && !seen[name] {
					seen[name] = true
					tables = append(tables, name)
				}
			}
		}
	}
	sort.Strings(tables)
	return op, tables
}

// tableName strips the punctuation an identifier picks up in situ — quotes, a
// trailing comma or paren — and rejects anything left that is not an
// identifier.
//
// A leading paren is a rejection rather than something to strip: `FROM (SELECT`
// is a sub-select, and trimming the paren would name a table "SELECT".
func tableName(field string) string {
	if strings.HasPrefix(field, "(") {
		return ""
	}
	name := strings.Trim(field, `"');,`)
	if name == "" {
		return ""
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '.', r == '$':
		default:
			return ""
		}
	}
	return name
}

// rowsFromTag reads the count off a command tag: "INSERT 0 1", "SELECT 3",
// "UPDATE 2". Tags without one ("BEGIN", "SET") report -1 rather than 0, so a
// fixture can tell "no rows" from "not a row-returning statement".
func rowsFromTag(tag string) int {
	fields := strings.Fields(tag)
	if len(fields) < 2 {
		return -1
	}
	count := 0
	for _, r := range fields[len(fields)-1] {
		if r < '0' || r > '9' {
			return -1
		}
		count = count*10 + int(r-'0')
	}
	return count
}
