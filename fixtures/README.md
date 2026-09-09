# AI fixture runtime configuration

Command, test, and lint fixtures can run without AI configuration. For an executed AI checklist, the CLI lazily captures the verification profile's authored layers and the saved settings from `~/.captain.yaml`. The fixture's flat overrides and grading prompt are added before Captain resolves the final runtime. Saved provider defaults therefore follow the final model selected by the fixture.

An authored review fixture can override selected settings while inheriting the verification profile:

````markdown
---
ai:
  model: api:gpt-5
  temperature: 0
  noCache: false
  maxTokens: 1000
---

# Review the change

- [ ] The changed behavior has focused test coverage.
````

Save this as `review.fixture.md` and run `gavel fixtures review.fixture.md`. This executes an AI review using the configured provider. `temperature: 0` and `noCache: false` are explicit overrides; omitting those keys inherits their values. JSON and YAML round trips preserve that distinction.

The profile is selected by `todos.verify.runtimeProfile`, or `todos.runtimeProfile` when the step has no selection. Its layers, `ai`, and `todos.verify` remain authored until the final fixture override is known. Profile constraints continue to restrict that final runtime.

## Complete grader snapshots

A generated TODO definition of done stores the complete grader in `ai.spec`. That snapshot replaces runner defaults and bypasses both current profile lookup and saved settings. An explicit library runner `Spec` has the same authority; a document's `ai.spec` takes precedence when both are present. Flat fixture fields still override the selected snapshot.

```yaml
ai:
  spec:
    model: gpt-5
    mode: api
    budget:
      maxTokens: 1000
      timeout: 2m
    memory:
      skipHooks: true
  temperature: 0
  noCache: false
```

Captain resolves authored standalone runtimes before provider construction. Complete snapshots are structurally composed with explicit flat overrides and validated without rewriting their supplied runtime fields. The grading prompt and response schema replace the source prompt; session history, attachments, and nested workflow instructions are removed. Other runtime fields, including permissions, memory, sandbox, setup, and fallback candidates, remain part of the spec sent to the provider. Shared resolution warnings and field provenance are retained in the fixture result metadata.

## Library callers

`RunnerOptions.ResolveSpec` and `verifier.Verifier.ResolveSpec` return `api.ResolveSpecOptions` containing raw `Layers` and an optional captured `Saved` snapshot. They run only for AI steps that have no authoritative `Spec`. `RunOptions.Runtime` carries these inputs to the AI step, which copies the layer slice before adding its own layers and calls `api.ResolveSpecLayers` once with `RequireModel: true`.

Library callers supply saved settings explicitly. Neither fixture parsing nor provider construction rereads them. `FixtureAIConfig.Temperature` and `NoCache` are pointers for Go callers; nil means omitted. `ToAgentConfig` accepts the resolved `api.Spec` and projects provider configuration after resolution.

The focused regression suites run without a model call or a real database:

```sh
go test ./fixtures ./fixtures/types -run '^TestFixtures$|^TestFixtureTypes$' -ginkgo.focus='AI fixture|lazy fixture runtime spec|serialized TODO grader runtime' -count=1
go test ./cmd/gavel -run '^TestGavelCLI$' -ginkgo.focus='standalone fixture runtime profiles' -count=1
```
