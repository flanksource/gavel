package ui

import (
	"context"
	"net/http"
	"slices"
	"strings"
)

// directStreamRefusal is the body of the 410 a browser gets for a direct
// EventSource on a stream route.
const directStreamRefusal = "stream routes are served through /api/events; reload the page"

// eventsSubKey marks a request the events hub dispatches for a sub. It lives in
// the request context, which a client cannot set, so no header can forge it.
type eventsSubKey struct{}

func withEventsSub(ctx context.Context) context.Context {
	return context.WithValue(ctx, eventsSubKey{}, true)
}

func isEventsSub(ctx context.Context) bool {
	marked, _ := ctx.Value(eventsSubKey{}).(bool)
	return marked
}

// refuseDirectBrowserStreams answers 410 Gone to a browser opening a native
// EventSource on a stream route the events hub can serve instead. A page loaded
// from a bundle older than /api/events holds one such connection per topic and,
// under the browser's six-connections-per-host cap, starves every ordinary
// request of the tab and its siblings. EventSource treats a non-200 response as
// fatal and does not reconnect, so the stale page releases the connection.
//
// Only browsers are refused — they alone send Sec-Fetch-Mode, so CLI and Go
// clients keep direct streams — and only GETs, so POST launch streams are left
// alone. Paths the hub cannot subscribe to (outside /api/, e.g. the /results/
// page's embedded test runner) have no replacement and stay served.
func refuseDirectBrowserStreams(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet &&
			r.Header.Get("Sec-Fetch-Mode") != "" &&
			slices.ContainsFunc(r.Header.Values("Accept"), func(v string) bool { return strings.Contains(v, "text/event-stream") }) &&
			isEventsSubscribablePath(r.URL.Path) &&
			!isEventsSub(r.Context()) {
			w.Header().Set("Cache-Control", "no-store")
			http.Error(w, directStreamRefusal, http.StatusGone)
			return
		}
		next.ServeHTTP(w, r)
	})
}
