package record

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

func init() { Implemented[KindANSI] = true }

// Cast is one PTY capture: the timed output stream a terminal would have
// rendered, plus what gavel made of it. The package that owns the PTY builds
// this; record only knows how to write and summarise it, which is what keeps
// this package free of any gavel import.
type Cast struct {
	Width, Height int
	Command       []string
	ExitCode      int
	Duration      time.Duration
	Events        []CastEvent
	// Snapshots is the settled screen sampled on an interval — the timeline a
	// viewer scrubs. Empty when the capture ran without screen tracking.
	Snapshots []CastSnapshot
	// Final is the visible viewport at end of stream, and Duplicates the lines
	// it repeats: the tell-tale of a redraw that left stale content behind.
	Final      string
	Duplicates []CastDuplicate
	// Truncated reports that the size caps stopped the event stream early. The
	// screen state is still the true final one — the caps drop recorded bytes,
	// not the terminal emulation that produced them.
	Truncated bool
}

// CastEvent is one chunk of output at an offset from the start of the run.
type CastEvent struct {
	Offset time.Duration
	Data   string
}

// CastSnapshot is the settled screen at an offset.
type CastSnapshot struct {
	Offset time.Duration
	Screen string
}

// CastDuplicate is a line the final screen shows more than once.
type CastDuplicate struct {
	Text  string
	Count int
}

// castHeader is asciinema v2's first line. The `_gavel` object carries what
// asciinema has no field for; players ignore keys they do not know, so the file
// stays playable with `asciinema play` while remaining a complete record.
type castHeader struct {
	Version   int               `json:"version"`
	Width     int               `json:"width"`
	Height    int               `json:"height"`
	Timestamp int64             `json:"timestamp,omitempty"`
	Duration  float64           `json:"duration"`
	Command   string            `json:"command,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Gavel     castExtras        `json:"_gavel"`
}

type castExtras struct {
	ExitCode   int             `json:"exit_code"`
	Snapshots  []CastSnapshot  `json:"snapshots,omitempty"`
	Final      string          `json:"final,omitempty"`
	Duplicates []CastDuplicate `json:"duplicates,omitempty"`
	Truncated  bool            `json:"truncated,omitempty"`
}

// MarshalJSON emits the [t_ms, screen] pair the dashboard scrubs, keeping the
// offsets in the same unit as the header's `duration_ms`.
func (s CastSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{s.Offset.Milliseconds(), s.Screen})
}

// MarshalJSON emits asciinema's canonical [time_seconds, "o", data] triple.
func (e CastEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{e.Offset.Seconds(), "o", e.Data})
}

// MarshalJSON emits the duplicate as an object, since a fixture asserts on it by
// name through the `cast.duplicates` CEL root.
func (d CastDuplicate) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"text": d.Text, "count": d.Count})
}

// WriteCast serialises a capture as asciinema v2: a header object on the first
// line and one event array per line after it. Line-delimited rather than a
// single document so a large cast streams into a player instead of being parsed
// whole.
func WriteCast(w io.Writer, cast Cast) error {
	buffered := bufio.NewWriter(w)
	encoder := json.NewEncoder(buffered)

	header := castHeader{
		Version:  2,
		Width:    cast.Width,
		Height:   cast.Height,
		Duration: cast.Duration.Seconds(),
		// A label a player prints, never something that is re-executed, so the
		// join needs no quoting.
		Command: strings.Join(cast.Command, " "),
		Gavel: castExtras{
			ExitCode:   cast.ExitCode,
			Snapshots:  cast.Snapshots,
			Final:      cast.Final,
			Duplicates: cast.Duplicates,
			Truncated:  cast.Truncated,
		},
	}
	if err := encoder.Encode(header); err != nil {
		return fmt.Errorf("write cast header: %w", err)
	}
	for _, event := range cast.Events {
		if err := encoder.Encode(event); err != nil {
			return fmt.Errorf("write cast event: %w", err)
		}
	}
	return buffered.Flush()
}

// SaveCast writes a capture under store and returns the reference that travels
// on the fixture's result.
func SaveCast(store *Store, label string, cast Cast) (Result, error) {
	file, result, err := store.Create(label, KindANSI)
	if err != nil {
		return Result{}, err
	}
	defer file.Close()

	if err := WriteCast(file, cast); err != nil {
		return result, err
	}
	if info, err := file.Stat(); err == nil {
		result.Bytes = info.Size()
	}
	result.Count = len(cast.Events)
	result.DurationMs = cast.Duration.Milliseconds()
	result.Truncated = cast.Truncated
	return result, nil
}

// CastCELVars builds the `cast` CEL root. As with the http root every key is
// present even on an empty capture, so `cast.has_duplicates == false` is a legal
// assertion rather than an evaluation error.
//
// `snapshots` is the count, not the timeline: a screen is kilobytes and a run is
// hundreds of them, and the artifact is where they belong. `final` is the one
// screen worth asserting on inline.
func CastCELVars(cast Cast, path string) map[string]any {
	duplicates := make([]map[string]any, 0, len(cast.Duplicates))
	for _, duplicate := range cast.Duplicates {
		duplicates = append(duplicates, map[string]any{"text": duplicate.Text, "count": duplicate.Count})
	}
	return map[string]any{
		"events":         len(cast.Events),
		"duration_ms":    cast.Duration.Milliseconds(),
		"exit_code":      cast.ExitCode,
		"width":          cast.Width,
		"height":         cast.Height,
		"final":          cast.Final,
		"snapshots":      len(cast.Snapshots),
		"duplicates":     duplicates,
		"has_duplicates": len(duplicates) > 0,
		"truncated":      cast.Truncated,
		"path":           path,
	}
}
