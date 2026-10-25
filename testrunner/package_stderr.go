package testrunner

import (
	"strings"

	"github.com/flanksource/gavel/testrunner/parsers"
)

// attachFailedPackageStderr sets each failed package's process stderr on that
// package's node in the tree, once. A failed test reporting stderr of its own
// ("Failed to import test file …") never gets the package stderr, which is
// where the process explains it (a bundler reload, a crash). A package whose
// failed test already carries that stderr is left alone.
func attachFailedPackageStderr(tree []parsers.Test, results parsers.TestSuiteResults) []parsers.Test {
	byPackage := map[string]string{}
	for _, suite := range results {
		stderr := stripExitStatus(strings.TrimSpace(suite.Stderr))
		if sum := suite.Tests.Sum(); stderr == "" || sum.Failed+sum.TimedOut == 0 {
			continue
		}
		if hasFailedTestWithStderr(suite.Tests, stderr) {
			continue
		}
		pkg := treePackagePath(suite.Tests[0].PackagePath)
		byPackage[pkg] = strings.TrimSpace(byPackage[pkg] + "\n" + stderr)
	}
	if len(byPackage) == 0 {
		return tree
	}
	var walk func(nodes []parsers.Test)
	walk = func(nodes []parsers.Test) {
		for i := range nodes {
			if len(nodes[i].Children) == 0 {
				continue
			}
			if stderr, ok := byPackage[nodes[i].PackagePath]; ok && strings.HasSuffix(nodes[i].Name, "/") {
				nodes[i].Stderr = stderr
			}
			walk(nodes[i].Children)
		}
	}
	walk(tree)
	return tree
}

func hasFailedTestWithStderr(tests parsers.Tests, stderr string) bool {
	for _, test := range tests {
		if test.Failed && strings.TrimSpace(test.Stderr) == stderr {
			return true
		}
		if hasFailedTestWithStderr(test.Children, stderr) {
			return true
		}
	}
	return false
}

// treePackagePath is the PackagePath parsers.BuildTestTree gives a package's
// directory node.
func treePackagePath(pkgPath string) string {
	if pkg := strings.TrimPrefix(pkgPath, "./"); pkg != "" {
		return pkg
	}
	return "."
}
