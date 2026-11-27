// Package missioncontrol reads what a Mission Control instance knows about the
// infrastructure hosted gavel runs on, for the platform dashboard, through its client SDK.
package missioncontrol

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/flanksource/mission-control-sdk/go/client"
)

// Source reads one Mission Control instance.
type Source struct {
	// URL is the Mission Control backend the client talks to, named in errors.
	URL    string
	Client *client.Client
	// Search is the Mission Control resource search selecting gavel's infrastructure,
	// e.g. "namespace=gavel".
	Search string
}

// New reads the Mission Control at url with the access token, selecting gavel's
// infrastructure with search.
func New(url, token, search string) *Source {
	return &Source{URL: url, Client: client.New(url, token), Search: search}
}

// explain wraps err from reading what with where it was read from, and which token to
// check when Mission Control refused it.
func (s *Source) explain(what string, err error) error {
	var server *client.ServerError
	if errors.As(err, &server) && (server.StatusCode == http.StatusUnauthorized || server.StatusCode == http.StatusForbidden) {
		return fmt.Errorf("read %s from Mission Control at %s: %w (check GAVEL_MISSION_CONTROL_TOKEN, a token with Mission Control's viewer role)", what, s.URL, err)
	}
	return fmt.Errorf("read %s from Mission Control at %s: %w", what, s.URL, err)
}
