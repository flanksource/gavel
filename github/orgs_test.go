package github

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchUserOrgs_OK(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		require.Equal(t, "/user/orgs", r.URL.Path)
		require.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`[
			{"login":"flanksource","avatar_url":"https://example.test/f.png"},
			{"login":"anthropics","avatar_url":"https://example.test/a.png"}
		]`))
	}))
	t.Cleanup(srv.Close)

	orgs, err := fetchUserOrgs(Options{Token: "tok"}, srv.URL)
	require.NoError(t, err)
	require.Len(t, orgs, 2)
	assert.Equal(t, "flanksource", orgs[0].Login)
	assert.Equal(t, "https://example.test/f.png", orgs[0].AvatarURL)
	assert.Equal(t, "anthropics", orgs[1].Login)
}

func TestFetchUserOrgs_EmptyListIsNotAnError(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(srv.Close)

	orgs, err := fetchUserOrgs(Options{Token: "tok"}, srv.URL)
	require.NoError(t, err)
	assert.Empty(t, orgs)
}

func TestFetchUserOrgs_ErrorsOnNon200(t *testing.T) {
	srv := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		nethttp.Error(w, "unauth", nethttp.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	_, err := fetchUserOrgs(Options{Token: "bad"}, srv.URL)
	assert.Error(t, err)
}

func TestResolveDefaultOrg_PrefersFirstOrg(t *testing.T) {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("/user/orgs", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write([]byte(`[
			{"login":"primary-org","avatar_url":""},
			{"login":"secondary-org","avatar_url":""}
		]`))
	})
	mux.HandleFunc("/user", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		// /user shouldn't be consulted when /user/orgs returned members.
		t.Errorf("/user was hit despite /user/orgs returning results")
		_, _ = w.Write([]byte(`{"login":"should-not-be-used"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	got, err := resolveDefaultOrg(Options{Token: "tok"}, srv.URL, nil)
	require.NoError(t, err)
	assert.Equal(t, "primary-org", got)
}

func TestResolveDefaultOrg_FallsBackToUserLogin(t *testing.T) {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("/user/orgs", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		// Solo developer: the token is valid but has no org memberships.
		_, _ = w.Write([]byte(`[]`))
	})
	mux.HandleFunc("/user", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write([]byte(`{"login":"solo-dev"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	got, err := resolveDefaultOrg(Options{Token: "tok"}, srv.URL, nil)
	require.NoError(t, err)
	assert.Equal(t, "solo-dev", got)
}

func TestResolveDefaultOrg_SkipsIgnoredOrgs(t *testing.T) {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("/user/orgs", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write([]byte(`[
			{"login":"noisy-legacy-org","avatar_url":""},
			{"login":"real-work-org","avatar_url":""}
		]`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	// With the first org ignored, the resolver should pick the second.
	got, err := resolveDefaultOrg(Options{Token: "tok"}, srv.URL, []string{"noisy-legacy-org"})
	require.NoError(t, err)
	assert.Equal(t, "real-work-org", got)
}

func TestResolveDefaultOrg_AllIgnoredFallsBackToUser(t *testing.T) {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("/user/orgs", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write([]byte(`[{"login":"only-org","avatar_url":""}]`))
	})
	mux.HandleFunc("/user", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = w.Write([]byte(`{"login":"solo-dev"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	got, err := resolveDefaultOrg(Options{Token: "tok"}, srv.URL, []string{"only-org"})
	require.NoError(t, err)
	assert.Equal(t, "solo-dev", got, "ignoring the only org should fall through to /user.login")
}

func TestResolveDefaultOrg_ErrorsWhenBothEndpointsFail(t *testing.T) {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("/user/orgs", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		nethttp.Error(w, "unauth", nethttp.StatusUnauthorized)
	})
	mux.HandleFunc("/user", func(w nethttp.ResponseWriter, r *nethttp.Request) {
		nethttp.Error(w, "unauth", nethttp.StatusUnauthorized)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	_, err := resolveDefaultOrg(Options{Token: "bad"}, srv.URL, nil)
	assert.Error(t, err)
}
