package record

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"time"
	"unicode/utf8"
)

// defaultBodies is the cap applied when a fixture enables body capture without
// naming a size. Bodies are the most likely place for a large payload and for a
// secret, so the default is deliberately small enough to read.
const defaultBodies = Size(256 * 1024)

// exchangeState is carried on the ProxyCtx between the request and response
// handlers: the request body has already been consumed by the transport by the
// time the response comes back, so it has to be captured on the way out.
type exchangeState struct {
	started time.Time
	post    *PostData
}

// capturePolicy is the part of a recorder's configuration that shapes an entry:
// how much of a body to keep, and which headers to blank on top of the built-in
// denylist. The proxy and the client transport record the same way, so the entry
// builder takes the policy rather than one recorder or the other.
type capturePolicy struct {
	Bodies Size
	Redact []string
}

// connectEntry describes one CONNECT tunnel. There is no method, path, status
// or body to report — the payload stayed encrypted — so the entry says what was
// actually observed and marks itself tunnelled rather than inventing a 200 that
// nobody sent. The status is 200 only because the proxy itself answered
// "200 Connection established".
func connectEntry(host string, started time.Time, elapsed time.Duration, in, out int64, err error) Entry {
	extras := &Extras{Tunnelled: true, BytesIn: in, BytesOut: out}
	status, statusText := http.StatusOK, "Connection established"
	if err != nil {
		extras.Error = err.Error()
		status, statusText = 0, ""
	}

	return Entry{
		StartedDateTime: started.Format(httpTimeFormat),
		Time:            float64(elapsed.Nanoseconds()) / 1e6,
		Request: Request{
			Method:      http.MethodConnect,
			URL:         "https://" + host,
			HTTPVersion: "HTTP/1.1",
			Cookies:     []Cookie{},
			Headers:     []NameValue{},
			QueryString: []NameValue{},
			HeadersSize: -1,
			BodySize:    out,
		},
		Response: Response{
			Status:      status,
			StatusText:  statusText,
			HTTPVersion: "HTTP/1.1",
			Cookies:     []Cookie{},
			Headers:     []NameValue{},
			Content:     Content{Size: in},
			HeadersSize: -1,
			BodySize:    in,
		},
		Timings: unmeasuredTimings(elapsed),
		Gavel:   extras,
		started: started,
	}
}

// exchangeEntry describes one request/response pair seen in the clear: a plain
// HTTP call, or a decrypted one under mitm.
func exchangeEntry(policy capturePolicy, req *http.Request, resp *http.Response, state exchangeState, err error) Entry {
	elapsed := time.Since(state.started)
	entry := Entry{
		StartedDateTime: state.started.Format(httpTimeFormat),
		Time:            float64(elapsed.Nanoseconds()) / 1e6,
		Request: Request{
			Method:      req.Method,
			URL:         req.URL.String(),
			HTTPVersion: req.Proto,
			Cookies:     []Cookie{},
			Headers:     headerList(req.Header, policy.Redact),
			QueryString: queryList(req.URL.RawQuery),
			PostData:    state.post,
			HeadersSize: -1,
			BodySize:    req.ContentLength,
		},
		Timings: unmeasuredTimings(elapsed),
		started: state.started,
	}

	if err != nil {
		entry.Gavel = &Extras{Error: err.Error()}
	}
	if resp == nil {
		entry.Response = Response{Cookies: []Cookie{}, Headers: []NameValue{}, HeadersSize: -1, BodySize: -1}
		return entry
	}

	content, truncated := captureResponseBody(policy, resp)
	entry.Response = Response{
		Status:      resp.StatusCode,
		StatusText:  http.StatusText(resp.StatusCode),
		HTTPVersion: resp.Proto,
		Cookies:     []Cookie{},
		Headers:     headerList(resp.Header, policy.Redact),
		Content:     content,
		RedirectURL: resp.Header.Get("Location"),
		HeadersSize: -1,
		BodySize:    resp.ContentLength,
	}
	if truncated {
		entry.Response.Content.Comment = "truncated by record.http.bodies"
	}
	return entry
}

// captureResponseBody reads up to the configured cap and puts back what it
// read, so the child still receives the whole response.
func captureResponseBody(policy capturePolicy, resp *http.Response) (Content, bool) {
	content := Content{Size: resp.ContentLength, MimeType: resp.Header.Get("Content-Type")}
	if policy.Bodies <= 0 || resp.Body == nil {
		return content, false
	}

	consumed, keep, truncated, err := readCapped(resp.Body, policy.Bodies.Or(defaultBodies))
	if err != nil {
		content.Comment = "body capture failed: " + err.Error()
		return content, false
	}
	resp.Body = replayBody{Reader: io.MultiReader(bytes.NewReader(consumed), resp.Body), Closer: resp.Body}

	if resp.ContentLength < 0 {
		content.Size = int64(len(consumed))
	}
	content.Text, content.Encoding = encodeBody(keep)
	return content, truncated
}

// readCapped reads one byte past the cap so it can tell a body that exactly
// fills the cap from one that overflows it.
//
// It returns two slices deliberately: `consumed` is everything taken off the
// stream and must be replayed in full or the child loses a byte, while `keep`
// is the prefix that goes into the artifact.
func readCapped(reader io.Reader, limit Size) (consumed, keep []byte, truncated bool, err error) {
	consumed, err = io.ReadAll(io.LimitReader(reader, int64(limit)+1))
	if err != nil {
		return nil, nil, false, err
	}
	if int64(len(consumed)) > int64(limit) {
		return consumed, consumed[:limit], true, nil
	}
	return consumed, consumed, false, nil
}

// encodeBody keeps text as text and base64s anything that is not valid UTF-8,
// which is what the HAR spec's `encoding` field is for.
func encodeBody(body []byte) (string, string) {
	if len(body) == 0 {
		return "", ""
	}
	if utf8.Valid(body) {
		return string(body), ""
	}
	return base64.StdEncoding.EncodeToString(body), "base64"
}

// replayBody hands back the bytes already read, then the rest of the original
// stream, while closing the original.
type replayBody struct {
	io.Reader
	io.Closer
}

// capturePostData reads the request body before the transport does and puts it
// back. Returns nil when the body is empty or unreadable — a missing postData
// is valid HAR, a wrong one is not.
func capturePostData(req *http.Request, limit Size) *PostData {
	consumed, keep, truncated, err := readCapped(req.Body, limit)
	if err != nil || len(consumed) == 0 {
		return nil
	}
	req.Body = replayBody{Reader: io.MultiReader(bytes.NewReader(consumed), req.Body), Closer: req.Body}

	text, encoding := encodeBody(keep)
	post := &PostData{MimeType: req.Header.Get("Content-Type"), Params: []NameValue{}, Text: text}
	if encoding == "base64" {
		// HAR has no encoding field on postData, so a binary body is described
		// rather than smuggled in as if it were text.
		post.Text = ""
		post.Comment = "binary body omitted"
	}
	if truncated {
		post.Comment = "truncated by record.http.bodies"
	}
	return post
}
