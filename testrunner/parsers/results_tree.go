package parsers

import "strings"

func (results TestSuiteResults) Tree() []Test {
	tree := BuildTestTree(results.All())
	for _, result := range results {
		if len(result.Tests) == 0 || result.Stdout == "" && result.Stderr == "" {
			continue
		}
		path := strings.TrimPrefix(result.Tests[0].PackagePath, "./")
		if path == "" {
			path = "."
		}
		attachPackageOutput(tree, path, result)
	}
	return tree
}

func attachPackageOutput(tests Tests, path string, result TestResults) {
	for i := range tests {
		if tests[i].PackagePath == path && strings.HasSuffix(tests[i].Name, "/") {
			tests[i].Command = strings.TrimSpace(tests[i].Command + "\n" + result.Command)
			tests[i].Stdout += result.Stdout
			tests[i].Stderr += result.Stderr
			return
		}
		attachPackageOutput(tests[i].Children, path, result)
	}
}
