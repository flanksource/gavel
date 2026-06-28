import { useEffect, useMemo, useState } from 'react';
import { Button, Combobox } from '@flanksource/clicky-ui/components';
import type { AcceptanceCriterion, CriteriaCatalogItem, CriterionResult, TodoItem, VerifyResult } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { inputClass, todoQuery } from './format';

// AcceptanceCriteria renders a todo's acceptance criteria as a structured,
// editable list (add / edit / remove / toggle, each auto-saved) with an AI
// "Draft" action and a "Verify" action that reviews the structured verdict.
export function AcceptanceCriteria({
  dir,
  provider,
  todo,
  onChanged,
}: {
  dir: string;
  provider: string;
  todo: TodoItem;
  onChanged: (todo: TodoItem) => void;
}) {
  const [criteria, setCriteria] = useState<AcceptanceCriterion[]>(todo.criteria ?? []);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [editing, setEditing] = useState<number | null>(null);
  const [draft, setDraft] = useState('');
  const [verifyBusy, setVerifyBusy] = useState(false);
  const [verdict, setVerdict] = useState<VerifyResult | null>(null);
  const [catalog, setCatalog] = useState<CriteriaCatalogItem[]>([]);

  // Load the standard-check catalog once so the add control can offer them.
  useEffect(() => {
    let active = true;
    fetch('/api/todos/criteria/catalog')
      .then(res => (res.ok ? res.json() : []))
      .then(items => active && setCatalog(items as CriteriaCatalogItem[]))
      .catch(() => {});
    return () => {
      active = false;
    };
  }, []);

  // Standard checks not already added, as grouped combobox options. Typing a
  // value not in this list adds a custom criterion (allowCustomValue).
  const addOptions = useMemo(() => {
    const added = new Set(criteria.filter(c => c.checkId).map(c => c.checkId));
    return catalog
      .filter(item => !added.has(item.id))
      .map(item => ({ value: item.id, label: item.description, group: item.category }));
  }, [catalog, criteria]);

  // Adopt the server's criteria whenever they change (a save returns the
  // re-parsed list); keep this independent of the verdict so showing a verdict
  // (which also refreshes the todo) does not wipe it.
  useEffect(() => {
    setCriteria(todo.criteria ?? []);
  }, [todo.criteria]);

  // Reset transient view state only when switching to a different todo.
  useEffect(() => {
    setEditing(null);
    setVerdict(null);
    setError('');
  }, [todo.ref]);

  // save persists the full criteria list and adopts the server's returned todo.
  async function save(next: AcceptanceCriterion[]): Promise<boolean> {
    if (busy) return false;
    setBusy(true);
    setError('');
    setCriteria(next); // optimistic
    try {
      const res = await fetch(`/api/todos/criteria?${todoQuery(dir, provider)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref, criteria: next }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Save failed');
      onChanged(data as TodoItem);
      return true;
    } catch (err: any) {
      setError(err?.message || 'Save failed');
      setCriteria(todo.criteria ?? []); // revert
      return false;
    } finally {
      setBusy(false);
    }
  }

  // addFromCombobox adds the just-selected standard checks and/or a typed custom
  // criterion. The combobox value is kept empty, so `next` holds exactly the
  // newly chosen values: catalog ids become static criteria, anything else a
  // custom one.
  function addFromCombobox(next: string[]) {
    const byId = new Map(catalog.map(c => [c.id, c]));
    const additions: AcceptanceCriterion[] = [];
    for (const v of next) {
      const item = byId.get(v);
      if (item) additions.push({ checkId: item.id, text: item.description });
      else if (v.trim()) additions.push({ text: v.trim() });
    }
    if (additions.length) save([...criteria, ...additions]);
  }

  function saveEdit(i: number) {
    const text = draft.trim();
    if (!text) return;
    // Editing the text makes it a custom criterion (drops any static check id).
    const next = criteria.map((c, idx) => (idx === i ? { text, done: c.done } : c));
    save(next).then(ok => ok && setEditing(null));
  }

  async function callTodoAction(path: string): Promise<void> {
    setBusy(true);
    setError('');
    try {
      const res = await fetch(`/api/todos/${path}?${todoQuery(dir, provider)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Request failed');
      onChanged(data as TodoItem);
    } catch (err: any) {
      setError(err?.message || 'Request failed');
    } finally {
      setBusy(false);
    }
  }

  async function runVerify() {
    if (verifyBusy) return;
    setVerifyBusy(true);
    setError('');
    try {
      const res = await fetch(`/api/todos/verify?${todoQuery(dir, provider)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Verification failed');
      setVerdict(data.result as VerifyResult);
      if (data.todo) onChanged(data.todo as TodoItem);
    } catch (err: any) {
      setError(err?.message || 'Verification failed');
    } finally {
      setVerifyBusy(false);
    }
  }

  const failedChecks = Object.entries(verdict?.checks ?? {}).filter(([, c]) => !c.pass);

  return (
    <section className="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <div className="flex flex-wrap items-center gap-2 border-b border-border bg-muted/30 px-3 py-2.5">
        <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground">
          <GavelIcon name="codicon:checklist" className="text-xs" />
        </span>
        <span className="min-w-0 flex-1 truncate text-xs font-semibold uppercase text-muted-foreground">
          Acceptance Criteria
        </span>
        {criteria.length > 0 && (
          <span className="rounded-full border border-border bg-background px-1.5 py-0.5 text-[11px] tabular-nums text-muted-foreground">
            {criteria.length}
          </span>
        )}
        <Button
          variant="ghost"
          size="sm"
          onClick={() => callTodoAction('criteria/generate')}
          loading={busy}
          disabled={busy || verifyBusy}
          title="Draft acceptance criteria with AI"
          className="h-7 gap-1 px-2 text-xs"
        >
          <GavelIcon name="codicon:sparkle" className="text-xs" />
          Draft with AI
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={runVerify}
          loading={verifyBusy}
          disabled={busy || verifyBusy}
          title="Verify the committed work against these criteria"
          className="h-7 gap-1 px-2 text-xs"
        >
          <GavelIcon name="codicon:beaker" className="text-xs" />
          Verify
        </Button>
      </div>

      <div className="space-y-2 px-3 py-3">
        {criteria.length === 0 && (
          <p className="text-sm text-muted-foreground">No acceptance criteria yet — add one below or draft with AI.</p>
        )}

        <ul className="space-y-1">
          {criteria.map((c, i) => (
            <li key={i} className="group flex items-start gap-2 rounded-md border border-transparent px-2 py-1.5 hover:border-border hover:bg-muted/40">
              <input
                type="checkbox"
                checked={!!c.done}
                disabled={busy}
                onChange={() => save(criteria.map((x, idx) => (idx === i ? { ...x, done: !x.done } : x)))}
                className="mt-1 h-4 w-4 shrink-0 accent-primary"
                aria-label={c.done ? 'Mark not done' : 'Mark done'}
              />
              {editing === i ? (
                <input
                  className={inputClass}
                  value={draft}
                  disabled={busy}
                  autoFocus
                  onChange={e => setDraft(e.currentTarget.value)}
                  onKeyDown={e => {
                    if (e.key === 'Enter') saveEdit(i);
                    else if (e.key === 'Escape') setEditing(null);
                  }}
                  onBlur={() => saveEdit(i)}
                  aria-label="Edit criterion"
                />
              ) : (
                <span className={`min-w-0 flex-1 text-sm ${c.done ? 'text-muted-foreground line-through' : ''}`}>
                  {c.checkId && (
                    <span className="mr-1.5 rounded bg-muted px-1 py-0.5 font-mono text-[10px] uppercase tracking-wide text-muted-foreground">
                      {c.checkId}
                    </span>
                  )}
                  {c.text}
                </span>
              )}
              {editing !== i && (
                <span className="flex shrink-0 items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
                  <IconButton icon="codicon:edit" label="Edit criterion" disabled={busy} onClick={() => { setDraft(c.text); setEditing(i); }} />
                  <IconButton icon="codicon:trash" label="Remove criterion" disabled={busy} onClick={() => save(criteria.filter((_, idx) => idx !== i))} />
                </span>
              )}
            </li>
          ))}
        </ul>

        <Combobox
          multiple
          value={[]}
          options={addOptions}
          onChange={addFromCombobox}
          disabled={busy}
          allowCustomValue
          placeholder="Add criteria — pick standard checks or type a custom one…"
          ariaLabel="Add acceptance criteria"
        />

        {error && <div className="text-xs text-red-600">{error}</div>}
        {verdict && <VerifyReview verdict={verdict} failedChecks={failedChecks} />}
      </div>
    </section>
  );
}

function IconButton({ icon, label, onClick, disabled }: { icon: string; label: string; onClick: () => void; disabled?: boolean }) {
  return (
    <Button
      variant="ghost"
      size="icon"
      type="button"
      onClick={onClick}
      disabled={disabled}
      title={label}
      aria-label={label}
      className="inline-flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
    >
      <GavelIcon name={icon} className="text-xs" />
    </Button>
  );
}

// VerifyReview renders the structured verdict: score + implemented badge,
// per-criterion met/unmet rows, failed static checks, and completeness.
function VerifyReview({ verdict, failedChecks }: { verdict: VerifyResult; failedChecks: [string, { pass: boolean }][] }) {
  const scoreColor = verdict.score >= 80 ? 'text-emerald-600' : verdict.score >= 60 ? 'text-amber-600' : 'text-red-600';
  return (
    <div className="space-y-2 rounded-md border border-border bg-muted/30 p-2.5">
      <div className="flex items-center gap-2 text-sm">
        <span className={`font-semibold ${scoreColor}`}>Score {verdict.score}/100</span>
        {verdict.implemented !== undefined && (
          <span className={`inline-flex items-center gap-1 font-medium ${verdict.implemented ? 'text-emerald-600' : 'text-red-600'}`}>
            <GavelIcon name={verdict.implemented ? 'codicon:pass' : 'codicon:error'} className="text-sm" />
            {verdict.implemented ? 'Implemented' : 'Not implemented'}
          </span>
        )}
      </div>
      {(verdict.acceptance_criteria?.length ?? 0) > 0 && (
        <ul className="space-y-1">
          {verdict.acceptance_criteria!.map((c, i) => <ResultRow key={i} result={c} />)}
        </ul>
      )}
      {failedChecks.length > 0 && (
        <div>
          <p className="text-xs font-semibold uppercase text-muted-foreground">Failed checks</p>
          <ul className="mt-1 space-y-0.5">
            {failedChecks.map(([id]) => (
              <li key={id} className="flex items-center gap-1.5 text-sm text-red-600">
                <GavelIcon name="codicon:error" className="text-xs" />
                {id}
              </li>
            ))}
          </ul>
        </div>
      )}
      {verdict.completeness?.summary && (
        <p className="text-xs text-muted-foreground">
          <span className="font-semibold">Completeness:</span> {verdict.completeness.summary}
        </p>
      )}
    </div>
  );
}

function ResultRow({ result }: { result: CriterionResult }) {
  return (
    <li className="text-sm">
      <span className="flex items-start gap-1.5">
        <GavelIcon
          name={result.met ? 'codicon:pass' : 'codicon:error'}
          className={`mt-0.5 shrink-0 text-sm ${result.met ? 'text-emerald-600' : 'text-red-600'}`}
        />
        <span className="min-w-0">{result.criterion}</span>
      </span>
      {result.evidence?.map((e, i) => (
        <span key={i} className="ml-5 block text-xs text-muted-foreground">
          {e.file}{e.line ? `:${e.line}` : ''} — {e.message}
        </span>
      ))}
    </li>
  );
}
