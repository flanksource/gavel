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
//	import _ "github.com/flanksource/gavel/fixtures/types"
//
//	runner, err := fixtures.NewRunner(fixtures.RunnerOptions{
//	    Paths: []string{"tests/ast_queries.md"},
//	})
//	if err != nil {
//	    t.Fatal(err)
//	}
//	runner.RunTesting(t)
//
// RunTesting mirrors markdown files, sections, tables, and table rows as nested
// t.Run nodes when passed a *testing.T. Ginkgo specs can use runner.RunGinkgo()
// for aggregate reporting or runner.RegisterGinkgoSpecs() during spec
// registration to mirror fixture nodes as Describe/It nodes. Gomega users can
// call runner.RunGomega(gomega.NewWithT(t)).
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
//	terminal: pty
//	timeout: 30s
//	os: linux
//	arch: amd64
//	skip: "! command -v docker"
//	---
//
// The fields terminal, os, arch, and skip can also be set per-test via
// table columns or command block YAML frontmatter (test-level overrides file-level).
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
// # Auto-Injected Variables
//
// The following variables are automatically available using shell-style $VAR
// syntax (e.g. $GIT_ROOT_DIR), Go template syntax (e.g. {{.GIT_ROOT_DIR}}),
// and as environment variables in executed commands:
//
//   - GIT_ROOT_DIR: nearest parent directory containing .git
//   - GO_ROOT_DIR: nearest parent directory containing go.mod
//   - ROOT_DIR: GIT_ROOT_DIR if available, else GO_ROOT_DIR, else working directory
//   - CWD: resolved working directory for the test
//   - GOOS: Go runtime OS (e.g. linux, darwin)
//   - GOARCH: Go runtime architecture (e.g. amd64, arm64)
//   - GOPATH: Go workspace path from environment
//
// Shell-style $VAR references to unknown variables pass through unchanged,
// so commands like "echo $HOME" work as expected in bash execution.
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
// # Params, Forms and if: Gates
//
// Front-matter (or a YAML fixture's top level) may declare params: as an
// object JSON Schema. It is lowered to a leading form step whose id is
// "params". Values from --param k=v, --params file.yaml or the UI start request
// are typed by the schema and filled with its defaults; when they satisfy it the
// step resolves without asking, otherwise it asks with them pre-filled.
//
// A kind: form step asks a person through the run's Prompter (the --ui
// server's prompt manager, a TTY form, or a host's) and blocks until they answer:
//
//	steps:
//	  - id: approval
//	    kind: form
//	    title: Purge {{.inspect.rowCount}} rows?
//	    schema: {type: object, required: [expected], properties: {expected: {type: integer}}}
//
// Its outputs are approved (false when cancelled or rejected, which does not
// fail the step), values (also spread at the top level), answeredBy and
// answeredAt. A property with format: connection (optionally
// x-connection-types: [sql_server]) is a picker over the connection catalog
// that commits the connection name. Without a prompter a form that must ask
// fails, naming the missing keys.
//
// Every completed step's outputs are bound at the root under its id, beside
// steps.<id>: approval.approved in CEL, {{.approval.expected}} in templates.
// An id used as a root must be a CEL identifier and may not shadow env, steps,
// outputs, fixture or tmp. A skipped step binds {skipped: true, reason}, so
// `false && approval.approved` short-circuits while a bare read fails.
//
// if: on any step is a CEL gate evaluated before the step renders. False skips
// the step ("skipped (if: <expr> = false)") without failing the fixture; an
// expression that does not evaluate errors it. Markdown rows take an If column.
//
// # Query Steps and Findings
//
// A kind: query step runs a commons-db query against a connection:
//
//	steps:
//	  - id: inspect
//	    kind: query
//	    connection: "{{.params.db}}"        # a name, connection://ref or inline spec
//	    query: SELECT guid, size FROM activities WHERE size >= {{.params.min}}
//	    params: {min: 20}                   # bound as driver arguments
//
// query is the text, an inline profile body ({query, params, columns, …}) or
// {profile: <name>, params} from ./.profiles, ~/.config/gavel/profiles or the
// host's store. connection is a catalog name or connection:// reference (the
// host's resolver, then the gavel layers and ./.connections), or an inline
// connection whose valueFrom references are hydrated. Outputs are rows, row,
// rowCount, columns and connection (name and type only).
//
// findings: turns rows into lint violations (models.Violation): either a CEL
// expression returning a list of maps, written over outputs.rows because
// gomplate types the roots as any —
//
//	findings: >-
//	  outputs.rows.filter(r, r.size >= 20).map(r, {"severity": "error",
//	    "message": r.guid, "locations": ["entity://Activity/" + r.guid]})
//
// — or a per-row mapper {when, severity, rule, message, code, locations} whose
// strings are templates over .row and whose when and severity may be CEL.
// Locations take models.ViolationLocation objects or the ParseLocation
// shortcuts file:line[:col], <gitref>:path:line[:col], entity://Kind/id and
// https://…#css=… / #xpath=…. An error finding fails the step; warnings and
// info warn it, and a warned step still evaluates its cel: and exports.
//
// See also: github.com/flanksource/gavel/cmd for CLI integration
package fixtures
