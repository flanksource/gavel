import { useEffect, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { FixtureEditor } from '@flanksource/clicky-ui/data';
import type { FixtureFenceOption, FixtureFenceSchemas } from '@flanksource/clicky-ui/data';
import { UiBeaker, UiCopy } from '@flanksource/clicky-ui/icons';
import type { TodoItem, TodoRunOptions, TodoSessionDetailResponse } from '../../types';
import { todoQuery } from './format';
import { AcceptanceCriteria } from './AcceptanceCriteria';
import { TodoVerificationAttempts } from './TodoVerificationAttempts';
import { fixtureFenceSchemasFromDocument } from './fixtureSchema';
import {
  PromptRunAdvancedDialog,
  PromptRunButton,
  verificationSpec,
} from './PromptRunButton';
import { defaultRunOptions, runSpec } from './run';

const GAVEL_FIXTURE_FENCES = [
  { info: 'yaml test', label: 'test', description: 'Gavel test options' },
  { info: 'yaml lint', label: 'lint', description: 'Gavel lint options' },
  { info: 'ai', label: 'ai', description: 'Reviewer instructions' },
  { info: 'exec', label: 'exec', description: 'Shell command or script' },
  { info: 'bash', label: 'bash', description: 'Bash command block' },
] satisfies readonly FixtureFenceOption[];

// The response of POST /api/todos/verification/run. Only the parts this view
// acts on are modelled: the recorded results are read back from the attempt
// list, not from this payload.
interface VerificationRunResponse {
  todo?: TodoItem;
  verification?: { allPassed?: boolean; error?: string };
  error?: string;
}

// TodoVerification renders the Verification tab: a FixtureEditor over the
// todo's "## Verification" fixture markdown (explicit Save, since the editor
// fires onChange on every keystroke), an "Add from body" shortcut that seeds
// the fixture from the todo's body, every recorded verification attempt, and
// the acceptance-criteria checklist underneath. Attempt results are read from
// the persisted prompt runs (`attempts`), never held in local state — a reload
// or a tab switch must not erase the evidence of a failed check.
export function TodoVerification({
  dir,
  todo,
  onChanged,
  attempts,
  attemptsError,
}: {
  dir: string;
  todo: TodoItem;
  onChanged: (todo: TodoItem) => void;
  attempts: TodoSessionDetailResponse | null;
  attemptsError?: string;
}) {
  const saved = todo.verificationMarkdown ?? '';
  const [fixture, setFixture] = useState(saved);
  const [schemas, setSchemas] = useState<FixtureFenceSchemas>({});
  const [busy, setBusy] = useState(false);
  const [runBusy, setRunBusy] = useState(false);
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [advancedOptions, setAdvancedOptions] = useState<TodoRunOptions>(defaultRunOptions);
  const [error, setError] = useState('');

  useEffect(() => {
    const controller = new AbortController();
    fetch('/api/todos/verification/schema', { signal: controller.signal })
      .then(async (res) => {
        if (!res.ok) throw new Error(await res.text());
        return res.json();
      })
      .then((doc) => setSchemas(fixtureFenceSchemasFromDocument(doc)))
      .catch((err) => {
        if (err?.name !== 'AbortError') setSchemas({});
      });
    return () => controller.abort();
  }, []);

  // Adopt the server's saved fixture whenever a different todo is shown, or
  // after this todo's own save round-trips back through props.
  useEffect(() => {
    setFixture(saved);
    setError('');
  }, [todo.ref, saved]);

  const dirty = fixture !== saved;

  async function save(): Promise<TodoItem | null> {
    if (busy || !dirty) return todo;
    setBusy(true);
    setError('');
    try {
      const res = await fetch(`/api/todos/verification/fixture?${todoQuery(dir)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref, fixture }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Save failed');
      const refreshed = data as TodoItem;
      onChanged(refreshed);
      return refreshed;
    } catch (err: any) {
      setError(err?.message || 'Save failed');
      return null;
    } finally {
      setBusy(false);
    }
  }

  async function runVerification(options: TodoRunOptions) {
    if (busy || runBusy) return;
    setRunBusy(true);
    setError('');
    try {
      const current = dirty ? await save() : todo;
      if (!current) return;
      const res = await fetch(`/api/todos/verification/run?${todoQuery(dir)}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        // The verification endpoint takes a bare api.Spec under `spec`, not a run
        // payload, so the spec half travels alone.
        body: JSON.stringify({ ref: current.ref, spec: verificationSpec(runSpec(options)) }),
      });
      const data = await res.json() as VerificationRunResponse;
      if (!res.ok) throw new Error(data.error || 'Verification failed');
      // A pre-execution failure (no fixture, no criteria, bad budget) creates no
      // prompt run, so the attempt list can never show it — the banner is its
      // only home.
      if (data.verification?.error) setError(data.verification.error);
      if (data.todo) onChanged(data.todo);
    } catch (err: any) {
      setError(err?.message || 'Verification failed');
    } finally {
      setRunBusy(false);
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
          <PromptRunButton
            scope="verification"
            label="Run verification"
            title="Run the persisted verification fixture"
            disabled={busy || runBusy}
            loading={runBusy}
            onRun={options => void runVerification(options)}
            onAdvanced={options => {
              setAdvancedOptions(options);
              setAdvancedOpen(true);
            }}
          />
        </div>
        <PromptRunAdvancedDialog
          scope="verification"
          open={advancedOpen}
          initial={advancedOptions}
          loading={runBusy}
          onClose={() => setAdvancedOpen(false)}
          onRun={options => {
            setAdvancedOpen(false);
            void runVerification(options);
          }}
        />
        <div className="px-3 py-3">
          <FixtureEditor
            value={fixture}
            onChange={setFixture}
            size="sm"
            schemas={schemas}
            allowedFences={GAVEL_FIXTURE_FENCES}
            placeholder="Write the verification fixture markdown…"
          />
        </div>
        {error && <div className="px-3 pb-3 text-xs text-red-600">{error}</div>}
      </section>

      <TodoVerificationAttempts dir={dir} todoRef={todo.ref} detail={attempts} error={attemptsError} />

      <AcceptanceCriteria dir={dir} todo={todo} onChanged={onChanged} />
    </div>
  );
}
