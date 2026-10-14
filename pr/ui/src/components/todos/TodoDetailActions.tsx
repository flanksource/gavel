import { useState, type ComponentType, type ReactNode } from 'react';
import { Button, DropdownMenu } from '@flanksource/clicky-ui/components';
import { UiCheck, UiCheckFilled, UiClose, UiCog, UiCopy, UiDebugStepOver, UiDotsVertical, UiEdit, UiGitPr, UiLinkExternal, UiListDashes, UiPass, UiPlay, UiRestart, UiStop, UiTrash, type IconProps } from '@flanksource/clicky-ui/icons';
import type { Project, TodoItem, TodoPriority, TodoStatus } from '../../types';
import { Spinner } from '../../icons/Spinner';
import { priorities, priorityIcon, statusIcon, statuses, statusLabel } from './format';
import type { TodoRunAction } from './run';
import { TodoTagRow } from './TodoTag';
import type { TagIndex } from './tagResolve';
import { LinkPRDialog } from './LinkPRDialog';
import { MoveSubmenu, SetParentSubmenu, UnlinkPRSubmenu } from './TodoActionSubmenus';

export function HeaderActionsMenu({
  todo,
  dir,
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
  parentCandidates,
  hasChildren,
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
  onSetParent,
  onRemoveParent,
}: {
  todo: TodoItem;
  /** The todo's workspace, which scopes its PR links. */
  dir: string;
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
  /** Top-level todos of this workspace, other than this one, that could become its parent. */
  parentCandidates: TodoItem[];
  /** A todo with children is a parent and cannot itself be given one. */
  hasChildren: boolean;
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
  onSetParent: (ref: string) => void;
  onRemoveParent: () => void;
}) {
  // Owned here, not by the menu content: that unmounts the moment the menu closes.
  const [linkingPR, setLinkingPR] = useState(false);
  return (
    <>
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
            {!hasChildren && parentCandidates.length > 0 && (
              <SetParentSubmenu
                disabled={busy}
                candidates={parentCandidates}
                currentParentId={todo.parentId}
                onSelect={onSetParent}
                onCloseParent={close}
              />
            )}
            {todo.parentId && (
              <MobileMenuItem
                icon={UiClose}
                label="Remove parent"
                disabled={busy}
                onClick={() => {
                  close();
                  onRemoveParent();
                }}
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
              icon={UiGitPr}
              label="Link PR…"
              detail="Pick a pull request or paste its URL"
              disabled={busy}
              onClick={() => {
                close();
                setLinkingPR(true);
              }}
            />
            <UnlinkPRSubmenu dir={dir} todoRef={todo.ref} disabled={busy} onCloseParent={close} />
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
    <LinkPRDialog open={linkingPR} dir={dir} todoRef={todo.ref} onClose={() => setLinkingPR(false)} />
    </>
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

