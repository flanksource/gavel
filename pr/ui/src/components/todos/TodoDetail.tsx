import { useEffect, useState, type ReactNode } from 'react';
import { Button, DropdownMenu } from '@flanksource/clicky-ui/components';
import type { Project, TodoItem, TodoPriority, TodoRunOptions, TodoStatus } from '../../types';
import { Markdown } from '../Markdown';
import { GavelIcon } from '../GavelIcon';
import { TodoTimeline } from './TodoTimeline';
import { TodoCommits } from './TodoCommits';
import { TodoSession } from './TodoSession';
import { TodoSessionTimer, useSessionStats } from './TodoSessionTimer';
import { priorities, statusClass, statuses, statusLabel, todoQuery } from './format';
import { TodoRunAdvancedDialog, TodoRunSplitButton, defaultRunOptions, useTodoRun } from './run';
import { TodoBodyEditor, TodoCommentBox, TodoTitleEditor } from './TodoCompose';
import { AcceptanceCriteria } from './AcceptanceCriteria';

export function TodoDetail({
  todo,
  loading,
  dir,
  provider,
  onChanged,
  onDeleted,
  workspaces = [],
  onTransferred,
}: {
  todo: TodoItem | null;
  loading: boolean;
  dir: string;
  provider: string;
  onChanged: (todo: TodoItem) => void;
  onDeleted: () => void;
  // workspaces/onTransferred are optional: the "Move to project" control only
  // renders where a caller wires them (the dashboard), not the compact menubar.
  workspaces?: Project[];
  onTransferred?: (toDir: string, todo: TodoItem) => void;
}) {
  const [busy, setBusy] = useState(false);
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [error, setError] = useState('');
  const [tab, setTab] = useState<'overview' | 'session'>('overview');
  const [editingTitle, setEditingTitle] = useState(false);
  const [editingBody, setEditingBody] = useState(false);
  const [draftTitle, setDraftTitle] = useState('');
  const [draftBody, setDraftBody] = useState('');
  const [copyState, setCopyState] = useState<'idle' | 'copied' | 'error'>('idle');
  const { runBusy, runMessage, runError, reset: resetRun, run } = useTodoRun(dir, provider);
  const isGrite = todo?.provider === 'grite';
  // Projects this todo can move to: every configured workspace except its own.
  const transferTargets = workspaces.filter(ws => !!ws.dir && ws.dir !== dir);
  // A todo is "closed" when completed (Grite also reports providerState).
  const closed = todo?.status === 'completed' || todo?.providerState === 'closed';
  const body = todo?.body?.trim() ?? '';
  const events = todo?.events ?? [];
  const fullTodoId = todo ? todoFullId(todo, provider) : '';
  const visibleLabels = todo ? todoHeaderLabels(todo) : [];
  const { stats: headerSessionStats } = useSessionStats(dir, provider, todo?.sessionId, !!todo?.sessionId);
  const sessionInProgress = !!todo && !!todo.sessionId && (headerSessionStats?.inProgress || (!headerSessionStats?.found && todo.status === 'in_progress'));

  useEffect(() => {
    setError('');
    resetRun();
    setAdvancedOpen(false);
    setTab('overview');
    setEditingTitle(false);
    setEditingBody(false);
    setCopyState('idle');
  }, [todo?.ref, resetRun]);

  useEffect(() => {
    if (copyState === 'idle') return;
    const timeout = window.setTimeout(() => setCopyState('idle'), 1600);
    return () => window.clearTimeout(timeout);
  }, [copyState]);

  // patch sends a partial update (status, priority, title, body, and/or comment)
  // and adopts the server's returned todo so the view reflects provider-side side
  // effects (labels, state, rewritten body, new comment). Resolves true on success.
  async function patch(payload: {
    status?: TodoStatus;
    priority?: TodoPriority;
    title?: string;
    body?: string;
    comment?: string;
  }): Promise<boolean> {
    if (!todo || busy) return false;
    setBusy(true);
    setError('');
    try {
      const response = await fetch(`/api/todos/item?${todoQuery(dir, provider)}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ref: todo.ref, ...payload }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || 'Update failed');
      onChanged(data as TodoItem);
      return true;
    } catch (err: any) {
      setError(err?.message || 'Update failed');
      return false;
    } finally {
      setBusy(false);
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
    const target = transferTargets.find(ws => ws.dir === toDir);
    setBusy(true);
    setError('');
    try {
      const response = await fetch('/api/todos/transfer', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ref: todo.ref,
          fromDir: dir,
          fromProvider: provider,
          toDir,
          toProvider: target?.todoProvider || 'auto',
        }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || 'Move failed');
      onTransferred(toDir, data.todo as TodoItem);
    } catch (err: any) {
      setError(err?.message || 'Move failed');
    } finally {
      setBusy(false);
    }
  }

  async function archiveTodo() {
    if (!todo || busy) return;
    if (!window.confirm(isGrite ? 'Archive this Grite issue?' : 'Delete this TODO file?')) return;
    setBusy(true);
    setError('');
    try {
      const params = new URLSearchParams(todoQuery(dir, provider));
      params.set('ref', todo.ref);
      const response = await fetch(`/api/todos/item?${params.toString()}`, { method: 'DELETE' });
      if (!response.ok) {
        const data = await response.json().catch(() => ({}));
        throw new Error(data.error || 'Archive failed');
      }
      onDeleted();
    } catch (err: any) {
      setError(err?.message || 'Archive failed');
    } finally {
      setBusy(false);
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

  async function runTodo(options?: TodoRunOptions) {
    if (!todo) return;
    const result = await run([todo.ref], options);
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
      <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
        <div className="text-center">
          <GavelIcon name="codicon:check" className="mb-2 text-4xl" />
          <p>Select a todo</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col">
      <div className="shrink-0 border-b border-border bg-background px-3 py-2 sm:px-4 sm:py-4">
        <div className="flex min-w-0 flex-col gap-2 sm:gap-3">
          <div className="grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-center gap-2">
            <div className="min-w-0">
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
                  <h1 className="min-w-0 flex-1 truncate text-xl font-semibold leading-8 text-foreground">
                    {todo.title}
                  </h1>
                  <span className="hidden sm:inline-flex">
                    <EditPencil label="Edit title" onClick={startEditTitle} disabled={busy} />
                  </span>
                </div>
              )}
            </div>

            <MobileHeaderMenu
              todo={todo}
              busy={busy}
              runBusy={runBusy}
              closed={closed}
              isGrite={isGrite}
              sessionInProgress={sessionInProgress}
              fullTodoId={fullTodoId}
              labels={visibleLabels}
              transferTargets={transferTargets}
              canTransfer={!!onTransferred}
              onCopy={copyFullId}
              onEditTitle={startEditTitle}
              onStatus={status => patch({ status })}
              onPriority={priority => patch({ priority })}
              onTransfer={transferTo}
              onResume={() => runTodo({ ...defaultRunOptions, resume: true })}
              onRun={() => runTodo(defaultRunOptions)}
              onAdvanced={() => setAdvancedOpen(true)}
              onVerify={() => patch({ status: 'verified' })}
              onToggleClosed={() => patch({ status: closed ? 'pending' : 'completed' })}
              onArchive={archiveTodo}
            />

            <div className="hidden min-w-0 flex-wrap items-center justify-end gap-1.5 sm:flex">
              {onTransferred && transferTargets.length > 0 && (
                <MoveMenu
                  disabled={busy}
                  targets={transferTargets}
                  onSelect={transferTo}
                />
              )}
              {todo.sessionId && (
                <Button
                  variant="ghost"
                  size="icon"
                  type="button"
                  onClick={() => runTodo({ ...defaultRunOptions, resume: true })}
                  disabled={busy || runBusy || sessionInProgress}
                  title={sessionInProgress ? 'Session is already running' : 'Resume prior agent session'}
                  className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50"
                  aria-label="Resume prior agent session"
                >
                  <GavelIcon name={runBusy ? 'svg-spinners:ring-resize' : 'codicon:debug-step-over'} className="text-sm" />
                </Button>
              )}
              <TodoRunSplitButton
                disabled={busy || runBusy || sessionInProgress}
                loading={runBusy}
                label={sessionInProgress ? 'Stop' : 'Run'}
                icon={sessionInProgress ? 'codicon:debug-stop' : 'codicon:play'}
                tone={sessionInProgress ? 'danger' : 'default'}
                title={sessionInProgress ? 'Stop is unavailable until session interrupt is supported' : 'Run todo'}
                onRun={runTodo}
                onAdvanced={() => setAdvancedOpen(true)}
              />
              <HeaderIconButton
                icon={busy ? 'svg-spinners:ring-resize' : 'octicon:check-circle-fill-16'}
                label={todo.status === 'verified' ? 'Already verified' : 'Mark verified'}
                onClick={() => patch({ status: 'verified' })}
                disabled={busy || todo.status === 'verified'}
                className="text-emerald-600 hover:text-emerald-700"
              />
              <HeaderIconButton
                icon={busy ? 'svg-spinners:ring-resize' : closed ? 'codicon:debug-restart' : 'codicon:pass'}
                label={closed ? 'Reopen todo' : 'Close todo'}
                onClick={() => patch({ status: closed ? 'pending' : 'completed' })}
                disabled={busy}
              />
              <HeaderIconButton
                icon={busy ? 'svg-spinners:ring-resize' : 'codicon:trash'}
                label={isGrite ? 'Archive issue' : 'Delete file'}
                onClick={archiveTodo}
                disabled={busy}
                className="hover:text-red-600"
              />
            </div>

            <div className="hidden min-w-0 flex-wrap items-center gap-2 text-xs text-muted-foreground sm:col-span-2 sm:flex">
              <Button
                variant="ghost"
                type="button"
                onClick={copyFullId}
                title={copyState === 'copied' ? 'Copied' : 'Copy full issue ID'}
                className="inline-flex h-auto min-w-0 max-w-full items-center gap-1.5 rounded border border-border bg-muted/20 px-2 py-1 text-left font-mono text-[11px] hover:bg-muted"
              >
                <GavelIcon
                  name={copyState === 'copied' ? 'codicon:check' : copyState === 'error' ? 'codicon:error' : 'codicon:copy'}
                  className={copyState === 'error' ? 'shrink-0 text-red-600' : 'shrink-0 text-muted-foreground'}
                />
                <span className="min-w-0 truncate">{fullTodoId}</span>
              </Button>
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
              <HeaderTags labels={visibleLabels} />
              {copyState === 'error' && <span className="text-red-600">Copy failed</span>}
            </div>
          </div>
          {(error || runError) && <div className="mt-2 text-xs text-red-600">{error || runError}</div>}
          {runMessage && !error && !runError && <div className="mt-2 text-xs text-emerald-600">{runMessage}</div>}
          {todo.sessionId && (
            <div className="hidden sm:block">
              <TodoSessionTimer dir={dir} provider={provider} sessionId={todo.sessionId} />
            </div>
          )}
        </div>
      </div>
      <TodoRunAdvancedDialog
        open={advancedOpen}
        onClose={() => setAdvancedOpen(false)}
        onRun={options => {
          setAdvancedOpen(false);
          runTodo(options);
        }}
        loading={runBusy}
        dir={dir}
        provider={provider}
        refs={[todo.ref]}
      />
      <div className="flex shrink-0 gap-1 border-b border-border bg-background px-4 pt-2">
        <DetailTab active={tab === 'overview'} onClick={() => setTab('overview')} icon="codicon:list-flat" label="Overview" />
        <DetailTab active={tab === 'session'} onClick={() => setTab('session')} icon="codicon:comment-discussion" label="Session" />
      </div>
      <div className="flex min-h-0 flex-1 flex-col bg-[#f4f6f9] dark:bg-[#0a1020]">
        {tab === 'session' ? (
          <TodoSession dir={dir} provider={provider} sessionId={todo.sessionId} active={tab === 'session'} />
        ) : (
          <div className="min-h-0 flex-1 overflow-y-auto px-4 py-3">
            {loading ? (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <GavelIcon name="svg-spinners:ring-resize" />
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
                    icon="codicon:markdown"
                    defaultOpen
                    resetKey={`${todo.ref}:body`}
                    action={<EditPencil label="Edit body" onClick={startEditBody} disabled={busy} />}
                  >
                    <Markdown text={body} className="text-sm" />
                  </TodoSection>
                ) : (
                  <TodoSection
                    title="Body"
                    icon="codicon:markdown"
                    defaultOpen
                    resetKey={`${todo.ref}:body-empty`}
                    action={<EditPencil label="Add body" onClick={startEditBody} disabled={busy} />}
                  >
                    <p className="text-sm text-muted-foreground">No body yet.</p>
                  </TodoSection>
                )}
                <AcceptanceCriteria dir={dir} provider={provider} todo={todo} onChanged={onChanged} />
                <TodoCommentBox
                  closed={closed}
                  busy={busy}
                  onComment={(text, reopen) => patch(reopen ? { status: 'pending', comment: text } : { comment: text })}
                />
                <TodoCommits dir={dir} provider={provider} todoRef={todo.ref} />
                {events.length > 0 && <TodoTimeline events={events} />}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

function todoFullId(todo: TodoItem, fallbackProvider: string): string {
  const filePath = todo.filePath?.trim();
  if (filePath?.startsWith('grite:')) return filePath;
  if (todo.id) return `${todo.provider || fallbackProvider || 'todo'}:${todo.id}`;
  return filePath || todo.ref;
}

function todoHeaderLabels(todo: TodoItem): string[] {
  return (todo.labels ?? [])
    .map(label => label.trim())
    .filter(label =>
      label &&
      !label.startsWith('status:') &&
      !label.startsWith('priority:') &&
      !label.startsWith('session:'),
    );
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
            <GavelIcon name={statusIcon(value)} className="text-xs" />
          </span>
          <span className={compact ? 'capitalize text-foreground' : ''}>{statusLabel(value)}</span>
          <GavelIcon name="codicon:chevron-down" className="text-[11px] opacity-70" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          {statuses.map(status => (
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
                <GavelIcon name={statusIcon(status)} className="text-xs" />
              </span>
              <span className="min-w-0 flex-1 capitalize text-foreground">{statusLabel(status)}</span>
              {status === value && <GavelIcon name="codicon:check" className="text-xs text-primary" />}
            </Button>
          ))}
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
            <GavelIcon name={priorityIcon(value)} className="text-xs" />
          </span>
          <span className={compact ? 'capitalize text-foreground' : ''}>{value}</span>
          <GavelIcon name="codicon:chevron-down" className="text-[11px] opacity-70" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          {priorities.map(priority => (
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
                <GavelIcon name={priorityIcon(priority)} className="text-xs" />
              </span>
              <span className="min-w-0 flex-1 capitalize text-foreground">{priority}</span>
              {priority === value && <GavelIcon name="codicon:check" className="text-xs text-primary" />}
            </Button>
          ))}
        </div>
      )}
    </DropdownMenu>
  );
}

function HeaderTags({ labels }: { labels: string[] }) {
  if (labels.length === 0) return null;
  const visible = labels.slice(0, 6);
  const overflow = labels.length - visible.length;
  return (
    <>
      {visible.map(label => (
        <span
          key={label}
          className="inline-flex max-w-[14rem] items-center truncate rounded border border-border bg-muted/20 px-2 py-1 text-[11px] text-muted-foreground"
          title={label}
        >
          <span className="truncate">{label}</span>
        </span>
      ))}
      {overflow > 0 && (
        <span className="inline-flex items-center rounded border border-border bg-muted/20 px-2 py-1 text-[11px] text-muted-foreground">
          +{overflow}
        </span>
      )}
    </>
  );
}

function MobileHeaderMenu({
  todo,
  busy,
  runBusy,
  closed,
  isGrite,
  sessionInProgress,
  fullTodoId,
  labels,
  transferTargets,
  canTransfer,
  onCopy,
  onEditTitle,
  onStatus,
  onPriority,
  onTransfer,
  onResume,
  onRun,
  onAdvanced,
  onVerify,
  onToggleClosed,
  onArchive,
}: {
  todo: TodoItem;
  busy: boolean;
  runBusy: boolean;
  closed: boolean;
  isGrite: boolean;
  sessionInProgress: boolean;
  fullTodoId: string;
  labels: string[];
  transferTargets: Project[];
  canTransfer: boolean;
  onCopy: () => void;
  onEditTitle: () => void;
  onStatus: (status: TodoStatus) => void;
  onPriority: (priority: TodoPriority) => void;
  onTransfer: (dir: string) => void;
  onResume: () => void;
  onRun: () => void;
  onAdvanced: () => void;
  onVerify: () => void;
  onToggleClosed: () => void;
  onArchive: () => void;
}) {
  return (
    <DropdownMenu
      align="right"
      menuLabel="Issue actions"
      menuClassName="w-72 max-h-[80vh] max-w-[calc(100vw-16px)] overflow-y-auto"
      className="sm:hidden"
      trigger={
        <Button
          variant="ghost"
          size="icon"
          type="button"
          title="Issue actions"
          aria-label="Issue actions"
          className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground"
        >
          <GavelIcon name="codicon:kebab-vertical" className="text-sm" />
        </Button>
      }
    >
      {close => (
        <div className="p-1 text-xs">
          <MobileMenuSection title="Issue">
            <MobileMenuItem
              icon="codicon:copy"
              label="Copy issue ID"
              detail={fullTodoId}
              onClick={() => {
                close();
                onCopy();
              }}
            />
            <MobileMenuItem
              icon="codicon:edit"
              label="Edit title"
              disabled={busy}
              onClick={() => {
                close();
                onEditTitle();
              }}
            />
          </MobileMenuSection>

          <MobileMenuSection title="Run">
            {todo.sessionId && (
              <MobileMenuItem
                icon="codicon:debug-step-over"
                label="Resume session"
                disabled={busy || runBusy || sessionInProgress}
                onClick={() => {
                  close();
                  onResume();
                }}
              />
            )}
            <MobileMenuItem
              icon={sessionInProgress ? 'codicon:debug-stop' : runBusy ? 'svg-spinners:ring-resize' : 'codicon:play'}
              label={sessionInProgress ? 'Stop unavailable' : 'Run todo'}
              detail={sessionInProgress ? 'Session interrupt is not supported yet' : undefined}
              disabled={busy || runBusy || sessionInProgress}
              onClick={() => {
                close();
                onRun();
              }}
            />
            <MobileMenuItem
              icon="codicon:gear"
              label="Advanced run"
              disabled={busy || runBusy}
              onClick={() => {
                close();
                onAdvanced();
              }}
            />
          </MobileMenuSection>

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

          {canTransfer && transferTargets.length > 0 && (
            <MobileMenuSection title="Move to">
              {transferTargets.map(target => (
                <MobileMenuItem
                  key={target.dir}
                  icon="codicon:folder"
                  label={target.name || target.dir}
                  detail={target.dir}
                  disabled={busy}
                  onClick={() => {
                    close();
                    onTransfer(target.dir);
                  }}
                />
              ))}
            </MobileMenuSection>
          )}

          <MobileMenuSection title="Actions">
            <MobileMenuItem
              icon="octicon:check-circle-fill-16"
              label={todo.status === 'verified' ? 'Already verified' : 'Mark verified'}
              disabled={busy || todo.status === 'verified'}
              onClick={() => {
                close();
                onVerify();
              }}
            />
            <MobileMenuItem
              icon={closed ? 'codicon:debug-restart' : 'codicon:pass'}
              label={closed ? 'Reopen todo' : 'Close todo'}
              disabled={busy}
              onClick={() => {
                close();
                onToggleClosed();
              }}
            />
            <MobileMenuItem
              icon="codicon:trash"
              label={isGrite ? 'Archive issue' : 'Delete file'}
              disabled={busy}
              danger
              onClick={() => {
                close();
                onArchive();
              }}
            />
          </MobileMenuSection>

          {labels.length > 0 && (
            <MobileMenuSection title="Tags">
              <div className="flex flex-wrap gap-1 px-2 py-1">
                {labels.map(label => (
                  <span
                    key={label}
                    className="max-w-full truncate rounded border border-border bg-muted/20 px-1.5 py-0.5 text-[11px] text-muted-foreground"
                    title={label}
                  >
                    {label}
                  </span>
                ))}
              </div>
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
  icon,
  label,
  detail,
  selected,
  disabled,
  danger,
  onClick,
}: {
  icon: string;
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
      <GavelIcon name={icon} className={`mt-0.5 shrink-0 text-sm ${danger ? 'text-red-600' : 'text-muted-foreground'}`} />
      <span className="min-w-0 flex-1">
        <span className={`block truncate font-medium capitalize ${danger ? 'text-red-600' : 'text-foreground'}`}>{label}</span>
        {detail && <span className="block truncate font-mono text-[10px] text-muted-foreground">{detail}</span>}
      </span>
      {selected && <GavelIcon name="codicon:check" className="mt-0.5 text-xs text-primary" />}
    </Button>
  );
}

function MoveMenu({
  disabled,
  targets,
  onSelect,
}: {
  disabled?: boolean;
  targets: Project[];
  onSelect: (dir: string) => void;
}) {
  return (
    <DropdownMenu
      align="right"
      menuLabel="Move todo to another project"
      menuClassName="w-72 max-w-[calc(100vw-24px)]"
      trigger={
        <Button
          variant="outline"
          type="button"
          disabled={disabled}
          className="inline-flex h-8 items-center gap-1.5 rounded-md px-2.5 text-xs disabled:opacity-50"
          title="Move todo to another project"
          aria-label="Move todo to another project"
        >
          <GavelIcon name="codicon:folder" className="text-xs" />
          Move to…
          <GavelIcon name="codicon:chevron-down" className="text-[11px] opacity-70" />
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
                onSelect(target.dir);
              }}
              className="flex h-auto w-full items-start justify-start gap-2 rounded px-2 py-1.5 text-left hover:bg-muted"
            >
              <GavelIcon name="codicon:folder" className="mt-0.5 shrink-0 text-sm text-muted-foreground" />
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

function HeaderIconButton({
  icon,
  label,
  onClick,
  disabled,
  className = '',
}: {
  icon: string;
  label: string;
  onClick: () => void;
  disabled?: boolean;
  className?: string;
}) {
  return (
    <Button
      variant="ghost"
      size="icon"
      type="button"
      onClick={onClick}
      disabled={disabled}
      title={label}
      aria-label={label}
      className={`inline-flex h-8 w-8 items-center justify-center rounded-md border border-border text-muted-foreground hover:bg-muted hover:text-foreground disabled:opacity-50 ${className}`}
    >
      <GavelIcon name={icon} className="text-sm" />
    </Button>
  );
}

function statusIcon(status: TodoStatus | string): string {
  switch (status) {
    case 'draft':
      return 'codicon:circle-large-outline';
    case 'in_progress':
      return 'svg-spinners:ring-resize';
    case 'failed':
      return 'octicon:x-circle-fill-16';
    case 'verified':
      return 'octicon:check-circle-fill-16';
    case 'completed':
      return 'codicon:pass';
    case 'skipped':
      return 'codicon:circle-slash';
    default:
      return 'codicon:circle-large-outline';
  }
}

function priorityIcon(priority: TodoPriority | string): string {
  switch (priority) {
    case 'high':
      return 'codicon:chevron-up';
    case 'low':
      return 'codicon:chevron-down';
    default:
      return 'codicon:circle-large-outline';
  }
}

function priorityBadgeClass(priority: TodoPriority | string): string {
  switch (priority) {
    case 'high':
      return 'border-red-500/25 bg-red-500/10 text-red-600';
    case 'low':
      return 'border-emerald-500/25 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400';
    default:
      return 'border-amber-500/25 bg-amber-500/10 text-amber-700 dark:text-amber-400';
  }
}

function DetailTab({ active, onClick, icon, label }: { active: boolean; onClick: () => void; icon: string; label: string }) {
  return (
    <Button
      variant="ghost"
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={`-mb-px h-auto inline-flex items-center gap-1.5 border-b-2 px-2.5 py-1.5 text-xs font-medium transition-colors ${
        active
          ? 'border-primary text-foreground'
          : 'border-transparent text-muted-foreground hover:text-foreground'
      }`}
    >
      <GavelIcon name={icon} className="text-sm" />
      {label}
    </Button>
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
      <GavelIcon name="codicon:edit" className="text-xs" />
    </Button>
  );
}

function TodoSection({
  title,
  icon,
  count,
  defaultOpen = false,
  resetKey,
  action,
  children,
}: {
  title: string;
  icon: string;
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
          <GavelIcon name={open ? 'codicon:chevron-down' : 'codicon:chevron-right'} className="shrink-0 text-xs text-muted-foreground" />
          <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground">
            <GavelIcon name={icon} className="text-xs" />
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
