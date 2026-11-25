import { useEffect, useState } from "react";
import { Button, Combobox, Field, InputField, SegmentedControl, Select } from "@flanksource/clicky-ui";
import { requestBackground } from "../messages";
import { captureFromParams, priorities, statuses, type Attachment, type Capture, type QueuedTodo, type Snapshot } from "./model";
import { TodoQueue } from "./TodoQueue";
import { TodoProjectPicker } from "./TodoProjectPicker";
import { TodoRunControls } from "./TodoRunControls";
import { runCaptureError, type TodoRun } from "./runtime";

export function TodoForm({ server, project, params, onDone, attachmentSource }: {
  server: string;
  project?: string;
  params: URLSearchParams;
  onDone: (type: "todo-saved" | "cancel") => void;
  attachmentSource?: { parentOrigin: string };
}) {
  const [draft, setDraft] = useState<Capture>(() => captureFromParams(params));
  const projectName = params.get("project") ?? project;
  const [snapshot, setSnapshot] = useState<Snapshot>({ queue: [] });
  const [snapshotScope, setSnapshotScope] = useState("");
  const [cacheError, setCacheError] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);
  const [converting, setConverting] = useState(0);
  const [saved, setSaved] = useState<QueuedTodo>();
  const [revision, setRevision] = useState(0);
  const update = (fields: Partial<Capture>) => setDraft(previous => ({ ...previous, ...fields }));

  useEffect(() => {
    let disposed = false;
    const load = async () => {
      const next = await requestBackground<Snapshot>({ type: "todo-snapshot", server, dir: draft.dir, sourceUrl: draft.sourceUrl, status: draft.status });
      if (disposed) return;
      setSnapshot(next);
      setSnapshotScope(JSON.stringify([draft.dir, draft.status]));
      setDraft(previous => {
        if (previous.dir) return previous;
        const preferred = next.projects?.find(entry => entry.name === projectName)?.dir ?? next.recentDir;
        return preferred ? { ...previous, dir: preferred } : previous;
      });
    };
    void load().catch(cause => { if (!disposed) setError(String(cause)); });
    const refresh = window.setTimeout(() => {
      const requests = [requestBackground({ type: "todo-refresh", server, dir: draft.dir })];
      if (draft.dir && draft.mode === "new") requests.push(requestBackground({ type: "todo-runtime-refresh", server, dir: draft.dir, status: draft.status }));
      void Promise.allSettled(requests).then(async results => {
        if (disposed) return;
        if (!disposed) setCacheError("");
        await load();
        const errors = results.flatMap(result => result.status === "rejected" ? [String(result.reason)] : []);
        if (!disposed && errors.length) setCacheError(`Using saved choices. Catalog refresh failed: ${errors.join("; ")}`);
      }).catch(cause => { if (!disposed) setCacheError(`Using saved choices. Catalog refresh failed: ${String(cause)}`); });
    }, 350);
    const interval = window.setInterval(() => { void load().catch(cause => { if (!disposed) setError(String(cause)); }); }, 3000);
    return () => { disposed = true; window.clearTimeout(refresh); window.clearInterval(interval); };
  }, [server, projectName, draft.dir, draft.status, draft.mode, draft.sourceUrl, revision]);

  async function addAttachment(blob: Blob, name: string) {
    setConverting(count => count + 1);
    try {
      const dataUrl = await new Promise<string>((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(String(reader.result));
        reader.onerror = () => reject(reader.error);
        reader.readAsDataURL(blob);
      });
      setDraft(previous => ({ ...previous, attachments: [...previous.attachments, { name, dataUrl }] }));
    } catch (cause) { setError(String(cause)); }
    finally { setConverting(count => count - 1); }
  }

  useEffect(() => {
    const listener = (event: MessageEvent) => {
      if (!attachmentSource || event.source !== window.parent || event.origin !== attachmentSource.parentOrigin) return;
      const data = event.data as { source?: string; type?: string; blob?: unknown; name?: string } | null;
      if (data?.source === "gavel-react-grab" && data.type === "attachment" && data.blob instanceof Blob) void addAttachment(data.blob, data.name || "screenshot.png");
    };
    const paste = (event: ClipboardEvent) => {
      const images = Array.from(event.clipboardData?.files ?? []).filter(file => file.type.startsWith("image/"));
      if (!images.length) return;
      event.preventDefault();
      images.forEach(file => void addAttachment(file, file.name || "screenshot.png"));
    };
    window.addEventListener("message", listener);
    window.addEventListener("paste", paste);
    if (attachmentSource) window.parent.postMessage({ source: "gavel-react-grab", type: "embed-ready" }, attachmentSource.parentOrigin === "null" ? "*" : attachmentSource.parentOrigin);
    return () => { window.removeEventListener("message", listener); window.removeEventListener("paste", paste); };
  }, [attachmentSource]);

  async function submit() {
    if (runError) { setError(runError); return; }
    setSaving(true);
    setError("");
    try {
      setSaved(await requestBackground<QueuedTodo>({ type: "todo-capture", server, capture: draft }));
      setRevision(value => value + 1);
    } catch (cause) { setError(String(cause)); }
    finally { setSaving(false); }
  }

  const ready = snapshotScope === JSON.stringify([draft.dir, draft.status]);
  const catalog = ready ? snapshot : undefined;
  const labels = [...new Set([...Object.keys(catalog?.labels?.counts ?? {}), ...(catalog?.labels?.definitions.map(entry => entry.name) ?? []), ...draft.labels])];
  const existing = draft.mode === "existing";
  const issues = catalog?.todos?.filter(todo => !["completed", "skipped"].includes(todo.status)) ?? [];
  const runError = runCaptureError(draft.run, catalog?.runContext, catalog?.steps);
  const chooseRun = (run: TodoRun | undefined) => {
    update({ run });
    void requestBackground({ type: "todo-run-choice", server, dir: draft.dir, run: run ?? null }).catch(cause => setError(String(cause)));
  };
  const recent = [...new Map([...snapshot.queue].reverse().filter(todo => todo.remoteRef && todo.dir === draft.dir && new URL(todo.sourceUrl || "https://capture.invalid").host === new URL(draft.sourceUrl || "https://capture.invalid").host).map(todo => [todo.remoteRef, todo])).values()].slice(0, 5);
  return (
    <main className="mx-auto max-w-4xl space-y-4 p-4 text-foreground">
      <header><h1 className="text-lg font-semibold">{saved ? "Saved on this device" : existing ? "Add to issue" : "New todo"}</h1><p className="break-all text-xs text-muted-foreground">{draft.sourceUrl}</p></header>
      {saved ? (
        <div role="status" className="space-y-3">
          <p className="text-sm">Your capture is saved. It will sync to Gavel when the server is available.</p>
          <Button onClick={() => onDone("todo-saved")}>Done</Button>
        </div>
      ) : (
        <form className="space-y-4" onSubmit={event => { event.preventDefault(); void submit(); }}>
          <SegmentedControl value={draft.mode} onChange={mode => update({ mode, run: undefined })} options={[{ id: "new", label: "New issue" }, { id: "existing", label: "Existing issue" }]} aria-label="Capture mode" />
          {cacheError && <p role="status" className="break-words text-xs text-muted-foreground">{cacheError}</p>}
          <TodoProjectPicker projects={snapshot.projects ?? []} dir={draft.dir} disabled={saving} onChange={dir => update({ dir, ref: "", run: undefined })} />
          {recent.length > 0 && <div className="flex flex-wrap items-center gap-2 text-xs"><span>Add to recent</span>{recent.map(todo => <Button key={todo.remoteRef} type="button" size="sm" variant="outline" onClick={() => update({ mode: "existing", ref: todo.remoteRef!, run: undefined })}>{todo.title}</Button>)}</div>}
          {existing ? (
            <>
              <Field label="Issue" htmlFor="todo-ref" required helper="Search saved issues or enter an issue reference.">
                <Combobox id="todo-ref" ariaLabel="Issue" required ariaRequired value={draft.ref} onChange={ref => update({ ref })}
                  options={issues.map(todo => ({ value: todo.ref, label: `${todo.shortId ?? todo.ref} — ${todo.title}` }))} />
              </Field>
            </>
          ) : (
            <>
              <Field label="Title" htmlFor="todo-title" required><InputField id="todo-title" required value={draft.title} onChange={title => update({ title })} /></Field>
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <Field label="Priority" htmlFor="todo-priority"><Select id="todo-priority" value={draft.priority} onChange={event => update({ priority: event.target.value as Capture["priority"] })}>{priorities.map(value => <option key={value}>{value}</option>)}</Select></Field>
                <Field label="Status" htmlFor="todo-status"><Select id="todo-status" value={draft.status} onChange={event => update({ status: event.target.value as Capture["status"] })}>{statuses.map(value => <option key={value}>{value}</option>)}</Select></Field>
              </div>
              <Field label="Labels" htmlFor="todo-labels"><Combobox id="todo-labels" ariaLabel="Labels" multiple variant="tags" separators={[","]} options={labels.map(value => ({ value, label: value, trailing: catalog?.labels?.counts?.[value], description: catalog?.labels?.definitions.find(entry => entry.name === value)?.description ?? "" }))} value={draft.labels} onChange={labels => update({ labels })} /></Field>
              <TodoRunControls key={draft.dir} run={draft.run} context={catalog?.runContext} steps={catalog?.steps} remembered={catalog?.runChoice} ready={ready} disabled={saving} error={runError} onChange={chooseRun} />
              <Field label="Pull request" htmlFor="todo-pr"><InputField id="todo-pr" type="url" value={draft.pr} onChange={pr => update({ pr })} placeholder="https://github.com/acme/repo/pull/123" /></Field>
            </>
          )}
          <Field label={existing ? "Comment" : "Description"} htmlFor="todo-body" helper="Markdown supported."><InputField as="textarea" id="todo-body" rows={10} value={draft.body} onChange={body => update({ body })} inputClassName="font-mono" /></Field>
          <Field label="Screenshots" htmlFor="todo-images" helper="Choose images or paste a screenshot into the form.">
            <InputField id="todo-images" type="file" accept="image/*" multiple onChange={(_value, event) => { Array.from(event.target.files ?? []).forEach(file => void addAttachment(file, file.name)); event.target.value = ""; }} />
            <Attachments attachments={draft.attachments} onRemove={index => update({ attachments: draft.attachments.filter((_, i) => i !== index) })} />
          </Field>
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onDone("cancel")} disabled={saving}>Cancel</Button>
            <Button type="submit" loading={saving} disabled={saving || converting > 0 || !!runError}>Save {existing ? "comment" : draft.run ? "todo and start run" : "todo"}</Button>
          </div>
        </form>
      )}
      {error && <p role="alert" className="break-words text-sm text-destructive">{error}</p>}
      <TodoQueue server={server} queue={snapshot.queue} onChange={() => setRevision(value => value + 1)} onError={setError} />
    </main>
  );
}

function Attachments({ attachments, onRemove }: { attachments: Attachment[]; onRemove: (index: number) => void }) {
  return <div className="flex flex-wrap gap-3">{attachments.map((attachment, index) => <div key={index} className="space-y-1"><img src={attachment.dataUrl} alt={attachment.name} className="max-h-32 max-w-full rounded border border-border" /><Button type="button" size="sm" variant="ghost" onClick={() => onRemove(index)}>Remove {attachment.name}</Button></div>)}</div>;
}
