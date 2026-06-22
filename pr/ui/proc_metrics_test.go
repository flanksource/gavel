package ui

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/flanksource/clicky/metrics"
	"github.com/flanksource/gavel/procfile"
)

func TestProcRunKey(t *testing.T) {
	started := time.Date(2026, 6, 19, 10, 15, 39, 291808000, time.FixedZone("EEST", 3*3600))
	cases := []struct {
		name    string
		project string
		ps      procfile.ProcState
		want    string
	}{
		{
			name:    "running process with started + pid",
			project: "Clicky UI",
			ps:      procfile.ProcState{Name: "storybook", PID: 7338, Started: &started},
			want:    "Clicky UI/storybook/2026-06-19T10:15:39.291808+03:00/7338",
		},
		{
			name:    "stopped process has no started and zero pid",
			project: "sink",
			ps:      procfile.ProcState{Name: "dev"},
			want:    "sink/dev/not-started/0",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := procRunKey(tc.project, tc.ps); got != tc.want {
				t.Fatalf("procRunKey(%q, %+v) = %q, want %q", tc.project, tc.ps, got, tc.want)
			}
		})
	}
}

// TestProcRunKeyMatchesWireStarted pins the one fragile assumption: the backend
// rebuilds the metric key from ProcState.Started with RFC3339Nano, which must
// equal the `started` string the browser receives (Go's time.Time JSON form)
// and feeds into the frontend's runKey. If these ever diverge, the gauge would
// poll an id the sampler never recorded.
func TestProcRunKeyMatchesWireStarted(t *testing.T) {
	started := time.Date(2026, 6, 19, 10, 15, 39, 291808000, time.FixedZone("EEST", 3*3600))
	ps := procfile.ProcState{Name: "storybook", PID: 7338, Started: &started}

	raw, err := json.Marshal(ps)
	if err != nil {
		t.Fatalf("marshal ProcState: %v", err)
	}
	var wire struct {
		Started string `json:"started"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatalf("unmarshal ProcState: %v", err)
	}

	wantKey := "p/storybook/" + wire.Started + "/7338"
	if got := procRunKey("p", ps); got != wantKey {
		t.Fatalf("procRunKey = %q, want %q (wire started %q)", got, wantKey, wire.Started)
	}
}

// TestProcMetricsEndpoint exercises the full request path the gauges use: a
// recorded series with a slash- and space-bearing id is requested as one
// URL-encoded segment and must round-trip through the mounted clicky handler's
// {id} wildcard back to the recorded points. This pins the encoded-segment
// assumption the whole fix rests on (a workspace name like "Clicky UI").
func TestProcMetricsEndpoint(t *testing.T) {
	s := &Server{procMetrics: metrics.NewMemory(metrics.MemoryConfig{})}
	handler := s.Handler()

	id := "Clicky UI/storybook/2026-06-19T10:15:39.291808+03:00/7338/cpu"
	// Record at now so the point falls inside the handler's ?since=10m window.
	if err := s.procMetrics.Record(metrics.RecordRequest{ID: id, At: time.Now(), Value: 42}); err != nil {
		t.Fatalf("record: %v", err)
	}

	// encodeURIComponent for a path segment: percent-encode everything (space ->
	// %20, etc.) including the slashes that separate the runKey parts.
	encoded := strings.ReplaceAll(url.PathEscape(id), "/", "%2F")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/api/proc/metrics/"+encoded+"?since=10m", nil))

	if rec.Code != 200 {
		t.Fatalf("status = %d, body %q", rec.Code, rec.Body.String())
	}
	var resp struct {
		ID     string `json:"id"`
		Points []struct {
			Value float64 `json:"value"`
		} `json:"points"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}
	if resp.ID != id {
		t.Fatalf("response id = %q, want %q (encoded segment did not decode back)", resp.ID, id)
	}
	if len(resp.Points) != 1 || resp.Points[0].Value != 42 {
		t.Fatalf("points = %+v, want one point of value 42", resp.Points)
	}
}
