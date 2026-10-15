---
cwd: $GIT_ROOT_DIR
timeout: 10m
traces:
  - kind: http_proxy
    params:
      paths: ["/api/captain/sessions/*", "/api/events", "/api/events/*/subs"]
      bodies: 1MiB
    require: count > 0
---

# Session inspector tails a live session

The session inspector on `/sessions/<id>` must only ask the backend for the tail of a live captain session. Its first read is windowed (`?tail=`), the follow stream resumes after the newest message it holds (`?follow=1&after=`), and scrolling up pages older history (`?before=`). It never re-reads the whole transcript.

`hack/session_tail_e2e.go` starts `gavel serve --dev` on a scratch Postgres, seeds a running session of 300 messages (`m-000` … `m-299`), and runs `pr/ui/e2e/session-tail.spec.ts`. The spec appends 5 live messages (`m-300` … `m-304`) and scrolls back to the first one. Chromium sends its requests through the recording proxy, and the recording keeps only the session's traffic:

- `GET /api/captain/sessions/<id>` is a direct read of the session.
- `POST /api/events/<conn>/subs` carries the follow stream's URL in its body, because pr/ui multiplexes every SSE stream over one `/api/events` connection.
- `GET /api/events` is that connection. Each of the session's frames is `event: s<N>/entry` with a message as its data.

A whole-transcript read of 300 messages is about 28 KB, so each read is held to under 12 KB, which fits one 100-message page.

## tails the session and pages history

```bash
go run hack/session_tail_e2e.go -gavel .bin/gavel -messages 300
```

The first read is the newest page: one `?tail=100` read returning 100 of the 300 messages.

* cel: http_proxy.requests.filter(r, r.method == "GET" && r.path == "/api/captain/sessions/0199f0aa-7a11-7000-8000-00000000a11e" && "tail" in r.query).size() == 1
* cel: http_proxy.requests.filter(r, r.method == "GET" && r.path == "/api/captain/sessions/0199f0aa-7a11-7000-8000-00000000a11e" && "tail" in r.query).all(r, r.query.tail == "100" && r.response_json.messages.size() == 100 && r.response_json.window.messages == 300 && r.response_size < 12000)

Every read is windowed, and scrolling up fetched at least the two pages of history before the tail, each at most one page of messages and bytes.

* cel: http_proxy.requests.filter(r, r.method == "GET" && r.path == "/api/captain/sessions/0199f0aa-7a11-7000-8000-00000000a11e").all(r, "tail" in r.query || "before" in r.query || ("follow" in r.query && "after" in r.query))
* cel: http_proxy.requests.filter(r, r.method == "GET" && r.path == "/api/captain/sessions/0199f0aa-7a11-7000-8000-00000000a11e" && "before" in r.query).size() >= 2
* cel: http_proxy.requests.filter(r, r.method == "GET" && r.path == "/api/captain/sessions/0199f0aa-7a11-7000-8000-00000000a11e" && "before" in r.query).all(r, r.query.limit == "100" && r.response_json.messages.size() <= 100 && r.response_size < 12000)

The follow stream resumes after the newest message the tail returned, so the hub connection carries exactly the 5 appended messages and none of the 300 already loaded. The hub stream is captured whole (under the 1 MiB cap), so the count is not cut short.

* cel: http_proxy.requests.filter(r, r.method == "POST" && r.body.contains("/api/captain/sessions/0199f0aa-7a11-7000-8000-00000000a11e")).size() > 0
* cel: http_proxy.requests.filter(r, r.method == "POST" && r.body.contains("/api/captain/sessions/0199f0aa-7a11-7000-8000-00000000a11e")).all(r, r.body.contains("after=m-299"))
* cel: http_proxy.requests.filter(r, r.method == "GET" && r.path == "/api/events").size() > 0
* cel: http_proxy.requests.filter(r, r.method == "GET" && r.path == "/api/events").all(r, r.response_size < 1048576)
* cel: http_proxy.requests.filter(r, r.method == "GET" && r.path == "/api/events").all(r, regexp.FindAll([r'event: s[0-9]+/entry\ndata: \{"id":"m-[0-9]+"', r.response_body]).size() == 5)
