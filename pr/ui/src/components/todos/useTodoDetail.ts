import { useEffect, useState } from 'react';
import type { TodoItem, TodoPriority, TodoRunOptions, TodoStatus } from '../../types';
import type { TodoDetailProps } from './TodoDetail';
import { useSessionStats } from './TodoSessionTimer';
import { loadLastTodoRunOptions, rememberTodoRunOptions, requestStepFor, runSpec, useTodoRun, useTodoRunContext } from './run';
import { loadPromptRunOptions, rememberPromptRunOptions, verificationSpec } from './PromptRunButton';
import type { PhaseRunOptions } from './TodoPhaseButton';
import type { TodoDetailTabKey } from './TodoDetailTabs';
import { useTodoSessionDetail } from './TodoSessionDetail';
import { verificationAttempts, verificationBadge } from './verificationReport';
import { TodoMutationError, useDeleteTodoMutation, useGithubPushTodoMutation, useTodoSessionStop, useTodoVerificationRun, useTransferTodoMutation, useUpdateTodoMutation } from './todoMutations';
import { useTodoTagCounts, useTodoTagIndex } from './tagQueries';
import { todoVisibleLabels } from './tagResolve';

export function useTodoDetail({ todo, dir, onChanged, onDeleted, workspaces = [], onTransferred }: TodoDetailProps) {
  const [advancedMode, setAdvancedMode] = useState<string | null>(null);
  const [runSelections, setRunSelections] = useState<PhaseRunOptions>({});
  const [verifySelection, setVerifySelection] = useState<TodoRunOptions | null>(null);
  const [error, setError] = useState('');
  const [tab, setTab] = useState<TodoDetailTabKey>('overview');
  const [editingTitle, setEditingTitle] = useState(false);
  const [editingBody, setEditingBody] = useState(false);
  const [draftTitle, setDraftTitle] = useState('');
  const [draftBody, setDraftBody] = useState('');
  const [copyState, setCopyState] = useState<'idle' | 'copied' | 'error'>('idle');
  const { runBusy, runMessage, runError, reset: resetRun, run } = useTodoRun(dir);
  const { context: runContext } = useTodoRunContext({ dir });
  const updateTodo = useUpdateTodoMutation(dir, `Failed to update todo ${todo?.ref || ''}`.trim());
  const deleteTodo = useDeleteTodoMutation(dir);
  const transferTodo = useTransferTodoMutation();
  const githubPushTodo = useGithubPushTodoMutation(dir);
  const busy = updateTodo.isPending || deleteTodo.isPending || transferTodo.isPending || githubPushTodo.isPending;
  // Projects this todo can move to: every configured workspace except its own.
  const transferTargets = workspaces.filter(ws => !!ws.dir && ws.dir !== dir);
  const closed = todo?.status === 'completed';
  const body = todo?.body?.trim() ?? '';
  const events = todo?.events ?? [];
  // One attempts-only poll feeds both the tab badge and the Verification tab, so
  // a failed check is visible before the tab is ever opened. It keeps polling
  // while the tab is closed, just more slowly.
  const { detail: verificationDetail, error: verificationError } = useTodoSessionDetail(
    dir,
    todo?.ref ?? '',
    undefined,
    !!todo?.ref,
    { attemptsOnly: true, intervalMs: tab === 'verification' ? 1500 : 15000 }
  );
  const verification = verificationBadge(verificationAttempts(verificationDetail));
  const verificationRun = useTodoVerificationRun(dir, todo?.ref ?? '');
  const sessionStop = useTodoSessionStop(dir, todo?.ref ?? '', todo?.sessionId);
  // The attempt the Stop control acts on. Only a live attempt the server says it
  // can interrupt qualifies; without one the header offers a disabled Stop that
  // says why, rather than one that looks live and does nothing.
  const stoppableAttempt = (verificationDetail?.attempts ?? []).find(attempt => attempt.processActive && attempt.canStop);
  const fullTodoId = todo ? todoFullId(todo) : '';
  const visibleLabels = todo ? todoVisibleLabels(todo) : [];
  const tagIndex = useTodoTagIndex(dir);
  // Same cached query as the index, so the picker can lead with the tags this
  // project actually uses without a second request.
  const tagCounts = useTodoTagCounts(dir);
  const viewSessionId = todo?.lookupSessionId || todo?.sessionId;
  const viewingHistoricalSession = !!todo?.lookupSessionId && todo.lookupSessionId !== todo.sessionId;
  const { stats: headerSessionStats } = useSessionStats({ dir, sessionId: todo?.sessionId, active: !!todo?.sessionId });
  const sessionInProgress = !!todo && !!todo.sessionId && (headerSessionStats?.inProgress || (!headerSessionStats?.found && todo.status === 'in_progress'));
  // A todo awaiting a human decision (plan review or a blocking question) must
  // route through TodoReviewBanner's approve/reject/answer flow, not have its
  // review/ask state silently bypassed by re-triggering a run from here.
  const awaitingHumanAction = todo?.status === 'review' || todo?.status === 'ask';
  const phaseOptions: PhaseRunOptions = {
    ...runSelections,
    verify: verifySelection ?? undefined,
  };
  // The running attempt names its own step, so the strip can say "Verify" rather
  // than a generic "Running" when a check is what is in flight.
  const runningPhaseLabel = stoppableAttempt?.step
    ? todo?.lifecycle?.steps.find(entry => entry.name === stoppableAttempt.step)?.label ?? stoppableAttempt.step
    : undefined;

  function changePhaseOptions(name: string, options: TodoRunOptions) {
    if (!runContext) return;
    if (name === 'verify') {
      setVerifySelection(rememberPromptRunOptions('verification', options, runContext));
      return;
    }
    const remembered = rememberTodoRunOptions(name, options);
    setRunSelections(previous => ({ ...previous, [name]: remembered }));
  }

  useEffect(() => {
    setError('');
    resetRun();
    setAdvancedMode(null);
    setTab(todo?.lookupSessionId ? 'session' : 'overview');
    setEditingTitle(false);
    setEditingBody(false);
    setCopyState('idle');
  }, [todo?.ref, todo?.lookupSessionId, resetRun]);

  useEffect(() => {
    if (!todo || !runContext) {
      setRunSelections({});
      setVerifySelection(null);
      return;
    }
    setRunSelections(Object.fromEntries(runContext.lifecycle.steps
      .filter(({ name }) => name !== 'verify')
      .map(({ name }) => [name, loadLastTodoRunOptions(name, runContext)])));
    setVerifySelection(loadPromptRunOptions('verification', runContext));
  }, [todo?.ref, runContext]);

  useEffect(() => {
    if (copyState === 'idle') return;
    const timeout = window.setTimeout(() => setCopyState('idle'), 1600);
    return () => window.clearTimeout(timeout);
  }, [copyState]);

  // patch sends a partial update (status, priority, title, body, and/or comment)
  // and adopts the server's returned todo so the view reflects server-side effects
  // (labels, state, rewritten body, new comment). Resolves true on success.
  async function patch(payload: {
    status?: TodoStatus;
    priority?: TodoPriority;
    title?: string;
    body?: string;
    comment?: string;
    // The complete replacement label set. TodoTagField builds it, reserved
    // lifecycle labels included, because the API replaces the whole set.
    labels?: string[];
  }): Promise<boolean> {
    if (!todo || busy) return false;
    setError('');
    try {
      const data = await updateTodo.mutateAsync({
        ref: todo.ref,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref, ...payload }),
      });
      onChanged(data);
      return true;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update todo');
      return false;
    }
  }

  function startEditTitle() {
    if (!todo) return;
    setDraftTitle(todo.title);
    setEditingTitle(true);
  }

  function startEditBody() {
    if (!todo) return;
    setDraftBody(todo.body ?? '');
    setTab('overview');
    setEditingBody(true);
  }

  async function saveTitle() {
    const title = draftTitle.trim();
    if (!title) return;
    if (await patch({ title })) setEditingTitle(false);
  }

  async function saveBody() {
    if (await patch({ body: draftBody })) setEditingBody(false);
  }

  async function transferTo(toDir: string) {
    if (!todo || busy || !toDir || !onTransferred) return;
    setError('');
    try {
      const { todo: moved } = await transferTodo.mutateAsync({ ref: todo.ref, fromDir: dir, toDir });
      onTransferred(toDir, moved);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to move todo');
    }
  }

  // A todo that was already pushed comes back as a 409 rather than silently
  // opening a duplicate; the retry rewrites that issue from the todo as it
  // stands now, overwriting whatever the issue body says.
  async function pushToGithub(update = false) {
    if (!todo || (busy && !update)) return;
    setError('');
    try {
      const { url } = await githubPushTodo.mutateAsync({ ref: todo.ref, update });
      window.open(url, '_blank', 'noopener');
    } catch (err) {
      if (!update && err instanceof TodoMutationError && err.status === 409) {
        if (window.confirm('This todo is already linked to a GitHub issue.\n\n'
          + "Rewrite that issue with the todo's current title, body, plan and verification?")) {
          await pushToGithub(true);
        }
        return;
      }
      setError(err instanceof Error ? err.message : 'Failed to push todo to GitHub');
    }
  }

  async function archiveTodo() {
    if (!todo || busy) return;
    if (!window.confirm('Archive this todo?')) return;
    setError('');
    try {
      await deleteTodo.mutateAsync(todo.ref);
      onDeleted();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to archive todo');
    }
  }

  async function copyFullId() {
    if (!fullTodoId) return;
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(fullTodoId);
      } else {
        const textarea = document.createElement('textarea');
        textarea.value = fullTodoId;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.focus();
        textarea.select();
        document.execCommand('copy');
        document.body.removeChild(textarea);
      }
      setCopyState('copied');
    } catch {
      setCopyState('error');
    }
  }

  // Enter a lifecycle step. All post to the same /api/todos/run endpoint,
  // naming the step; verify's mutation invalidates the attempt list too, so
  // the Verification tab it switches to shows the fresh evidence immediately.
  async function runPhase(name: string, options = phaseOptions[name] ?? loadLastTodoRunOptions(name, runContext)) {
    if (!todo) return;
    if (name === 'verify') {
      setError('');
      try {
        await verificationRun.mutateAsync({
          ref: todo.ref,
          runtimeProfile: options.runtimeProfile,
          spec: verificationSpec(runSpec(options)),
        });
        setTab('verification');
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Verification run failed');
      }
      return;
    }
    await runTodo({ ...options, step: name });
  }

  function submitAdvanced(options: TodoRunOptions) {
    if (!runContext) return;
    const step = requestStepFor(options);
    const remembered = step === 'verify'
      ? rememberPromptRunOptions('verification', options, runContext)
      : rememberTodoRunOptions(step, options, true);
    if (step === 'verify') setVerifySelection(remembered);
    else setRunSelections(previous => ({ ...previous, [step]: remembered }));
    setAdvancedMode(null);
    void runPhase(step, remembered);
  }

  async function stopRun() {
    if (!stoppableAttempt) return;
    setError('');
    try {
      await sessionStop.mutateAsync(stoppableAttempt.promptRunId);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not stop the attempt');
    }
  }

  async function runTodo(options?: TodoRunOptions) {
    if (!todo) return;
    const result = await run(todo.ref, options);
    if (result?.status === 'started') {
      onChanged({
        ...todo,
        status: 'in_progress',
        lastRun: new Date().toISOString(),
        // Adopt the run's session id so the Session tab follows the new run.
        sessionId: result.sessionId || todo.sessionId,
      });
      // Surface the live session as soon as a run starts.
      setTab('session');
    }
  }

  return {
    advancedMode, setAdvancedMode, runSelections, setRunSelections, error, tab, setTab,
    editingTitle, setEditingTitle, editingBody, setEditingBody, draftTitle, setDraftTitle,
    draftBody, setDraftBody, copyState, runBusy, runMessage, runError, runContext,
    busy, transferTargets, closed, body, events, verificationDetail, verificationError,
    verification, verificationRun, sessionStop, stoppableAttempt, fullTodoId, visibleLabels,
    tagIndex, tagCounts, viewSessionId, viewingHistoricalSession, sessionInProgress,
    awaitingHumanAction, phaseOptions, runningPhaseLabel, changePhaseOptions, patch,
    startEditTitle, startEditBody, saveTitle, saveBody, transferTo, pushToGithub,
    archiveTodo, copyFullId, runPhase, stopRun, runTodo, submitAdvanced,
  };
}

function todoFullId(todo: TodoItem): string {
  return todo.id ? `todo:${todo.id}` : todo.ref;
}
