package record

import (
	"net/url"
	"sort"
	"strconv"
)

// celRequestCap bounds the per-request detail exposed to CEL. The aggregates
// above it are always exact — a fixture asserting `http.entries == 300` still
// gets 300 — but the artifact on disk, not the CEL variable, is where a long
// tail of requests belongs.
const celRequestCap = 200

// HTTPCELVars builds the `http` CEL root from a fixture's slice of the
// recording. Keys are stable and always present, including on an empty
// recording, so `http.errors == 0` is a legal assertion for a fixture that made
// no calls rather than an evaluation error.
//
// Map keys are strings throughout: `http.statuses["200"]` reads the same as
// `http.methods["GET"]`, and CEL has no integer map literal indexing that would
// make the alternative worth the inconsistency.
func HTTPCELVars(entries []Entry, path string) map[string]any {
	methods := map[string]int{}
	statuses := map[string]int{}
	hostSet := map[string]bool{}
	requests := make([]map[string]any, 0, min(len(entries), celRequestCap))

	var errors int
	var bytesIn, bytesOut int64

	for _, entry := range entries {
		host, urlPath := splitURL(entry.Request.URL)
		hostSet[host] = true
		methods[entry.Request.Method]++
		statuses[strconv.Itoa(entry.Response.Status)]++

		if entry.Gavel != nil {
			bytesIn += entry.Gavel.BytesIn
			bytesOut += entry.Gavel.BytesOut
			if entry.Gavel.Error != "" {
				errors++
			}
		}
		// A 4xx/5xx the server actually returned is an error the fixture is
		// usually asserting *about*, so it counts here alongside transport
		// failures rather than being silently distinct.
		if entry.Response.Status >= 400 {
			errors++
		}

		if len(requests) < celRequestCap {
			requests = append(requests, requestVars(entry, host, urlPath))
		}
	}

	return map[string]any{
		"entries":   len(entries),
		"hosts":     sortedKeys(hostSet),
		"methods":   methods,
		"statuses":  statuses,
		"errors":    errors,
		"bytes_in":  bytesIn,
		"bytes_out": bytesOut,
		"requests":  requests,
		"path":      path,
	}
}

func requestVars(entry Entry, host, urlPath string) map[string]any {
	vars := map[string]any{
		"method":           entry.Request.Method,
		"url":              entry.Request.URL,
		"host":             host,
		"path":             urlPath,
		"status":           entry.Response.Status,
		"duration_ms":      entry.Time,
		"mime":             entry.Response.Content.MimeType,
		"request_headers":  headerMap(entry.Request.Headers),
		"response_headers": headerMap(entry.Response.Headers),
		"tunnelled":        entry.Gavel != nil && entry.Gavel.Tunnelled,
		"error":            "",
	}
	if entry.Gavel != nil {
		vars["error"] = entry.Gavel.Error
	}
	return vars
}

// headerMap flattens HAR's ordered pairs into the lookup CEL wants. Repeated
// headers keep the first value; a fixture that needs every value of a repeated
// header should assert against the artifact, not the summary.
func headerMap(headers []NameValue) map[string]string {
	out := make(map[string]string, len(headers))
	for _, header := range headers {
		if _, exists := out[header.Name]; !exists {
			out[header.Name] = header.Value
		}
	}
	return out
}

// splitURL pulls out the host and path. A CONNECT entry's URL is synthesised as
// `https://<host:port>` and so yields an empty path, which is honest: the proxy
// never saw one.
func splitURL(raw string) (host, path string) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", ""
	}
	return hostOnly(parsed.Host), parsed.Path
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
