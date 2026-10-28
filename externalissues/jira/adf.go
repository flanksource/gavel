package jira

import "encoding/json"

type ADFNode struct {
	Type    string          `json:"type"`
	Version int             `json:"version,omitempty"`
	Text    string          `json:"text,omitempty"`
	Attrs   json.RawMessage `json:"attrs,omitempty"`
	Marks   []ADFMark       `json:"marks,omitempty"`
	Content []ADFNode       `json:"content,omitempty"`
}

type ADFMark struct {
	Type  string          `json:"type"`
	Attrs json.RawMessage `json:"attrs,omitempty"`
}

type adfAttrs struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Text       string `json:"text"`
	Alt        string `json:"alt"`
	URL        string `json:"url"`
	Href       string `json:"href"`
	Collection string `json:"collection"`
	Language   string `json:"language"`
	ShortName  string `json:"shortName"`
	Level      int    `json:"level"`
}
