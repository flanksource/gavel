// Package jsonb prepares encoded JSON for a Postgres jsonb column.
package jsonb

import "bytes"

// nullEscape is how encoding/json renders U+0000: the six characters
// backslash, 'u', '0', '0', '0', '0'. Spelled byte by byte because writing it
// as a string literal in Go source is itself an escape. The JSON spec fixes the
// form — lowercase "u", four hex digits — and encoding/json never emits a raw
// NUL byte inside a string literal, so this is the only shape it can take.
var nullEscape = []byte{'\\', 'u', '0', '0', '0', '0'}

// StripNullEscapes removes every U+0000 escape from an encoded JSON document.
//
// Postgres accepts the escape in a json value but not in jsonb, because jsonb
// decodes escapes into text and PostgreSQL text cannot hold a NUL byte. Writing
// one fails the whole statement with SQLSTATE 22P05, "unsupported Unicode
// escape sequence". Task snapshots carry raw subprocess stdout/stderr, where a
// stray NUL is ordinary — binary output, a partial write, a tool padding its
// buffer — so the byte has to go somewhere before it reaches the column.
// Dropping it is the only option available: jsonb genuinely cannot represent
// it, and the surrounding output is worth keeping.
//
// The scan tracks string state and consumes each backslash escape as a unit.
// That is the whole reason this is not a bytes.Replace: a source string holding
// those six characters as literal text marshals to a doubled backslash followed
// by u0000, whose trailing six bytes match the pattern. Replacing them blind
// would leave a dangling backslash and turn a valid document into a parse error.
func StripNullEscapes(encoded []byte) []byte {
	if !bytes.Contains(encoded, nullEscape) {
		return encoded
	}
	stripped := make([]byte, 0, len(encoded))
	inString := false
	for i := 0; i < len(encoded); {
		current := encoded[i]
		if !inString {
			inString = current == '"'
			stripped = append(stripped, current)
			i++
			continue
		}
		if current == '\\' {
			if bytes.HasPrefix(encoded[i:], nullEscape) {
				i += len(nullEscape)
				continue
			}
			// Copy the backslash together with the character it escapes, so an
			// escaped backslash cannot be re-read as the start of an escape.
			stripped = append(stripped, current)
			i++
			if i < len(encoded) {
				stripped = append(stripped, encoded[i])
				i++
			}
			continue
		}
		if current == '"' {
			inString = false
		}
		stripped = append(stripped, current)
		i++
	}
	return stripped
}
