package externalissues

import (
	"github.com/flanksource/clicky"
	"github.com/flanksource/clicky/api"
)

func (p Person) PrettyShort() api.Textable {
	label := p.Name
	if label == "" {
		label = p.ID
	}
	return clicky.Link(p.URL).Add(clicky.Avatar(label, p.AvatarURL))
}

func (s Status) PrettyShort() api.Textable {
	if s.Category == "" {
		return s.Reference.PrettyShort()
	}
	classes := "bg-slate-200 text-slate-800"
	switch s.Category {
	case StatusInProgress:
		classes = "bg-blue-100 text-blue-800"
	case StatusDone:
		classes = "bg-green-100 text-green-800"
	}
	return clicky.Badge(s.Name, classes)
}

func (r Reference) PrettyShort() api.Textable {
	label := r.Name
	if label == "" {
		label = r.Key
	}
	if label == "" {
		label = r.ID
	}
	return clicky.Link(r.URL).AddText(label)
}
