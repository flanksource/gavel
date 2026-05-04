package main

import "testing"

func TestGavelMCPProfileExcludesLintAndTestSubcommands(t *testing.T) {
	excluded := gavelMCPExcludedCommands()
	for _, want := range []string{"^lint ", "^test "} {
		if !containsString(excluded, want) {
			t.Fatalf("missing MCP exclude %q in %v", want, excluded)
		}
	}
}

func TestGavelMCPProfileIgnoresWorkflowOnlyParams(t *testing.T) {
	ignored := gavelMCPGlobalIgnoredParams()
	for _, want := range []string{
		"--ui",
		"--triage",
		"--sync-todos",
		"--todo-template",
		"--todos-dir",
		"--baseline",
		"--work-dir",
	} {
		if !containsString(ignored, want) {
			t.Fatalf("missing MCP ignored param %q in %v", want, ignored)
		}
	}
}

func TestGavelMCPProfileIgnoresTestRunnerParams(t *testing.T) {
	ignored := gavelMCPTestIgnoredParams()
	for _, want := range []string{
		"addr",
		"bench",
		"detach",
		"diagnostics",
		"fixture-files",
		"fixtures",
		"idle-timeout",
		"lint",
		"lint-timeout",
		"nodes",
		"recursive",
		"skip-hooks",
	} {
		if !containsString(ignored, want) {
			t.Fatalf("missing test MCP ignored param %q in %v", want, ignored)
		}
	}
}

func TestGavelMCPProfileIgnoresCommitWorkflowParams(t *testing.T) {
	ignored := gavelMCPCommitIgnoredParams()
	for _, want := range []string{
		"force",
		"interactive",
		"lint",
		"lint-secrets",
		"max-files",
		"max-lines",
		"model",
		"precommit",
		"push",
		"tree",
	} {
		if !containsString(ignored, want) {
			t.Fatalf("missing commit MCP ignored param %q in %v", want, ignored)
		}
	}
}

func TestGavelMCPProfileIgnoresPRListDashboardParams(t *testing.T) {
	ignored := gavelMCPPRListIgnoredParams()
	for _, want := range []string{
		"all",
		"interval",
		"menu-bar",
		"persist-port",
		"port",
	} {
		if !containsString(ignored, want) {
			t.Fatalf("missing pr list MCP ignored param %q in %v", want, ignored)
		}
	}
}

func TestGavelMCPProfileIgnoresPRStatusFormatParams(t *testing.T) {
	ignored := gavelMCPPRStatusIgnoredParams()
	for _, want := range []string{
		"csv",
		"dump-schema",
		"filter",
		"html",
		"json",
		"markdown",
		"output",
		"pdf",
		"pretty",
		"slack",
		"table",
		"tree",
		"yaml",
		"interval",
	} {
		if !containsString(ignored, want) {
			t.Fatalf("missing pr status MCP ignored param %q in %v", want, ignored)
		}
	}
}

func containsString(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
