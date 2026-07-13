package main

import "testing"

func TestLegacyFileBackedTODOSyncCLISurfacesAreRemoved(t *testing.T) {
	for _, command := range prCmd.Commands() {
		if command.Name() == "fix" {
			t.Fatal("retired gavel pr fix command is still registered")
		}
	}

	for _, test := range []struct {
		path  []string
		flags []string
	}{
		{path: []string{"pr", "status"}, flags: []string{"sync-todos"}},
		{path: []string{"lint"}, flags: []string{"sync-todos", "group-by"}},
		{path: []string{"test"}, flags: []string{"sync-todos", "todos-dir", "todo-template"}},
	} {
		command, _, err := rootCmd.Find(test.path)
		if err != nil {
			t.Fatalf("find %v: %v", test.path, err)
		}
		for _, flag := range test.flags {
			if command.Flags().Lookup(flag) != nil {
				t.Errorf("%v still exposes retired --%s", test.path, flag)
			}
		}
	}
}
