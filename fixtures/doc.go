// Package fixtures provides a comprehensive testing framework for AST queries and CLI commands
// using markdown-based test definitions with CEL (Common Expression Language) validation.
//
// The fixtures package enables declarative testing through two formats:
//   - Markdown tables (traditional format)
//   - Command blocks (expressive format with frontmatter)
//
// # Markdown Table Format
//
// Define tests in markdown tables:
//
//	| Test Name | Query | Expected Count | CEL Validation |
//	|-----------|-------|----------------|----------------|
//	| Find Controllers | *Controller* | 2 | nodes.all(n, n.type_name.endsWith("Controller")) |
//	| Complex Methods | cyclomatic(*) > 10 | - | nodes.exists(n, n.cyclomatic_complexity > 15) |
//
// # Command Block Format
//
// Define tests with command blocks and frontmatter:
//
//	### command: json output validation
//
//	```bash
//	ast * --format json
//	```
//
//	```frontmatter
//	cwd: ./examples
//	exitCode: 0
//	env:
//	  DEBUG: "true"
//	```
//
//	Validations:
//	* cel: stdout.contains("node_type")
//	* cel: json.length > 0
//	* not contains: ERROR
//
// # CEL Validation
//
// Fixtures use CEL (Common Expression Language) for validation:
//
//	// Validate all nodes match criteria
//	nodes.all(n, n.cyclomatic_complexity < 10)
//
//	// Check for specific properties
//	stdout.contains("node_type") && exitCode == 0
//
//	// Complex conditions
//	nodes.filter(n, n.node_type == "method").size() > 5
//
// # Running Fixtures
//
// Run fixture files programmatically:
//
//	runner := fixtures.NewRunner()
//	results, _ := runner.RunFixtureFile("tests/ast_queries.md")
//	for _, result := range results {
//	    fmt.Printf("%s: %s\n", result.Name, result.Status)
//	}
//
// # YAML Front-matter
//
// Configure fixture execution with YAML front-matter:
//
//	---
//	build: "go build -o myapp"
//	exec: "./myapp"
//	env:
//	  LOG_LEVEL: "debug"
//	timeout: 30s
//	---
//
// # Validation Types
//
//   - cel: Full CEL expression validation
//   - contains: Simple string contains check
//   - regex: Regular expression matching
//   - not: Negation of any validation type
//   - json: JSON path validation
//
// # File Expansion
//
// Fixtures support glob patterns to run tests across multiple files:
//
//	---
//	files: "**/*.go"
//	---
//
// This creates one test per matching file with template variables:
//   - {file}: Relative file path
//   - {filename}: Filename without extension
//   - {dir}: Directory containing file
//
// See also: github.com/flanksource/gavel/cmd for CLI integration
package fixtures
