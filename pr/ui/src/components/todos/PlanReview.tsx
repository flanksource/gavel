import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { UiChevronDown, UiChevronUp, UiClose, UiEye } from '@flanksource/clicky-ui/icons';
import type { TodoItem, TodoRunOptions } from '../../types';
import type { WorkspaceTodos } from './useWorkspaceTodos';
import type { TodoEntry } from './todoGroup';
import { flattenTodos } from './todoGroup';
import { buildReviewQueue, nextIndexAfterRemoval, reviewableCount } from './reviewQueue';
import { AnswerBox, PlanApproveButtons, QuestionsPanel, usePlanActions } from './planActions';
import { loadLastTodoRunOptions } from './run';
import { statusClass, statusLabel } from './format';

// Plan Review mode: a queue over the todos waiting for a human — approve their
// plans (`review`) or answer their blocking questions (`ask`) — layered on the
// existing detail pane (no route; navigation reuses todos.select so URLs stay
// /todos/{ref}). useReviewMode is the client-only state; the two components are
// the navbar entry pill and the sticky action bar.

function selectedIndex(queue: TodoEntry[], selected: WorkspaceTodos['selected']): number {
  if (!selected) return -1;
  return queue.findIndex(e => e.todo.ref === selected.ref && e.workspace.dir === selected.dir);
}

export function useReviewMode(todos: WorkspaceTodos) {
  const [active, setActive] = useState(false);
  const queue = useMemo(
    () => buildReviewQueue(flattenTodos(todos.workspaces, todos.byDir)),
    [todos.workspaces, todos.byDir],
  );
  const count = reviewableCount(todos.aggregate);
  const index = selectedIndex(queue, todos.selected);

  const goToEntry = useCallback(
    (entry: TodoEntry | undefined) => {
      if (!entry) return;
      todos.select({ dir: entry.workspace.dir, ref: entry.todo.ref });
    },
    [todos],
  );

  const enter = useCallback(() => {
    setActive(true);
    const inQueue = index >= 0;
    if (!inQueue) goToEntry(queue[0]);
  }, [index, queue, goToEntry]);

  const exit = useCallback(() => setActive(false), []);
  const goNext = useCallback(() => goToEntry(queue[Math.min((index < 0 ? -1 : index) + 1, queue.length - 1)]), [queue, index, goToEntry]);
  const goPrev = useCallback(() => goToEntry(queue[Math.max((index < 0 ? 1 : index) - 1, 0)]), [queue, index, goToEntry]);

  // After acting on the current todo (it will leave the queue on the next
  // refresh), jump to the successor in the post-removal queue so the reviewer
  // keeps moving without waiting for the round-trip.
  const advanceAfterAction = useCallback(() => {
    const remaining = queue.filter((_, i) => i !== index);
    const nextIdx = nextIndexAfterRemoval(index < 0 ? 0 : index, remaining.length);
    if (nextIdx < 0) {
      setActive(false);
      return;
    }
    goToEntry(remaining[nextIdx]);
  }, [queue, index, goToEntry]);

  // Nothing left to review → drop out of the mode.
  useEffect(() => {
    if (active && queue.length === 0) setActive(false);
  }, [active, queue.length]);

  return { active, count, queue, index, enter, exit, goNext, goPrev, advanceAfterAction };
}

export type ReviewMode = ReturnType<typeof useReviewMode>;

// TodoReviewButton is the navbar entry point: an amber "Review · N" pill that
// only appears when something is waiting. Clicking it enters review mode.
export function TodoReviewButton({ review }: { review: ReviewMode }) {
  if (review.count === 0) return null;
  return (
    <Button
      type="button"
      variant="ghost"
      onClick={review.enter}
      title="Review plans and answer questions"
      className="inline-flex h-8 items-center gap-1 rounded-md border border-amber-500/40 bg-amber-500/10 px-2 text-xs font-medium text-amber-700 hover:bg-amber-500/20"
    >
      <UiEye className="text-xs" />
      Review · {review.count}
    </Button>
  );
}

// PlanReviewBar is the sticky strip above the detail pane while review mode is
// active: position, status chip, prev/next, exit, and the inline approve/answer
// action for the current todo. Renders nothing when inactive.
export function PlanReviewBar({ review, todos }: { review: ReviewMode; todos: WorkspaceTodos }) {
  const sel = todos.selected;
  const detail = todos.detail && sel && todos.detail.ref === sel.ref ? todos.detail : null;
  const queueTodo = review.queue[review.index]?.todo;
  const current: TodoItem | null = detail ?? queueTodo ?? null;
  const { busy, error, reset, approve, answer, reject, revise } = usePlanActions(sel?.dir ?? '');
  const [answerText, setAnswerText] = useState('');
  const [changesText, setChangesText] = useState('');
  const [showChanges, setShowChanges] = useState(false);
  const answerRef = useRef<HTMLTextAreaElement | null>(null);
  const changesRef = useRef<HTMLTextAreaElement | null>(null);

  // Reset the drafts and any stale error whenever the focused todo changes.
  useEffect(() => {
    setAnswerText('');
    setChangesText('');
    setShowChanges(false);
    reset();
  }, [sel?.ref, sel?.dir, reset]);

  const ref = current?.ref ?? '';
  const status = current?.status;

  const onApprove = useCallback(
    async (run: boolean, options?: TodoRunOptions) => {
      if (!ref) return;
      const result = await approve(ref, { run, options: run ? options ?? loadLastTodoRunOptions('run') : undefined });
      if (!result) return;
      todos.updateItem(result.todo);
      review.advanceAfterAction();
      todos.refresh();
    },
    [ref, approve, todos, review],
  );

  const onAnswer = useCallback(async () => {
    if (!ref || !answerText.trim()) return;
    const result = await answer(ref, answerText);
    if (!result) return;
    todos.updateItem(result.todo);
    setAnswerText('');
    review.advanceAfterAction();
    todos.refresh();
  }, [ref, answerText, answer, todos, review]);

  const onReject = useCallback(async () => {
    if (!ref) return;
    const result = await reject(ref);
    if (!result) return;
    todos.updateItem(result.todo);
    review.advanceAfterAction();
    todos.refresh();
  }, [ref, reject, todos, review]);

  const onRevise = useCallback(async () => {
    if (!ref || !changesText.trim()) return;
    const result = await revise(ref, changesText);
    if (!result) return;
    todos.updateItem(result.todo);
    setChangesText('');
    setShowChanges(false);
    review.advanceAfterAction();
    todos.refresh();
  }, [ref, changesText, revise, todos, review]);

  // Keyboard shortcuts while review mode is active. Editable targets keep their
  // own keys (the answer textarea handles Cmd/Ctrl+Enter itself); only Escape is
  // honoured from within a field, to blur it.
  useEffect(() => {
    if (!review.active) return;
    const onKey = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null;
      const editable = !!target && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable);
      if (e.key === 'Escape') {
        if (editable && target) {
          target.blur();
        } else {
          review.exit();
        }
        return;
      }
      if (editable || e.metaKey || e.ctrlKey || e.altKey) return;
      switch (e.key) {
        case 'j':
        case 'ArrowDown':
        case 'n':
          e.preventDefault();
          review.goNext();
          break;
        case 'k':
        case 'ArrowUp':
        case 'p':
          e.preventDefault();
          review.goPrev();
          break;
        case 's':
          e.preventDefault();
          review.goNext();
          break;
        case 'a':
          if (status === 'review') {
            e.preventDefault();
            void onApprove(e.shiftKey);
          }
          break;
        case 'A':
          if (status === 'review') {
            e.preventDefault();
            void onApprove(true);
          }
          break;
        case 'r':
          if (status === 'ask') {
            e.preventDefault();
            answerRef.current?.focus();
          }
          break;
        case 'x':
          if (status === 'review') {
            e.preventDefault();
            void onReject();
          }
          break;
        case 'c':
          if (status === 'review') {
            e.preventDefault();
            setShowChanges(true);
            // Focus after the box mounts.
            setTimeout(() => changesRef.current?.focus(), 0);
          }
          break;
      }
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [review, status, onApprove, onReject]);

  if (!review.active) return null;

  const total = review.queue.length;
  const position = review.index >= 0 ? review.index + 1 : Math.min(1, total);

  return (
    <div className="flex shrink-0 flex-col gap-2 border-b border-amber-500/30 bg-amber-500/5 px-3 py-2">
      <div className="flex flex-wrap items-center gap-2">
        <span className="inline-flex items-center gap-1 text-xs font-semibold uppercase tracking-wide text-amber-700">
          <UiEye className="text-xs" />
          Plan Review
        </span>
        <div className="inline-flex items-center gap-0.5">
          <Button type="button" variant="ghost" size="icon" onClick={review.goPrev} title="Previous (k / ↑)" className="h-7 w-7 text-muted-foreground hover:bg-muted">
            <UiChevronUp className="text-xs" />
          </Button>
          <span className="min-w-[3rem] text-center text-xs tabular-nums text-muted-foreground">{total === 0 ? '—' : `${position} / ${total}`}</span>
          <Button type="button" variant="ghost" size="icon" onClick={review.goNext} title="Next (j / ↓)" className="h-7 w-7 text-muted-foreground hover:bg-muted">
            <UiChevronDown className="text-xs" />
          </Button>
        </div>
        {status && (
          <span className={`inline-flex items-center rounded px-1.5 py-0.5 text-[11px] font-medium ${statusClass(status)}`}>
            {statusLabel(status)}
          </span>
        )}
        <span className="min-w-0 flex-1 truncate text-xs text-muted-foreground" title={current?.title}>
          {current?.title}
        </span>
        {status === 'review' && (
          <PlanApproveButtons
            busy={busy}
            onApprove={(run, options) => void onApprove(run, options)}
            onReject={() => void onReject()}
            onRequestChanges={() => setShowChanges(v => !v)}
          />
        )}
        <span className="hidden text-[11px] text-muted-foreground lg:inline">a approve · ⇧a run · c changes · x reject · r reply · s skip · esc exit</span>
        <Button type="button" variant="ghost" size="icon" onClick={review.exit} title="Exit review (Esc)" aria-label="Exit review" className="h-7 w-7 text-muted-foreground hover:bg-muted">
          <UiClose className="text-xs" />
        </Button>
      </div>
      {total === 0 && <div className="text-xs text-muted-foreground">All caught up — no plans or questions waiting.</div>}
      {status === 'review' && showChanges && (
        <AnswerBox
          value={changesText}
          onChange={setChangesText}
          onSubmit={() => void onRevise()}
          busy={busy}
          label="Send changes & re-plan"
          placeholder="Describe what to change in the plan… (Cmd/Ctrl+Enter to send)"
          inputRef={el => (changesRef.current = el)}
        />
      )}
      {status === 'ask' && current && (
        <div className="flex flex-col gap-2">
          {current.questions && current.questions.length > 0 && <QuestionsPanel questions={current.questions} />}
          <AnswerBox
            value={answerText}
            onChange={setAnswerText}
            onSubmit={() => void onAnswer()}
            busy={busy}
            inputRef={el => (answerRef.current = el)}
          />
        </div>
      )}
      {error && <div className="text-xs text-red-600">{error}</div>}
    </div>
  );
}
