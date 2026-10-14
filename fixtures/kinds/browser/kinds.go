// Package browser registers the fixture step kinds that drive a real browser:
//
//   - browser: starts one Chrome on a profile and exports its CDP endpoint;
//   - playwright: runs a Playwright spec (or, in the background, a node page
//     watcher) attached to that Chrome;
//   - agent-browser: runs agent-browser commands attached to that Chrome.
//
// All three share one Chrome process and profile: only the browser kind
// launches it, the others attach over BROWSER_CDP_URL.
package browser

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flanksource/clicky/task"
	"github.com/flanksource/gavel/fixtures"
	"github.com/invopop/jsonschema"
)

func init() {
	fixtures.RegisterKind(Browser{})
	fixtures.RegisterKind(Playwright{})
	fixtures.RegisterKind(AgentBrowser{})
}

// schemaOf reflects a kind's config struct into the JSON schema its steps are
// validated against; unknown keys are rejected.
func schemaOf(config any) json.RawMessage {
	reflector := jsonschema.Reflector{DoNotReference: true, ExpandedStruct: true}
	raw, err := json.Marshal(reflector.Reflect(config))
	if err != nil {
		panic(fmt.Sprintf("browser kinds: reflect schema of %T: %v", config, err))
	}
	return raw
}

// The browser kinds render their own config: inline scripts must never be
// templated (their `${…}` template literals are TypeScript, not fixture
// variables), so the executor's generic config rendering is opted out of.
func (Browser) TemplatesOwnConfig()      {}
func (Playwright) TemplatesOwnConfig()   {}
func (AgentBrowser) TemplatesOwnConfig() {}

// render expands one non-shell config value — a URL, a path, an agent-browser
// argument — through the fixture scope: `{{.env.X}}`, `{{.tmp}}`, then `$X`.
func render(step *fixtures.StepContext, field, value string) (string, error) {
	if value == "" {
		return "", nil
	}
	out, err := step.Expand(value)
	if err != nil {
		return "", fmt.Errorf("%s step %q: %s: %w", step.Spec.Kind, step.Spec.ID, field, err)
	}
	return out, nil
}

// fixtureDir is the directory relative paths in a step resolve against: the
// fixture file's directory, else the run's working directory.
func fixtureDir(step *fixtures.StepContext) string {
	if step.Test.SourceDir != "" {
		return step.Test.SourceDir
	}
	return step.WorkDir
}

func resolvePath(step *fixtures.StepContext, path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(fixtureDir(step), path)
}

func requireEnv(step *fixtures.StepContext, name, hint string) (string, error) {
	value, ok := step.Scope.Env.Get(name)
	if !ok || value == "" {
		return "", fmt.Errorf("%s step %q: %s is not set; %s", step.Spec.Kind, step.Spec.ID, name, hint)
	}
	return value, nil
}

// environ is the fixture Env as a map, for clicky exec's WithExactEnv.
func environ(step *fixtures.StepContext) map[string]string { return step.Scope.Env.Map() }

func stepName(step *fixtures.StepContext) string {
	if step.Spec.Name != "" {
		return step.Spec.Name
	}
	if step.Spec.ID != "" {
		return step.Spec.ID
	}
	return step.Spec.Kind
}

func newResult(step *fixtures.StepContext, start time.Time) fixtures.FixtureResult {
	return fixtures.FixtureResult{
		Name:   stepName(step),
		Type:   step.Spec.Kind,
		Status: task.StatusPASS,
		Start:  &start,
		Test:   step.Test,
	}
}

func finish(result fixtures.FixtureResult) fixtures.FixtureResult {
	if result.Start != nil {
		result.Duration = time.Since(*result.Start)
	}
	return result
}

func randomHex(bytes int) string {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("crypto/rand: %v", err))
	}
	return hex.EncodeToString(buf)
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir() && info.Mode()&0o111 != 0
}
