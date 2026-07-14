import { useCallback, useEffect, useState, type KeyboardEvent } from 'react';
import { Button, DropdownMenu } from '@flanksource/clicky-ui/components';
import { UiCancel, UiCheck, UiChevronDown, UiEdit, UiEye, UiPlay, UiQuestion } from '@flanksource/clicky-ui/icons';
import type { TodoItem, TodoQuestion, TodoRunOptions } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { inputClass, todoQuery } from './format';
import {
  defaultRunOptions,
  loadLastTodoRunOptions,
  rememberTodoRunOptions,
  runButtonQualifierForOptions,
  TodoRunDropdownContent,
  useTodoRunContext,
} from './run';

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
export function usePlanActions(dir: string) {
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
        const res = await fetch(`/api/todos/plan/approve?${todoQuery(dir)}`, {
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
    [dir, busy],
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
        const res = await fetch(`/api/todos/answer?${todoQuery(dir)}`, {
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
    [dir, busy],
  );

  const reject = useCallback(
    async (ref: string): Promise<PlanApproveResult | null> => {
      if (busy || !ref.trim()) return null;
      setBusy(true);
      setError('');
      try {
        const res = await fetch(`/api/todos/plan/reject?${todoQuery(dir)}`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ ref: ref.trim() }),
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Reject failed');
        return data as PlanApproveResult;
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Reject failed');
        return null;
      } finally {
        setBusy(false);
      }
    },
    [dir, busy],
  );

  const revise = useCallback(
    async (ref: string, feedback: string): Promise<PlanAnswerResult | null> => {
      const trimmed = feedback.trim();
      if (busy || !ref.trim() || !trimmed) return null;
      setBusy(true);
      setError('');
      try {
        const options = loadLastTodoRunOptions('plan');
        const res = await fetch(`/api/todos/plan/revise?${todoQuery(dir)}`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ ref: ref.trim(), feedback: trimmed, options }),
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Revise failed');
        return data as PlanAnswerResult;
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Revise failed');
        return null;
      } finally {
        setBusy(false);
      }
    },
    [dir, busy],
  );

  return { busy, error, reset, approve, answer, reject, revise };
}

// PlanApproveButtons is the review action pair: primary "Approve & Run" (→ the
// implementing run) and secondary "Approve" (→ pending, run later). Rendered
// whenever a todo is in `review`.
export function PlanApproveButtons({
  busy,
  onApprove,
  onReject,
  onRequestChanges,
  size = 'sm',
}: {
  busy?: boolean;
  onApprove: (run: boolean, options?: TodoRunOptions) => void;
  // Optional secondary actions: reject the plan (→ pending) and request changes
  // (reveal a feedback box that re-plans on the same session). Rendered only when
  // the caller wires them.
  onReject?: () => void;
  onRequestChanges?: () => void;
  size?: 'sm' | 'default';
}) {
  const Icon = busy ? Spinner : UiPlay;
  const context = useTodoRunContext(!busy);
  const [selectedOptions, setSelectedOptions] = useState<TodoRunOptions | null>(null);
  const runOptions = selectedOptions ?? loadLastTodoRunOptions('run', context);
  const runLabel = `Approve & Run ${runButtonQualifierForOptions(runOptions, context)}`;

  function approveRun(options: TodoRunOptions, advanced = false) {
    const remembered = rememberTodoRunOptions('run', options, advanced);
    setSelectedOptions(remembered);
    onApprove(true, remembered);
  }

  return (
    <div className="inline-flex flex-wrap items-center gap-1.5">
      <div className="inline-flex h-8 shrink-0 items-stretch rounded-md border border-border bg-background">
        <Button
          variant="ghost"
          type="button"
          disabled={busy}
          onClick={() => approveRun(runOptions)}
          className="inline-flex h-8 items-center gap-1 rounded-none border-r border-border px-2 text-xs font-medium text-foreground hover:bg-muted disabled:opacity-50"
          title="Approve the plan and start the implementing run"
        >
          <Icon className="text-xs" />
          <span>{runLabel}</span>
        </Button>
        <DropdownMenu
          align="right"
          menuLabel="Approve and run options"
          menuClassName="max-h-[70vh] w-[320px] max-w-[calc(100vw-24px)] overflow-y-auto"
          trigger={
            <Button
              variant="ghost"
              size="icon"
              type="button"
              disabled={busy}
              title="Approve & Run options"
              aria-label="Approve & Run options"
              className="h-8 w-7 rounded-none text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
            >
              <UiChevronDown className="text-xs" />
            </Button>
          }
        >
          {close => (
            <TodoRunDropdownContent
              context={context}
              initialAction="run"
              closeParent={close}
              onSelect={(_selectedAction, options, advanced) => approveRun(options, advanced)}
              showAdvanced={false}
            />
          )}
        </DropdownMenu>
      </div>
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
      {onRequestChanges && (
        <Button
          type="button"
          size={size}
          variant="outline"
          disabled={busy}
          onClick={onRequestChanges}
          className="inline-flex items-center gap-1 disabled:opacity-50"
          title="Send the agent feedback and re-plan on the same session"
        >
          <UiEdit className="text-xs" />
          Request changes
        </Button>
      )}
      {onReject && (
        <Button
          type="button"
          size={size}
          variant="ghost"
          disabled={busy}
          onClick={onReject}
          className="inline-flex items-center gap-1 text-red-600 hover:bg-red-500/10 hover:text-red-700 disabled:opacity-50"
          title="Reject the plan; the todo returns to pending"
        >
          <UiCancel className="text-xs" />
          Reject
        </Button>
      )}
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
  onChanged,
}: {
  todo: TodoItem;
  dir: string;
  onChanged: (todo: TodoItem) => void;
}) {
  const { busy, error, reset, approve, answer, reject, revise } = usePlanActions(dir);
  const [answerText, setAnswerText] = useState('');
  const [changesText, setChangesText] = useState('');
  const [showChanges, setShowChanges] = useState(false);

  useEffect(() => {
    setAnswerText('');
    setChangesText('');
    setShowChanges(false);
    reset();
  }, [todo.ref, reset]);

  const onApprove = useCallback(
    async (run: boolean, options?: TodoRunOptions) => {
      const result = await approve(todo.ref, { run, options: run ? options ?? loadLastTodoRunOptions('run') : undefined });
      if (result) onChanged(result.todo);
    },
    [approve, todo.ref, onChanged],
  );

  const onReject = useCallback(async () => {
    const result = await reject(todo.ref);
    if (result) onChanged(result.todo);
  }, [reject, todo.ref, onChanged]);

  const onRevise = useCallback(async () => {
    const result = await revise(todo.ref, changesText);
    if (result) {
      setChangesText('');
      setShowChanges(false);
      onChanged(result.todo);
    }
  }, [revise, todo.ref, changesText, onChanged]);

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
          <PlanApproveButtons
            busy={busy}
            onApprove={(run, options) => void onApprove(run, options)}
            onReject={() => void onReject()}
            onRequestChanges={() => setShowChanges(v => !v)}
          />
        </div>
        {showChanges && (
          <AnswerBox
            value={changesText}
            onChange={setChangesText}
            onSubmit={() => void onRevise()}
            busy={busy}
            label="Send changes & re-plan"
            placeholder="Describe what to change in the plan… (Cmd/Ctrl+Enter to send)"
            autoFocus
          />
        )}
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
  label = 'Send & resume',
  placeholder = "Answer the agent's questions… (Cmd/Ctrl+Enter to send)",
}: {
  value: string;
  onChange: (v: string) => void;
  onSubmit: () => void;
  busy?: boolean;
  autoFocus?: boolean;
  inputRef?: (el: HTMLTextAreaElement | null) => void;
  // label/placeholder let the same box serve answering an `ask` and requesting
  // changes on a `review` plan.
  label?: string;
  placeholder?: string;
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
        placeholder={placeholder}
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
          {label}
        </Button>
      </div>
    </div>
  );
}
