// Tests hosted mode's route allowlist: listed routes are served, run endpoints answer
// 501 without running anything, and every other route is 404, through the hub too.
package ui

import (
	"encoding/json"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func hostedTestServer() string {
	srv := &Server{}
	srv.SetHostedMode(true)
	return newDashboardTestServer(srv).URL
}

func hostedError(resp *http.Response) string {
	var body struct {
		Error string `json:"error"`
	}
	Expect(json.NewDecoder(resp.Body).Decode(&body)).To(Succeed())
	return body.Error
}

var _ = Describe("hosted mode routes", func() {
	It("serves the allowlisted routes", func() {
		base := hostedTestServer()
		for _, path := range []string{"/api/settings/schema", "/favicon.svg", "/manifest.webmanifest", "/prs", "/todos/abc"} {
			resp := doRequest(newRequest(http.MethodGet, base+path, nil, nil))
			Expect(resp.StatusCode).To(Equal(http.StatusOK), path)
		}
	})

	It("answers 404 with the reason for every route it doesn't list", func() {
		base := hostedTestServer()
		for _, route := range []struct{ method, path string }{
			{http.MethodGet, "/api/activity"},
			{http.MethodGet, "/api/activity/cache"},
			{http.MethodPost, "/api/captain/sessions/abc/message"},
			{http.MethodGet, "/api/sessions"},
			{http.MethodPost, "/api/todos/session/focus"},
			{http.MethodGet, "/api/git/metrics"},
			{http.MethodGet, "/api/projects/git-summary"},
			{http.MethodGet, "/api/projects/app/diff"},
			{http.MethodGet, "/api/tests"},
			{http.MethodGet, "/api/proc/status"},
			{http.MethodGet, "/api/openapi.json"},
			{http.MethodGet, "/api/v1/openapi.json"},
			{http.MethodGet, "/api/v1/project"},
			{http.MethodPost, "/api/v1/todo/abc/push"},
			{http.MethodGet, "/api/v1/tasks"},
			{http.MethodPost, "/api/chat"},
			{http.MethodGet, "/api/settings/gavel"},
			{http.MethodGet, "/api/settings/prompts"},
			{http.MethodPost, "/api/todos/land"},
			{http.MethodPost, "/todos/new"},
			{http.MethodGet, "/debug/pprof/"},
			{http.MethodGet, "/metrics"},
			{http.MethodPost, "/api/todos/new"},
			{http.MethodGet, "/api/v1/todo"},
			{http.MethodGet, "/api/v1/todo/abc"},
			{http.MethodGet, "/api/projects/app"},
			{http.MethodGet, "/api/settings/runtime-presets"},
		} {
			resp := doRequest(newRequest(route.method, base+route.path, nil, nil))
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound), route.method+" "+route.path)
			Expect(hostedError(resp)).To(Equal(route.method + " " + route.path + " is not available in hosted mode"))
		}
		resp := doRequest(newRequest(http.MethodGet, base+"/results/abc", nil, nil))
		Expect(resp.StatusCode).To(Equal(http.StatusNotFound), "a page path that isn't a dashboard page")
	})

	It("answers agent run endpoints with 501 instead of running them", func() {
		base := hostedTestServer()
		for _, path := range []string{"/api/v1/todo/abc/run", "/api/v1/todo/abc/plan", "/api/v1/todo/abc/triage"} {
			resp := doRequest(newRequest(http.MethodPost, base+path, nil, strings.NewReader(`{}`)))
			Expect(resp.StatusCode).To(Equal(http.StatusNotImplemented), path)
			Expect(hostedError(resp)).To(Equal("agent runs are not available in hosted mode yet"))
		}
	})

	It("answers project changes with 400, since hosted projects are repositories", func() {
		base := hostedTestServer()
		for _, route := range []struct{ method, path string }{
			{http.MethodPost, "/api/projects"},
			{http.MethodPut, "/api/projects/app"},
			{http.MethodDelete, "/api/projects/app"},
		} {
			resp := doRequest(newRequest(route.method, base+route.path, nil, strings.NewReader(`{"dir":"/etc"}`)))
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest), route.method+" "+route.path)
			Expect(hostedError(resp)).To(Equal("projects are repos in hosted mode"))
		}
	})

	It("answers project actions without touching the host's projects", func() {
		base := hostedTestServer()
		for _, test := range []struct {
			path, body string
			status     int
			message    string
		}{
			{"/api/projects/app/actions", `{"action":"test"}`, http.StatusNotImplemented, "project actions are not available in hosted mode yet"},
			{"/api/projects/app/actions", `{"action":"lint"}`, http.StatusNotImplemented, "project actions are not available in hosted mode yet"},
			{"/api/projects/app/actions", `{"action":"commit"}`, http.StatusBadRequest, "commits come from agent runs in hosted mode"},
			{"/api/projects/app/commit-queue", `{}`, http.StatusBadRequest, "commits come from agent runs in hosted mode"},
			{"/api/projects/app/branch/pr", `{}`, http.StatusBadRequest, "commits come from agent runs in hosted mode"},
		} {
			resp := doRequest(newRequest(http.MethodPost, base+test.path, nil, strings.NewReader(test.body)))
			Expect(resp.StatusCode).To(Equal(test.status), test.path+" "+test.body)
			Expect(hostedError(resp)).To(Equal(test.message))
		}
	})

	It("closes a hub subscription to a route it doesn't list with the 404", func() {
		events := openDashboardEvents(hostedTestServer())
		Expect(events.subscribe("activity", "/api/activity/stream").StatusCode).To(Equal(http.StatusNoContent))
		frame := events.nextDataFrame()
		Expect(frame.Event).To(Equal("activity/__closed"))
		Expect(frame.Data).To(ContainSubstring(`"status":404`))
		Expect(frame.Data).To(ContainSubstring("GET /api/activity/stream is not available in hosted mode"))
	})

	It("leaves every route in place outside hosted mode", func() {
		base := newDashboardTestServer(&Server{}).URL
		resp := doRequest(newRequest(http.MethodGet, base+"/api/openapi.json", nil, nil))
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
	})
})

var _ = Describe("hosted route allowlist", func() {
	It("names only routes the dashboard registers", func() {
		original := hostedPassThrough
		DeferCleanup(func() { hostedPassThrough = original })
		hostedPassThrough = append(append([]string{}, original...), "GET /api/prs/typo")
		srv := &Server{}
		srv.SetHostedMode(true)
		Expect(func() { srv.Handler() }).To(PanicWith(ContainSubstring(`"GET /api/prs/typo"`)))
	})
})
