import { lazy, Suspense, useEffect, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { UiSave } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../../icons/Spinner';
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
}

// TodoPlan shows the plan a plan-mode run produced — recovered from the Claude
// session by the /api/todos/session/plan endpoint — in an editable markdown editor,
// and saves edits back to the plan file so a human can refine the plan before
// approving it. It is mounted only while the Plan tab is active.
export function TodoPlan({
  dir,
  provider,
  sessionId,
  active,
}: {
  dir: string;
  provider: string;
  sessionId: string | undefined;
  active: boolean;
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
    if (!active || !sessionId) return;

    let cancelled = false;
    setLoading(true);
    const params = new URLSearchParams(todoQuery(dir, provider));
    params.set('sessionId', sessionId);
    fetch(`/api/todos/session/plan?${params.toString()}`)
      .then(res => (res.ok ? res.json() : Promise.reject(new Error(`plan request failed (${res.status})`))))
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
  }, [active, sessionId, dir, provider]);

  async function save() {
    if (saving || !path || draft === loaded) return;
    setSaving(true);
    setError('');
    try {
      const res = await fetch('/api/todos/session/plan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ sessionId, path, content: draft }),
      });
      if (!res.ok) throw new Error(`save failed (${res.status})`);
      const data = (await res.json()) as PlanResponse;
      const next = data.content ?? draft;
      setLoaded(next);
      setDraft(next);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setSaving(false);
    }
  }

  if (!sessionId) {
    return <PlanEmpty message="This todo has no agent session yet." />;
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
    return <PlanEmpty message="No plan yet. Run this todo in Plan mode to produce one." />;
  }

  const dirty = draft !== loaded;
  return (
    <div className="flex min-h-0 flex-1 flex-col gap-2 px-4 py-3">
      <div className="flex shrink-0 items-center justify-between gap-2">
        <div className="min-w-0 truncate text-xs text-muted-foreground" title={path}>
          {path}
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
