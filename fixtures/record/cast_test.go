package record

import (
	"bufio"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleCast() Cast {
	return Cast{
		Width:    80,
		Height:   24,
		Command:  []string{"/bin/sh", "-c", "echo hi"},
		ExitCode: 3,
		Duration: 1500 * time.Millisecond,
		Events: []CastEvent{
			{Offset: 10 * time.Millisecond, Data: "hi\r\n"},
			{Offset: 1200 * time.Millisecond, Data: "bye\r\n"},
		},
		Snapshots:  []CastSnapshot{{Offset: 500 * time.Millisecond, Screen: "hi"}},
		Final:      "hi\nbye",
		Duplicates: []CastDuplicate{{Text: "hi", Count: 2}},
	}
}

// asciinema v2 is line-delimited: a header object, then one array per event.
// Anything else stops `asciinema play` from opening the artifact at all.
func TestWriteCastIsLineDelimitedAsciinemaV2(t *testing.T) {
	var out strings.Builder
	require.NoError(t, WriteCast(&out, sampleCast()))

	scanner := bufio.NewScanner(strings.NewReader(out.String()))
	require.True(t, scanner.Scan(), "expected a header line")

	var header struct {
		Version  int     `json:"version"`
		Width    int     `json:"width"`
		Height   int     `json:"height"`
		Duration float64 `json:"duration"`
		Command  string  `json:"command"`
		Gavel    struct {
			ExitCode   int    `json:"exit_code"`
			Final      string `json:"final"`
			Duplicates []struct {
				Text  string `json:"text"`
				Count int    `json:"count"`
			} `json:"duplicates"`
			Snapshots [][]json.RawMessage `json:"snapshots"`
		} `json:"_gavel"`
	}
	require.NoError(t, json.Unmarshal(scanner.Bytes(), &header))
	assert.Equal(t, 2, header.Version)
	assert.Equal(t, 80, header.Width)
	assert.Equal(t, 24, header.Height)
	assert.InDelta(t, 1.5, header.Duration, 0.001)
	assert.Equal(t, "/bin/sh -c echo hi", header.Command)
	// The exit code has no home in the asciinema header, and losing it would
	// mean the artifact could not say whether the run failed.
	assert.Equal(t, 3, header.Gavel.ExitCode)
	assert.Equal(t, "hi\nbye", header.Gavel.Final)
	require.Len(t, header.Gavel.Duplicates, 1)
	assert.Equal(t, 2, header.Gavel.Duplicates[0].Count)
	require.Len(t, header.Gavel.Snapshots, 1)

	var events [][]json.RawMessage
	for scanner.Scan() {
		var event []json.RawMessage
		require.NoError(t, json.Unmarshal(scanner.Bytes(), &event), "each line after the header is an event array")
		events = append(events, event)
	}
	require.NoError(t, scanner.Err())
	require.Len(t, events, 2)
	assert.JSONEq(t, `0.01`, string(events[0][0]), "offsets are seconds, as the format requires")
	assert.JSONEq(t, `"o"`, string(events[0][1]))
	assert.JSONEq(t, `"hi\r\n"`, string(events[0][2]))
}

func TestCastCELVarsSummarisesTheCapture(t *testing.T) {
	vars := CastCELVars(sampleCast(), "/tmp/rec.cast.json")

	assert.Equal(t, 2, vars["events"])
	assert.Equal(t, int64(1500), vars["duration_ms"])
	assert.Equal(t, 3, vars["exit_code"])
	assert.Equal(t, "hi\nbye", vars["final"])
	// A count rather than the timeline: a screen is kilobytes and a run holds
	// hundreds of them, so the scrubbable version stays in the artifact.
	assert.Equal(t, 1, vars["snapshots"])
	assert.Equal(t, true, vars["has_duplicates"])
	assert.Equal(t, "/tmp/rec.cast.json", vars["path"])
}

// Every key is present on an empty capture too, so `cast.has_duplicates == false`
// is an assertion rather than an evaluation error.
func TestCastCELVarsAreCompleteWhenNothingWasCaptured(t *testing.T) {
	vars := CastCELVars(Cast{}, "")

	for _, key := range []string{"events", "duration_ms", "exit_code", "width", "height",
		"final", "snapshots", "duplicates", "has_duplicates", "truncated", "path"} {
		assert.Contains(t, vars, key)
	}
	assert.Equal(t, false, vars["has_duplicates"])
	assert.Empty(t, vars["duplicates"])
}
