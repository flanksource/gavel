package model

type CreatePRInput struct {
	Title string
	Body  string
	Head  string
	Base  string
	Draft bool
}

type CreatePRResult struct {
	Number int    `json:"number"`
	URL    string `json:"html_url"`
	Title  string `json:"title"`
	State  string `json:"state"`
	NodeID string `json:"node_id"`
	Base   string `json:"-"`
}
