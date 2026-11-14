package testui

const (
	TestTaskGroupName = "Running tests across packages"
	LintTaskGroupName = "Running linters"
	// PromptLintTaskGroupName holds one task per `gavel lint prompt` review run.
	PromptLintTaskGroupName = "lint prompt"
)

// liveTaskGroupNames are the clicky groups the UI surfaces as virtual task tests.
var liveTaskGroupNames = []string{TestTaskGroupName, LintTaskGroupName, PromptLintTaskGroupName}
