package record

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func exchange(method, url string, status int, mime string) Entry {
	return Entry{
		Request:  Request{Method: method, URL: url},
		Response: Response{Status: status, Content: Content{MimeType: mime}},
		Time:     12.5,
	}
}

func TestHTTPCELVarsAggregates(t *testing.T) {
	entries := []Entry{
		exchange(http.MethodGet, "http://api.example.com/users?page=2", 200, "application/json"),
		exchange(http.MethodGet, "http://api.example.com/users/7", 404, "application/json"),
		exchange(http.MethodPost, "http://uploads.example.com/blobs", 201, "application/octet-stream"),
	}

	vars := HTTPCELVars(entries, ".gavel/recordings/rec-x.har")

	assert.Equal(t, 3, vars["entries"])
	assert.Equal(t, []string{"api.example.com", "uploads.example.com"}, vars["hosts"])
	assert.Equal(t, map[string]int{"GET": 2, "POST": 1}, vars["methods"])
	assert.Equal(t, map[string]int{"200": 1, "201": 1, "404": 1}, vars["statuses"])
	// The 404 is the only failure: 200 and 201 are successes and no entry
	// carries a transport error.
	assert.Equal(t, 1, vars["errors"])
	assert.Equal(t, ".gavel/recordings/rec-x.har", vars["path"])

	requests, ok := vars["requests"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, requests, 3)
	assert.Equal(t, "/users", requests[0]["path"])
	assert.Equal(t, "api.example.com", requests[0]["host"])
	assert.Equal(t, 200, requests[0]["status"])
	assert.Equal(t, "application/json", requests[0]["mime"])
}

func TestHTTPCELVarsCountsTunnelBytesAndErrors(t *testing.T) {
	started := time.Now()
	entries := []Entry{
		connectEntry("api.example.com:443", started, 40*time.Millisecond, 900, 120, nil),
		connectEntry("down.example.com:443", started, time.Millisecond, 0, 0, assert.AnError),
	}

	vars := HTTPCELVars(entries, "")

	assert.Equal(t, int64(900), vars["bytes_in"])
	assert.Equal(t, int64(120), vars["bytes_out"])
	// One transport failure. The successful tunnel reports status 200 and the
	// failed one status 0, so neither is counted twice by the >= 400 rule.
	assert.Equal(t, 1, vars["errors"])

	requests := vars["requests"].([]map[string]any)
	assert.Equal(t, true, requests[0]["tunnelled"])
	assert.Equal(t, "", requests[0]["error"])
	assert.Equal(t, assert.AnError.Error(), requests[1]["error"])
}

func TestHTTPCELVarsOnEmptyRecordingIsStillAssertable(t *testing.T) {
	vars := HTTPCELVars(nil, "")

	// Every key present, so `http.errors == 0` is a legal assertion for a
	// fixture that made no calls rather than a CEL "no such key" error.
	for _, key := range []string{"entries", "hosts", "methods", "statuses", "errors", "bytes_in", "bytes_out", "requests", "path"} {
		assert.Contains(t, vars, key)
	}
	assert.Equal(t, 0, vars["entries"])
	assert.Empty(t, vars["requests"])
}

func TestHTTPCELVarsCapsRequestDetailButNotCounts(t *testing.T) {
	entries := make([]Entry, celRequestCap+25)
	for i := range entries {
		entries[i] = exchange(http.MethodGet, "http://api.example.com/", 200, "")
	}

	vars := HTTPCELVars(entries, "")

	assert.Equal(t, celRequestCap+25, vars["entries"], "aggregate count stays exact past the cap")
	assert.Len(t, vars["requests"], celRequestCap)
}

func TestHeaderMapKeepsFirstValueOfRepeatedHeader(t *testing.T) {
	got := headerMap([]NameValue{
		{Name: "Set-Cookie", Value: "a=1"},
		{Name: "Set-Cookie", Value: "b=2"},
		{Name: "Content-Type", Value: "text/plain"},
	})

	assert.Equal(t, map[string]string{"Set-Cookie": "a=1", "Content-Type": "text/plain"}, got)
}
