package cache

import (
	"encoding/json"
	nethttp "net/http"
	"time"

	"github.com/flanksource/commons/logger"
	"gorm.io/gorm/clause"
)

// CachedHTTPResponse describes the result of a cache lookup.
//
// On a hit, callers send an If-None-Match (and/or If-Modified-Since) request
// to the upstream. If the upstream returns 304, the caller serves Body
// directly without re-decoding. If the upstream returns 200 the caller calls
// StoreHTTP with the new ETag/body.
type CachedHTTPResponse struct {
	Body         []byte
	ETag         string
	LastModified string
	StatusCode   int
	Headers      nethttp.Header
	FetchedAt    time.Time
}

// HasValidator reports whether the cached entry has an ETag or Last-Modified
// header — i.e., whether a conditional request will produce a meaningful 304.
func (r *CachedHTTPResponse) HasValidator() bool {
	return r != nil && (r.ETag != "" || r.LastModified != "")
}

// LookupHTTP returns the cached response for (url, method) or nil if absent.
// A disabled store always returns nil.
func (s *Store) LookupHTTP(url, method string) *CachedHTTPResponse {
	if s == nil || s.disabled {
		return nil
	}
	if method == "" {
		method = "GET"
	}
	var entry HTTPCacheEntry
	res := s.gorm().Where("url = ? AND method = ?", url, method).First(&entry)
	if res.Error != nil {
		return nil
	}
	if entry.ExpiresAt != nil && time.Now().After(*entry.ExpiresAt) {
		// Expired hard cap. Treat as miss but keep the row so the ETag can
		// still be sent on the next request — that's what TouchHTTP is for.
		return nil
	}
	return &CachedHTTPResponse{
		Body:         entry.Body,
		ETag:         entry.ETag,
		LastModified: entry.LastModified,
		StatusCode:   entry.StatusCode,
		Headers:      decodeHeaders(entry.Headers),
		FetchedAt:    entry.FetchedAt,
	}
}

// StoreHTTP upserts a cache entry. expiresAt is optional — pass nil for no
// expiry (revalidation only via ETag).
func (s *Store) StoreHTTP(url, method string, statusCode int, body []byte, headers nethttp.Header, expiresAt *time.Time) {
	if s == nil || s.disabled {
		return
	}
	if method == "" {
		method = "GET"
	}
	entry := HTTPCacheEntry{
		URL:          url,
		Method:       method,
		ETag:         headers.Get("ETag"),
		LastModified: headers.Get("Last-Modified"),
		StatusCode:   statusCode,
		Body:         body,
		Headers:      encodeHeaders(headers),
		FetchedAt:    time.Now(),
		ExpiresAt:    expiresAt,
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	// Upsert: postgres ON CONFLICT (url, method) DO UPDATE.
	res := s.gorm().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "url"}, {Name: "method"}},
		UpdateAll: true,
	}).Create(&entry)
	if res.Error != nil {
		logger.Warnf("github cache upsert failed for %s %s: %v", method, url, res.Error)
	}
}

// TouchHTTP updates only the FetchedAt timestamp on an existing entry. Use
// this after a successful 304 to keep the entry warm without rewriting the
// body.
func (s *Store) TouchHTTP(url, method string) {
	if s == nil || s.disabled {
		return
	}
	if method == "" {
		method = "GET"
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	s.gorm().Model(&HTTPCacheEntry{}).
		Where("url = ? AND method = ?", url, method).
		Update("fetched_at", time.Now())
}

// relevantHeaderKeys controls which response headers we persist alongside
// the body. Keep this list small — headers are stored as JSON and we don't
// want every cookie/server header bloating the cache.
var relevantHeaderKeys = []string{
	"Content-Type",
	"Link",
	"ETag",
	"Last-Modified",
}

func encodeHeaders(h nethttp.Header) []byte {
	if len(h) == 0 {
		return nil
	}
	out := make(map[string]string, len(relevantHeaderKeys))
	for _, k := range relevantHeaderKeys {
		if v := h.Get(k); v != "" {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	b, _ := json.Marshal(out)
	return b
}

func decodeHeaders(b []byte) nethttp.Header {
	if len(b) == 0 {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	h := make(nethttp.Header, len(m))
	for k, v := range m {
		h.Set(k, v)
	}
	return h
}
