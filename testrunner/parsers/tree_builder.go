package parsers

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/flanksource/clicky/api"
	"github.com/flanksource/commons/logger"
	"github.com/samber/lo"
)

// suiteBuilder is a helper struct for building suite hierarchies
type suiteBuilder struct {
	path     []string
	children map[string]*suiteBuilder
	tests    []Test
}

// BuildTestTree converts flat test results into a hierarchical tree structure
// using native Test.Children field. Tree hierarchy: directory > package > suite > test
func BuildTestTree(tests []Test) []Test {
	if len(tests) == 0 {
		return nil
	}

	// Build directory tree structure
	dirMap := make(map[string]*Test)
	rootDirs := []string{}

	for _, test := range tests {
		// Parse package path into directory hierarchy
		pkgPath := strings.TrimPrefix(test.PackagePath, "./")
		if pkgPath == "" {
			pkgPath = "."
		}
		parts := strings.Split(pkgPath, "/")

		// Build directory nodes
		var currentPath string
		for i, part := range parts {
			if i > 0 {
				currentPath += "/"
			}
			currentPath += part

			// Create directory node if it doesn't exist
			if dirMap[currentPath] == nil {
				dirMap[currentPath] = &Test{
					Name:        part + "/",
					PackagePath: currentPath,
					Children:    []Test{},
				}

				// Track root directories
				if i == 0 {
					rootDirs = append(rootDirs, currentPath)
				}
			}
		}

		// Add test to its leaf directory
		leafPath := currentPath
		if leafDir := dirMap[leafPath]; leafDir != nil {
			leafDir.Children = append(leafDir.Children, test)
		}
	}

	// Build suite hierarchies within each directory's test children
	// Store the suite-built children separately
	suiteTrees := make(map[string][]Test)
	for path, dir := range dirMap {
		suiteTrees[path] = buildSuiteTree(dir.Children)
	}

	// Clear all children before reconstruction
	for _, dir := range dirMap {
		dir.Children = []Test{}
	}

	// Reconstruct directory hierarchy by processing deepest paths first
	paths := make([]string, 0, len(dirMap))
	for path := range dirMap {
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool {
		// Sort by depth (number of slashes) descending
		depthI := strings.Count(paths[i], "/")
		depthJ := strings.Count(paths[j], "/")
		if depthI != depthJ {
			return depthI > depthJ
		}
		return paths[i] > paths[j]
	})

	for _, path := range paths {
		dir := dirMap[path]

		// Start with suite-built children, then add subdirectories
		// (suite trees contain package/suite nodes from tests, not subdirectories)
		newChildren := make([]Test, 0, len(suiteTrees[path])+len(dir.Children))
		newChildren = append(newChildren, suiteTrees[path]...)
		newChildren = append(newChildren, dir.Children...)
		dir.Children = newChildren

		// Add this directory to its parent
		lastSlash := strings.LastIndex(path, "/")
		if lastSlash > 0 {
			parentPath := path[:lastSlash]
			if parent, exists := dirMap[parentPath]; exists {
				parent.Children = append(parent.Children, *dir)
			}
		}
	}

	// Collect root level directories
	sort.Strings(rootDirs)
	result := make([]Test, 0, len(rootDirs))
	for _, path := range rootDirs {
		if node := dirMap[path]; node != nil {
			result = append(result, *node)
		}
	}

	// Propagate failure status from children to parents
	propagateFailureStatus(result)

	return result
}

// buildSuiteTree organizes tests by their File path hierarchy, then Suite hierarchy.
// It receives tests that belong to the same package directory (from BuildTestTree),
// so file paths are made relative to the common package path.
func buildSuiteTree(tests []Test) []Test {
	// Group tests by file
	byFile := make(map[string][]Test)
	withoutFile := []Test{}

	// Find common package path prefix to make file paths relative
	var commonPkgPath string
	for _, test := range tests {
		if test.PackagePath != "" {
			commonPkgPath = strings.TrimPrefix(test.PackagePath, "./")
			break
		}
	}

	for _, test := range tests {
		if test.File != "" {
			relFile := makeFileRelativeToPackage(test.File, commonPkgPath)
			byFile[relFile] = append(byFile[relFile], test)
		} else {
			withoutFile = append(withoutFile, test)
		}
	}

	if len(byFile) == 0 {
		return withoutFile
	}

	// Build directory tree from relative file paths
	dirMap := make(map[string]*Test)
	rootDirs := []string{}
	rootFiles := []Test{}

	// Sort file names for consistent ordering
	fileNames := make([]string, 0, len(byFile))
	for file := range byFile {
		fileNames = append(fileNames, file)
	}
	sort.Strings(fileNames)

	for _, relFilePath := range fileNames {
		fileTests := byFile[relFilePath]

		// Parse file path into directory + filename
		dir := filepath.Dir(relFilePath)
		filename := filepath.Base(relFilePath)

		// Build directory nodes for the file's path (relative to package)
		if dir != "." && dir != "" {
			parts := strings.Split(dir, "/")
			var currentPath string
			for i, part := range parts {
				if i > 0 {
					currentPath += "/"
				}
				currentPath += part

				if dirMap[currentPath] == nil {
					dirMap[currentPath] = &Test{
						Name:     part + "/",
						Children: []Test{},
					}
					if i == 0 {
						rootDirs = append(rootDirs, currentPath)
					}
				}
			}
		}

		// Build suite hierarchy for tests in this file
		var children []Test
		hasSuites := false
		for _, t := range fileTests {
			if len(t.Suite) > 0 {
				hasSuites = true
				break
			}
		}
		if hasSuites {
			children = buildSuiteHierarchy(fileTests)
		} else {
			children = fileTests
		}

		// Create file node - use original file path from first test
		originalFile := ""
		if len(fileTests) > 0 {
			originalFile = fileTests[0].File
		}

		// Clear File on children since parent file node shows it
		clearFileOnDescendants(children)

		fileNode := Test{
			Name:     filename,
			File:     originalFile,
			Children: children,
		}

		// Add file to its directory or to root
		if dir != "." && dir != "" {
			dirMap[dir].Children = append(dirMap[dir].Children, fileNode)
		} else {
			rootFiles = append(rootFiles, fileNode)
		}
	}

	// Reconstruct directory hierarchy by processing deepest paths first
	paths := make([]string, 0, len(dirMap))
	for path := range dirMap {
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool {
		depthI := strings.Count(paths[i], "/")
		depthJ := strings.Count(paths[j], "/")
		if depthI != depthJ {
			return depthI > depthJ
		}
		return paths[i] > paths[j]
	})

	for _, path := range paths {
		dir := dirMap[path]
		lastSlash := strings.LastIndex(path, "/")
		if lastSlash > 0 {
			parentPath := path[:lastSlash]
			if parent, exists := dirMap[parentPath]; exists {
				parent.Children = append(parent.Children, *dir)
			}
		}
	}

	// Collect root level directories and files
	rootDirs = lo.Uniq(rootDirs)
	sort.Strings(rootDirs)
	result := withoutFile
	for _, path := range rootDirs {
		if node := dirMap[path]; node != nil {
			result = append(result, *node)
		}
	}
	result = append(result, rootFiles...)

	return result
}

// buildSuiteHierarchy creates nested suite structure from Suite paths
func buildSuiteHierarchy(tests []Test) []Test {
	root := &suiteBuilder{children: make(map[string]*suiteBuilder)}

	// Build tree from suite paths
	for _, test := range tests {
		current := root
		for _, suiteName := range test.Suite {
			if current.children[suiteName] == nil {
				current.children[suiteName] = &suiteBuilder{
					path:     append(current.path, suiteName),
					children: make(map[string]*suiteBuilder),
				}
			}
			current = current.children[suiteName]
		}
		current.tests = append(current.tests, test)
	}

	// Convert to Test tree
	return suiteBuilderToTests(root)
}

// suiteBuilderToTests converts suiteBuilder tree to Test tree
func suiteBuilderToTests(builder *suiteBuilder) []Test {
	var result []Test

	// Get sorted child suite names
	suiteNames := make([]string, 0, len(builder.children))
	for name := range builder.children {
		suiteNames = append(suiteNames, name)
	}
	sort.Strings(suiteNames)

	// Process child suites
	for _, name := range suiteNames {
		child := builder.children[name]
		suiteNode := Test{
			Name:  name,
			Suite: child.path,
		}

		// Recursively build children
		suiteNode.Children = suiteBuilderToTests(child)

		// If this suite has no nested suites, the recursive call returns the tests
		// Otherwise, add tests after the nested suites
		if len(child.children) > 0 {
			// Has nested suites, append tests after them
			for _, test := range child.tests {
				suiteNode.Children = append(suiteNode.Children, test)
			}
		}

		result = append(result, suiteNode)
	}

	// Add tests without child suites (leaf level)
	if len(builder.children) == 0 {
		result = append(result, builder.tests...)
	}

	return result
}

// BuildTestTreeWithVerbosity builds a tree with different detail levels
// - verbosity 0: Directory tree with counts only
// - verbosity 1: Directory tree + Suite levels with counts
// - verbosity 2: Full tree including individual tests
func BuildTestTreeWithVerbosity(tests []Test, verbosity int) []Test {
	tree := BuildTestTree(tests)
	if verbosity >= 2 {
		return tree
	}
	return pruneTreeByVerbosity(tree, verbosity)
}

// pruneTreeByVerbosity removes details based on verbosity level
// At verbosity 0: Filters out passing tests, shows only failing/skipped tests
// At verbosity 1: Shows directory + suite structure with all test counts
// At verbosity 2: Shows full tree including individual tests
func pruneTreeByVerbosity(tests []Test, verbosity int) []Test {
	result := []Test{}

	for _, test := range tests {

		if !logger.V(1).Enabled() {
			result = append(result, test.Filter(TestFilter{
				ExcludePassed: true,
				ExcludeSpecs:  true,
				SlowTests:     lo.ToPtr(10 * time.Second),
			}))
		} else {
			result = append(result, test)
		}
	}
	return result
}

// isContainerNode checks if a test is a directory, file, or suite node
func isContainerNode(test Test) bool {
	return isDirectoryNode(test) || isFileNode(test) || isSuiteNode(test)
}

// isDirectoryNode checks if test represents a directory
func isDirectoryNode(test Test) bool {
	return strings.HasSuffix(test.Name, "/")
}

// isFileNode checks if test represents a source file
func isFileNode(test Test) bool {
	return test.File != "" && test.Name == filepath.Base(test.File) && len(test.Children) > 0
}

// isSuiteNode checks if test represents a suite container (not an individual test in a suite)
func isSuiteNode(test Test) bool {
	if len(test.Suite) == 0 || test.Name == "" || isDirectoryNode(test) || isFileNode(test) {
		return false
	}
	// A suite node's name matches the last element of its suite path
	// Individual tests have a suite path but their name is different
	return test.Name == test.Suite[len(test.Suite)-1]
}


// GetTestTreeAsTreeNodes converts Test tree to TreeNode slice
func GetTestTreeAsTreeNodes(tests []Test) []api.TreeNode {
	nodes := make([]api.TreeNode, len(tests))
	for i := range tests {
		nodes[i] = tests[i]
	}
	return nodes
}

// propagateFailureStatus traverses the tree bottom-up and sets Failed=true on parent nodes
// if any child has Failed=true.
func propagateFailureStatus(tests []Test) {
	for i := range tests {
		propagateFailureStatusRecursive(&tests[i])
	}
}

func propagateFailureStatusRecursive(test *Test) {
	if len(test.Children) == 0 {
		return
	}

	// Process children first (bottom-up)
	for i := range test.Children {
		propagateFailureStatusRecursive(&test.Children[i])
	}

	// Set parent failed if any child failed
	for _, child := range test.Children {
		if child.Failed {
			test.Failed = true
			break
		}
	}
}

// clearFileOnDescendants clears the File field on all descendant tests,
// keeping only the Line number so they display as `:line` instead of `file:line`.
func clearFileOnDescendants(tests []Test) {
	for i := range tests {
		tests[i].File = ""
		clearFileOnDescendants(tests[i].Children)
	}
}

// makeFileRelativeToPackage extracts the relative file path from an absolute or relative file path.
// It handles cases where:
// - filePath is absolute (e.g., /Users/moshe/project/pkg/test_test.go)
// - filePath is relative with ../ prefix (e.g., ../other/pkg/test_test.go)
// - filePath is relative and matches package path prefix
// - filePath only contains the filename
func makeFileRelativeToPackage(filePath, pkgPath string) string {
	if filePath == "" {
		return ""
	}

	// If file path is absolute or has ../ prefix, extract the part after the package path
	if filepath.IsAbs(filePath) || strings.HasPrefix(filePath, "..") {
		// Try to find the package path suffix in the file path
		if pkgPath != "" {
			// Look for the package path anywhere in the file path
			if idx := strings.Index(filePath, "/"+pkgPath+"/"); idx != -1 {
				return strings.TrimPrefix(filePath[idx+1:], pkgPath+"/")
			}
			if idx := strings.Index(filePath, "/"+pkgPath); idx != -1 && strings.HasSuffix(filePath, pkgPath) {
				return filepath.Base(filePath)
			}
		}
		// Fall back to just the filename
		return filepath.Base(filePath)
	}

	// For relative paths, try to strip the package path prefix
	if pkgPath != "" && strings.HasPrefix(filePath, pkgPath+"/") {
		return strings.TrimPrefix(filePath, pkgPath+"/")
	}
	if pkgPath != "" && strings.HasPrefix(filePath, pkgPath) {
		result := strings.TrimPrefix(filePath, pkgPath)
		result = strings.TrimPrefix(result, "/")
		if result == "" {
			return filepath.Base(filePath)
		}
		return result
	}

	// If no prefix matches, return as-is (already relative)
	if filePath == "" {
		return filepath.Base(filePath)
	}
	return filePath
}
