package parsers

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

// RenderDetail converts provider-owned Test.Detail values into a displayable
// clicky text tree without changing how those values marshal to JSON.
func RenderDetail(detail any) api.Text {
	switch d := detail.(type) {
	case nil:
		return api.Text{}
	case api.Pretty:
		return d.Pretty()
	case api.Text:
		return d
	case api.Textable:
		return clicky.Text("").Add(d)
	case string:
		return clicky.Text(d)
	case json.RawMessage:
		return clicky.Text("").Add(api.CodeBlock("json", string(d)))
	default:
		raw, err := json.MarshalIndent(d, "", "  ")
		if err != nil {
			return clicky.Text(fmt.Sprintf("%v", d))
		}
		return clicky.Text("").Add(api.CodeBlock("json", string(raw)))
	}
}
