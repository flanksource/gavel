package github

import (
	"context"
	"fmt"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github/activity"
	"github.com/flanksource/gavel/github/cache"
)

const apiBaseURL = "https://api.github.com"

// cachedGetResult is the output of cachedGet — either fresh bytes from the
// upstream (FromCache=false) or previously-cached bytes served after a 304
// (FromCache=true).
type cachedGetResult struct {
	Body      []byte
	FromCache bool
}

// cachedGet performs a GET against the GitHub API with ETag-aware caching.
// path is the API-relative path (e.g. "/repos/owner/name/actions/runs/42").
// extraHeaders are applied to the outgoing request; caller-supplied Accept
// headers take precedence over the default.
//
// The flow:
//  1. Look up the (absolute URL, GET) in the cache.
//  2. If a cached entry with a validator (ETag or Last-Modified) exists,
//     attach If-None-Match / If-Modified-Since.
//  3. Issue the request.
//  4. On 304, touch the cache entry and return cached Body.
//  5. On 2xx, store the new body + ETag and return it.
//  6. On 5xx with a stale cached body, log a warning and return the stale
//     body so the caller can degrade gracefully.
func cachedGet(ctx context.Context, token, path string, extraHeaders map[string]string) (*cachedGetResult, error) {
	absURL := apiBaseURL + path
	store := cache.Shared()
	entry := store.LookupHTTP(absURL, "GET")

	client := newClient(token)
	req := client.R(ctx)
	for k, v := range extraHeaders {
		req = req.Header(k, v)
	}
	if entry != nil && entry.HasValidator() {
		if entry.ETag != "" {
			req = req.Header("If-None-Match", entry.ETag)
		}
		if entry.LastModified != "" {
			req = req.Header("If-Modified-Since", entry.LastModified)
		}
	}

	logger.Tracef("cachedGet: GET %s (cached=%t)", path, entry != nil)
	start := time.Now()
	resp, err := req.Get(absURL)
	if err != nil {
		// Network error: fall back to stale cache if available.
		if entry != nil {
			logger.Warnf("cachedGet: network error on %s, serving stale cache: %v", path, err)
			activity.Shared().Record(activity.Entry{
				Method: "GET", URL: path, Kind: activity.KindREST,
				Duration: time.Since(start), SizeBytes: len(entry.Body),
				FromCache: true, Error: err.Error(),
			})
			return &cachedGetResult{Body: entry.Body, FromCache: true}, nil
		}
		activity.Shared().Record(activity.Entry{
			Method: "GET", URL: path, Kind: activity.KindREST,
			Duration: time.Since(start), Error: err.Error(),
		})
		return nil, fmt.Errorf("GET %s: %w", path, err)
	}

	if resp.StatusCode == 304 && entry != nil {
		store.TouchHTTP(absURL, "GET")
		logger.Tracef("cachedGet: 304 served from cache: %s", path)
		activity.Shared().Record(activity.Entry{
			Method: "GET", URL: path, Kind: activity.KindREST,
			StatusCode: 304, Duration: time.Since(start),
			SizeBytes: len(entry.Body), FromCache: true,
		})
		return &cachedGetResult{Body: entry.Body, FromCache: true}, nil
	}

	body, err := resp.AsString()
	if err != nil {
		activity.Shared().Record(activity.Entry{
			Method: "GET", URL: path, Kind: activity.KindREST,
			StatusCode: resp.StatusCode, Duration: time.Since(start),
			Error: err.Error(),
		})
		return nil, fmt.Errorf("read body %s: %w", path, err)
	}
	bodyBytes := []byte(body)

	if resp.StatusCode >= 500 && entry != nil {
		logger.Warnf("cachedGet: %d on %s, serving stale cache", resp.StatusCode, path)
		activity.Shared().Record(activity.Entry{
			Method: "GET", URL: path, Kind: activity.KindREST,
			StatusCode: resp.StatusCode, Duration: time.Since(start),
			SizeBytes: len(entry.Body), FromCache: true,
		})
		return &cachedGetResult{Body: entry.Body, FromCache: true}, nil
	}
	if !resp.IsOK() {
		activity.Shared().Record(activity.Entry{
			Method: "GET", URL: path, Kind: activity.KindREST,
			StatusCode: resp.StatusCode, Duration: time.Since(start),
			SizeBytes: len(bodyBytes),
			Error:     fmt.Sprintf("status %d", resp.StatusCode),
		})
		return nil, fmt.Errorf("GET %s: status %d: %s", path, resp.StatusCode, body)
	}

	store.StoreHTTP(absURL, "GET", resp.StatusCode, bodyBytes, resp.Header, nil)
	activity.Shared().Record(activity.Entry{
		Method: "GET", URL: path, Kind: activity.KindREST,
		StatusCode: resp.StatusCode, Duration: time.Since(start),
		SizeBytes: len(bodyBytes), FromCache: false,
	})
	return &cachedGetResult{Body: bodyBytes, FromCache: false}, nil
}
