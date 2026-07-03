import { useCallback, useEffect, useState, type KeyboardEvent } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { UiCheck, UiEye, UiPlay, UiQuestion } from '@flanksource/clicky-ui/icons';
import type { TodoItem, TodoQuestion, TodoRunOptions } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { inputClass, todoQuery } from './format';
import { defaultRunOptions } from './run';

// Plan-review actions shared by the TodoDetail surfacing and Plan Review mode:
// approving a reviewed plan (optionally chaining the implementing run) and
// answering an agent's blocking questions. Both hit the domain endpoints that
// validate the todo's state server-side, so a stale view gets a 409 rather than
// silently clobbering a concurrent change.

// Server response of POST /api/todos/plan/approve.
export interface PlanApproveResult {
  todo: TodoItem;
  run?: { status: string; message?: string; sessionId?: string };
}

// Server response of POST /api/todos/answer.
export interface PlanAnswerResult {
  todo: TodoItem;
  sessionId?: string;
  status: string;
}

// usePlanActions exposes approve/answer against a workspace, with shared
// busy/error state so a bar or a detail pane can drive either action.
export function usePlanActions(dir: string, provider: string) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const reset = useCallback(() => setError(''), []);

  const approve = useCallback(
    async (ref: string, opts: { run?: boolean; options?: TodoRunOptions } = {}): Promise<PlanApproveResult | null> => {
      if (busy || !ref.trim()) return null;
      setBusy(true);
      setError('');
      try {
        const body: Record<string, unknown> = { ref: ref.trim(), run: !!opts.run };
        if (opts.run) body.options = opts.options ?? defaultRunOptions;
        const res = await fetch(`/api/todos/plan/approve?${todoQuery(dir, provider)}`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body),
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Approve failed');
        return data as PlanApproveResult;
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Approve failed');
        return null;
      } finally {
        setBusy(false);
      }
    },
    [dir, provider, busy],
  );

  const answer = useCallback(
    async (ref: string, text: string, options?: TodoRunOptions): Promise<PlanAnswerResult | null> => {
      const trimmed = text.trim();
      if (busy || !ref.trim() || !trimmed) return null;
      setBusy(true);
      setError('');
      try {
        const body: Record<string, unknown> = { ref: ref.trim(), answer: trimmed };
        if (options) body.options = options;
        const res = await fetch(`/api/todos/answer?${todoQuery(dir, provider)}`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body),
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Answer failed');
        return data as PlanAnswerResult;
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Answer failed');
        return null;
      } finally {
        setBusy(false);
      }
    },
    [dir, provider, busy],
  );

  return { busy, error, reset, approve, answer };
}

// PlanApproveButtons is the review action pair: primary "Approve & Run" (→ the
// implementing run) and secondary "Approve" (→ pending, run later). Rendered
// whenever a todo is in `review`.
export function PlanApproveButtons({
  busy,
  onApprove,
  size = 'sm',
}: {
  busy?: boolean;
  onApprove: (run: boolean) => void;
  size?: 'sm' | 'default';
}) {
  const Icon = busy ? Spinner : UiPlay;
  return (
    <div className="inline-flex items-center gap-1.5">
      <Button
        type="button"
        size={size}
        variant="default"
        disabled={busy}
        onClick={() => onApprove(true)}
        className="inline-flex items-center gap-1 bg-amber-600 text-white hover:bg-amber-700 disabled:opacity-50"
        title="Approve the plan and start the implementing run"
      >
        <Icon className="text-xs" />
        Approve &amp; Run
      </Button>
      <Button
        type="button"
        size={size}
        variant="outline"
        disabled={busy}
        onClick={() => onApprove(false)}
        className="inline-flex items-center gap-1 disabled:opacity-50"
        title="Approve the plan (moves to pending); run it later"
      >
        <UiCheck className="text-xs" />
        Approve
      </Button>
    </div>
  );
}

// QuestionsPanel lists an agent's blocking questions (with any context and
// suggested options) above the answer box, so a reviewer sees exactly what the
// agent is stuck on.
export function QuestionsPanel({ questions }: { questions: TodoQuestion[] }) {
  if (questions.length === 0) return null;
  return (
    <ul className="flex flex-col gap-2">
      {questions.map((q, i) => (
        <li key={i} className="rounded-md border border-purple-500/30 bg-purple-500/5 p-2 text-sm">
          <div className="flex items-start gap-1.5">
            <UiQuestion className="mt-0.5 shrink-0 text-purple-600" />
            <div className="min-w-0">
              <div className="font-medium text-foreground">{q.text}</div>
              {q.context && <div className="mt-0.5 text-xs text-muted-foreground">{q.context}</div>}
              {q.options && q.options.length > 0 && (
                <div className="mt-1 flex flex-wrap gap-1">
                  {q.options.map((opt, j) => (
                    <span key={j} className="rounded border border-border bg-muted px-1.5 py-0.5 text-[11px] text-muted-foreground">
                      {opt}
                    </span>
                  ))}
                </div>
              )}
            </div>
          </div>
        </li>
      ))}
    </ul>
  );
}

// TodoReviewBanner surfaces the approve/answer action inline in the detail view,
// so a `review`/`ask` todo can be acted on without entering Plan Review mode. It
// is self-contained (owns its answer draft + action state) so callers inject it
// with a single line; onChanged refreshes the todo after a successful action.
export function TodoReviewBanner({
  todo,
  dir,
  provider,
  onChanged,
}: {
  todo: TodoItem;
  dir: string;
  provider: string;
  onChanged: (todo: TodoItem) => void;
}) {
  const { busy, error, reset, approve, answer } = usePlanActions(dir, provider);
  const [answerText, setAnswerText] = useState('');

  useEffect(() => {
    setAnswerText('');
    reset();
  }, [todo.ref, reset]);

  const onApprove = useCallback(
    async (run: boolean) => {
      const result = await approve(todo.ref, { run, options: defaultRunOptions });
      if (result) onChanged(result.todo);
    },
    [approve, todo.ref, onChanged],
  );

  const onAnswer = useCallback(async () => {
    const result = await answer(todo.ref, answerText);
    if (result) {
      setAnswerText('');
      onChanged(result.todo);
    }
  }, [answer, todo.ref, answerText, onChanged]);

  if (todo.status === 'review') {
    return (
      <div className="flex flex-col gap-2 border-b border-amber-500/30 bg-amber-500/5 px-4 py-2">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <span className="inline-flex items-center gap-1.5 text-sm font-medium text-amber-700">
            <UiEye className="text-xs" />
            Plan ready for review
          </span>
          <PlanApproveButtons busy={busy} onApprove={run => void onApprove(run)} />
        </div>
        {error && <div className="text-xs text-red-600">{error}</div>}
      </div>
    );
  }

  if (todo.status === 'ask') {
    return (
      <div className="flex flex-col gap-2 border-b border-purple-500/30 bg-purple-500/5 px-4 py-2">
        <span className="inline-flex items-center gap-1.5 text-sm font-medium text-purple-700">
          <UiQuestion className="text-xs" />
          The agent needs your input to continue
        </span>
        {todo.questions && todo.questions.length > 0 && <QuestionsPanel questions={todo.questions} />}
        <AnswerBox value={answerText} onChange={setAnswerText} onSubmit={() => void onAnswer()} busy={busy} />
        {error && <div className="text-xs text-red-600">{error}</div>}
      </div>
    );
  }

  return null;
}

// AnswerBox is the textarea + submit for answering an `ask` todo. Cmd/Ctrl+Enter
// submits; the value is controlled by the caller so a queue can reset it per
// todo. `autoFocus`/`inputRef` let review mode focus it with the `r` key.
export function AnswerBox({
  value,
  onChange,
  onSubmit,
  busy,
  autoFocus,
  inputRef,
}: {
  value: string;
  onChange: (v: string) => void;
  onSubmit: () => void;
  busy?: boolean;
  autoFocus?: boolean;
  inputRef?: (el: HTMLTextAreaElement | null) => void;
}) {
  const onKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      onSubmit();
    }
  };
  return (
    <div className="flex flex-col gap-1.5">
      <textarea
        ref={inputRef}
        autoFocus={autoFocus}
        className={`${inputClass} h-auto min-h-[4rem] resize-y`}
        placeholder="Answer the agent's questions… (Cmd/Ctrl+Enter to send)"
        value={value}
        onChange={e => onChange(e.currentTarget.value)}
        onKeyDown={onKeyDown}
      />
      <div className="flex items-center justify-end">
        <Button
          type="button"
          size="sm"
          variant="default"
          disabled={busy || !value.trim()}
          onClick={onSubmit}
          className="inline-flex items-center gap-1 bg-purple-600 text-white hover:bg-purple-700 disabled:opacity-50"
        >
          {busy ? <Spinner className="text-xs" /> : <UiCheck className="text-xs" />}
          Send &amp; resume
        </Button>
      </div>
    </div>
  );
}
