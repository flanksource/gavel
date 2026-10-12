# Triage newly created TODOs

The New Todo dialog, React Grab form and Create from PR dialog enable **Triage after creation** by default. Triage chooses a specific title and taxonomy labels and searches other TODOs in the same workspace. Turn the checkbox off to create the TODO without starting triage.

The TODO, attachments and PR verification are saved before triage starts. Closing the creation window does not cancel an admitted run. If triage cannot start, the window links to the saved TODO and offers **Retry triage**; retrying starts a run on that TODO instead of creating another one.

Title and labels are applied automatically when the work stays separate. Proposals to close a duplicate, merge into an existing TODO or make the new TODO a child require approval in the TODO session. The approval shows the source and target titles and IDs, the rationale, and the proposed content. The existing target survives a merge. Child assignments use the existing single-level hierarchy. Denying a proposal keeps the new TODO separate. Canceled review makes no relationship changes. Changes to either TODO invalidate a pending proposal and require another triage pass.

## Configuration

`triage.new` is an auxiliary lifecycle step and can also be selected manually:

```bash
gavel todos run <todo-id> --step triage.new
```

The built-in prompt defaults to `api:luna:high`. Override it independently of bulk triage:

```yaml
todos:
  triageNew:
    model: api:luna:high
```

The lifecycle prompt reference is `todos.triage.new`; its settings descriptor and configuration path are `todos.triageNew`. The separate `todos.triage` contract continues to serve backlog triage.

## Creation API

`POST /api/todos/new` accepts `triage` as an optional boolean in JSON, multipart, URL-encoded forms or query parameters. Omitting it is equivalent to `false`. An explicit body value takes precedence over the query value.

The creation request and response are included in `/api/openapi.json` from their Go wire types.

```json
{"title":"Investigate parser crash","body":"Reproduction details","triage":true}
```

A successful creation returns HTTP 201 and the saved `todo`. When requested, `triage` contains `status`, `sessionId` and `promptRunId` after admission, or `status: "failed"` with `error` if admission fails. A failed admission still represents a successful creation. Retry with `POST /api/todos/run` and `{ "ref": "<saved-id>", "dir": "<workspace>", "step": "triage.new" }`.

The completed prompt result requires nonempty `title`, at least one allowed `labels` value, `summary`, `endStatus` and `action`. Relationship results require `target` and `rationale`; merge results must preserve the combined body and any existing verification and plan. Relationship application also requires the exact proposal returned by the approved `triage_review` call.
