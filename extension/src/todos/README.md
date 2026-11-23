# Offline todo capture

The extension bundles the todo form in `todo-relay.html`. **Add to gavel todo**, **Screenshot to gavel todo**, and the popup footer's sync error, which opens the saved captures, use this local page; the grabber's menu captures a todo. Neither the playground nor Gavel needs to be running to save a new todo or a comment on an existing issue.

Build and load it:

```bash
pnpm --dir pr/ui --filter @flanksource/gavel-extension build
```

Load `extension/dist` from the Gavel repository unpacked through `chrome://extensions` with Developer mode enabled, or reload an existing installation. Configure the Gavel server in the popup. The popup's capture button opens a standalone form with saved-capture history; grabbing an element opens the same form in the page's dialog.

The background worker owns the `gavel-todo-capture` IndexedDB database. Its `cache` store isolates project catalogs by server and labels, existing issues, runtime catalogs, and run choices by server and workspace directory. Lifecycle steps are also keyed by the new todo's status. Opening the Gavel grabber warms project and label catalogs. Opening a form reads saved choices immediately and refreshes them from Gavel; failed refreshes leave the saved catalog intact and show the error. The searchable project picker displays project names and workspace directories and accepts an explicit directory when no catalog is available. Labels and existing issues have searchable pickers too. The form remembers the workspace last used for a source host and offers recent synced issues for additional comments.

**Save todo** and **Save comment** confirm only after the record and screenshot data URLs have been written to the database's `queue` store. Delivery starts in the background and retries pending work every minute through a Chrome alarm, including after the dialog closes or the worker restarts. **Sync now** runs the same synchronizer. Success records the remote reference and releases the queued image bytes. The form supports title, Markdown description/comment, workspace, priority, assignable status, existing issue reference, labels, a PR link, and pasted or selected screenshots.

**Start a new run** enables the selected **Run after creation** lifecycle step. The shared runtime bar offers mode, model, effort, and the permissions, workspace, commit and budget fields supported by that runtime; presets come from the workspace's Gavel run context. Step defaults remain inherited, while explicit runtime edits and presets are saved with the capture. **Save todo and start run** works offline after that project's run catalog has been cached: syncing sends the exact step, spec and presets with the todo, including multipart submissions with screenshots. A status change that invalidates the chosen step is shown explicitly and blocks saving until the step is changed or the run is disabled.

If Gavel creates the todo but cannot start its run, the saved capture keeps the remote reference and shows the run error. **Retry run** starts only the requested step on that existing todo; it never creates the todo again. An interrupted run retry is held for review to avoid starting a duplicate run.

An unavailable server leaves the capture pending. An HTTP validation or authentication rejection remains visible with its original data and a retry action. Gavel's todo endpoints do not accept an idempotency key: if a mutation loses its response, returns a server error, or is interrupted by a worker restart, its state becomes `uncertain`. Automatic delivery pauses for that record. Check Gavel, then use **Checked Gavel — retry** only if the capture was not delivered.

Run the persistence, server isolation, recovery, attachment, and validation tests:

```bash
pnpm --dir pr/ui --filter @flanksource/gavel-extension test src/todos src/todo-relay-target.test.ts host/host-todo-modal.test.ts
```
