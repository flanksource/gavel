package activity

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecorder_RecordAndSnapshot(t *testing.T) {
	r := New()
	r.Record(Entry{Kind: KindREST, URL: "/repos/x/y", StatusCode: 200, Duration: 100 * time.Millisecond, SizeBytes: 1024, FromCache: false})
	r.Record(Entry{Kind: KindREST, URL: "/repos/x/y", StatusCode: 304, Duration: 20 * time.Millisecond, SizeBytes: 0, FromCache: true})
	r.Record(Entry{Kind: KindGraphQL, URL: "/graphql", StatusCode: 200, Duration: 200 * time.Millisecond, SizeBytes: 2048, FromCache: false})

	entries, stats := r.Snapshot(0)

	assert.Len(t, entries, 3)
	// Newest-first ordering
	assert.Equal(t, KindGraphQL, entries[0].Kind)
	assert.Equal(t, KindREST, entries[2].Kind)

	assert.Equal(t, int64(3), stats.Total)
	assert.Equal(t, int64(1), stats.CacheHits)
	assert.Equal(t, int64(0), stats.Errors)
	assert.Equal(t, int64(3072), stats.TotalBytes)
	assert.Equal(t, int64(320*time.Millisecond), stats.TotalNs)

	assert.Equal(t, int64(2), stats.ByKind[KindREST].Total)
	assert.Equal(t, int64(1), stats.ByKind[KindREST].CacheHits)
	assert.Equal(t, int64(1), stats.ByKind[KindGraphQL].Total)
}

func TestRecorder_ErrorCounted(t *testing.T) {
	r := New()
	r.Record(Entry{Kind: KindREST, StatusCode: 500})
	r.Record(Entry{Kind: KindREST, StatusCode: 200, Error: "context deadline exceeded"})
	r.Record(Entry{Kind: KindREST, StatusCode: 200})

	_, stats := r.Snapshot(0)
	assert.Equal(t, int64(3), stats.Total)
	assert.Equal(t, int64(2), stats.Errors)
}

func TestRecorder_RingBufferWraps(t *testing.T) {
	r := New()
	// Fill past capacity to force wrap.
	overflow := ringCapacity + 50
	for i := range overflow {
		r.Record(Entry{Kind: KindREST, URL: fmt.Sprintf("/req/%d", i), StatusCode: 200})
	}

	entries, stats := r.Snapshot(0)
	assert.Len(t, entries, ringCapacity, "ring should be capped")
	assert.Equal(t, int64(overflow), stats.Total, "stats should reflect every Record call, not just retained entries")

	// Newest entry should be the last URL we wrote.
	assert.Equal(t, fmt.Sprintf("/req/%d", overflow-1), entries[0].URL)
	// Oldest retained should be (overflow - ringCapacity).
	assert.Equal(t, fmt.Sprintf("/req/%d", overflow-ringCapacity), entries[ringCapacity-1].URL)
}

func TestRecorder_SnapshotLimit(t *testing.T) {
	r := New()
	for i := range 10 {
		r.Record(Entry{Kind: KindREST, URL: fmt.Sprintf("/req/%d", i), StatusCode: 200})
	}
	entries, _ := r.Snapshot(3)
	assert.Len(t, entries, 3)
	assert.Equal(t, "/req/9", entries[0].URL)
	assert.Equal(t, "/req/7", entries[2].URL)
}

func TestRecorder_Reset(t *testing.T) {
	r := New()
	r.Record(Entry{Kind: KindREST, StatusCode: 200, SizeBytes: 100})
	r.Reset()

	entries, stats := r.Snapshot(0)
	assert.Empty(t, entries)
	assert.Equal(t, int64(0), stats.Total)
	assert.Equal(t, int64(0), stats.TotalBytes)
	assert.Empty(t, stats.ByKind)
}

func TestRecorder_ConcurrentRecord(t *testing.T) {
	r := New()
	const goroutines = 50
	const perGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for i := range perGoroutine {
				r.Record(Entry{Kind: KindREST, URL: fmt.Sprintf("/g%d/%d", id, i), StatusCode: 200, SizeBytes: 10})
			}
		}(g)
	}
	wg.Wait()

	_, stats := r.Snapshot(0)
	assert.Equal(t, int64(goroutines*perGoroutine), stats.Total)
	assert.Equal(t, int64(goroutines*perGoroutine*10), stats.TotalBytes)
}

func TestScrubURL(t *testing.T) {
	tests := map[string]string{
		"":                                            "",
		"https://api.github.com/repos/x/y":            "https://api.github.com/repos/x/y",
		"https://api.github.com/x?access_token=abc":   "https://api.github.com/x?access_token=REDACTED",
		"https://api.github.com/x?token=abc&page=2":   "https://api.github.com/x?page=2&token=REDACTED",
		"https://api.github.com/x?client_secret=zzz":  "https://api.github.com/x?client_secret=REDACTED",
	}
	for in, want := range tests {
		assert.Equal(t, want, scrubURL(in), "input=%q", in)
	}
}

func TestShared_Singleton(t *testing.T) {
	a := Shared()
	b := Shared()
	assert.Same(t, a, b)
}
