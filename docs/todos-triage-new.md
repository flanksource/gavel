# Triage newly created TODOs

The New Todo dialog, React Grab form and Create from PR dialog have a **Run after creation** selector listing the lifecycle steps that apply to the new TODO (`GET /api/todos/new/steps`). Choose `Triage new` to triage the TODO once it is saved, or `None` to start nothing.

The TODO, attachments and PR verification are saved before the step starts. Closing the creation window does not cancel an admitted run. If the step cannot start, the window links to the saved TODO and offers **Retry**; retrying starts a run on that TODO instead of creating another one.

`triage.new` writes through the todo tools while it runs: it retitles the TODO with `todo_update`, labels it from the workspace taxonomy with `todo_labels`, searches the workspace with `todo_search` and `todo_get`, and links or parents it with `todo_link` and `todo_set_parent`. Closing it as a duplicate (`todo_close` with `duplicateOf`) or merging it into an existing TODO (`todo_merge`) is an ask-policy call that a person approves in the TODO session; the approval shows the call's arguments. The existing TODO survives a merge. Denying the call leaves the new TODO separate. Merges across TODOs linked to different PRs are refused. The run ends with the event `kept`, `duplicate`, `merged` or `child`.

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

`GET /api/todos/new/steps?dir=<workspace>&status=<status>` lists the lifecycle steps that apply to a TODO that is about to be created: `{"steps":[{"name","label","prompt","auxiliary"}]}` in definition order, auxiliary steps included. The workspace's lifecycle is evaluated for a new TODO with the given status (default `draft`), medium priority, no labels, an empty body, no plan, no verification and no runs. A status that cannot be assigned returns HTTP 400.

`POST /api/todos/new` accepts `step` as an optional step name in JSON, multipart, URL-encoded forms or query parameters. Omitting it starts nothing. The step must be one the steps endpoint lists for the new TODO's status; any other name returns HTTP 400 naming the applicable steps, and nothing is saved. An explicit body value takes precedence over the query value.

Both endpoints are included in `/api/openapi.json` from their Go wire types.

```json
{"title":"Investigate parser crash","body":"Reproduction details","step":"triage.new"}
```

A successful creation returns HTTP 201 and the saved `todo`. When a step was named, `run` contains `status`, `sessionId` and `promptRunId` after admission, or `status: "failed"` with `error` if admission fails. A failed admission still represents a successful creation. Retry with `POST /api/todos/run` and `{ "ref": "<saved-id>", "dir": "<workspace>", "step": "<step>" }`.

The run ends with the unified result envelope `{success, event, comment}`, where `event` is one of `kept`, `duplicate`, `merged` or `child`. Its `lifecycle_outcome` event lists every write the tools made as `tools: [{name, ref}]`.
