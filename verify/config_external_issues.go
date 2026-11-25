package verify

type ExternalIssuesConfig struct {
	Jira JiraConfig `yaml:"jira,omitempty" json:"jira,omitempty"`
}

type JiraConfig struct {
	IgnoreFields []string `yaml:"ignoreFields,omitempty" json:"ignoreFields,omitempty" merge:"append,unique"`
}
