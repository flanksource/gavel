package record

// Result is one recording, referenced rather than inlined. It carries the
// counts and digest needed to render a verdict and to write CEL assertions,
// with the payload left on disk under Path.
//
// The indirection is the same rule RunArtifact documents: a FixtureResult is
// persisted into a Captain prompt run's result_json and re-served on every
// dashboard poll, so a full cast or HAR per fixture would make that
// quadratically expensive.
type Result struct {
	Kind Kind `json:"kind"`
	// ID is the artifact's file stem, which GET /api/tests/recording resolves.
	ID string `json:"id"`
	// Path is the absolute location on the host that ran the fixture, which is
	// not necessarily the host rendering the result.
	Path   string `json:"path,omitempty"`
	Format string `json:"format,omitempty"` // asciinema-v2 | har-1.2 | jsonl

	Bytes int64 `json:"bytes,omitempty"`
	// Count is events, entries or statements depending on Kind.
	Count      int   `json:"count,omitempty"`
	Errors     int   `json:"errors,omitempty"`
	DurationMs int64 `json:"duration_ms,omitempty"`
	// Truncated marks an artifact that hit a size or event cap. The fixture's
	// assertions still ran against what was captured.
	Truncated bool `json:"truncated,omitempty"`
	// Error records a recorder that failed. The fixture itself may still have
	// passed, so this is reported alongside the verdict rather than as one.
	Error string `json:"error,omitempty"`
}
