package record

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteHARShape(t *testing.T) {
	started := time.Date(2026, 8, 2, 10, 30, 0, 500*int(time.Millisecond), time.UTC)
	entry := connectEntry("api.github.com:443", started, 250*time.Millisecond, 4096, 512, nil)

	var buffer bytes.Buffer
	require.NoError(t, WriteHAR(&buffer, []Entry{entry}))

	var har HAR
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &har))

	assert.Equal(t, "1.2", har.Log.Version)
	assert.Equal(t, "gavel", har.Log.Creator.Name)
	require.Len(t, har.Log.Entries, 1)

	got := har.Log.Entries[0]
	assert.Equal(t, "2026-08-02T10:30:00.500Z", got.StartedDateTime)
	assert.InDelta(t, 250.0, got.Time, 0.001)
	assert.Equal(t, http.MethodConnect, got.Request.Method)
	assert.Equal(t, "https://api.github.com:443", got.Request.URL)
	assert.Equal(t, http.StatusOK, got.Response.Status)

	require.NotNil(t, got.Gavel)
	assert.True(t, got.Gavel.Tunnelled, "a CONNECT entry must say the payload was never decrypted")
	assert.Equal(t, int64(4096), got.Gavel.BytesIn)
	assert.Equal(t, int64(512), got.Gavel.BytesOut)

	assert.InDelta(t, 250.0, got.Timings.Receive, 0.001)
	assert.InDelta(t, -1.0, got.Timings.Connect, 0.001, "an unmeasured phase must be -1, not a made-up 0")
}

func TestWriteHAREmptyEntriesIsStillValid(t *testing.T) {
	var buffer bytes.Buffer
	require.NoError(t, WriteHAR(&buffer, nil))

	var raw map[string]any
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &raw))
	log, ok := raw["log"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, []any{}, log["entries"], "entries must serialise as [] rather than null")
}

func TestConnectEntryRecordsDialFailure(t *testing.T) {
	entry := connectEntry("unreachable.invalid:443", time.Now(), time.Millisecond, 0, 0, assert.AnError)

	require.NotNil(t, entry.Gavel)
	assert.Equal(t, assert.AnError.Error(), entry.Gavel.Error)
	assert.Zero(t, entry.Response.Status, "a tunnel that never opened must not claim a 200")
}

func TestHeaderListRedactsSecrets(t *testing.T) {
	header := http.Header{}
	header.Set("Authorization", "Bearer sk-live-abcdef")
	header.Set("Cookie", "session=abc")
	header.Set("X-My-Token", "shhh")
	header.Set("Accept", "application/json")

	list := headerList(header, []string{"X-My-Token"})

	byName := map[string]string{}
	for _, pair := range list {
		byName[pair.Name] = pair.Value
	}
	assert.Equal(t, redactedValue, byName["Authorization"], "a bearer token must never reach disk")
	assert.Equal(t, redactedValue, byName["Cookie"])
	assert.Equal(t, redactedValue, byName["X-My-Token"], "`redact:` names must be honoured on top of the built-in list")
	assert.Equal(t, "application/json", byName["Accept"], "an innocuous header must survive")
}

func TestHeaderListIsSortedForStableDiffs(t *testing.T) {
	header := http.Header{}
	header.Set("Zeta", "1")
	header.Set("Alpha", "2")
	header.Add("Alpha", "1")

	assert.Equal(t, []NameValue{
		{Name: "Alpha", Value: "1"},
		{Name: "Alpha", Value: "2"},
		{Name: "Zeta", Value: "1"},
	}, headerList(header, nil))
}

func TestQueryListSplitsPairs(t *testing.T) {
	assert.Equal(t, []NameValue{
		{Name: "page", Value: "2"},
		{Name: "q", Value: "flanksource"},
	}, queryList("page=2&q=flanksource"))
	assert.Empty(t, queryList(""))
}
