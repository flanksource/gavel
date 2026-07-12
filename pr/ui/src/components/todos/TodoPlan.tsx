import { lazy, Suspense, useEffect, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { UiSave } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../../icons/Spinner';
import type { TodoItem } from '../../types';
import { inputClass, todoQuery } from './format';

// MdxEditorField lazily pulls in the heavy @mdxeditor/editor (the same markdown
// field the run dialog uses), so it is code-split and rendered under Suspense with a
// plain-textarea fallback.
const MdxEditorField = lazy(() =>
  import('@flanksource/clicky-ui/mdx-editor').then(m => ({ default: m.MdxEditorField })),
);

interface PlanResponse {
  found: boolean;
  path?: string;
  content?: string;
  onDisk?: boolean;
  slug?: string;
  ref?: string;
  version?: number;
  todo?: TodoItem;
}

// TodoPlan shows the latest immutable Captain revision selected on the issue.
// Human edits append another database revision; they never rewrite an agent's
// local plan file.
export function TodoPlan({
  dir,
  provider,
  todo,
  active,
  onChanged,
}: {
  dir: string;
  provider: string;
  todo: TodoItem;
  active: boolean;
  onChanged?: (todo: TodoItem) => void;
}) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [found, setFound] = useState(false);
  const [path, setPath] = useState('');
  const [loaded, setLoaded] = useState(''); // server content — the save baseline
  const [draft, setDraft] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    setError('');
    setFound(false);
    setPath('');
    setLoaded('');
    setDraft('');
    if (!active || !todo.ref) return;

    let cancelled = false;
    setLoading(true);
    const params = new URLSearchParams(todoQuery(dir, provider));
    params.set('ref', todo.ref);
    fetch(`/api/todos/session/plan?${params.toString()}`)
      .then(async res => {
        const data = await res.json().catch(() => null);
        if (!res.ok) {
          const message = data && typeof data === 'object' && 'error' in data
            ? String(data.error)
            : `plan request failed (${res.status})`;
          throw new Error(message);
        }
        return data as PlanResponse;
      })
      .then((data: PlanResponse) => {
        if (cancelled) return;
        setFound(!!data.found);
        setPath(data.path ?? '');
        setLoaded(data.content ?? '');
        setDraft(data.content ?? '');
      })
      .catch(err => !cancelled && setError(err instanceof Error ? err.message : String(err)))
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [active, todo.ref, dir, provider]);

  async function save() {
    if (saving || draft === loaded) return;
    setSaving(true);
    setError('');
    try {
      const params = new URLSearchParams(todoQuery(dir, provider));
      const res = await fetch(`/api/todos/session/plan?${params.toString()}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref, version: todo.version, content: draft }),
      });
      if (!res.ok) {
        let detail = `save failed (${res.status})`;
        try {
          const data = await res.json();
          detail = data.error || detail;
        } catch {
          // Keep the status fallback when the response is not JSON.
        }
        throw new Error(detail);
      }
      const data = (await res.json()) as PlanResponse;
      const next = data.content ?? draft;
      setLoaded(next);
      setDraft(next);
      if (data.todo) onChanged?.(data.todo);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setSaving(false);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center gap-2 px-4 py-3 text-sm text-muted-foreground">
        <Spinner />
        Loading plan
      </div>
    );
  }
  if (!found) {
    if (error) return <div role="alert" className="px-4 py-3 text-sm text-red-600">{error}</div>;
    return <PlanEmpty message="No plan yet. Run this todo in Plan mode to produce one." />;
  }

  const dirty = draft !== loaded;
  return (
    <div className="flex min-h-0 flex-1 flex-col gap-2 px-4 py-3">
      <div className="flex shrink-0 items-center justify-between gap-2">
        <div className="min-w-0 truncate text-xs text-muted-foreground" title={path}>
          {path || 'PostgreSQL plan revision'}
        </div>
        <Button size="sm" variant="outline" disabled={!dirty || saving} onClick={() => void save()}>
          {saving ? <Spinner /> : <UiSave />}
          Save
        </Button>
      </div>
      {error && <div className="shrink-0 text-xs text-red-600">{error}</div>}
      <div className="min-h-0 flex-1 overflow-auto">
        <Suspense
          fallback={
            <textarea
              className={`${inputClass} h-auto min-h-[16rem] resize-y font-mono`}
              value={draft}
              onChange={e => setDraft(e.currentTarget.value)}
            />
          }
        >
          <MdxEditorField value={draft} onChange={setDraft} className="min-h-[16rem]" />
        </Suspense>
      </div>
    </div>
  );
}

function PlanEmpty({ message }: { message: string }) {
  return <div className="px-4 py-3 text-sm text-muted-foreground">{message}</div>;
}
