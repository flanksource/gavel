package model

type MergeReadiness struct {
	State            string   `json:"state"`
	Reasons          []string `json:"reasons,omitempty"`
	WaitingForBuilds bool     `json:"-"`
	WaitingForMerge  bool     `json:"-"`
}

type StatusSnapshot struct {
	PR   *PRInfo
	Runs map[int64]*WorkflowRun
}

type StatusOptions struct {
	Logs     bool
	TailLogs int
}
