import { Button } from '@flanksource/clicky-ui/components';
import { Markdown } from '@flanksource/clicky-ui/data';
import { UiArrowLeft, UiCheck, UiCopy, UiError, UiMarkdown } from '@flanksource/clicky-ui/icons';
import type { Project, TodoItem } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { TodoTimeline } from './TodoTimeline';
import { TodoCommits } from './TodoCommits';
import { TodoSession } from './TodoSession';
import { TodoPlan } from './TodoPlan';
import { statusClass, statusIcon, statusLabel } from './format';
import { TodoRunAdvancedDialog } from './TodoRunAdvancedDialog';
import { defaultRunOptions, loadLastTodoRunOptions } from './run';
import { TodoPhaseButton, TodoPhaseTicks } from './TodoPhaseButton';
import { TodoRunStatusStrip } from './TodoRunStatusStrip';
import { TodoBodyEditor, TodoCommentBox, TodoTitleEditor } from './TodoCompose';
import { TodoVerification } from './TodoVerification';
import { TodoDetailTabs } from './TodoDetailTabs';
import { TodoReviewBanner } from './planActions';
import { TodoNavigationControls, type TodoNavigationControlsProps } from './TodoNavigationControls';
import { TodoTagField } from './TodoTag';
import { useTodoDetail } from './useTodoDetail';
import { EditPencil, ExternalIssueLink, PriorityMenu, StatusMenu, TodoSection } from './TodoDetailFields';
import { HeaderActionsMenu } from './TodoDetailActions';

export interface TodoDetailProps {
  todo: TodoItem | null;
  loading: boolean;
  loadError?: string;
  dir: string;
  onChanged: (todo: TodoItem) => void;
  onDeleted: () => void;
  onBack?: () => void;
  workspaces?: Project[];
  onTransferred?: (toDir: string, todo: TodoItem) => void;
  navigation?: TodoNavigationControlsProps;
}

export function TodoDetail(props: TodoDetailProps) {
  const { todo, loading, loadError, dir, onChanged, onBack, onTransferred, navigation } = props;
  const {
    advancedMode, setAdvancedMode, runSelections, setRunSelections, error, tab, setTab,
    editingTitle, setEditingTitle, editingBody, setEditingBody, draftTitle, setDraftTitle,
    draftBody, setDraftBody, copyState, runBusy, runMessage, runError, runContext,
    busy, transferTargets, closed, body, events, verificationDetail, verificationError,
    verification, verificationRun, sessionStop, stoppableAttempt, fullTodoId, visibleLabels,
    tagIndex, tagCounts, viewSessionId, viewingHistoricalSession, sessionInProgress,
    awaitingHumanAction, phaseOptions, runningPhaseLabel, changePhaseOptions, patch,
    startEditTitle, startEditBody, saveTitle, saveBody, transferTo, pushToGithub,
    archiveTodo, copyFullId, runPhase, stopRun, runTodo, submitAdvanced,
  } = useTodoDetail(props);

  if (!todo) {
    return (
      <div className="flex h-full min-h-0 flex-col text-sm text-muted-foreground">
        {(onBack || navigation) && (
          <div className="flex shrink-0 items-center justify-between border-b border-border px-3 py-2">
            <div>
              {onBack && (
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
              )}
            </div>
            {navigation && <TodoNavigationControls {...navigation} />}
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
              {navigation && <TodoNavigationControls {...navigation} />}
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
                  <TodoPhaseTicks todo={todo} />
                  <TodoPhaseButton
                    todo={todo}
                    sessionInProgress={sessionInProgress}
                    context={runContext}
                    options={phaseOptions}
                    disabled={busy || !runContext}
                    busy={runBusy || verificationRun.isPending}
                    onRunStep={runPhase}
                    onReview={() => setTab('overview')}
                    onOptionsChange={changePhaseOptions}
                    onAdvanced={setAdvancedMode}
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
        onRun={submitAdvanced}
        loading={runBusy || verificationRun.isPending}
        initialMode={advancedMode ?? 'run'}
        nextStep={todo.lifecycle?.next ?? null}
        title={`${runContext?.lifecycle.steps.find(step => step.name === advancedMode)?.label ?? advancedMode ?? 'Run'} todo`}
        dir={dir}
        refID={todo.ref}
      />
      <TodoReviewBanner todo={todo} dir={dir} onChanged={onChanged} />
      <TodoDetailTabs tab={tab} onSelect={setTab} verification={verification} />
      <div className="flex min-h-0 flex-1 flex-col bg-muted/30">
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
