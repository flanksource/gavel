// Package httpx is the one place gavel builds an http.Client. Going through it
// is what makes `record: clients` possible: a client constructed anywhere else
// is invisible to the recorder, because there is no proxy in front of gavel's
// own process to catch it.
package httpx

import (
	"net/http"
	"time"

	"github.com/flanksource/gavel/fixtures/record"
)

// Shared is the default client for gavel's own outbound calls. It has no
// timeout, matching http.DefaultClient, so callers that need one use Client.
var Shared = Client(0)

// Client returns a recordable client with the given timeout. Zero means none.
//
// The transport is wrapped unconditionally: whether anything is recorded is
// decided per request by whether a fixture has a recording open, so a client
// built at init costs one atomic load per request for the runs that record
// nothing.
func Client(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout, Transport: record.MaybeWrap(http.DefaultTransport)}
}
