import { useCallback, useEffect, useId, useState, type KeyboardEvent } from 'react';
import { Button, DropdownMenu } from '@flanksource/clicky-ui/components';
import { UiCancel, UiCheck, UiChevronDown, UiEdit, UiEye, UiPlay, UiQuestion } from '@flanksource/clicky-ui/icons';
import type { TodoItem, TodoQuestion, TodoRunOptions } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { inputClass, todoQuery } from './format';
import { defaultRunOptions, loadLastTodoRunOptions, rememberTodoRunOptions, runButtonQualifierForOptions, TodoRunDropdownContent, useTodoRunContext } from './run';

export interface PlanApproveResult {
  todo: TodoItem;
  run?: { status: string; message?: string; sessionId?: string };
}

export interface PlanAnswerResult {
  todo: TodoItem;
  sessionId?: string;
  status: string;
}

export type TodoQuestionSelections = Record<number, string>;
export interface TodoAnswerInput {
  answer?: string;
  answers?: Record<string, string>;
  options?: TodoRunOptions;
}
export function buildTodoAnswerInput(questions: TodoQuestion[], selections: TodoQuestionSelections, details: string): TodoAnswerInput {
  const questionTexts = questions.map((question, index) => question.text.trim() || `Question ${index + 1}`);
  const textCounts = questionTexts.reduce<Record<string, number>>((counts, text) => {
    counts[text] = (counts[text] ?? 0) + 1;
    return counts;
  }, {});
  const trimmedDetails = details.trim();
  const answers: Record<string, string> = {};

  questions.forEach((_question, index) => {
    const selection = selections[index]?.trim();
    if (!selection) return;
    const text = questionTexts[index];
    const needsNumber = textCounts[text] > 1 || (trimmedDetails && text === 'Additional details');
    answers[needsNumber ? `${text} (${index + 1})` : text] = selection;
  });

  if (Object.keys(answers).length === 0) return { answer: trimmedDetails };
  if (trimmedDetails) answers['Additional details'] = trimmedDetails;
  return { answers };
}

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
    async (ref: string, input: TodoAnswerInput): Promise<PlanAnswerResult | null> => {
      const trimmedAnswer = input.answer?.trim() ?? '';
      const normalizedAnswers = Object.fromEntries(
        Object.entries(input.answers ?? {})
          .map(([question, value]) => [question.trim(), value.trim()])
          .filter(([question, value]) => question && value),
      );
      const hasAnswers = Object.keys(normalizedAnswers).length > 0;
      if (busy || !ref.trim() || (!trimmedAnswer && !hasAnswers)) return null;
      if (trimmedAnswer && hasAnswers) {
        setError('Answer must use either free text or structured selections, not both');
        return null;
      }
      setBusy(true);
      setError('');
      try {
        const body: Record<string, unknown> = { ref: ref.trim() };
        if (hasAnswers) body.answers = normalizedAnswers;
        else body.answer = trimmedAnswer;
        if (input.options) body.options = input.options;
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

export function PlanApproveButtons({
  busy,
  onApprove,
  onReject,
  onRequestChanges,
  size = 'sm',
}: {
  busy?: boolean;
  onApprove: (run: boolean, options?: TodoRunOptions) => void;
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

export function QuestionsPanel({
  questions,
  selections,
  onSelectionChange,
  disabled,
}: { questions: TodoQuestion[]; selections: TodoQuestionSelections; onSelectionChange: (questionIndex: number, option: string) => void; disabled?: boolean }) {
  const groupId = useId();
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
                <div className="mt-1 flex flex-col gap-1">
                  {q.options.map((opt, j) => (
                    <label
                      key={j}
                      className={`flex items-center gap-1.5 rounded border border-border bg-muted px-1.5 py-1 text-[11px] text-muted-foreground ${disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer hover:bg-muted/80'}`}
                    >
                      <input
                        type="radio"
                        name={`${groupId}-question-${i}`}
                        value={opt}
                        checked={selections[i] === opt}
                        disabled={disabled}
                        onChange={() => onSelectionChange(i, opt)}
                      />
                      <span>{opt}</span>
                    </label>
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
  const [questionSelections, setQuestionSelections] = useState<TodoQuestionSelections>({});
  const [changesText, setChangesText] = useState('');
  const [showChanges, setShowChanges] = useState(false);

  useEffect(() => {
    setAnswerText('');
    setQuestionSelections({});
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
    const result = await answer(todo.ref, buildTodoAnswerInput(todo.questions ?? [], questionSelections, answerText));
    if (result) {
      setAnswerText('');
      setQuestionSelections({});
      onChanged(result.todo);
    }
  }, [answer, todo.ref, todo.questions, questionSelections, answerText, onChanged]);

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
        {todo.questions && todo.questions.length > 0 && (
          <QuestionsPanel
            questions={todo.questions}
            selections={questionSelections}
            onSelectionChange={(questionIndex, option) => {
              setQuestionSelections(previous => ({ ...previous, [questionIndex]: option }));
            }}
            disabled={busy}
          />
        )}
        <AnswerBox
          value={answerText}
          onChange={setAnswerText}
          onSubmit={() => void onAnswer()}
          busy={busy}
          canSubmit={answerText.trim().length > 0 || Object.values(questionSelections).some(value => value.trim().length > 0)}
        />
        {error && <div className="text-xs text-red-600">{error}</div>}
      </div>
    );
  }

  return null;
}

export function AnswerBox({
  value,
  onChange,
  onSubmit,
  busy,
  autoFocus,
  inputRef,
  canSubmit = value.trim().length > 0,
  label = 'Send & resume',
  placeholder = "Answer the agent's questions… (Cmd/Ctrl+Enter to send)",
}: {
  value: string;
  onChange: (v: string) => void;
  onSubmit: () => void;
  busy?: boolean;
  autoFocus?: boolean;
  inputRef?: (el: HTMLTextAreaElement | null) => void;
  canSubmit?: boolean;
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
          disabled={busy || !canSubmit}
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
