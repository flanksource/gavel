// Signs the state that carries a GitHub App install from gavel to GitHub and back, so the
// setup callback knows which org and user started it and that nobody changed it.
package installations

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// stateLifetime is how long a GitHub install link stays valid.
const stateLifetime = 15 * time.Minute

var errInvalidState = errors.New("the GitHub install link isn't valid; start the install again from gavel")

// State is who started an install, and until when (Unix seconds) it may finish.
type State struct {
	Org     uuid.UUID `json:"org"`
	User    uuid.UUID `json:"user"`
	Expires int64     `json:"exp"`
}

// SignState returns base64url(json) + "." + base64url(HMAC-SHA256(key, payload)).
func SignState(key []byte, org, user uuid.UUID, now time.Time) string {
	data, _ := json.Marshal(State{Org: org, User: user, Expires: now.Add(stateLifetime).Unix()})
	payload := base64.RawURLEncoding.EncodeToString(data)
	return payload + "." + base64.RawURLEncoding.EncodeToString(stateMAC(key, payload))
}

// VerifyState checks state's signature and expiry and returns what it carries.
func VerifyState(key []byte, state string, now time.Time) (State, error) {
	payload, signature, ok := strings.Cut(state, ".")
	if !ok {
		return State{}, errInvalidState
	}
	mac, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil || !hmac.Equal(mac, stateMAC(key, payload)) {
		return State{}, errInvalidState
	}
	data, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return State{}, errInvalidState
	}
	var parsed State
	if err := json.Unmarshal(data, &parsed); err != nil {
		return State{}, errInvalidState
	}
	if now.Unix() > parsed.Expires {
		return State{}, errors.New("the GitHub install link expired; start the install again from gavel")
	}
	return parsed, nil
}

func stateMAC(key []byte, payload string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	return mac.Sum(nil)
}
