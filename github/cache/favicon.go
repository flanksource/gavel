package cache

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/flanksource/commons/logger"
	"github.com/flanksource/gavel/github/activity"
	"golang.org/x/net/html"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FaviconCache stores favicons fetched from repo homepage URLs. An entry with
// empty Data but a non-zero FetchedAt is a negative cache entry — we looked,
// found nothing usable, and don't want to re-hit the site until TTL expiry.
type FaviconCache struct {
	Homepage  string    `gorm:"primaryKey;size:1024"`
	IconURL   string    `gorm:"size:1024"`
	Data      []byte    // empty for negative entries
	MimeType  string    `gorm:"size:64"`
	FetchedAt time.Time `gorm:"index"`
	ExpiresAt time.Time `gorm:"index"`
}

const (
	faviconTTL         = 7 * 24 * time.Hour
	faviconMaxBytes    = 64 * 1024
	faviconHTTPTimeout = 5 * time.Second
)

// normalizeHomepage reduces a homepage URL to scheme://host (lowercased host,
// no path, no trailing slash) so that multiple repos pointing at the same site
// share one cache entry. Returns an error for relative or non-http(s) URLs.
func normalizeHomepage(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("empty homepage")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse homepage %q: %w", raw, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("homepage missing host: %q", raw)
	}
	return u.Scheme + "://" + strings.ToLower(u.Host), nil
}

// GetFavicon returns a cached favicon for the given homepage. hit indicates
// whether a cache row (positive or negative) was found and is still fresh.
// data is empty for negative hits and misses; callers decide whether to 404
// or trigger a fetch.
func (s *Store) GetFavicon(homepage string) (data []byte, mime string, hit bool, err error) {
	if s.Disabled() {
		return nil, "", false, nil
	}
	norm, err := normalizeHomepage(homepage)
	if err != nil {
		return nil, "", false, err
	}
	var row FaviconCache
	if err := s.gorm().Where("homepage = ?", norm).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", false, nil
		}
		return nil, "", false, fmt.Errorf("favicon cache read %q: %w", norm, err)
	}
	if time.Now().After(row.ExpiresAt) {
		return nil, "", false, nil
	}
	return row.Data, row.MimeType, true, nil
}

// FetchFavicon resolves and downloads the favicon for homepage, persists the
// result (positive or negative) with faviconTTL, and returns the bytes. A nil
// error with empty data means "site has no usable favicon" — callers should
// treat that as a 404, not a 500.
func (s *Store) FetchFavicon(ctx context.Context, homepage string) (data []byte, mime string, err error) {
	norm, err := normalizeHomepage(homepage)
	if err != nil {
		return nil, "", err
	}

	start := time.Now()
	iconURL, iconData, iconMime, fetchErr := discoverFavicon(ctx, norm)
	// Size cap: if upstream returned oversized bytes, demote to negative entry.
	if len(iconData) > faviconMaxBytes {
		logger.Debugf("favicon for %s exceeds %d bytes, caching as negative", norm, faviconMaxBytes)
		iconData = nil
		iconMime = ""
	}

	activity.Shared().Record(activity.Entry{
		Method:     http.MethodGet,
		URL:        norm,
		Kind:       activity.KindFavicon,
		StatusCode: statusFor(fetchErr, iconData),
		Duration:   time.Since(start),
		SizeBytes:  len(iconData),
		Error:      errString(fetchErr),
	})

	// Persist even on failure — a negative entry prevents repeated fetches
	// for sites without a favicon or that are temporarily broken.
	if err := s.saveFavicon(norm, iconURL, iconData, iconMime); err != nil {
		logger.Warnf("favicon cache save %s: %v", norm, err)
	}

	if fetchErr != nil {
		return nil, "", fetchErr
	}
	return iconData, iconMime, nil
}

func statusFor(err error, data []byte) int {
	if err != nil {
		return 0
	}
	if len(data) == 0 {
		return http.StatusNotFound
	}
	return http.StatusOK
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (s *Store) saveFavicon(homepage, iconURL string, data []byte, mime string) error {
	if s.Disabled() {
		return nil
	}
	row := FaviconCache{
		Homepage:  homepage,
		IconURL:   iconURL,
		Data:      data,
		MimeType:  mime,
		FetchedAt: time.Now(),
		ExpiresAt: time.Now().Add(faviconTTL),
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.gorm().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "homepage"}},
		UpdateAll: true,
	}).Create(&row).Error
}

// discoverFavicon fetches homepage HTML, walks <link rel="icon"> tags to pick
// the best candidate, and falls back to /favicon.ico. Returns the resolved
// icon URL alongside its bytes so the cache can record where the icon came
// from (useful for debugging).
func discoverFavicon(ctx context.Context, homepage string) (iconURL string, data []byte, mime string, err error) {
	client := &http.Client{Timeout: faviconHTTPTimeout}

	candidates, htmlErr := parseLinkCandidates(ctx, client, homepage)
	if htmlErr != nil {
		logger.Debugf("favicon html fetch %s: %v", homepage, htmlErr)
	}
	// Always append /favicon.ico as the final fallback so sites that don't
	// declare a <link rel="icon"> still resolve.
	candidates = append(candidates, homepage+"/favicon.ico")

	var lastErr error
	for _, u := range candidates {
		body, contentType, fetchErr := fetchIcon(ctx, client, u)
		if fetchErr != nil {
			lastErr = fetchErr
			continue
		}
		if len(body) == 0 {
			continue
		}
		return u, body, contentType, nil
	}
	if lastErr != nil {
		return "", nil, "", lastErr
	}
	// No candidates returned a body: treat as "no favicon found" (nil error).
	return "", nil, "", nil
}

type iconCandidate struct {
	url  string
	size int // max of width/height in sizes="WxH"; 0 when unspecified
}

// parseLinkCandidates downloads homepage, parses the HTML and returns icon
// candidate URLs ordered by declared size (largest first). Missing size hints
// sort last so "any" / default icons are tried after explicit high-res ones.
func parseLinkCandidates(ctx context.Context, client *http.Client, homepage string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, homepage, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gavel-favicon/1")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("homepage GET %s: status %d", homepage, resp.StatusCode)
	}
	// Cap HTML read to 512KB — anything larger is almost certainly a non-HTML
	// response and we're wasting bandwidth.
	body := io.LimitReader(resp.Body, 512*1024)
	doc, err := html.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	base, err := url.Parse(homepage)
	if err != nil {
		return nil, err
	}

	var cands []iconCandidate
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && strings.EqualFold(n.Data, "link") {
			if c, ok := extractIconCandidate(n, base); ok {
				cands = append(cands, c)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	// Sort by size descending; unspecified sizes go last.
	sortBySize(cands)
	out := make([]string, len(cands))
	for i, c := range cands {
		out[i] = c.url
	}
	return out, nil
}

func extractIconCandidate(n *html.Node, base *url.URL) (iconCandidate, bool) {
	var rel, href, sizes string
	for _, a := range n.Attr {
		switch strings.ToLower(a.Key) {
		case "rel":
			rel = strings.ToLower(a.Val)
		case "href":
			href = a.Val
		case "sizes":
			sizes = a.Val
		}
	}
	if href == "" {
		return iconCandidate{}, false
	}
	// Accept "icon", "shortcut icon", "apple-touch-icon", "mask-icon".
	isIcon := false
	for r := range strings.FieldsSeq(rel) {
		switch r {
		case "icon", "shortcut", "apple-touch-icon", "mask-icon":
			isIcon = true
		}
	}
	if !isIcon {
		return iconCandidate{}, false
	}
	resolved, err := base.Parse(href)
	if err != nil {
		return iconCandidate{}, false
	}
	return iconCandidate{url: resolved.String(), size: maxDimension(sizes)}, true
}

func maxDimension(sizes string) int {
	best := 0
	for token := range strings.FieldsSeq(sizes) {
		if strings.EqualFold(token, "any") {
			continue
		}
		w, h, ok := strings.Cut(strings.ToLower(token), "x")
		if !ok {
			continue
		}
		wi, _ := strconv.Atoi(w)
		hi, _ := strconv.Atoi(h)
		if wi > best {
			best = wi
		}
		if hi > best {
			best = hi
		}
	}
	return best
}

func sortBySize(cands []iconCandidate) {
	// Stable insertion sort — lists are short (typically < 8 entries).
	for i := 1; i < len(cands); i++ {
		for j := i; j > 0 && cands[j].size > cands[j-1].size; j-- {
			cands[j], cands[j-1] = cands[j-1], cands[j]
		}
	}
}

func fetchIcon(ctx context.Context, client *http.Client, iconURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, iconURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "gavel-favicon/1")
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("icon GET %s: status %d", iconURL, resp.StatusCode)
	}
	// Cap at faviconMaxBytes+1 so the caller can detect oversize.
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, io.LimitReader(resp.Body, faviconMaxBytes+1)); err != nil {
		return nil, "", fmt.Errorf("read icon %s: %w", iconURL, err)
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = http.DetectContentType(buf.Bytes())
	}
	// Reject HTML / text payloads — some servers return a 200 HTML error page
	// for missing favicons.
	if strings.Contains(strings.ToLower(mime), "text/html") {
		return nil, "", fmt.Errorf("icon %s returned html content-type", iconURL)
	}
	return buf.Bytes(), mime, nil
}
