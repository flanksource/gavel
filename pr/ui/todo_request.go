package ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/flanksource/gavel/todos"
)

// decodeTodoRequest decodes a todo write endpoint's body strictly: exactly one
// JSON object, with no key the payload does not declare. An unknown key is a
// client that believes it configured something this server never read, and
// the decoder's own message says what was wrong with the input — a bare
// "invalid json" sent the caller back to guess.
func decodeTodoRequest(r *http.Request, payload any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(payload); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("invalid request body: expected one JSON object")
	}
	return nil
}

// writeTodoJSON encodes a response body under an explicit status.
func writeTodoJSON(w http.ResponseWriter, status int, body any) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body) //nolint:errcheck
}

// continuationFailureStatus is the status for a continuation whose transition
// committed but whose run did not start. A run refused because a live one
// already owns the todo is a conflict the client resolves by retrying with
// force; anything else is the server failing to do what it had already agreed
// to. Neither is a bad request: the request was accepted, and part of it took.
func continuationFailureStatus(err error) int {
	var owned *todos.ErrRunOwnedElsewhere
	if errors.As(err, &owned) {
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}
