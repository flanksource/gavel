package main

import (
	"strings"
	"testing"

	"github.com/flanksource/gavel/testrunner"
)

func TestLegacyFileBackedTODOCLIPathsFailBeforeRuntimeWork(t *testing.T) {
	tests := []struct {
		name string
		run  func() error
	}{
		{name: "pr fix", run: func() error { return runPRFix(prFixCmd, nil) }},
		{name: "pr status", run: func() error {
			_, err := runPRStatus(PRStatusOptions{SyncTodos: ".todos"})
			return err
		}},
		{name: "lint", run: func() error {
			_, err := runLint(LintOptions{SyncTodos: ".todos"})
			return err
		}},
		{name: "test", run: func() error {
			_, err := runTests(testrunner.RunOptions{SyncTodos: true})
			return err
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			if err == nil {
				t.Fatal("expected retired runtime error")
			}
			for _, want := range []string{"retired", "PostgreSQL", "gavel todos create", "gavel todos sync", "explicit import/export", "gavel todos import-grite"} {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error = %q, want %q", err, want)
				}
			}
		})
	}
}
