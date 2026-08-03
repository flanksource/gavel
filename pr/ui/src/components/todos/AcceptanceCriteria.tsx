import { useEffect, useState, type ComponentType } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@flanksource/clicky-ui/components';
import { UiEdit, UiListDashes, UiTrash, type IconProps } from '@flanksource/clicky-ui/icons';
import type { AcceptanceCriterion, TodoItem } from '../../types';
import { inputClass, todoQuery } from './format';
import { optimisticallySetTodoCaches, setTodoCaches, todoMutationJSON } from './todoMutations';

// AcceptanceCriteria renders a todo's acceptance criteria as a structured,
// editable list (add / edit / remove / toggle, each auto-saved). Verification is
// run from the fixture panel above so there is one definition-of-done path.
export function AcceptanceCriteria({
  dir,
  todo,
  onChanged,
}: {
  dir: string;
  todo: TodoItem;
  onChanged: (todo: TodoItem) => void;
}) {
  const queryClient = useQueryClient();
  const [criteria, setCriteria] = useState<AcceptanceCriterion[]>(todo.criteria ?? []);
  const [editing, setEditing] = useState<number | null>(null);
  const [draft, setDraft] = useState('');
  const [adding, setAdding] = useState('');
  const updateMutation = useMutation({
    mutationKey: ['todos', 'criteria', 'update', { dir: dir.trim(), ref: todo.ref }],
    mutationFn: (next: AcceptanceCriterion[]) => todoMutationJSON<TodoItem>(
      `/api/todos/criteria?${todoQuery(dir)}`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref, criteria: next }),
      },
      'Acceptance criteria update failed',
    ),
    onMutate: (next) => {
      const previous = criteria;
      setCriteria(next);
      return { previous, rollback: optimisticallySetTodoCaches(queryClient, dir, { ...todo, criteria: next }) };
    },
    onError: (_error, _next, context) => {
      context?.rollback();
      if (context) setCriteria(context.previous);
    },
    onSuccess: async (updated) => {
      setCriteria(updated.criteria ?? []);
      await setTodoCaches(queryClient, dir, updated);
      onChanged(updated);
    },
  });
  const busy = updateMutation.isPending;

  // Adopt the server's criteria whenever they change (a save returns the
  // re-parsed list); keep this independent of the verdict so showing a verdict
  // (which also refreshes the todo) does not wipe it.
  useEffect(() => {
    setCriteria(todo.criteria ?? []);
  }, [todo.criteria]);

  // Reset transient view state only when switching to a different todo.
  useEffect(() => {
    setEditing(null);
    updateMutation.reset();
  }, [todo.ref]);

  // save persists the full criteria list and adopts the server's returned todo.
  async function save(next: AcceptanceCriterion[]): Promise<boolean> {
    if (busy) return false;
    try {
      await updateMutation.mutateAsync(next);
      return true;
    } catch {
      return false;
    }
  }

  // addCustom appends a typed custom criterion, clearing the input on success.
  function addCustom() {
    const text = adding.trim();
    if (!text) return;
    save([...criteria, { text }]).then(ok => ok && setAdding(''));
  }

  function saveEdit(i: number) {
    const text = draft.trim();
    if (!text) return;
    // Editing the text makes it a custom criterion (drops any static check id).
    const next = criteria.map((c, idx) => (idx === i ? { text, done: c.done } : c));
    save(next).then(ok => ok && setEditing(null));
  }

  return (
    <section className="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <div className="flex flex-wrap items-center gap-2 border-b border-border bg-muted/30 px-3 py-2.5">
        <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground">
          <UiListDashes className="text-xs" />
        </span>
        <span className="min-w-0 flex-1 truncate text-xs font-semibold uppercase text-muted-foreground">
          Acceptance Criteria
        </span>
        {criteria.length > 0 && (
          <span className="rounded-full border border-border bg-background px-1.5 py-0.5 text-[11px] tabular-nums text-muted-foreground">
            {criteria.length}
          </span>
        )}
      </div>

      <div className="space-y-2 px-3 py-3">
        {criteria.length === 0 && (
          <p className="text-sm text-muted-foreground">No acceptance criteria yet — add one below.</p>
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
                  <IconButton icon={UiEdit} label="Edit criterion" disabled={busy} onClick={() => { setDraft(c.text); setEditing(i); }} />
                  <IconButton icon={UiTrash} label="Remove criterion" disabled={busy} onClick={() => save(criteria.filter((_, idx) => idx !== i))} />
                </span>
              )}
            </li>
          ))}
        </ul>

        <input
          className={inputClass}
          value={adding}
          disabled={busy}
          onChange={e => setAdding(e.currentTarget.value)}
          onKeyDown={e => {
            if (e.key === 'Enter') addCustom();
          }}
          placeholder="Add a criterion…"
          aria-label="Add acceptance criterion"
        />

        {updateMutation.error && <div className="text-xs text-red-600">{updateMutation.error.message}</div>}
      </div>
    </section>
  );
}

function IconButton({ icon, label, onClick, disabled }: { icon: ComponentType<IconProps>; label: string; onClick: () => void; disabled?: boolean }) {
  const Icon = icon;
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
      <Icon className="text-xs" />
    </Button>
  );
}
