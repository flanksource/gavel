package azuredevops

import "time"

type repositoryInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"defaultBranch"`
	Project       struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
}

type commitRef struct {
	ID string `json:"commitId"`
}

type pullRequest struct {
	ID           int            `json:"pullRequestId"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Status       string         `json:"status"`
	Draft        bool           `json:"isDraft"`
	Source       string         `json:"sourceRefName"`
	Target       string         `json:"targetRefName"`
	MergeStatus  string         `json:"mergeStatus"`
	MergeFailure string         `json:"mergeFailureMessage"`
	MergeCommit  commitRef      `json:"lastMergeCommit"`
	SourceCommit commitRef      `json:"lastMergeSourceCommit"`
	TargetCommit commitRef      `json:"lastMergeTargetCommit"`
	Repository   repositoryInfo `json:"repository"`
	CreatedBy    struct {
		ID     string `json:"id"`
		Name   string `json:"displayName"`
		Login  string `json:"uniqueName"`
		Avatar string `json:"imageUrl"`
	} `json:"createdBy"`
}

type policyEvaluation struct {
	ID            string `json:"evaluationId"`
	Status        string `json:"status"`
	Configuration struct {
		ID       int  `json:"id"`
		Enabled  bool `json:"isEnabled"`
		Blocking bool `json:"isBlocking"`
		Deleted  bool `json:"isDeleted"`
		Type     struct {
			ID   string `json:"id"`
			Name string `json:"displayName"`
		} `json:"type"`
	} `json:"configuration"`
}

type build struct {
	ID            int64     `json:"id"`
	Number        string    `json:"buildNumber"`
	Status        string    `json:"status"`
	Result        string    `json:"result"`
	SourceBranch  string    `json:"sourceBranch"`
	SourceVersion string    `json:"sourceVersion"`
	QueueTime     time.Time `json:"queueTime"`
	Definition    struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"definition"`
}

type timelineRecord struct {
	ID         string        `json:"id"`
	ParentID   string        `json:"parentId"`
	Identifier string        `json:"identifier"`
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	State      string        `json:"state"`
	Result     string        `json:"result"`
	Order      int           `json:"order"`
	Attempt    int           `json:"attempt"`
	Started    time.Time     `json:"startTime"`
	Finished   time.Time     `json:"finishTime"`
	Log        *logReference `json:"log"`
}

type logReference struct {
	ID    int64 `json:"id"`
	Lines int64 `json:"lineCount"`
}
