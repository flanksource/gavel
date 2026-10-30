package lifecycle

import "fmt"

type Selection struct {
	HeadSHA    string `json:"headSha"`
	Definition string `json:"definition"`
	Step       string `json:"step"`
	EventID    string `json:"eventId"`
}

func SelectAction(snapshot Snapshot, actions []Action, definition string, selection Selection) (Action, error) {
	if selection.HeadSHA == "" || selection.HeadSHA != snapshot.PR.HeadSHA {
		return Action{}, fmt.Errorf("PR head changed; refresh the action preview")
	}
	if selection.Definition == "" || selection.Definition != definition {
		return Action{}, fmt.Errorf("lifecycle definition changed; refresh the action preview")
	}
	for _, action := range actions {
		if action.Step.Name == selection.Step && action.Event.ID == selection.EventID {
			return action, nil
		}
	}
	return Action{}, fmt.Errorf("selected lifecycle context no longer applies; refresh the action preview")
}
