package jira

import (
	"encoding/json"

	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

func semanticFieldObject(values map[string]any) (fieldView, bool, error) {
	var text api.Textable
	if _, ok := values["displayName"].(string); ok {
		data, err := json.Marshal(values)
		if err != nil {
			return fieldView{}, true, err
		}
		var person personPayload
		if err := json.Unmarshal(data, &person); err != nil {
			return fieldView{}, true, err
		}
		text = person.person().PrettyShort()
	} else if name, ok := values["name"].(string); ok {
		color, colored := values["colorName"].(string)
		if category, ok := values["statusCategory"].(map[string]any); ok {
			color, colored = category["colorName"].(string)
		}
		if colored {
			classes := "bg-slate-200 text-slate-800"
			switch color {
			case "yellow":
				classes = "bg-amber-100 text-amber-800"
			case "green":
				classes = "bg-green-100 text-green-800"
			case "blue":
				classes = "bg-blue-100 text-blue-800"
			}
			text = clicky.Badge(name, classes)
		}
	}
	if text == nil {
		return fieldView{}, false, nil
	}
	return fieldView{short: text, full: text}, true, nil
}
