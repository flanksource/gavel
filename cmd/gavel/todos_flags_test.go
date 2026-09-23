package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The flag surface is what the CLI owns; the flag *semantics* are tested in
// todos/ops, which is where the logic lives now that every surface shares it.
func TestTodosLabelFlagsRegistered(t *testing.T) {
	for _, flag := range []string{"label", "clear-labels"} {
		assert.NotNil(t, todosEditCmd.Flags().Lookup(flag), "missing todos edit --%s", flag)
	}
	assert.NotNil(t, todosCreateCmd.Flags().Lookup("label"), "missing todos create --label")

	for _, flag := range []string{"color", "icon", "description", "global"} {
		assert.NotNil(t, todosLabelsSetCmd.Flags().Lookup(flag), "missing todos labels set --%s", flag)
	}
}
