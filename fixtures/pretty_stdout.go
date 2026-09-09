package fixtures

import (
	"encoding/json"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/clicky/formatters"
)

type clickyOutputEnvelope struct {
	Pretty      json.RawMessage `json:"pretty"`
	TracePretty json.RawMessage `json:"tracePretty"`
}

func prettyFixtureStdout(output string) api.Textable {
	return api.ANSIText{Content: fixtureStdoutContent(output)}
}

func fixtureStdoutContent(output string) string {
	var envelope clickyOutputEnvelope
	if err := json.Unmarshal([]byte(output), &envelope); err == nil {
		for _, raw := range []json.RawMessage{envelope.Pretty, envelope.TracePretty} {
			var document formatters.ClickyDocument
			if len(raw) > 0 && json.Unmarshal(raw, &document) == nil && document.Version == 1 && document.Node.Plain != "" {
				return document.Node.Plain
			}
		}
	}
	return output
}
