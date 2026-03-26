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
// # Custom Columns as Template Variables
//
// Unrecognized table column headers become template variables usable in exec, args, and build.
// Custom keys in YAML frontmatter provide global defaults, overridable per-row.
//
// Prefer markdown tables over command blocks unless commands are multi-line or need
// per-test setup/teardown. Tables are more compact and easier to scan.
//
//	---
//	exec: bash
//	args: ["-c", "curl {{.flags}} {{.baseUrl}}{{.path}}"]
//	baseUrl: https://api.example.com
//	flags: "-s"
//	---
//
//	| Name | path | CEL Validation |
//	|------|------|----------------|
//	| get users | /users | json.size() > 0 |
//	| get health | /health | json.status == "ok" |
//
// Priority order (highest to lowest): TemplateVars (file expansion) > Properties (table columns) > Metadata (frontmatter)
//
// # Command Block Format
//
// Define tests with command blocks and frontmatter.
// Use command blocks only when tests need multi-line scripts, setup/teardown,
// or per-test YAML config that tables cannot express:
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
//	cwd: ./testdir
//	timeout: 30s
//	---
//
// # Working Directory (CWD) Resolution
//
// The working directory for test execution is resolved with the following priority:
//
//  1. Test-level CWD (per-test frontmatter block or table "cwd"/"dir"/"working directory" column)
//  2. File-level CWD (YAML front-matter at top of fixture file)
//  3. SourceDir (directory containing the fixture markdown file)
//  4. Runner WorkDir (passed via RunOptions)
//
// Relative CWD paths are resolved from SourceDir (the fixture file's directory).
// Absolute CWD paths are used directly.
//
// Environment variables from frontmatter and per-test config are passed to the
// executed command.
//
// # Auto-Injected Root Directory Variables
//
// The following variables are automatically computed from the working directory
// and available as both gomplate template variables (e.g. {{.GIT_ROOT_DIR}}) and
// environment variables in executed commands:
//
//   - GIT_ROOT_DIR: nearest parent directory containing .git
//   - GO_ROOT_DIR: nearest parent directory containing go.mod
//   - ROOT_DIR: GIT_ROOT_DIR if available, else GO_ROOT_DIR, else working directory
//
// User-defined env vars with the same name take precedence over auto-injected values.
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
