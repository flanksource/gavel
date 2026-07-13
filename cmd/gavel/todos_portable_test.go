package main

import (
	"testing"

	"github.com/flanksource/gavel/todos/portable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortableTodoCommandsAreExplicitAndFileScoped(t *testing.T) {
	assert.Same(t, todosCmd, todosImportCmd.Parent())
	assert.Same(t, todosCmd, todosExportCmd.Parent())
	require.NotNil(t, todosImportCmd.Flags().Lookup("dir"))
	assert.Equal(t, portable.DefaultDirectory, todosImportCmd.Flags().Lookup("dir").DefValue)
	require.NotNil(t, todosExportCmd.Flags().Lookup("dir"))
	require.NotNil(t, todosExportCmd.Flags().Lookup("force"))
	assert.Nil(t, todosImportCmd.Flags().Lookup("provider"), "import must not select a runtime provider")
	assert.Nil(t, todosExportCmd.Flags().Lookup("provider"), "export must not select a runtime provider")
}
