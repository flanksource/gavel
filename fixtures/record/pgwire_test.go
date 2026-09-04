package record

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sniffed struct {
	typ  byte
	body string
}

// wireMessage builds one framed protocol message: type byte, big-endian length
// covering the length field itself, then the body.
func wireMessage(typ byte, body string) []byte {
	msg := make([]byte, 5+len(body))
	msg[0] = typ
	binary.BigEndian.PutUint32(msg[1:5], uint32(4+len(body)))
	copy(msg[5:], body)
	return msg
}

func collect(seen *[]sniffed) *pgSniffer {
	return &pgSniffer{onMessage: func(typ byte, body []byte) {
		*seen = append(*seen, sniffed{typ: typ, body: string(body)})
	}}
}

func TestSnifferFramesMessagesAcrossChunkBoundaries(t *testing.T) {
	stream := append(wireMessage('Q', "SELECT 1"), wireMessage('Q', "SELECT 2")...)

	// A TCP read boundary lands anywhere, including mid-header, so the sniffer
	// has to reassemble rather than assume one write is one message.
	for _, chunk := range []int{1, 3, 7, len(stream)} {
		t.Run("chunk", func(t *testing.T) {
			var seen []sniffed
			sniffer := collect(&seen)

			for offset := 0; offset < len(stream); offset += chunk {
				end := min(offset+chunk, len(stream))
				n, err := sniffer.Write(stream[offset:end])
				require.NoError(t, err)
				assert.Equal(t, end-offset, n, "a sniffer must report every byte written or io.Copy stops")
			}

			assert.Equal(t, []sniffed{{'Q', "SELECT 1"}, {'Q', "SELECT 2"}}, seen)
		})
	}
}

func TestSnifferSkipsAnOversizedMessageAndResumes(t *testing.T) {
	var seen []sniffed
	sniffer := collect(&seen)

	oversized := make([]byte, 5+maxWireMessage)
	oversized[0] = 'd' // CopyData
	binary.BigEndian.PutUint32(oversized[1:5], uint32(4+maxWireMessage))

	// Split so the skip spans several writes, which is what a real multi-megabyte
	// COPY payload does.
	half := len(oversized) / 2
	_, err := sniffer.Write(oversized[:half])
	require.NoError(t, err)
	_, err = sniffer.Write(oversized[half:])
	require.NoError(t, err)
	_, err = sniffer.Write(wireMessage('Q', "SELECT 1"))
	require.NoError(t, err)

	assert.Equal(t, []sniffed{{'Q', "SELECT 1"}}, seen,
		"the oversized message is dropped from the capture but the stream stays framed")
}

func TestSnifferStopsOnAnImpossibleLength(t *testing.T) {
	var seen []sniffed
	sniffer := collect(&seen)

	// A length below the 4 bytes it counts itself means the stream is no longer
	// at a frame boundary; emitting anything after that would be garbage.
	broken := []byte{'Q', 0, 0, 0, 1, 'x', 'y'}
	_, err := sniffer.Write(broken)
	require.NoError(t, err)

	assert.Empty(t, seen)
}

func TestSessionRecordsPendingWorkOnReadyForQuery(t *testing.T) {
	recorder := &SQLRecorder{}
	session := newPGSession(recorder, false)

	session.frontend('Q', []byte("SELECT 1\x00"))
	// No CommandComplete: a statement inside a failed batch never gets its own
	// answer, and dropping it would hide it from the capture entirely.
	session.backend('Z', []byte{'I'})

	statements := recorder.Statements()
	require.Len(t, statements, 1)
	assert.Equal(t, "SELECT 1", statements[0].SQL)
	assert.Equal(t, -1, statements[0].Rows)
}

func TestSessionIgnoresAnUnmatchedCompletion(t *testing.T) {
	recorder := &SQLRecorder{}
	session := newPGSession(recorder, false)

	// The server answers messages the sniffer never saw a request for — a
	// ParameterStatus burst at startup, say. Nothing to complete, nothing to add.
	session.backend('C', []byte("SELECT 1\x00"))

	assert.Empty(t, recorder.Statements())
}
