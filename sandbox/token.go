// Run tokens authenticate a sandbox job to the run API: an HMAC-signed run ID, organization and
// expiry, so a job can only read its own spec and source and report its own results.
package sandbox

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ErrInvalidRunToken wraps every reason VerifyRunToken refuses a token.
var ErrInvalidRunToken = errors.New("invalid run token")

type runTokenClaims struct {
	Run     uuid.UUID `json:"run"`
	Org     uuid.UUID `json:"org"`
	Expires int64     `json:"exp"`
}

// SignRunToken returns a token for run in org that expires at exp:
// base64url(json) "." base64url(HMAC-SHA256 of the first part).
func SignRunToken(key []byte, run, org uuid.UUID, exp time.Time) string {
	data, _ := json.Marshal(runTokenClaims{Run: run, Org: org, Expires: exp.Unix()})
	payload := base64.RawURLEncoding.EncodeToString(data)
	return payload + "." + base64.RawURLEncoding.EncodeToString(runTokenMAC(key, payload))
}

// VerifyRunToken checks token's signature and expiry at now, and returns its run and organization.
func VerifyRunToken(key []byte, token string, now time.Time) (uuid.UUID, uuid.UUID, error) {
	payload, encodedSignature, ok := strings.Cut(token, ".")
	if !ok {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: run token is malformed", ErrInvalidRunToken)
	}
	signature, err := base64.RawURLEncoding.DecodeString(encodedSignature)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: run token is malformed", ErrInvalidRunToken)
	}
	if !hmac.Equal(signature, runTokenMAC(key, payload)) {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: run token signature doesn't match", ErrInvalidRunToken)
	}
	data, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: run token is malformed", ErrInvalidRunToken)
	}
	var claims runTokenClaims
	if err := json.Unmarshal(data, &claims); err != nil {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: run token is malformed", ErrInvalidRunToken)
	}
	if expires := time.Unix(claims.Expires, 0).UTC(); !now.Before(expires) {
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: run token expired at %s", ErrInvalidRunToken, expires.Format(time.RFC3339))
	}
	return claims.Run, claims.Org, nil
}

func runTokenMAC(key []byte, payload string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	return mac.Sum(nil)
}
