package testrunner

// Re-export types from parsers for backwards compatibility.
// Types are now defined in parsers package to avoid circular dependencies.
import (
	"github.com/flanksource/gavel/testrunner/parsers"
)

// Framework is a re-export from parsers package.
type Framework = parsers.Framework

// TestFailure is a re-export from parsers package.
type TestFailure = parsers.Test

// ExecutionResult is a re-export from parsers package.
type ExecutionResult = parsers.ExecutionResult

// Re-export constants
const (
	GoTest = parsers.GoTest
	Ginkgo = parsers.Ginkgo
)
