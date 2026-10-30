import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button } from '@flanksource/clicky-ui/components';
import type { PRItem, Project } from '../types';

interface Action { step: { name: string }; event: { id: string; type: string; body: string; author: string }; reason: string }
interface Run { id: string; sessionId: string; specProfile: string; state: string; error?: string; resultText?: string }
interface Context { headSha: string; definition: string; actions: Action[]; runs: Run[] }
interface Prepared { spec: unknown; specHash: string }

async function read<T>(url: string, body?: unknown): Promise<T> {
  const response = await fetch(url, body ? { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) } : undefined);
  if (!response.ok) throw new Error(await response.text());
  return response.json() as Promise<T>;
}

export function PRLifecycleActions({ pr, projects }: { pr: PRItem; projects: Project[] }) {
  const [open, setOpen] = useState(false);
  const [selected, setSelected] = useState<Action>();
  const [prepared, setPrepared] = useState<Prepared>();
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const project = projects.find(item => item.repos.includes(pr.repo));
  const params = new URLSearchParams({ project: project?.name ?? '', repo: pr.repo, number: String(pr.number) });
  const url = `/api/pr/lifecycle?${params}`;
  const query = useQuery({ queryKey: ['pr-lifecycle', url], queryFn: () => read<Context>(url), enabled: open && !!project });
  async function act(operation: 'preview' | 'start' | 'cancel', action?: Action, runId?: string) {
    setBusy(true); setError('');
    if (operation === 'preview') { setPrepared(undefined); setSelected(action); }
    try {
      const result = await read<Prepared>(url, { operation, headSha: query.data?.headSha, definition: query.data?.definition, step: action?.step.name, eventId: action?.event.id, specHash: prepared?.specHash, runId });
      if (operation === 'preview') setPrepared(result);
      else { setPrepared(undefined); setSelected(undefined); await query.refetch(); }
    } catch (cause) { setError(cause instanceof Error ? cause.message : String(cause)); }
    finally { setBusy(false); }
  }
  const active = query.data?.runs.some(run => ['pending', 'running', 'waiting'].includes(run.state));
  return <details className="my-4 rounded border border-border p-3" onToggle={event => setOpen(event.currentTarget.open)}>
    <summary className="cursor-pointer text-sm font-medium">Lifecycle actions</summary>
    {open && <div className="mt-3 space-y-3">
      {!project && <p className="text-sm text-muted-foreground">Register this repository with a local project to run lifecycle actions.</p>}
      <p className="text-xs text-muted-foreground">Actions start manually. CI failures require structured Gavel results for the current PR head.</p>
      {(error || query.error) && <p role="alert" className="break-words text-sm text-destructive">{error || query.error?.message}</p>}
      {query.isLoading && <p role="status">Loading lifecycle context…</p>}
      {project && <Button size="sm" variant="outline" disabled={busy} onClick={() => { setPrepared(undefined); void query.refetch(); }}>Refresh lifecycle context</Button>}
      {query.data?.actions.map(action => <div key={`${action.step.name}:${action.event.id}`} className="rounded border border-border p-2">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <span className="text-sm">{action.step.name} · {action.event.type}{action.event.author ? ` · ${action.event.author}` : ''}</span>
          <Button size="sm" variant="outline" disabled={busy || active} onClick={() => void act('preview', action)}>Preview {action.step.name}</Button>
        </div>
        {action.reason && <p className="mt-1 break-words font-mono text-xs text-muted-foreground">{action.reason}</p>}
        {action.event.body && <p className="mt-2 max-h-32 overflow-y-auto whitespace-pre-wrap break-words text-xs">{action.event.body}</p>}
      </div>)}
      {query.data && query.data.actions.length === 0 && <p className="text-sm">No lifecycle steps match this PR.</p>}
      {prepared && selected && <section aria-label="PR lifecycle preview" className="space-y-2">
        <h3 className="text-sm font-medium">Resolved prompt · {selected.step.name}</h3>
        <pre className="max-h-80 overflow-auto rounded bg-muted p-3 text-xs">{JSON.stringify(prepared.spec, null, 2)}</pre>
        <Button disabled={busy || active} onClick={() => void act('start', selected)}>Start {selected.step.name}</Button>
      </section>}
      {!!query.data?.runs?.length && <section aria-label="Lifecycle history" className="space-y-2">
        <h3 className="text-sm font-medium">Lifecycle history</h3>
        {query.data.runs.map(run => <div key={run.id} className="rounded border border-border p-2 text-sm">
          <div className="flex flex-wrap items-center gap-2"><a className="underline" href={`/sessions/${run.sessionId}?project=${encodeURIComponent(project?.name ?? '')}`}>{run.specProfile} · {run.state}</a>
            {['pending', 'running', 'waiting'].includes(run.state) && <Button size="sm" variant="outline" disabled={busy} onClick={() => void act('cancel', undefined, run.id)}>Cancel {run.specProfile}</Button>}
          </div>
          {run.error && <p className="mt-1 break-words text-destructive">{run.error}</p>}
          {run.resultText && <p className="mt-1 max-h-40 overflow-y-auto whitespace-pre-wrap break-words">{run.resultText}</p>}
        </div>)}
      </section>}
    </div>}
  </details>;
}
