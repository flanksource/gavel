import { useEffect, useState, type ComponentType, type ReactNode } from 'react';
import { Button, DropdownMenu } from '@flanksource/clicky-ui/components';
import { Markdown } from '@flanksource/clicky-ui/data';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import { UiArrowLeft, UiCheck, UiCheckFilled, UiChevronDown, UiChevronRight, UiCog, UiCopy, UiDebugStepOver, UiDotsVertical, UiEdit, UiError, UiFolder, UiLinkExternal, UiListDashes, UiMarkdown, UiPass, UiPlay, UiRestart, UiStop, UiTrash } from '@flanksource/clicky-ui/icons';
import type { Project, TodoExternalIssue, TodoItem, TodoPriority, TodoRunOptions, TodoStatus } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { RepoIcon } from '../RepoIcon';
import { TodoTimeline } from './TodoTimeline';
import { TodoCommits } from './TodoCommits';
import { TodoSession } from './TodoSession';
import { TodoPlan } from './TodoPlan';
import { useSessionStats } from './TodoSessionTimer';
import { priorities, priorityBadgeClass, priorityIcon, statusClass, statusIcon, statuses, statusLabel } from './format';
import { TodoRunAdvancedDialog } from './TodoRunAdvancedDialog';
import { defaultRunOptions, loadLastTodoRunOptions, rememberTodoRunOptions, rememberTodoRunOptionsForMode, runActionConfig, runSpec, type TodoRunAction, useTodoRun, useTodoRunContext } from './run';
import { loadPromptRunOptions, rememberPromptRunOptions, verificationSpec } from './PromptRunButton';
import { ALL_PHASES, phaseSignals, phaseState, type PhaseId, type PhaseState } from './phaseMachine';
import { TodoPhaseButton, TodoPhaseTicks, type PhaseRunOptions } from './TodoPhaseButton';
import { TodoRunStatusStrip } from './TodoRunStatusStrip';
import { TodoBodyEditor, TodoCommentBox, TodoTitleEditor } from './TodoCompose';
import { TodoVerification } from './TodoVerification';
import { TodoDetailTabs, type TodoDetailTabKey } from './TodoDetailTabs';
import { useTodoSessionDetail } from './TodoSessionDetail';
import { verificationAttempts, verificationBadge } from './verificationAttempts';
import { TodoReviewBanner } from './planActions';
import { TodoMutationError, useDeleteTodoMutation, useGithubPushTodoMutation, useTodoSessionStop, useTodoVerificationRun, useTransferTodoMutation, useUpdateTodoMutation } from './todoMutations';
import { TodoTagField, TodoTagRow } from './TodoTag';
import { useTodoTagCounts, useTodoTagIndex } from './tagQueries';
import { todoVisibleLabels, type TagIndex } from './tagResolve';

export function TodoDetail({
  todo,
  loading,
  loadError,
  dir,
  onChanged,
  onDeleted,
  onBack,
  workspaces = [],
  onTransferred,
}: {
  todo: TodoItem | null;
  loading: boolean;
  loadError?: string;
  dir: string;
  onChanged: (todo: TodoItem) => void;
  onDeleted: () => void;
  // onBack renders a back arrow in the header; supplied only by the single-column
  // menubar (the dashboard is master-detail and needs no back navigation).
  onBack?: () => void;
  // workspaces/onTransferred are optional: the "Move to project" control only
  // renders where a caller wires them (the dashboard), not the compact menubar.
  workspaces?: Project[];
  onTransferred?: (toDir: string, todo: TodoItem) => void;
}) {
  const [advancedMode, setAdvancedMode] = useState<TodoRunAction | null>(null);
  const [runSelections, setRunSelections] = useState<Partial<Record<TodoRunAction, TodoRunOptions>>>({});
  // Verify's runtime is stored apart from plan/run because it is a different
  // kind of run: a bare spec posted to the verification endpoint, with its own
  // persisted history (PromptRunScope 'verification'), not a driver + runMode.
  const [verifySelection, setVerifySelection] = useState<TodoRunOptions | null>(null);
  const [error, setError] = useState('');
  const [tab, setTab] = useState<TodoDetailTabKey>('overview');
  const [editingTitle, setEditingTitle] = useState(false);
  const [editingBody, setEditingBody] = useState(false);
  const [draftTitle, setDraftTitle] = useState('');
  const [draftBody, setDraftBody] = useState('');
  const [copyState, setCopyState] = useState<'idle' | 'copied' | 'error'>('idle');
  const { runBusy, runMessage, runError, reset: resetRun, run } = useTodoRun(dir);
  const { context: runContext } = useTodoRunContext();
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
  // Where the todo stands in plan -> run -> verify. This is what decides the
  // header's one primary action, replacing the old pair of peer buttons that
  // offered Plan and Run identically in every state.
  const machineState: PhaseState = todo
    ? phaseState(phaseSignals(todo, {
      sessionInProgress,
      verificationFailing: verification.failing,
      verificationAttempted: verification.count > 0,
    }))
    : 'draft';
  const phaseOptions: PhaseRunOptions = {
    plan: runSelections.plan,
    run: runSelections.run,
    verify: verifySelection ?? undefined,
  };
  // The running attempt names its own step, so the strip can say "Verify" rather
  // than a generic "Running" when a check is what is in flight.
  const runningPhaseLabel = stoppableAttempt?.step
    ? ALL_PHASES.find(entry => entry.id === stoppableAttempt.step)?.label
    : undefined;

  function changePhaseOptions(id: PhaseId, options: TodoRunOptions) {
    if (!runContext) return;
    if (id === 'verify') {
      setVerifySelection(rememberPromptRunOptions('verification', options, runContext));
      return;
    }
    const remembered = rememberTodoRunOptions(id, options);
    setRunSelections(previous => ({ ...previous, [id]: remembered }));
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
    setRunSelections({
      run: loadLastTodoRunOptions('run', runContext),
      plan: loadLastTodoRunOptions('plan', runContext),
    });
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

  // Enter a phase. plan/run/triage dispatch an agent run; verify posts the
  // definition-of-done fixture, which is a different endpoint entirely — the
  // machine hides that split from the header, it does not erase it.
  async function runPhase(id: PhaseId) {
    if (!todo) return;
    if (id === 'verify') {
      setError('');
      try {
        const options = phaseOptions.verify;
        const data = await verificationRun.mutateAsync({
          ref: todo.ref,
          spec: verificationSpec(options ? runSpec(options) : {}),
        });
        if (data.todo) onChanged(data.todo);
        setTab('verification');
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Verification run failed');
      }
      return;
    }
    await runTodo(phaseOptions[id] ?? loadLastTodoRunOptions(id, runContext));
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

  if (!todo) {
    return (
      <div className="flex h-full min-h-0 flex-col text-sm text-muted-foreground">
        {onBack && (
          <div className="shrink-0 border-b border-border px-3 py-2">
            <Button
              variant="ghost"
              size="icon"
              type="button"
              onClick={onBack}
              title="Back to todos"
              aria-label="Back to todos"
              className="inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-muted hover:text-foreground"
            >
              <UiArrowLeft className="text-base" />
            </Button>
          </div>
        )}
        <div className="flex min-h-0 flex-1 items-center justify-center px-4">
          {loading ? (
            <div className="flex items-center gap-2"><Spinner /> Loading todo</div>
          ) : loadError ? (
            <div role="alert" className="max-w-lg text-center text-destructive">
              <UiError className="mb-2 text-4xl" />
              <p>{loadError}</p>
            </div>
          ) : (
            <div className="text-center">
              <UiCheck className="mb-2 text-4xl" />
              <p>Select a todo</p>
            </div>
          )}
        </div>
      </div>
    );
  }

  const HeaderStatusIcon = statusIcon(todo.status);

  return (
    <div className="flex h-full flex-col">
      <div className="shrink-0 border-b border-border bg-background px-3 py-2 md:px-4 md:py-4">
        <div className="flex min-w-0 flex-col gap-2 md:gap-3">
          <div className="grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-center gap-2">
            <div className="flex min-w-0 items-center gap-2">
              {onBack && (
                <Button
                  variant="ghost"
                  size="icon"
                  type="button"
                  onClick={onBack}
                  title="Back to todos"
                  aria-label="Back to todos"
                  className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground"
                >
                  <UiArrowLeft className="text-base" />
                </Button>
              )}
              <div className="min-w-0 flex-1">
                {editingTitle ? (
                  <TodoTitleEditor
                    value={draftTitle}
                    busy={busy}
                    onChange={setDraftTitle}
                    onSave={saveTitle}
                    onCancel={() => setEditingTitle(false)}
                  />
                ) : (
                  <div className="group flex min-w-0 items-center gap-2">
                    <span
                      className={`inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border md:hidden ${statusClass(todo.status)}`}
                      title={statusLabel(todo.status)}
                      aria-label={statusLabel(todo.status)}
                      role="img"
                    >
                      <HeaderStatusIcon className="text-xs" />
                    </span>
                    <h1 className="min-w-0 flex-1 truncate text-xl font-semibold leading-8 text-foreground">
                      {todo.title}
                    </h1>
                    <span className="hidden md:flex">
                      <EditPencil label="Edit title" onClick={startEditTitle} disabled={busy} />
                    </span>
                  </div>
                )}
              </div>
            </div>

            <HeaderActionsMenu
              todo={todo}
              busy={busy}
              runBusy={runBusy}
              closed={closed}
              sessionInProgress={sessionInProgress}
              awaitingHumanAction={awaitingHumanAction}
              fullTodoId={fullTodoId}
              labels={visibleLabels}
              tags={tagIndex}
              transferTargets={transferTargets}
              canTransfer={!!onTransferred}
              className="md:hidden"
              onCopy={copyFullId}
              onEditTitle={startEditTitle}
              onStatus={status => patch({ status })}
              onPriority={priority => patch({ priority })}
              onTransfer={transferTo}
              onResume={() => runTodo({ ...defaultRunOptions, resume: true })}
              onRun={() => runTodo(runSelections.run ?? loadLastTodoRunOptions('run', runContext))}
              onPlan={() => runTodo(runSelections.plan ?? loadLastTodoRunOptions('plan', runContext))}

              onStop={stoppableAttempt ? stopRun : undefined}
              onAdvanced={setAdvancedMode}
              onVerify={() => patch({ status: 'verified' })}
              onToggleClosed={() => patch({ status: closed ? 'pending' : 'completed' })}
              onArchive={archiveTodo}
              onPushToGithub={pushToGithub}
            />

            <div className="hidden min-w-0 flex-wrap items-center justify-end gap-1.5 md:flex">
              {/* While a run is live the cluster becomes what is true right now —
                  which phase, which model, how long, how much context — plus the
                  one action that applies. Otherwise: where the todo stands, and
                  the single control for the step that comes next. */}
              {sessionInProgress ? (
                <TodoRunStatusStrip
                  dir={dir}
                  sessionId={todo.sessionId}
                  phaseLabel={runningPhaseLabel}
                  stopping={sessionStop.isPending || stoppableAttempt?.stopping}
                  onStop={stoppableAttempt ? stopRun : undefined}
                />
              ) : (
                <>
                  <TodoPhaseTicks state={machineState} />
                  <TodoPhaseButton
                    state={machineState}
                    context={runContext}
                    options={phaseOptions}
                    disabled={busy || !runContext}
                    busy={runBusy || verificationRun.isPending}
                    onRunPhase={runPhase}
                    onReview={() => setTab('overview')}
                    onOptionsChange={changePhaseOptions}
                    onAdvanced={id => setAdvancedMode(id === 'verify' ? 'run' : id)}
                  />
                </>
              )}
              <HeaderActionsMenu
                todo={todo}
                busy={busy}
                runBusy={runBusy}
                closed={closed}
                sessionInProgress={sessionInProgress}
                awaitingHumanAction={awaitingHumanAction}
                fullTodoId={fullTodoId}
                labels={visibleLabels}
                tags={tagIndex}
                transferTargets={transferTargets}
                canTransfer={!!onTransferred}
                showRunActions={false}
                showStatusPriority={false}
                showLabels={false}
                onCopy={copyFullId}
                onEditTitle={startEditTitle}
                onStatus={status => patch({ status })}
                onPriority={priority => patch({ priority })}
                onTransfer={transferTo}
                onResume={() => runTodo({ ...defaultRunOptions, resume: true })}
                onRun={() => runTodo(runSelections.run ?? loadLastTodoRunOptions('run', runContext))}
                onPlan={() => runTodo(runSelections.plan ?? loadLastTodoRunOptions('plan', runContext))}

                onStop={stoppableAttempt ? stopRun : undefined}
                onAdvanced={setAdvancedMode}
                onVerify={() => patch({ status: 'verified' })}
                onToggleClosed={() => patch({ status: closed ? 'pending' : 'completed' })}
                onArchive={archiveTodo}
                onPushToGithub={pushToGithub}
              />
            </div>

            <div className="hidden min-w-0 flex-wrap items-center gap-2 text-xs text-muted-foreground md:col-span-2 md:flex">
              <StatusMenu
                value={todo.status}
                disabled={busy}
                compact
                onSelect={status => patch({ status })}
              />
              <PriorityMenu
                value={todo.priority}
                disabled={busy}
                compact
                onSelect={priority => patch({ priority })}
              />
              <Button
                variant="ghost"
                type="button"
                onClick={copyFullId}
                title={copyState === 'copied' ? 'Copied' : 'Copy full issue ID'}
                className="hidden h-auto min-w-0 max-w-full items-center gap-1.5 rounded border border-border bg-muted/20 px-2 py-1 text-left font-mono text-[11px] hover:bg-muted md:flex"
              >
                {copyState === 'copied' ? (
                  <UiCheck className="shrink-0 text-muted-foreground" />
                ) : copyState === 'error' ? (
                  <UiError className="shrink-0 text-red-600" />
                ) : (
                  <UiCopy className="shrink-0 text-muted-foreground" />
                )}
                <span className="min-w-0 truncate">{fullTodoId}</span>
              </Button>
              {todo.externalIssue && <ExternalIssueLink issue={todo.externalIssue} />}
              <span className="hidden items-center gap-2 md:flex">
                {todo && (
                  <TodoTagField
                    todo={todo}
                    index={tagIndex}
                    counts={tagCounts}
                    disabled={busy}
                    onChange={labels => void patch({ labels })}
                  />
                )}
              </span>
              {copyState === 'error' && <span className="hidden text-red-600 md:block">Copy failed</span>}
            </div>
          </div>
          {(error || runError) && <div className="mt-2 text-xs text-red-600">{error || runError}</div>}
          {runMessage && !error && !runError && <div className="mt-2 text-xs text-emerald-600">{runMessage}</div>}
        </div>
      </div>
      <TodoRunAdvancedDialog
        open={advancedMode !== null}
        onClose={() => setAdvancedMode(null)}
        onRun={options => {
          const action = advancedMode
            ?? (options.prompt === 'triage'
              ? 'triage'
              : options.plan || options.runMode === 'plan' ? 'plan' : 'run');
          const remembered = rememberTodoRunOptionsForMode(options, true);
          setRunSelections(previous => ({ ...previous, [action]: remembered }));
          setAdvancedMode(null);
          runTodo(remembered);
        }}
        loading={runBusy}
        initialMode={advancedMode ?? 'run'}
        title={runActionConfig[advancedMode ?? 'run'].title}
        dir={dir}
        refID={todo.ref}
      />
      <TodoReviewBanner todo={todo} dir={dir} onChanged={onChanged} />
      <TodoDetailTabs tab={tab} onSelect={setTab} verification={verification} />
      <div className="flex min-h-0 flex-1 flex-col bg-[#f4f6f9] dark:bg-[#0a1020]">
        {tab === 'session' ? (
          <TodoSession
            dir={dir}
            sessionId={viewSessionId}
            active={tab === 'session'}
            todo={todo}
            onChanged={onChanged}
            onResume={() => runTodo({ ...defaultRunOptions, resume: true })}
            resumeDisabled={busy || runBusy || sessionInProgress || viewingHistoricalSession}
            onRun={runTodo}
            onAdvanced={setAdvancedMode}
            runOptions={runSelections.run}
            planOptions={runSelections.plan}
            onRunOptionsChange={options => setRunSelections(previous => ({ ...previous, run: options }))}
            onPlanOptionsChange={options => setRunSelections(previous => ({ ...previous, plan: options }))}
            runBusy={runBusy}
            runDisabled={busy || runBusy || awaitingHumanAction}
          />
        ) : tab === 'plan' ? (
          <TodoPlan dir={dir} todo={todo} active={tab === 'plan'} onChanged={onChanged} />
        ) : tab === 'verification' ? (
          <div className="min-h-0 flex-1 overflow-y-auto px-4 py-3">
            <TodoVerification
              dir={dir}
              todo={todo}
              onChanged={onChanged}
              attempts={verificationDetail}
              attemptsError={verificationError}
            />
          </div>
        ) : (
          <div className="min-h-0 flex-1 overflow-y-auto px-4 py-3">
            {loading ? (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Spinner />
                Loading
              </div>
            ) : (
              <div className="space-y-3">
                {editingBody ? (
                  <TodoBodyEditor
                    value={draftBody}
                    busy={busy}
                    onChange={setDraftBody}
                    onSave={saveBody}
                    onCancel={() => setEditingBody(false)}
                  />
                ) : body ? (
                  <TodoSection
                    title="Body"
                    icon={UiMarkdown}
                    defaultOpen
                    resetKey={`${todo.ref}:body`}
                    action={<EditPencil label="Edit body" onClick={startEditBody} disabled={busy} />}
                  >
                    <Markdown text={body} className="text-sm" />
                  </TodoSection>
                ) : (
                  <TodoSection
                    title="Body"
                    icon={UiMarkdown}
                    defaultOpen
                    resetKey={`${todo.ref}:body-empty`}
                    action={<EditPencil label="Add body" onClick={startEditBody} disabled={busy} />}
                  >
                    <p className="text-sm text-muted-foreground">No body yet.</p>
                  </TodoSection>
                )}
                <TodoCommentBox
                  closed={closed}
                  busy={busy}
                  onComment={(text, reopen) => patch(reopen ? { status: 'pending', comment: text } : { comment: text })}
                />
                <TodoCommits dir={dir} todoRef={todo.ref} />
                {events.length > 0 && <TodoTimeline events={events} />}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

function todoFullId(todo: TodoItem): string {
  if (todo.id) return `todo:${todo.id}`;
  return todo.ref;
}


function StatusMenu({
  value,
  disabled,
  compact = false,
  onSelect,
}: {
  value: TodoStatus;
  disabled?: boolean;
  compact?: boolean;
  onSelect: (status: TodoStatus) => void;
}) {
  const ValueIcon = statusIcon(value);
  return (
    <DropdownMenu
      align="left"
      menuLabel="Update todo status"
      menuClassName="w-56"
      trigger={
        <Button
          variant="ghost"
          type="button"
          disabled={disabled}
          className={compact
            ? 'inline-flex h-7 items-center gap-1.5 rounded-md border border-border bg-muted/20 px-2 text-[11px] font-medium text-muted-foreground hover:bg-muted disabled:opacity-50'
            : `inline-flex h-8 items-center gap-1.5 rounded-full border px-2.5 text-xs font-semibold capitalize ${statusClass(value)} disabled:opacity-50`}
          title="Update status"
          aria-label="Update todo status"
        >
          {compact && <span>Status</span>}
          <span className={compact ? `inline-flex h-4 w-4 items-center justify-center rounded-full border ${statusClass(value)}` : ''}>
            <ValueIcon className="text-xs" />
          </span>
          <span className={compact ? 'capitalize text-foreground' : ''}>{statusLabel(value)}</span>
          <UiChevronDown className="text-[11px] opacity-70" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          {statuses.map(status => {
            const StatusItemIcon = statusIcon(status);
            return (
            <Button
              key={status}
              variant="ghost"
              type="button"
              disabled={disabled}
              onClick={() => {
                close();
                if (status !== value) onSelect(status);
              }}
              className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
            >
              <span className={`inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border ${statusClass(status)}`}>
                <StatusItemIcon className="text-xs" />
              </span>
              <span className="min-w-0 flex-1 capitalize text-foreground">{statusLabel(status)}</span>
              {status === value && <UiCheck className="text-xs text-primary" />}
            </Button>
            );
          })}
        </div>
      )}
    </DropdownMenu>
  );
}

function PriorityMenu({
  value,
  disabled,
  compact = false,
  onSelect,
}: {
  value: TodoPriority;
  disabled?: boolean;
  compact?: boolean;
  onSelect: (priority: TodoPriority) => void;
}) {
  const ValueIcon = priorityIcon(value);
  return (
    <DropdownMenu
      align="left"
      menuLabel="Update todo priority"
      menuClassName="w-44"
      trigger={
        <Button
          variant="ghost"
          type="button"
          disabled={disabled}
          className={compact
            ? 'inline-flex h-7 items-center gap-1.5 rounded-md border border-border bg-muted/20 px-2 text-[11px] font-medium text-muted-foreground hover:bg-muted disabled:opacity-50'
            : `inline-flex h-8 items-center gap-1.5 rounded-full border px-2.5 text-xs font-semibold capitalize ${priorityBadgeClass(value)} disabled:opacity-50`}
          title="Update priority"
          aria-label="Update todo priority"
        >
          {compact && <span>Severity</span>}
          <span className={compact ? `inline-flex h-4 w-4 items-center justify-center rounded-full border ${priorityBadgeClass(value)}` : ''}>
            <ValueIcon className="text-xs" />
          </span>
          <span className={compact ? 'capitalize text-foreground' : ''}>{value}</span>
          <UiChevronDown className="text-[11px] opacity-70" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          {priorities.map(priority => {
            const PriorityItemIcon = priorityIcon(priority);
            return (
            <Button
              key={priority}
              variant="ghost"
              type="button"
              disabled={disabled}
              onClick={() => {
                close();
                if (priority !== value) onSelect(priority);
              }}
              className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
            >
              <span className={`inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-full border ${priorityBadgeClass(priority)}`}>
                <PriorityItemIcon className="text-xs" />
              </span>
              <span className="min-w-0 flex-1 capitalize text-foreground">{priority}</span>
              {priority === value && <UiCheck className="text-xs text-primary" />}
            </Button>
            );
          })}
        </div>
      )}
    </DropdownMenu>
  );
}

// ExternalIssueLink names the tracker issue this todo was pushed to. It is what
// makes the list's Linked/Not-linked filter legible: the filter says a link
// exists, this says which issue.
function ExternalIssueLink({ issue }: { issue: TodoExternalIssue }) {
  return (
    <a
      href={issue.url}
      target="_blank"
      rel="noreferrer"
      title={`Open ${issue.repo}#${issue.number} on GitHub`}
      className="inline-flex h-auto min-w-0 items-center gap-1.5 rounded border border-border bg-muted/20 px-2 py-1 text-[11px] text-muted-foreground hover:bg-muted hover:text-foreground"
    >
      <RepoIcon repo={issue.repo} size={12} />
      <span className="min-w-0 truncate">{issue.repo}#{issue.number}</span>
      <UiLinkExternal className="shrink-0 text-[10px]" />
    </a>
  );
}


function HeaderActionsMenu({
  todo,
  busy,
  runBusy,
  closed,
  sessionInProgress,
  awaitingHumanAction,
  fullTodoId,
  labels,
  tags,
  transferTargets,
  canTransfer,
  className,
  showRunActions = true,
  showStatusPriority = true,
  showLabels = true,
  onCopy,
  onEditTitle,
  onStatus,
  onPriority,
  onTransfer,
  onResume,
  onRun,
  onPlan,
  onStop,
  onAdvanced,
  onVerify,
  onToggleClosed,
  onArchive,
  onPushToGithub,
}: {
  todo: TodoItem;
  busy: boolean;
  runBusy: boolean;
  closed: boolean;
  sessionInProgress: boolean;
  awaitingHumanAction: boolean;
  fullTodoId: string;
  labels: string[];
  tags: TagIndex;
  transferTargets: Project[];
  canTransfer: boolean;
  className?: string;
  showRunActions?: boolean;
  showStatusPriority?: boolean;
  showLabels?: boolean;
  onCopy: () => void;
  onEditTitle: () => void;
  onStatus: (status: TodoStatus) => void;
  onPriority: (priority: TodoPriority) => void;
  onTransfer: (dir: string) => void;
  onResume: () => void;
  onRun: () => void;
  onPlan: () => void;
  /** Interrupt the run in flight. Absent when nothing is stoppable. */
  onStop?: (() => void) | undefined;
  onAdvanced: (action: TodoRunAction) => void;
  onVerify: () => void;
  onToggleClosed: () => void;
  onArchive: () => void;
  /** Open a GitHub issue for this todo and link the two. */
  onPushToGithub: () => void;
}) {
  return (
    <DropdownMenu
      align="right"
      menuLabel="Issue actions"
      menuClassName="w-72 max-h-[80vh] max-w-[calc(100vw-16px)] overflow-y-auto"
      className={className}
      trigger={
        <Button
          variant="ghost"
          size="icon"
          type="button"
          title="Issue actions"
          aria-label="Issue actions"
          className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <UiDotsVertical className="text-sm" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          <MobileMenuSection title="Issue">
            <MobileMenuItem
              icon={UiCopy}
              label="Copy issue ID"
              detail={fullTodoId}
              onClick={() => {
                close();
                onCopy();
              }}
            />
            <MobileMenuItem
              icon={UiEdit}
              label="Edit title"
              disabled={busy}
              onClick={() => {
                close();
                onEditTitle();
              }}
            />
          </MobileMenuSection>

          {showRunActions && (
            <MobileMenuSection title="Run">
              {todo.sessionId && (
                <MobileMenuItem
                  icon={UiDebugStepOver}
                  label="Resume session"
                  detail={awaitingHumanAction ? 'Resolve the pending plan review or question first' : undefined}
                  disabled={busy || runBusy || sessionInProgress || awaitingHumanAction}
                  onClick={() => {
                    close();
                    onResume();
                  }}
                />
              )}
              {/* Stop is a real action — /api/todos/session/stop — whenever the
                  running attempt reports it can be interrupted. It stays
                  disabled without one, and says why, rather than claiming the
                  feature does not exist. */}
              <MobileMenuItem
                icon={sessionInProgress ? UiStop : runBusy ? Spinner : UiPlay}
                label={sessionInProgress ? 'Stop run' : 'Run todo'}
                detail={sessionInProgress
                  ? (onStop ? undefined : 'This run cannot be interrupted')
                  : awaitingHumanAction ? 'Resolve the pending plan review or question first' : undefined}
                disabled={busy || runBusy || awaitingHumanAction || (sessionInProgress ? !onStop : false)}
                onClick={() => {
                  close();
                  if (sessionInProgress) onStop?.();
                  else onRun();
                }}
              />
              <MobileMenuItem
                icon={UiListDashes}
                label="Plan todo"
                detail={awaitingHumanAction ? 'Resolve the pending plan review or question first' : undefined}
                disabled={busy || runBusy || sessionInProgress || awaitingHumanAction}
                onClick={() => {
                  close();
                  onPlan();
                }}
              />
              <MobileMenuItem
                icon={UiCog}
                label="Advanced run"
                detail={awaitingHumanAction ? 'Resolve the pending plan review or question first' : undefined}
                disabled={busy || runBusy || awaitingHumanAction}
                onClick={() => {
                  close();
                  onAdvanced('run');
                }}
              />
            </MobileMenuSection>
          )}

          {showStatusPriority && (
            <MobileMenuSection title="Status">
              {statuses.map(status => (
                <MobileMenuItem
                  key={status}
                  icon={statusIcon(status)}
                  label={statusLabel(status)}
                  selected={status === todo.status}
                  disabled={busy || status === todo.status}
                  onClick={() => {
                    close();
                    onStatus(status);
                  }}
                />
              ))}
            </MobileMenuSection>
          )}

          {showStatusPriority && (
            <MobileMenuSection title="Severity">
              {priorities.map(priority => (
                <MobileMenuItem
                  key={priority}
                  icon={priorityIcon(priority)}
                  label={priority}
                  selected={priority === todo.priority}
                  disabled={busy || priority === todo.priority}
                  onClick={() => {
                    close();
                    onPriority(priority);
                  }}
                />
              ))}
            </MobileMenuSection>
          )}

          <MobileMenuSection title="Actions">
            {canTransfer && transferTargets.length > 0 && (
              <MoveSubmenu
                disabled={busy}
                targets={transferTargets}
                onSelect={onTransfer}
                onCloseParent={close}
              />
            )}
            <MobileMenuItem
              icon={UiCheckFilled}
              label={todo.status === 'verified' ? 'Already verified' : 'Mark verified'}
              disabled={busy || todo.status === 'verified'}
              onClick={() => {
                close();
                onVerify();
              }}
            />
            <MobileMenuItem
              icon={closed ? UiRestart : UiPass}
              label={closed ? 'Reopen todo' : 'Mark complete'}
              disabled={busy}
              onClick={() => {
                close();
                onToggleClosed();
              }}
            />
            <MobileMenuItem
              icon={UiLinkExternal}
              label="Push to GitHub"
              disabled={busy}
              onClick={() => {
                close();
                onPushToGithub();
              }}
            />
            <MobileMenuItem
              icon={UiTrash}
              label="Archive todo"
              disabled={busy}
              danger
              onClick={() => {
                close();
                onArchive();
              }}
            />
          </MobileMenuSection>

          {showLabels && labels.length > 0 && (
            <MobileMenuSection title="Tags">
              <TodoTagRow
                labels={labels}
                index={tags}
                max={12}
                size="xxs"
                className="flex-wrap px-2 py-1"
              />
            </MobileMenuSection>
          )}
        </div>
      )}
    </DropdownMenu>
  );
}

function MobileMenuSection({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div>
      <div className="px-2 pb-0.5 pt-1.5 text-[10px] font-semibold uppercase tracking-wide text-muted-foreground">
        {title}
      </div>
      {children}
    </div>
  );
}

function MobileMenuItem({
  icon: Icon,
  label,
  detail,
  selected,
  disabled,
  danger,
  onClick,
}: {
  icon: ComponentType<IconProps>;
  label: string;
  detail?: string;
  selected?: boolean;
  disabled?: boolean;
  danger?: boolean;
  onClick: () => void;
}) {
  return (
    <Button
      variant="ghost"
      type="button"
      disabled={disabled}
      onClick={onClick}
      className="flex h-auto w-full items-start justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted disabled:opacity-50"
    >
      <Icon className={`mt-0.5 shrink-0 text-sm ${danger ? 'text-red-600' : 'text-muted-foreground'}`} />
      <span className="min-w-0 flex-1">
        <span className={`block truncate font-medium ${danger ? 'text-red-600' : 'text-foreground'}`}>{label}</span>
        {detail && <span className="block truncate font-mono text-[10px] text-muted-foreground">{detail}</span>}
      </span>
      {selected && <UiCheck className="mt-0.5 text-xs text-primary" />}
    </Button>
  );
}

function MoveSubmenu({
  disabled,
  targets,
  onSelect,
  onCloseParent,
}: {
  disabled?: boolean;
  targets: Project[];
  onSelect: (dir: string) => void;
  onCloseParent: () => void;
}) {
  return (
    <DropdownMenu
      align="right"
      menuLabel="Move todo to another project"
      menuClassName="w-72 max-w-[calc(100vw-24px)]"
      trigger={
        <Button
          variant="ghost"
          type="button"
          disabled={disabled}
          className="flex h-auto w-full items-center justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted disabled:opacity-50"
          title="Move todo to another project"
          aria-label="Move todo to another project"
        >
          <UiFolder className="shrink-0 text-sm text-muted-foreground" />
          <span className="min-w-0 flex-1">
            <span className="block truncate font-medium text-foreground">Move to</span>
            <span className="block truncate text-[11px] text-muted-foreground">Choose project</span>
          </span>
          <UiChevronRight className="text-[11px] text-muted-foreground" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          {targets.map(target => (
            <Button
              key={target.dir}
              variant="ghost"
              type="button"
              disabled={disabled}
              onClick={() => {
                close();
                onCloseParent();
                onSelect(target.dir);
              }}
              className="flex h-auto w-full items-start justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
            >
              <UiFolder className="mt-0.5 shrink-0 text-sm text-muted-foreground" />
              <span className="min-w-0 flex-1">
                <span className="block truncate font-medium text-foreground">{target.name || target.dir}</span>
                <span className="block truncate font-mono text-[10px] text-muted-foreground">{target.dir}</span>
              </span>
            </Button>
          ))}
        </div>
      )}
    </DropdownMenu>
  );
}

function EditPencil({ label, onClick, disabled }: { label: string; onClick: () => void; disabled?: boolean }) {
  return (
    <Button
      variant="ghost"
      size="icon"
      type="button"
      onClick={onClick}
      disabled={disabled}
      title={label}
      aria-label={label}
      className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
    >
      <UiEdit className="text-xs" />
    </Button>
  );
}

function TodoSection({
  title,
  icon: Icon,
  count,
  defaultOpen = false,
  resetKey,
  action,
  children,
}: {
  title: string;
  icon: ComponentType<IconProps>;
  count?: number;
  defaultOpen?: boolean;
  resetKey: string;
  // action renders a control (e.g. an edit pencil) to the right of the header,
  // outside the toggle button so it doesn't nest interactive elements.
  action?: ReactNode;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(defaultOpen);

  useEffect(() => {
    setOpen(defaultOpen);
  }, [defaultOpen, resetKey]);

  return (
    <section className="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <div className="flex w-full min-w-0 items-center gap-2 border-b border-border bg-muted/30 pr-2">
        <Button
          variant="ghost"
          type="button"
          onClick={() => setOpen(o => !o)}
          className="flex h-auto min-w-0 flex-1 items-center justify-start gap-2 rounded-none px-3 py-2.5 text-left hover:bg-muted/70"
          aria-expanded={open}
        >
          {open ? <UiChevronDown className="shrink-0 text-xs text-muted-foreground" /> : <UiChevronRight className="shrink-0 text-xs text-muted-foreground" />}
          <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground">
            <Icon className="text-xs" />
          </span>
          <span className="min-w-0 flex-1 truncate text-xs font-semibold uppercase tracking-wide text-muted-foreground">{title}</span>
          {typeof count === 'number' && (
            <span className="rounded-full border border-border bg-background px-1.5 py-0.5 text-[11px] tabular-nums text-muted-foreground">
              {count}
            </span>
          )}
        </Button>
        {action}
      </div>
      {open && <div className="px-3 py-3">{children}</div>}
    </section>
  );
}
