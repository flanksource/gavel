import { useCallback, useEffect, useId, useState, type KeyboardEvent } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Button } from '@flanksource/clicky-ui/components';
import { UiCancel, UiCheck, UiEdit, UiEye, UiQuestion } from '@flanksource/clicky-ui/icons';
import type { TodoItem, TodoQuestion, TodoRunOptions } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { inputClass, todoQuery } from './format';
import { defaultRunOptions, loadLastTodoRunOptions, normalizeRunOptions, requestStepFor, type TodoRunAction } from './run';
import { PromptRunAdvancedDialog, PromptRunButton } from './PromptRunButton';
import { invalidateTodoWorkflowCaches, todoMutationJSON } from './todoMutations';

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

// runRequestOptions sanitizes a TodoRunOptions for the plan-review endpoints'
// `options` field. /api/todos/plan/approve, /plan/revise, and /answer all
// continue a run through the same seam as /api/todos/run (see todo_review.go's
// continuation), which decodes just as strictly and rejects driver/runMode/
// prompt at the top level — only step/spec/resume/force may travel.
//
// `action`, when given, pins the step to the phase the endpoint always
// continues into (approve always runs, revise always re-plans) via the same
// normalizeRunOptions/requestStepFor pairing run.tsx itself uses to build
// /api/todos/run's body, so the two can never disagree on what a chosen
// options object dispatches as. Left out (answer, whose continuation resumes
// whatever step the turn was already on), the fields are stripped without
// asserting a step of its own — the server's own resolution decides, and a
// mismatched guess here would otherwise make the continuation reject the
// request outright.
function runRequestOptions(options: TodoRunOptions, action?: TodoRunAction): TodoRunOptions {
  const normalized = action ? normalizeRunOptions(action, options) : options;
  return {
    step: action ? requestStepFor(normalized) : undefined,
    spec: normalized.spec,
    resume: normalized.resume,
    force: normalized.force,
  };
}

function usePlanActionMutation<TResult extends { todo: TodoItem }, TVariables>(
  dir: string,
  action: string,
  request: (variables: TVariables) => { path: string; body: Record<string, unknown>; context: string },
) {
  const client = useQueryClient();
  return useMutation({
    mutationKey: ['todos', 'plan-action', action, { dir: dir.trim() }],
    mutationFn: (variables: TVariables) => {
      const { path, body, context } = request(variables);
      return todoMutationJSON<TResult>(
        `${path}?${todoQuery(dir)}`,
        { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) },
        context,
      );
    },
    onSuccess: result => invalidateTodoWorkflowCaches(client, dir, result.todo),
  });
}

export function usePlanActions(dir: string) {
  const [validationError, setValidationError] = useState('');
  const approveMutation = usePlanActionMutation<PlanApproveResult, { ref: string; opts: { run?: boolean; options?: TodoRunOptions } }>(
    dir,
    'approve',
    ({ ref, opts }) => {
      const body: Record<string, unknown> = { ref, run: !!opts.run };
      if (opts.run) body.options = runRequestOptions(opts.options ?? defaultRunOptions, 'run');
      return { path: '/api/todos/plan/approve', body, context: `Failed to approve plan for todo ${ref}` };
    },
  );
  const answerMutation = usePlanActionMutation<PlanAnswerResult, { ref: string; body: Record<string, unknown> }>(
    dir,
    'answer',
    ({ ref, body }) => ({ path: '/api/todos/answer', body, context: `Failed to answer todo ${ref}` }),
  );
  const rejectMutation = usePlanActionMutation<PlanApproveResult, string>(
    dir,
    'reject',
    ref => ({ path: '/api/todos/plan/reject', body: { ref }, context: `Failed to reject plan for todo ${ref}` }),
  );
  const reviseMutation = usePlanActionMutation<PlanAnswerResult, { ref: string; feedback: string }>(
    dir,
    'revise',
    ({ ref, feedback }) => ({
      path: '/api/todos/plan/revise',
      body: { ref, feedback, options: runRequestOptions(loadLastTodoRunOptions('plan'), 'plan') },
      context: `Failed to request plan changes for todo ${ref}`,
    }),
  );
  const mutations = [approveMutation, answerMutation, rejectMutation, reviseMutation];
  const busy = mutations.some(mutation => mutation.isPending);
  const mutationError = mutations.find(mutation => mutation.error)?.error;
  const error = validationError || (mutationError instanceof Error ? mutationError.message : '');
  const reset = useCallback(() => {
    setValidationError('');
    for (const mutation of mutations) mutation.reset();
  }, [approveMutation.reset, answerMutation.reset, rejectMutation.reset, reviseMutation.reset]);

  const approve = useCallback(async (ref: string, opts: { run?: boolean; options?: TodoRunOptions } = {}) => {
    const cleaned = ref.trim();
    if (busy || !cleaned) return null;
    setValidationError('');
    try { return await approveMutation.mutateAsync({ ref: cleaned, opts }); } catch { return null; }
  }, [approveMutation.mutateAsync, busy]);

  const answer = useCallback(async (ref: string, input: TodoAnswerInput) => {
    const trimmedAnswer = input.answer?.trim() ?? '';
    const normalizedAnswers = Object.fromEntries(Object.entries(input.answers ?? {})
      .map(([question, value]) => [question.trim(), value.trim()]).filter(([question, value]) => question && value));
    const hasAnswers = Object.keys(normalizedAnswers).length > 0;
    if (busy || !ref.trim() || (!trimmedAnswer && !hasAnswers)) return null;
    if (trimmedAnswer && hasAnswers) {
      setValidationError('Answer must use either free text or structured selections, not both');
      return null;
    }
    setValidationError('');
    const body: Record<string, unknown> = { ref: ref.trim() };
    if (hasAnswers) body.answers = normalizedAnswers;
    else body.answer = trimmedAnswer;
    if (input.options) body.options = runRequestOptions(input.options);
    try { return await answerMutation.mutateAsync({ ref: ref.trim(), body }); } catch { return null; }
  }, [answerMutation.mutateAsync, busy]);

  const reject = useCallback(async (ref: string) => {
    const cleaned = ref.trim();
    if (busy || !cleaned) return null;
    setValidationError('');
    try { return await rejectMutation.mutateAsync(cleaned); } catch { return null; }
  }, [rejectMutation.mutateAsync, busy]);

  const revise = useCallback(async (ref: string, feedback: string) => {
    const cleanedRef = ref.trim();
    const cleanedFeedback = feedback.trim();
    if (busy || !cleanedRef || !cleanedFeedback) return null;
    setValidationError('');
    try { return await reviseMutation.mutateAsync({ ref: cleanedRef, feedback: cleanedFeedback }); } catch { return null; }
  }, [reviseMutation.mutateAsync, busy]);

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
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [advancedOptions, setAdvancedOptions] = useState<TodoRunOptions>(defaultRunOptions);

  return (
    <div className="inline-flex flex-wrap items-center gap-1.5">
      <PromptRunButton
        scope="approval"
        label="Approve & Run"
        title="Approve the plan and start the implementing run"
        disabled={busy}
        loading={busy}
        onRun={options => onApprove(true, options)}
        onAdvanced={options => {
          setAdvancedOptions(options);
          setAdvancedOpen(true);
        }}
      />
      <PromptRunAdvancedDialog
        scope="approval"
        open={advancedOpen}
        initial={advancedOptions}
        onClose={() => setAdvancedOpen(false)}
        onRun={options => {
          onApprove(true, options);
          setAdvancedOpen(false);
        }}
        loading={busy}
      />
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
