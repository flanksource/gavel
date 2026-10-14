// Talks to a hosted gavel as the logged-in CLI user: the Clerk access token as a bearer
// token and the chosen organization as X-Gavel-Org, refreshing the token when it expires.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/flanksource/gavel/api"
	"github.com/flanksource/gavel/pr/ui"
)

// gavelServerHTTPClient reaches hosted gavel servers. It is a package variable so tests
// can reach a test server.
var gavelServerHTTPClient = &http.Client{Timeout: 30 * time.Second}

// gavelServer is a logged-in session with a hosted gavel.
type gavelServer struct {
	session *api.Session
}

// openGavelServer loads the login, refreshing and saving its tokens when they expired.
func openGavelServer(ctx context.Context) (gavelServer, error) {
	path, err := api.CredentialPath()
	if err != nil {
		return gavelServer{}, err
	}
	notLoggedIn := errors.New("not logged in to a gavel server; run gavel login --server <url>")
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return gavelServer{}, notLoggedIn
	}
	session, err := api.LoadSession(path)
	if err != nil {
		return gavelServer{}, err
	}
	if session.Server == "" {
		return gavelServer{}, notLoggedIn
	}
	client, err := api.NewAuthClient(ctx, session.Config, authHTTPClient)
	if err != nil {
		return gavelServer{}, err
	}
	updated, _, err := client.WhoAmI(ctx, session)
	if err != nil {
		return gavelServer{}, err
	}
	if err := api.SaveSession(path, updated); err != nil {
		return gavelServer{}, err
	}
	return gavelServer{session: updated}, nil
}

// call sends body (if any) as JSON to the server's path and decodes a JSON response into
// output (if any). Failures carry the server's reason.
func (g gavelServer) call(ctx context.Context, method, path string, body, output any) error {
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(g.session.Server, "/")+path, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+g.session.Token.AccessToken)
	if g.session.Org != "" {
		req.Header.Set("X-Gavel-Org", g.session.Org)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := gavelServerHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach %s: %w", g.session.Server, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return serverError(resp.StatusCode, data)
	}
	if output == nil || len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, output); err != nil {
		return fmt.Errorf("invalid response from %s: %w", g.session.Server, err)
	}
	return nil
}

// serverError reads {"error", "message"} into one error, or falls back to the status.
func serverError(status int, data []byte) error {
	var failure struct{ Error, Message string }
	if json.Unmarshal(data, &failure) == nil && failure.Error != "" {
		if failure.Message != "" {
			return fmt.Errorf("%s: %s", failure.Error, failure.Message)
		}
		return errors.New(failure.Error)
	}
	return fmt.Errorf("the gavel server answered HTTP %d", status)
}

// serverAuthConfig reads how server signs in, refusing a server without Clerk.
func serverAuthConfig(ctx context.Context, server string) (ui.AuthConfigResponse, error) {
	u, err := url.Parse(server)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return ui.AuthConfigResponse{}, fmt.Errorf("--server must be a URL like https://gavel.example.com, got %q", server)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(server, "/")+"/api/auth/config", nil)
	if err != nil {
		return ui.AuthConfigResponse{}, err
	}
	resp, err := gavelServerHTTPClient.Do(req)
	if err != nil {
		return ui.AuthConfigResponse{}, fmt.Errorf("cannot reach %s: %w", server, err)
	}
	defer resp.Body.Close()
	var cfg ui.AuthConfigResponse
	if resp.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&cfg) != nil {
		return ui.AuthConfigResponse{}, fmt.Errorf("%s didn't answer /api/auth/config (HTTP %d); is it a gavel server?", server, resp.StatusCode)
	}
	if cfg.Mode != authClerk || cfg.Issuer == "" || cfg.ClientID == "" {
		return ui.AuthConfigResponse{}, fmt.Errorf("%s doesn't accept logins (auth mode %q); it must run gavel serve --auth clerk", server, cfg.Mode)
	}
	return cfg, nil
}
