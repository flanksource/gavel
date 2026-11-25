import { useState } from "react";
import { Button } from "@flanksource/clicky-ui";
import { requestBackground } from "../messages";
import type { QueuedTodo } from "./model";

export function TodoQueue({ server, queue, onChange, onError }: { server: string; queue: QueuedTodo[]; onChange: () => void; onError: (message: string) => void }) {
  const [busy, setBusy] = useState(false);
  async function sync(retryId?: string) {
    setBusy(true);
    try {
      await requestBackground({ type: "todo-sync", server, ...(retryId ? { retryId } : {}) });
      onChange();
    } catch (cause) { onError(String(cause)); }
    finally { setBusy(false); }
  }
  if (!queue.length) return null;
  return <section className="space-y-3 border-t border-border pt-4" aria-label="Saved captures">
    <div className="flex items-center justify-between gap-2"><h2 className="text-sm font-semibold">Saved captures</h2><Button type="button" size="sm" variant="outline" loading={busy} onClick={() => void sync()}>Sync now</Button></div>
    <ul className="space-y-3">{[...queue].reverse().filter((todo, index) => todo.state !== "synced" || index < 20).map(todo => <li key={todo.id} className="space-y-1 text-sm">
      <div className="flex flex-wrap items-center justify-between gap-2"><span>{todo.mode === "existing" ? `Comment on ${todo.ref}` : todo.title}</span><span className="text-xs text-muted-foreground">{todo.state === "run_error" ? "Todo saved; run could not start" : todo.state}</span></div>
      <p className="break-all text-xs text-muted-foreground">{todo.dir}</p>
      {todo.remoteRef && <a href={`${server}/todos/${encodeURIComponent(todo.remoteRef)}?dir=${encodeURIComponent(todo.dir)}`} target="_blank" rel="noreferrer" className="text-xs text-primary underline">Open {todo.remoteRef}</a>}
      {todo.error && <p className="break-words text-xs text-destructive">{todo.error}</p>}
      {todo.runResult?.status === "started" && <p className="text-xs text-muted-foreground">{todo.run?.step} started{todo.runResult.sessionId ? ` · ${todo.runResult.sessionId}` : ""}</p>}
      {todo.state === "uncertain" && <p className="text-xs">Delivery was interrupted. Check Gavel before retrying; a retry could {todo.remoteRef ? "start another run" : "create a duplicate"}.</p>}
      {(todo.state === "error" || todo.state === "uncertain" || todo.state === "run_error") && <Button size="sm" type="button" variant="outline" disabled={busy} onClick={() => void sync(todo.id)}>{todo.state === "uncertain" ? "Checked Gavel — retry" : todo.remoteRef ? `Retry ${todo.run?.step ?? "run"}` : "Retry delivery"}</Button>}
    </li>)}</ul>
  </section>;
}
