package record

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// HAR 1.2 is hand-rolled rather than imported. The only Go implementation worth
// borrowing (martian/v3/har) would pull martian, martian/log, messageview and
// proxyutil into gavel for what is, at this level of use, a struct definition.
//
// The shape follows the spec closely enough that Chrome DevTools, Firefox and
// har-viewer all open the output; fields gavel has nothing to say about are
// emitted as their spec-mandated empty values rather than omitted, because some
// viewers reject a missing key where they tolerate an empty one.

const harVersion = "1.2"

// HAR is the root object written to a .har file.
type HAR struct {
	Log Log `json:"log"`
}

type Log struct {
	Version string  `json:"version"`
	Creator Creator `json:"creator"`
	Entries []Entry `json:"entries"`
}

type Creator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Entry is one recorded exchange. In connect mode it is one CONNECT tunnel; in
// mitm mode it is one decrypted request/response pair.
type Entry struct {
	StartedDateTime string   `json:"startedDateTime"`
	Time            float64  `json:"time"`
	Request         Request  `json:"request"`
	Response        Response `json:"response"`
	Cache           struct{} `json:"cache"`
	Timings         Timings  `json:"timings"`
	ServerIPAddress string   `json:"serverIPAddress,omitempty"`

	// Gavel carries what the HAR spec has no field for. The spec reserves
	// `_`-prefixed keys for exactly this, and viewers ignore them.
	Gavel *Extras `json:"_gavel,omitempty"`

	// started is the wall clock of the request, kept unformatted so entries can
	// be attributed to a fixture by time slice without reparsing the string.
	started time.Time
}

// Extras are gavel's own annotations on an entry.
type Extras struct {
	// Tunnelled marks an entry recorded in connect mode: the proxy saw the
	// CONNECT and counted its bytes, but the payload stayed encrypted, so there
	// is no method, path, status or body to report.
	Tunnelled bool  `json:"tunnelled,omitempty"`
	BytesIn   int64 `json:"bytes_in,omitempty"`
	BytesOut  int64 `json:"bytes_out,omitempty"`
	// Error is the transport failure, when the exchange never completed.
	Error string `json:"error,omitempty"`
}

type Request struct {
	Method      string      `json:"method"`
	URL         string      `json:"url"`
	HTTPVersion string      `json:"httpVersion"`
	Cookies     []Cookie    `json:"cookies"`
	Headers     []NameValue `json:"headers"`
	QueryString []NameValue `json:"queryString"`
	PostData    *PostData   `json:"postData,omitempty"`
	HeadersSize int64       `json:"headersSize"`
	BodySize    int64       `json:"bodySize"`
}

type Response struct {
	Status      int         `json:"status"`
	StatusText  string      `json:"statusText"`
	HTTPVersion string      `json:"httpVersion"`
	Cookies     []Cookie    `json:"cookies"`
	Headers     []NameValue `json:"headers"`
	Content     Content     `json:"content"`
	RedirectURL string      `json:"redirectURL"`
	HeadersSize int64       `json:"headersSize"`
	BodySize    int64       `json:"bodySize"`
}

type Content struct {
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
	// Encoding is "base64" when Text is not valid UTF-8 text.
	Encoding string `json:"encoding,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

type PostData struct {
	MimeType string      `json:"mimeType"`
	Params   []NameValue `json:"params"`
	Text     string      `json:"text,omitempty"`
	Comment  string      `json:"comment,omitempty"`
}

type NameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Timings is the HAR breakdown. gavel measures the exchange end to end and
// reports the rest as -1, the spec's "not applicable" value, rather than
// inventing a split it did not observe.
type Timings struct {
	Blocked float64 `json:"blocked"`
	DNS     float64 `json:"dns"`
	Connect float64 `json:"connect"`
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
	SSL     float64 `json:"ssl"`
}

// unmeasuredTimings reports only what was actually observed: the total, as
// receive time, and -1 everywhere else.
func unmeasuredTimings(total time.Duration) Timings {
	return Timings{
		Blocked: -1, DNS: -1, Connect: -1, Send: -1, Wait: -1, SSL: -1,
		Receive: float64(total.Nanoseconds()) / 1e6,
	}
}

// redactedHeaders are blanked out on every artifact, not merely hidden from the
// CEL variables. connection.SetupConnection deliberately injects cloud
// credentials into a fixture's environment, so a MITM HAR of an AWS or GitHub
// call would otherwise contain a usable signed credential on disk. There is no
// option to turn this off.
var redactedHeaders = map[string]bool{
	"authorization":       true,
	"proxy-authorization": true,
	"cookie":              true,
	"set-cookie":          true,
	"x-api-key":           true,
	"x-auth-token":        true,
}

const redactedValue = "[redacted]"

// headerList converts Go headers into HAR's ordered name/value pairs, blanking
// the denylist plus any extra names the fixture asked for. Sorted so a golden
// comparison of two runs is stable.
func headerList(header http.Header, extra []string) []NameValue {
	deny := map[string]bool{}
	for _, name := range extra {
		deny[strings.ToLower(strings.TrimSpace(name))] = true
	}

	var out []NameValue
	for name, values := range header {
		lower := strings.ToLower(name)
		for _, value := range values {
			if redactedHeaders[lower] || deny[lower] {
				value = redactedValue
			}
			out = append(out, NameValue{Name: name, Value: value})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Value < out[j].Value
	})
	return out
}

// queryList splits a URL's query into HAR's name/value pairs. Query values are
// never redacted by name: a token in a query string is already in the URL, which
// the entry records in full.
func queryList(rawQuery string) []NameValue {
	var out []NameValue
	for _, pair := range strings.Split(rawQuery, "&") {
		if pair == "" {
			continue
		}
		name, value, _ := strings.Cut(pair, "=")
		out = append(out, NameValue{Name: name, Value: value})
	}
	return out
}

// WriteHAR serialises entries as a HAR 1.2 document. Indented because these are
// read by humans and diffed in pull requests as often as they are opened in
// devtools.
func WriteHAR(w io.Writer, entries []Entry) error {
	if entries == nil {
		entries = []Entry{}
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(HAR{Log: Log{
		Version: harVersion,
		Creator: Creator{Name: "gavel", Version: harVersion},
		Entries: entries,
	}})
}
