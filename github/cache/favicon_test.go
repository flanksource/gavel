package cache

import (
	"context"
	"net"
	nethttp "net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type publicTestTransport struct {
	target *url.URL
	base   nethttp.RoundTripper
}

func (t publicTestTransport) RoundTrip(req *nethttp.Request) (*nethttp.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = t.target.Scheme
	clone.URL.Host = t.target.Host
	return t.base.RoundTrip(clone)
}

func publicTestClient(t *testing.T, srv *httptest.Server) (*nethttp.Client, string) {
	t.Helper()
	target, err := url.Parse(srv.URL)
	require.NoError(t, err)
	client := srv.Client()
	client.Transport = publicTestTransport{target: target, base: client.Transport}
	return client, "http://example.com"
}

func TestNormalizeHomepage(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"plain https", "https://example.com", "https://example.com", false},
		{"trailing slash stripped", "https://example.com/", "https://example.com", false},
		{"path stripped", "https://example.com/docs/guide", "https://example.com", false},
		{"uppercased host lowered", "https://EXAMPLE.com/foo", "https://example.com", false},
		{"http preserved", "http://local.test", "http://local.test", false},
		{"empty", "", "", true},
		{"whitespace only", "   ", "", true},
		{"ftp rejected", "ftp://example.com", "", true},
		{"scheme missing", "example.com", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeHomepage(tc.in)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestValidateFaviconURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "public hostname", url: "https://example.com/icon.png"},
		{name: "public IP", url: "http://8.8.8.8/icon.png"},
		{name: "loopback IPv4", url: "http://127.0.0.1/icon.png", wantErr: true},
		{name: "loopback IPv6", url: "http://[::1]/icon.png", wantErr: true},
		{name: "private IPv4", url: "http://10.0.0.1/icon.png", wantErr: true},
		{name: "private IPv6", url: "http://[fd00::1]/icon.png", wantErr: true},
		{name: "link local", url: "http://169.254.169.254/latest/meta-data", wantErr: true},
		{name: "carrier grade NAT", url: "http://100.64.0.1/icon.png", wantErr: true},
		{name: "unsupported scheme", url: "file:///etc/passwd", wantErr: true},
		{name: "user information", url: "https://user@example.com/icon.png", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := validateFaviconURL(tc.url)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewFaviconHTTPClient_BlocksRestrictedAddress(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		hits++
		w.WriteHeader(nethttp.StatusOK)
	}))
	defer srv.Close()

	target, err := url.Parse(srv.URL)
	require.NoError(t, err)
	for _, targetURL := range []string{srv.URL, "http://localhost:" + target.Port()} {
		_, err := newFaviconHTTPClient().Get(targetURL)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "restricted address")
	}
	assert.Zero(t, hits)
}

func TestNewFaviconHTTPClient_BlocksRestrictedRedirect(t *testing.T) {
	client := newFaviconHTTPClient()
	target, err := url.Parse("http://127.0.0.1/favicon.ico")
	require.NoError(t, err)

	err = client.CheckRedirect(&nethttp.Request{URL: target}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "restricted address")
}

func TestIsRestrictedFaviconIP(t *testing.T) {
	tests := map[string]bool{
		"8.8.8.8":         false,
		"2001:4860::8888": false,
		"0.0.0.0":         true,
		"127.0.0.1":       true,
		"10.0.0.1":        true,
		"169.254.1.1":     true,
		"224.0.0.1":       true,
		"100.64.0.1":      true,
		"::1":             true,
		"fd00::1":         true,
	}
	for raw, want := range tests {
		t.Run(raw, func(t *testing.T) {
			assert.Equal(t, want, isRestrictedFaviconIP(net.ParseIP(raw)))
		})
	}
}

func TestMaxDimension(t *testing.T) {
	tests := map[string]int{
		"":                    0,
		"any":                 0,
		"16x16":               16,
		"32X32":               32,
		"180x180":             180,
		"16x16 32x32 180x180": 180,
		"16x16 any":           16,
		"bogus":               0,
	}
	for in, want := range tests {
		assert.Equal(t, want, maxDimension(in), "input=%q", in)
	}
}

func TestSortBySize_LargestFirstUnspecifiedLast(t *testing.T) {
	cands := []iconCandidate{
		{url: "a", size: 0},
		{url: "b", size: 32},
		{url: "c", size: 180},
		{url: "d", size: 16},
	}
	sortBySize(cands)
	assert.Equal(t, []string{"c", "b", "d", "a"},
		[]string{cands[0].url, cands[1].url, cands[2].url, cands[3].url})
}

func TestParseLinkCandidates_PicksLargestFirst(t *testing.T) {
	page := `<!doctype html><html><head>
		<link rel="icon" href="/small.png" sizes="16x16">
		<link rel="apple-touch-icon" href="/big.png" sizes="180x180">
		<link rel="shortcut icon" href="/mid.png" sizes="32x32">
		<link rel="stylesheet" href="/ignored.css">
	</head></html>`
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(page))
	}))
	defer srv.Close()

	client, homepage := publicTestClient(t, srv)
	got, err := parseLinkCandidates(context.Background(), client, homepage)
	require.NoError(t, err)

	// Expect ordering: 180x180, 32x32, 16x16. Stylesheet must be absent.
	require.Len(t, got, 3)
	assert.Contains(t, got[0], "/big.png")
	assert.Contains(t, got[1], "/mid.png")
	assert.Contains(t, got[2], "/small.png")
	for _, u := range got {
		assert.NotContains(t, u, "ignored.css")
	}
}

func TestDiscoverFavicon_FallsBackToFaviconIco(t *testing.T) {
	var iconHits int
	mux := nethttp.NewServeMux()
	mux.HandleFunc("/", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		// Homepage has no <link rel="icon"> at all.
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<!doctype html><html><head></head><body></body></html>`))
	})
	mux.HandleFunc("/favicon.ico", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		iconHits++
		w.Header().Set("Content-Type", "image/x-icon")
		_, _ = w.Write([]byte{0x00, 0x00, 0x01, 0x00}) // ICO magic prefix
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, homepage := publicTestClient(t, srv)
	iconURL, data, mime, err := discoverFaviconWithClient(context.Background(), client, homepage)
	require.NoError(t, err)
	assert.Equal(t, 1, iconHits, "/favicon.ico should be fetched as fallback")
	assert.Contains(t, iconURL, "/favicon.ico")
	assert.Equal(t, "image/x-icon", mime)
	assert.NotEmpty(t, data)
}

func TestDiscoverFavicon_PrefersLinkTagOverFallback(t *testing.T) {
	page := `<!doctype html><html><head>
		<link rel="icon" href="/explicit.png" sizes="64x64">
	</head></html>`
	var explicitHits, fallbackHits int
	mux := nethttp.NewServeMux()
	mux.HandleFunc("/", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(page))
	})
	mux.HandleFunc("/explicit.png", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		explicitHits++
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("PNGDATA"))
	})
	mux.HandleFunc("/favicon.ico", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		fallbackHits++
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client, homepage := publicTestClient(t, srv)
	iconURL, data, mime, err := discoverFaviconWithClient(context.Background(), client, homepage)
	require.NoError(t, err)
	assert.Equal(t, 1, explicitHits)
	assert.Equal(t, 0, fallbackHits, "explicit link must win; fallback never hit")
	assert.Contains(t, iconURL, "/explicit.png")
	assert.Equal(t, "image/png", mime)
	assert.Equal(t, "PNGDATA", string(data))
}

func TestFetchIcon_RejectsHTMLResponse(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>404</html>"))
	}))
	defer srv.Close()

	client, homepage := publicTestClient(t, srv)
	_, _, err := fetchIcon(context.Background(), client, homepage+"/favicon.ico")
	assert.Error(t, err, "html content-type must be rejected")
}
