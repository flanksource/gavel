import { useEffect, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { FixtureEditor } from '@flanksource/clicky-ui/data';
import { UiBeaker, UiCopy } from '@flanksource/clicky-ui/icons';
import type { TodoItem } from '../../types';
import { todoQuery } from './format';
import { AcceptanceCriteria } from './AcceptanceCriteria';

// TodoVerification renders the Verification tab: a FixtureEditor over the
// todo's "## Verification" fixture markdown (explicit Save, since the editor
// fires onChange on every keystroke), an "Add from body" shortcut that seeds
// the fixture from the todo's body, and the existing acceptance-criteria
// checklist underneath.
export function TodoVerification({
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
  const saved = todo.verificationMarkdown ?? '';
  const [fixture, setFixture] = useState(saved);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  // Adopt the server's saved fixture whenever a different todo is shown, or
  // after this todo's own save round-trips back through props.
  useEffect(() => {
    setFixture(saved);
    setError('');
  }, [todo.ref, saved]);

  const dirty = fixture !== saved;

  async function save() {
    if (busy || !dirty) return;
    setBusy(true);
    setError('');
    try {
      const res = await fetch(`/api/todos/verification/fixture?${todoQuery(dir, provider)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref, fixture }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Save failed');
      onChanged(data as TodoItem);
    } catch (err: any) {
      setError(err?.message || 'Save failed');
    } finally {
      setBusy(false);
    }
  }

  function addFromBody() {
    const body = todo.body ?? '';
    if (!body.trim()) return;
    if (fixture.trim() && !window.confirm('Replace the current verification fixture with the todo body?')) return;
    setFixture(body);
  }

  return (
    <div className="space-y-3">
      <section className="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
        <div className="flex flex-wrap items-center gap-2 border-b border-border bg-muted/30 px-3 py-2.5">
          <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground">
            <UiBeaker className="text-xs" />
          </span>
          <span className="min-w-0 flex-1 truncate text-xs font-semibold uppercase text-muted-foreground">
            Verification Fixture
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={addFromBody}
            disabled={busy || !todo.body?.trim()}
            title="Seed the fixture from the todo body"
            className="h-7 gap-1 px-2 text-xs"
          >
            <UiCopy className="text-xs" />
            Add from body
          </Button>
          <Button
            size="sm"
            onClick={save}
            loading={busy}
            disabled={busy || !dirty}
            title="Save the verification fixture"
            className="h-7 gap-1 px-2 text-xs"
          >
            Save
          </Button>
        </div>
        <div className="px-3 py-3">
          <FixtureEditor
            value={fixture}
            onChange={setFixture}
            size="sm"
            placeholder="Write the verification fixture markdown…"
          />
        </div>
        {error && <div className="px-3 pb-3 text-xs text-red-600">{error}</div>}
      </section>

      <AcceptanceCriteria dir={dir} provider={provider} todo={todo} onChanged={onChanged} />
    </div>
  );
}
