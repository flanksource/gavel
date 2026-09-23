package ui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
)

// The multiplexed events stream exists because a browser allows only six
// HTTP/1.1 connections per host, shared by every tab. Each dashboard tab used
// to hold one EventSource per topic (PR list, proc status, task runs, …), so a
// second tab or a reload starved ordinary fetches until they queued forever.
// /api/events carries every topic of a tab over one connection:
//
//	GET    /api/events                  → text/event-stream; first frame
//	                                      `event: __hello` /
//	                                      `data: {"conn":"<id>","build":"<ui build id>"}`
//	POST   /api/events/{conn}/subs      {"id","path"} → 204 once the sub runs
//	DELETE /api/events/{conn}/subs/{id} → 204 once the sub's handler stopped
//
// A sub runs the stream handler registered for its path in-process, through
// the server's own top-level handler, and re-emits each of its events as
// `event: <subId>/<name>` on the connection. When the handler ends on its own
// the sub emits `event: <subId>/__closed` / `data: {"status":…}`.

const eventsPingInterval = 15 * time.Second

var eventsSubIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

var errEventsConnClosed = errors.New("events connection closed")

// eventHub tracks the open /api/events connections of one handler tree. mux
// resolves a sub's path to a registered route; root is the full top-level
// handler (middlewares included) the sub is served through; build is the UI
// build id every hello reports (see ui_build.go).
type eventHub struct {
	mux   *http.ServeMux
	root  http.Handler
	build string

	mu    sync.Mutex
	conns map[string]*eventConn
}

// eventConn is one open /api/events response and the subs multiplexed onto it.
type eventConn struct {
	id     string
	ctx    context.Context
	cancel context.CancelFunc

	writeMu sync.Mutex
	w       http.ResponseWriter
	flusher http.Flusher
	closed  bool

	mu         sync.Mutex
	subs       map[string]*eventSub
	subsClosed bool
	wg         sync.WaitGroup
}

type eventSub struct {
	id     string
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

// registerEventRoutes mounts the multiplexed events routes on mux. root must
// be the handler the server actually serves (mux wrapped in its middlewares) so
// subs see exactly what a direct request to the same path would.
func registerEventRoutes(mux *http.ServeMux, root http.Handler, build string) {
	if build == "" {
		panic("ui: registerEventRoutes needs a UI build id")
	}
	hub := &eventHub{mux: mux, root: root, build: build, conns: map[string]*eventConn{}}
	mux.HandleFunc("GET /api/events", hub.handleStream)
	mux.HandleFunc("POST /api/events/{conn}/subs", hub.handleSubscribe)
	mux.HandleFunc("DELETE /api/events/{conn}/subs/{id}", hub.handleUnsubscribe)
}

func (h *eventHub) handleStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx, cancel := context.WithCancel(r.Context())
	conn := &eventConn{id: uuid.NewString(), ctx: ctx, cancel: cancel, w: w, flusher: flusher, subs: map[string]*eventSub{}}
	h.mu.Lock()
	h.conns[conn.id] = conn
	h.mu.Unlock()
	defer h.closeConn(conn)

	hello, err := json.Marshal(eventsHello{Conn: conn.id, Build: h.build})
	if err != nil {
		panic(fmt.Sprintf("marshal events hello: %v", err))
	}
	if conn.write(fmt.Appendf(nil, "event: __hello\ndata: %s\n\n", hello)) != nil {
		return
	}
	ticker := time.NewTicker(eventsPingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if conn.write([]byte(": ping\n\n")) != nil {
				return
			}
		}
	}
}

// closeConn forgets conn, stops its writes, cancels every sub and waits for
// their goroutines: the response writer is invalid once handleStream returns.
func (h *eventHub) closeConn(conn *eventConn) {
	h.mu.Lock()
	delete(h.conns, conn.id)
	h.mu.Unlock()

	conn.writeMu.Lock()
	conn.closed = true
	conn.writeMu.Unlock()

	conn.mu.Lock()
	conn.subsClosed = true
	conn.mu.Unlock()
	conn.cancel()
	conn.wg.Wait()
}

func (h *eventHub) lookup(id string) (*eventConn, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conn, ok := h.conns[id]
	return conn, ok
}

// eventsHello is the payload of the __hello frame. Build lets a page loaded
// from an older bundle notice it is stale and reload.
type eventsHello struct {
	Conn  string `json:"conn"`
	Build string `json:"build"`
}

type eventSubRequest struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

func (h *eventHub) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	conn, ok := h.lookup(r.PathValue("conn"))
	if !ok {
		writeEventsError(w, http.StatusNotFound, fmt.Errorf("unknown events connection %q", r.PathValue("conn")))
		return
	}
	var body eventSubRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		writeEventsError(w, http.StatusBadRequest, fmt.Errorf("decode subscription: %w", err))
		return
	}
	if !eventsSubIDPattern.MatchString(body.ID) {
		writeEventsError(w, http.StatusBadRequest, fmt.Errorf("sub id %q must match %s", body.ID, eventsSubIDPattern))
		return
	}
	req, err := h.subRequest(r, body.Path)
	if err != nil {
		writeEventsError(w, http.StatusBadRequest, err)
		return
	}
	status, err := conn.startSub(body.ID, req, h.root)
	if err != nil {
		writeEventsError(w, status, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *eventHub) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	conn, ok := h.lookup(r.PathValue("conn"))
	if !ok {
		writeEventsError(w, http.StatusNotFound, fmt.Errorf("unknown events connection %q", r.PathValue("conn")))
		return
	}
	id := r.PathValue("id")
	conn.mu.Lock()
	sub, ok := conn.subs[id]
	conn.mu.Unlock()
	if !ok {
		writeEventsError(w, http.StatusNotFound, fmt.Errorf("unknown sub %q on events connection %q", id, conn.id))
		return
	}
	sub.cancel()
	<-sub.done
	w.WriteHeader(http.StatusNoContent)
}

// startSub registers the sub and launches it. The returned status is the HTTP
// status to answer the subscribe request with when err is non-nil.
func (c *eventConn) startSub(id string, req *http.Request, root http.Handler) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.subsClosed {
		return http.StatusNotFound, fmt.Errorf("events connection %q is closed", c.id)
	}
	if _, exists := c.subs[id]; exists {
		return http.StatusConflict, fmt.Errorf("sub %q is already active on events connection %q", id, c.id)
	}
	// The marker exempts the sub request from refuseDirectBrowserStreams: it is
	// the hub-served replacement for the direct stream that guard refuses.
	ctx, cancel := context.WithCancel(withEventsSub(c.ctx))
	sub := &eventSub{id: id, ctx: ctx, cancel: cancel, done: make(chan struct{})}
	c.subs[id] = sub
	c.wg.Add(1)
	go c.runSub(sub, req.WithContext(ctx), root)
	return 0, nil
}

func (c *eventConn) removeSub(sub *eventSub) {
	c.mu.Lock()
	if c.subs[sub.id] == sub {
		delete(c.subs, sub.id)
	}
	c.mu.Unlock()
}

// write sends one complete frame on the connection and flushes it. Every
// writer (hello, pings, every sub) goes through here, so frames never
// interleave. A failed write means the client is gone: the connection is
// cancelled, which tears down its subs.
func (c *eventConn) write(frame []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.closed {
		return errEventsConnClosed
	}
	if _, err := c.w.Write(frame); err != nil {
		c.cancel()
		return fmt.Errorf("write to events connection %q: %w", c.id, err)
	}
	c.flusher.Flush()
	return nil
}

func writeEventsError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	writeJSONError(w, status, err)
}
