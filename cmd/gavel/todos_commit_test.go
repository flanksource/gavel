package main

import "testing"

func TestTodosRunCommitFlagRegistered(t *testing.T) {
	flag := todosRunCmd.Flags().Lookup("commit")
	if flag == nil {
		t.Fatal("expected todos run --commit flag to be registered")
	}
	if flag.DefValue != "true" {
		t.Fatalf("expected --commit default true, got %q", flag.DefValue)
	}
}
