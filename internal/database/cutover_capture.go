package database

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	captainmigrations "github.com/flanksource/captain/migrations"
	commonsdb "github.com/flanksource/commons-db/db"
)

const (
	legacyCaptainCutoverName          = "captain-legacy-session-v1"
	legacyCaptainReportKey            = "captain-session-cache-v1"
	legacyTranscriptFreshnessSnapshot = "transcript-freshness.json"
)

// CaptureLegacyCaptainCutoverArtifact copies Captain's validated in-database
// rollback archive into a second, private filesystem generation before Gavel
// applies its own HCL. The generation name binds the canonical durable report
// and captured archive schema; exact retries reuse only an artifact whose
// report and snapshots still match the current capture.
func CaptureLegacyCaptainCutoverArtifact(
	ctx context.Context,
	dsn string,
	artifactRoot string,
	report *captainmigrations.LegacySessionCutoverReport,
) (_ *CutoverArtifactResult, resultErr error) {
	if report == nil {
		return nil, errors.New("Captain legacy cutover report is nil")
	}
	if report.CutoverKey != legacyCaptainReportKey {
		return nil, fmt.Errorf("unexpected Captain cutover key %q", report.CutoverKey)
	}
	if report.LegacySessionsTable != "captain_sessions_legacy_v1" {
		return nil, fmt.Errorf("unexpected Captain session rollback table %q", report.LegacySessionsTable)
	}
	if report.LegacyPromptsTable != nil && *report.LegacyPromptsTable != "captain_session_prompts_legacy_v1" {
		return nil, fmt.Errorf("unexpected Captain prompt rollback table %q", *report.LegacyPromptsTable)
	}

	db, err := commonsdb.NewDB(dsn)
	if err != nil {
		return nil, fmt.Errorf("open Captain cutover artifact database: %w", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close Captain cutover artifact database: %w", err))
		}
	}()
	if err := validateCapturedCutoverReport(ctx, db, report); err != nil {
		return nil, err
	}

	sessionRows, err := captureArchiveRows(ctx, db, report.LegacySessionsTable, "path")
	if err != nil {
		return nil, err
	}
	if int64(len(sessionRows)) != report.LegacySessionRows || captureRowsChecksum(sessionRows) != report.LegacySessionsChecksum {
		return nil, errors.New("Captain session rollback archive no longer matches its durable validation report")
	}
	var promptRows []json.RawMessage
	if report.LegacyPromptsTable != nil {
		promptRows, err = captureArchiveRows(ctx, db, *report.LegacyPromptsTable, "session_id")
		if err != nil {
			return nil, err
		}
		if report.LegacyPromptsChecksum == nil || int64(len(promptRows)) != report.LegacyPromptRows ||
			captureRowsChecksum(promptRows) != *report.LegacyPromptsChecksum {
			return nil, errors.New("Captain prompt rollback archive no longer matches its durable validation report")
		}
	}

	schema, err := captureArchiveColumns(ctx, db, report)
	if err != nil {
		return nil, err
	}
	freshness := captureLegacyTranscriptFreshness(sessionRows)
	sessionsJSON, err := json.MarshalIndent(sessionRows, "", "  ")
	if err != nil {
		return nil, err
	}
	promptsJSON, err := json.MarshalIndent(promptRows, "", "  ")
	if err != nil {
		return nil, err
	}
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}
	freshnessJSON, err := json.MarshalIndent(freshness, "", "  ")
	if err != nil {
		return nil, err
	}

	canonicalReport, canonicalReportJSON, err := canonicalCapturedCutoverReport(report)
	if err != nil {
		return nil, err
	}
	generation := captureArtifactGeneration(canonicalReportJSON, schemaJSON)
	request := CutoverArtifactRequest{
		Cutover:    legacyCaptainCutoverName,
		Generation: generation,
		CreatedAt:  report.CompletedAt,
		Metadata: map[string]any{
			"captainCutoverKey": report.CutoverKey,
			"legacySessionRows": report.LegacySessionRows,
			"legacyPromptRows":  report.LegacyPromptRows,
			"transcriptSummary": freshness.Summary,
		},
		Report: canonicalReport,
		Rollback: map[string]any{
			"sessionArchive":        report.LegacySessionsTable,
			"promptArchive":         report.LegacyPromptsTable,
			"legacySessionChecksum": report.LegacySessionsChecksum,
			"legacyPromptChecksum":  report.LegacyPromptsChecksum,
			"nativeSessionChecksum": report.NativeSessionsChecksum,
			"nativePromptChecksum":  report.NativePromptRunsChecksum,
			"guard":                 "stop all writers and prove no post-cutover native writes before restoring table names",
		},
		RollbackInstructions: captureRollbackGuide(report),
		Snapshots: []CutoverArtifactSnapshot{
			{Name: "legacy-sessions.json", ContentType: "application/json", Data: sessionsJSON},
			{Name: "legacy-session-prompts.json", ContentType: "application/json", Data: promptsJSON},
			{Name: "legacy-schema.json", ContentType: "application/json", Data: schemaJSON},
			{Name: legacyTranscriptFreshnessSnapshot, ContentType: "application/json", Data: freshnessJSON},
		},
	}
	root, err := captureArtifactRoot(artifactRoot)
	if err != nil {
		return nil, err
	}
	target := filepath.Join(root, generation)
	if existing, err := reuseCapturedArtifact(target, request); existing != nil || err != nil {
		return existing, err
	}
	artifact, err := WriteCutoverArtifact(target, request)
	if errors.Is(err, ErrCutoverArtifactCommitted) {
		return reuseCapturedArtifact(target, request)
	}
	return artifact, err
}

type capturedArchiveSchema struct {
	Tables []capturedArchiveTable `json:"tables"`
}

type capturedArchiveTable struct {
	Name        string                      `json:"name"`
	Columns     []capturedArchiveColumn     `json:"columns"`
	Constraints []capturedArchiveConstraint `json:"constraints"`
	Indexes     []string                    `json:"indexes"`
}

type capturedArchiveColumn struct {
	Ordinal  int    `json:"ordinal"`
	Name     string `json:"name"`
	DataType string `json:"dataType"`
	Nullable bool   `json:"nullable"`
	Default  string `json:"default,omitempty"`
}

type capturedArchiveConstraint struct {
	Name       string `json:"name"`
	Definition string `json:"definition"`
}

func captureArchiveColumns(ctx context.Context, db *sql.DB, report *captainmigrations.LegacySessionCutoverReport) (capturedArchiveSchema, error) {
	tables := []string{report.LegacySessionsTable}
	if report.LegacyPromptsTable != nil {
		tables = append(tables, *report.LegacyPromptsTable)
	}
	result := capturedArchiveSchema{Tables: make([]capturedArchiveTable, 0, len(tables))}
	for _, table := range tables {
		captured := capturedArchiveTable{Name: table}
		rows, err := db.QueryContext(ctx, `
			SELECT a.attnum, a.attname, pg_catalog.format_type(a.atttypid, a.atttypmod),
			       NOT a.attnotnull, COALESCE(pg_get_expr(d.adbin, d.adrelid), '')
			FROM pg_catalog.pg_attribute a
			JOIN pg_catalog.pg_class c ON c.oid = a.attrelid
			JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
			LEFT JOIN pg_catalog.pg_attrdef d ON d.adrelid = a.attrelid AND d.adnum = a.attnum
			WHERE n.nspname = 'public' AND c.relname = $1
			  AND a.attnum > 0 AND NOT a.attisdropped
			ORDER BY a.attnum`, table)
		if err != nil {
			return result, err
		}
		for rows.Next() {
			var column capturedArchiveColumn
			if err := rows.Scan(&column.Ordinal, &column.Name, &column.DataType, &column.Nullable, &column.Default); err != nil {
				rows.Close()
				return result, err
			}
			captured.Columns = append(captured.Columns, column)
		}
		if err := rows.Close(); err != nil {
			return result, err
		}
		if len(captured.Columns) == 0 {
			return result, fmt.Errorf("Captain rollback table public.%s has no columns", table)
		}
		constraintRows, err := db.QueryContext(ctx, `
			SELECT con.conname, pg_get_constraintdef(con.oid, true)
			FROM pg_catalog.pg_constraint con
			JOIN pg_catalog.pg_class c ON c.oid = con.conrelid
			JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
			WHERE n.nspname = 'public' AND c.relname = $1
			ORDER BY con.conname`, table)
		if err != nil {
			return result, err
		}
		for constraintRows.Next() {
			var constraint capturedArchiveConstraint
			if err := constraintRows.Scan(&constraint.Name, &constraint.Definition); err != nil {
				constraintRows.Close()
				return result, err
			}
			captured.Constraints = append(captured.Constraints, constraint)
		}
		if err := constraintRows.Close(); err != nil {
			return result, err
		}
		indexRows, err := db.QueryContext(ctx, "SELECT indexdef FROM pg_indexes WHERE schemaname = 'public' AND tablename = $1 ORDER BY indexname", table)
		if err != nil {
			return result, err
		}
		for indexRows.Next() {
			var definition string
			if err := indexRows.Scan(&definition); err != nil {
				indexRows.Close()
				return result, err
			}
			captured.Indexes = append(captured.Indexes, definition)
		}
		if err := indexRows.Close(); err != nil {
			return result, err
		}
		result.Tables = append(result.Tables, captured)
	}
	return result, nil
}

type capturedFreshness struct {
	Summary capturedFreshnessSummary `json:"summary"`
	Files   []capturedFreshnessFile  `json:"files"`
}

type capturedFreshnessSummary struct {
	Total    int `json:"total"`
	Fresh    int `json:"fresh"`
	Stale    int `json:"stale"`
	Missing  int `json:"missing"`
	Unproven int `json:"unproven"`
}

type capturedFreshnessFile struct {
	Path          string `json:"path"`
	Status        string `json:"status"`
	ArchivedSize  int64  `json:"archivedSize,omitempty"`
	CurrentSize   int64  `json:"currentSize,omitempty"`
	ArchivedMTime int64  `json:"archivedMtimeUnixNano,omitempty"`
	CurrentMTime  int64  `json:"currentMtimeUnixNano,omitempty"`
	Error         string `json:"error,omitempty"`
}

func captureLegacyTranscriptFreshness(rows []json.RawMessage) capturedFreshness {
	result := capturedFreshness{Files: make([]capturedFreshnessFile, 0, len(rows))}
	for _, raw := range rows {
		decoder := json.NewDecoder(strings.NewReader(string(raw)))
		decoder.UseNumber()
		var data map[string]any
		if err := decoder.Decode(&data); err != nil {
			result.Summary.Unproven++
			result.Files = append(result.Files, capturedFreshnessFile{Status: "unproven", Error: err.Error()})
			continue
		}
		file := capturedFreshnessFile{
			Path:          capturedString(data["path"]),
			ArchivedSize:  capturedInt64(data["size"]),
			ArchivedMTime: capturedInt64(data["mod_unix"]),
		}
		info, err := os.Stat(file.Path)
		switch {
		case file.Path == "" || file.ArchivedMTime <= 0:
			file.Status = "unproven"
			result.Summary.Unproven++
		case errors.Is(err, os.ErrNotExist):
			file.Status = "missing"
			file.Error = err.Error()
			result.Summary.Missing++
		case err != nil:
			file.Status = "unproven"
			file.Error = err.Error()
			result.Summary.Unproven++
		default:
			file.CurrentSize = info.Size()
			file.CurrentMTime = info.ModTime().UnixNano()
			if file.CurrentSize == file.ArchivedSize && file.CurrentMTime == file.ArchivedMTime {
				file.Status = "fresh"
				result.Summary.Fresh++
			} else {
				file.Status = "stale"
				result.Summary.Stale++
			}
		}
		result.Files = append(result.Files, file)
	}
	result.Summary.Total = len(result.Files)
	sort.Slice(result.Files, func(i, j int) bool { return result.Files[i].Path < result.Files[j].Path })
	return result
}

func captureArchiveRows(ctx context.Context, db *sql.DB, table, orderColumn string) ([]json.RawMessage, error) {
	if (table != "captain_sessions_legacy_v1" || orderColumn != "path") &&
		(table != "captain_session_prompts_legacy_v1" || orderColumn != "session_id") {
		return nil, fmt.Errorf("unsupported Captain rollback relation %q", table)
	}
	query := fmt.Sprintf("SELECT to_jsonb(t)::text FROM public.%s t ORDER BY %s", table, orderColumn)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []json.RawMessage
	for rows.Next() {
		var value []byte
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		result = append(result, append(json.RawMessage(nil), value...))
	}
	return result, rows.Err()
}

func validateCapturedCutoverReport(ctx context.Context, db *sql.DB, report *captainmigrations.LegacySessionCutoverReport) error {
	var sessionRows, promptRows, importedSessions, importedPrompts int64
	var sessionTable, sessionChecksum, nativeChecksum string
	var promptTable, promptChecksum, nativePromptChecksum sql.NullString
	var completedAt, updatedAt time.Time
	err := db.QueryRowContext(ctx, "SELECT legacy_sessions_table, legacy_prompts_table, legacy_session_rows, legacy_prompt_rows, imported_session_rows, imported_prompt_run_rows, legacy_sessions_checksum, legacy_prompts_checksum, native_sessions_checksum, native_prompt_runs_checksum, completed_at, updated_at FROM public.captain_legacy_session_cutovers WHERE cutover_key = $1", report.CutoverKey).Scan(
		&sessionTable, &promptTable,
		&sessionRows, &promptRows, &importedSessions, &importedPrompts,
		&sessionChecksum, &promptChecksum, &nativeChecksum, &nativePromptChecksum,
		&completedAt, &updatedAt,
	)
	if err != nil {
		return fmt.Errorf("read durable Captain cutover report: %w", err)
	}
	if sessionTable != report.LegacySessionsTable || !capturedOptionalEqual(promptTable, report.LegacyPromptsTable) ||
		sessionRows != report.LegacySessionRows || promptRows != report.LegacyPromptRows ||
		importedSessions != report.ImportedSessionRows || importedPrompts != report.ImportedPromptRunRows ||
		sessionChecksum != report.LegacySessionsChecksum || nativeChecksum != report.NativeSessionsChecksum ||
		!capturedOptionalEqual(promptChecksum, report.LegacyPromptsChecksum) ||
		!capturedOptionalEqual(nativePromptChecksum, report.NativePromptRunsChecksum) ||
		!completedAt.Equal(report.CompletedAt) || !updatedAt.Equal(report.UpdatedAt) {
		return errors.New("returned Captain cutover report does not match its durable database row")
	}
	return nil
}

func captureRowsChecksum(rows []json.RawMessage) string {
	hash := sha256.New()
	for _, row := range rows {
		hash.Write(row)
		hash.Write([]byte{'\n'})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func captureArtifactRoot(root string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, ".gavel", "migrations", "captain-sessions")
	}
	return filepath.Abs(root)
}

func canonicalCapturedCutoverReport(report *captainmigrations.LegacySessionCutoverReport) (any, []byte, error) {
	encoded, err := json.Marshal(report)
	if err != nil {
		return nil, nil, fmt.Errorf("encode canonical Captain cutover report: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var canonical any
	if err := decoder.Decode(&canonical); err != nil {
		return nil, nil, fmt.Errorf("decode canonical Captain cutover report: %w", err)
	}
	canonicalJSON, err := marshalCutoverArtifactJSON(cutoverReportFile, canonical)
	if err != nil {
		return nil, nil, err
	}
	return canonical, canonicalJSON, nil
}

func captureArtifactGeneration(canonicalReportJSON, schemaJSON []byte) string {
	fingerprint := "report-sha256=" + cutoverSHA256(canonicalReportJSON) + "\n" +
		"legacy-schema-sha256=" + cutoverSHA256(schemaJSON) + "\n"
	return "sha256-" + cutoverSHA256([]byte(fingerprint))
}

func reuseCapturedArtifact(target string, request CutoverArtifactRequest) (*CutoverArtifactResult, error) {
	if _, err := os.Lstat(target); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	manifest, err := VerifyCutoverArtifact(target)
	if err != nil {
		return nil, err
	}
	if manifest.Cutover != request.Cutover || manifest.Generation != request.Generation || manifest.Generation != filepath.Base(target) {
		return nil, fmt.Errorf("existing Captain cutover artifact %s has a mismatched manifest", target)
	}
	expectedReport, err := marshalCutoverArtifactJSON(cutoverReportFile, request.Report)
	if err != nil {
		return nil, err
	}
	actualReport, err := readPrivateCutoverFile(filepath.Join(target, cutoverReportFile))
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(actualReport, expectedReport) {
		return nil, fmt.Errorf("%w: existing Captain cutover artifact report does not match the current durable report", ErrCutoverArtifactInvalid)
	}
	expectedSnapshots := make(map[string]CutoverArtifactSnapshot, len(request.Snapshots))
	for _, snapshot := range request.Snapshots {
		expectedSnapshots[snapshot.Name] = snapshot
	}
	actualSnapshots := make(map[string]CutoverArtifactFile, len(request.Snapshots))
	for _, file := range manifest.Files {
		if file.Role == "snapshot" {
			actualSnapshots[file.Name] = file
		}
	}
	if len(actualSnapshots) != len(expectedSnapshots) {
		return nil, fmt.Errorf("%w: existing Captain cutover artifact snapshots do not match the current capture", ErrCutoverArtifactInvalid)
	}
	for name, expected := range expectedSnapshots {
		actualMetadata, ok := actualSnapshots[name]
		expectedContentType := strings.TrimSpace(expected.ContentType)
		if expectedContentType == "" {
			expectedContentType = "application/octet-stream"
		}
		if !ok || actualMetadata.ContentType != expectedContentType {
			return nil, fmt.Errorf("%w: existing Captain cutover artifact snapshot %s does not match the current capture", ErrCutoverArtifactInvalid, name)
		}
		// Transcript freshness records what was observable when the cutover
		// evidence was committed. Later file changes must not invalidate the
		// immutable database archive/report that makes the retry safe.
		if name == legacyTranscriptFreshnessSnapshot {
			continue
		}
		actual, err := readPrivateCutoverFile(filepath.Join(target, name))
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(actual, expected.Data) {
			return nil, fmt.Errorf("%w: existing Captain cutover artifact snapshot %s does not match the current capture", ErrCutoverArtifactInvalid, name)
		}
	}
	return verifiedCutoverArtifactResult(target)
}

func captureRollbackGuide(report *captainmigrations.LegacySessionCutoverReport) string {
	prompt := "none"
	if report.LegacyPromptsTable != nil {
		prompt = *report.LegacyPromptsTable
	}
	return fmt.Sprintf(
		"# Captain legacy session cutover rollback\n\n"+
			"Rollback tables retained in PostgreSQL:\n\n- sessions: public.%s\n- prompts: public.%s\n\n"+
			"Do not run an automatic table rename. Stop every Gavel and Captain writer, verify report.json and checksums.sha256, "+
			"and prove no native Captain rows or Gavel references were written after completedAt. If native writes exist, restore "+
			"the whole database from a coordinated backup. Otherwise a database operator may restore the archived schema/data "+
			"in one transaction. Retain this artifact and the archive tables for auditability.\n",
		report.LegacySessionsTable, prompt,
	)
}

func capturedOptionalEqual(actual sql.NullString, expected *string) bool {
	if expected == nil {
		return !actual.Valid
	}
	return actual.Valid && actual.String == *expected
}

func capturedString(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func capturedInt64(value any) int64 {
	number, ok := value.(json.Number)
	if !ok {
		return 0
	}
	parsed, _ := number.Int64()
	return parsed
}
