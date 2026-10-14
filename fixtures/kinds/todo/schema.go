package todo

const serverAndDirProperties = `
    "server": {"type": "string", "description": "Base URL of the gavel dashboard; defaults to $GAVEL_URL from the fixture env."},
    "dir": {"type": "string", "description": "Workspace directory the dashboard resolves the todo in; omitted resolves the todo globally."}`

const todoSchema = `{
  "type": "object",
  "additionalProperties": false,
  "description": "Create, read, or review a TODO through a running gavel dashboard. Exactly one of create, approve, reject, get; none means get.",
  "properties": {` + serverAndDirProperties + `,
    "ref": {"type": "string", "description": "TODO id, short id or alias; defaults to $TODO_ID. Not allowed with create."},
    "create": {
      "type": "object",
      "additionalProperties": false,
      "description": "POST /api/todos/new. Exports TODO_ID and TODO_URL.",
      "properties": {
        "title": {"type": "string"},
        "body": {"type": "string"},
        "labels": {"type": "array", "items": {"type": "string"}},
        "status": {"type": "string", "description": "Initial status; the server defaults to draft (pending with autoSave)."},
        "triage": {"type": "boolean", "description": "Start a triage.new run after saving; the step then waits for it by default."},
        "autoSave": {"type": "boolean"}
      }
    },
    "approve": {"type": "boolean", "description": "POST /api/todos/plan/approve for a todo in review."},
    "reject": {"type": "boolean", "description": "POST /api/todos/plan/reject for a todo in review."},
    "get": {"type": "boolean", "description": "Read the todo (the default)."}
  }
}`

const promptSchema = `{
  "type": "object",
  "additionalProperties": false,
  "required": ["step"],
  "description": "Launch one lifecycle step of a TODO through POST /api/todos/run and follow the attempt it admitted.",
  "properties": {` + serverAndDirProperties + `,
    "step": {"type": "string", "minLength": 1, "description": "Lifecycle step: plan, run, verify, triage, triage.new, or any step the project's lifecycle declares.", "examples": ["plan", "run", "verify", "triage", "triage.new"]},
    "todo": {"type": "string", "description": "TODO id, short id or alias; defaults to $TODO_ID."},
    "spec": {"type": "object", "description": "Captain api.Spec overrides (model, effort, budget, permissions, …) sent as the run's spec."},
    "force": {"type": "boolean", "description": "Dispatch even when another live run owns the todo."},
    "resume": {"type": "boolean", "description": "Continue the todo's prior agent session."},
    "reuseBranch": {"type": "boolean", "description": "Continue on the branch the previous run left behind."}
  }
}`
