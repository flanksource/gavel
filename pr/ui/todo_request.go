package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/flanksource/gavel/todos"
	"github.com/flanksource/gavel/verify"
)

// decodeTodoRequest requires exactly one JSON object and applies the shared
// unknown-field transition policy to every todo write endpoint.
func decodeTodoRequest(r *http.Request, payload any) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 || data[0] != '{' {
		return fmt.Errorf("invalid request body: expected one JSON object")
	}
	if err := verify.DecodeJSON(data, payload, verify.DecodeOptions{Source: r.Method + " " + r.URL.Path + " request body"}); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
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
