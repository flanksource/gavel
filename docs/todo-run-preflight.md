# Previewing TODO runs

`gavel todos run <ref> --dry-run` resolves the selected lifecycle step and runtime profile, renders the prompt, and validates the complete Captain run input. It prints capability warnings, the selected profile, the ordered input layers, and the resolved spec. The dashboard uses the same resolution for `POST /api/todos/run/preview` and displays its warnings in the advanced run dialog and the initial session preview.

```sh
gavel todos run <ref> --step plan --runtime-profile reviewer --dry-run
```

The HTTP preview accepts the same `ref`, `step`, `runtimeProfile`, and `spec` fields as a run. A successful response includes `prompt`, `specYaml`, `trace`, field `provenance`, and optional `warnings: string[]`. Empty warnings are omitted. Warnings describe unsupported runtime capabilities or missing runtime-mode configuration; they do not turn a valid preview into an error. Changing the request replaces its warnings with the latest preview verdict.

Invalid request input fails preview with HTTP 400 and fails CLI resolution with an error. Checks include runnable work, a valid deadline, prepared compatible attachments, model and permission validity, enforceable per-tool policies, verifier declarations and judge files, commit ownership, sandbox configuration, and the effective model/usage/input constraints supplied by resolution. Input-token estimates use Captain's approximate text-size guard, not a provider tokenizer.

Each home and project configuration file is structurally validated before merging. Prompt, lifecycle and selected catalog layers are also checked before request overrides: a valid request cannot conceal an invalid temperature, budget, permission value or referenced preset. These configuration errors return HTTP 500 with the source and field, and the CLI identifies them as lifecycle configuration errors. An unknown profile explicitly requested by the caller returns HTTP 400.

Defaults and raw profiles remain partial until the request is composed. The run dialog can display a configured runtime whose tool policy requires a different model; selecting a compatible model lets the complete preview succeed. Final runtime checks include the primary model and its fallbacks. Compact selectors such as `api:sonnet` retain their existing pin precedence over a separate mode field.

The built-in plan and triage prompts declare per-tool permissions. Selecting a runtime that cannot enforce those declarations now fails during preview; previously Captain rejected it only when execution began. Choose a runtime that supports the declared policy, or author a compatible policy explicitly for the intended runtime. Preflight never removes a policy to make a run succeed.

Preview constructs the real run configuration and hooks, but does not construct a provider, prepare a workspace, invoke hooks, run fixture factories, create an approval broker, or admit/persist a run. Dispatch validates its actual input again before admission, including a caller-supplied provider. The dashboard approval callback is present during validation; its factory is bound only after Captain supplies the admitted session and run identities.

Command and registered fixture verification can run without a generating model. A judge prompt still needs its required runtime. Preview checks fixture registration without executing the fixture or its setup, so a successful preview does not claim that commands, credentials, provider calls, or the definition of done will succeed.

The lifecycle captures `~/.captain.yaml` once for saved AI settings and runtime catalog discovery. Saved values fill gaps after every authored layer, using the final primary and fallback providers. A project budget or explicit request value wins over saved settings. Explicit `false`, `0`, empty strings and empty collections retain their meaning; clearing a required generating field produces an error. Rendering and dispatch use the resolved runtime without loading defaults again.

For example, a saved Captain configuration can supply the model and a cost ceiling:

```yaml
# ~/.captain.yaml
ai:
  defaultModel: agent:claude-sonnet-5
  budgetUSD: 1
```

A project can then override the ceiling for one operation:

```yaml
# .gavel.yaml
todos:
  run:
    budget:
      cost: 2
```

`gavel config --resolve` displays the complete operation specifications and their field sources. The run preview reports authored layers in `trace`; `provenance` identifies saved, catalog, authored and contextual sources without inventing saved-default layers. When a value is normalized, its original source is retained alongside `normalizedBy`.

A generating run with no configured model fails with configuration guidance. During the migration window, a named model with no configured mode warns and uses Captain's existing provider default. Effort comes from explicit settings, saved settings or the model catalog. The run dialog keeps requests sparse until the operator changes a setting; the first catalog entry is never an implicit selection.

This preflight uses the constraints produced by the current layer fold. Captain's restrictive permission-constraint channel remains separate migration work.
