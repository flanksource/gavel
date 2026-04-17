package tsc

import (
	"crypto/sha256"
	"encoding/hex"
)

func scriptHash(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:6])
}
