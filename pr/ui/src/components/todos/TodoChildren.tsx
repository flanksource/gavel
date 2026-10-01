import { useMemo, useState, type MouseEvent, type ReactNode } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { UiAdd, UiCoins, UiListTree, UiTimer } from '@flanksource/clicky-ui/icons';
import { TODO_PHASES, type TodoItem } from '../../types';
import { buildRoute, emptyRouteState } from '../../routes';
import { useNow } from '../../useNow';
import { RelativeTime } from '../RelativeTime';
import { CreateTodoDialog } from './CreateTodoDialog';
import { StatusIcon, TodoDiffBadge } from './format';
import { TodoPhaseStrip } from './TodoPhaseCell';
import { formatCost, formatDuration } from './TodoSessionTimer';
import { childHasLivePhase, childPhaseTotals, childRollup, type TodoFamilyStatus } from './todoFamily';

const ROLLUP_TITLE = 'Latest run of each phase, summed over every child';

// A route link that a plain click resolves in place through `onOpen`, so the
// selection changes without a page load. Modified clicks stay with the browser
// (new tab, copy link). `data-app-link` keeps the menubar's external-link
// interceptor from sending it out of the popover.
export function TodoRouteLink({ todoRef, onOpen, className, title, children }: {
  todoRef: string;
  onOpen?: (ref: string) => void;
  className?: string;
  title?: string;
  children: ReactNode;
}) {
  const href = buildRoute({ ...emptyRouteState(), tab: 'todos', selectedPath: todoRef });
  const open = (event: MouseEvent<HTMLAnchorElement>) => {
    if (!onOpen || event.defaultPrevented || event.button !== 0) return;
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    event.preventDefault();
    onOpen(todoRef);
  };
  return (
    <a href={href} data-app-link="" className={className} title={title} onClick={open}>
      {children}
    </a>
  );
}

function ChildRow({ child, nowMs, onOpenTodo }: { child: TodoItem; nowMs: number; onOpenTodo?: (ref: string) => void }) {
  const { durationMs, costUsd } = childPhaseTotals(child, nowMs);
  const cost = formatCost(costUsd);
  return (
    <li className="flex flex-wrap items-center gap-x-3 gap-y-1 px-3 py-2 text-xs">
      <StatusIcon status={child.status} />
      <TodoRouteLink
        todoRef={child.ref}
        onOpen={onOpenTodo}
        title={child.title}
        className="min-w-0 flex-1 basis-40 truncate text-sm font-medium text-foreground hover:underline"
      >
        {child.title}
      </TodoRouteLink>
      <TodoPhaseStrip todo={child} phases={TODO_PHASES} />
      {durationMs > 0 && <span className="tabular-nums text-muted-foreground" title="Time across phases">{formatDuration(durationMs)}</span>}
      {cost && <span className="tabular-nums text-muted-foreground" title="Cost across phases">{cost}</span>}
      {child.lastRun && (
        <RelativeTime iso={child.lastRun} title={new Date(child.lastRun).toLocaleString()} className="text-[11px] text-muted-foreground" />
      )}
      {child.diff && <TodoDiffBadge diff={child.diff} />}
    </li>
  );
}

interface ChildrenCardProps {
  canAdd: boolean;
  childTodos: TodoItem[];
  familyStatus: TodoFamilyStatus;
  onOpenTodo?: (ref: string) => void;
  onAdd: () => void;
  nowMs: number;
}

function ChildrenBody({ familyStatus, childTodos, nowMs, onOpenTodo }: Pick<ChildrenCardProps, 'familyStatus' | 'childTodos' | 'nowMs' | 'onOpenTodo'>) {
  if (familyStatus.state === 'loading') {
    return <p role="status" className="px-3 py-4 text-xs text-muted-foreground">Loading children…</p>;
  }
  if (familyStatus.state === 'error') {
    return <p role="alert" className="px-3 py-4 text-xs text-red-600">Could not list the children: {familyStatus.message}</p>;
  }
  if (childTodos.length === 0) return <p className="px-3 py-4 text-xs text-muted-foreground">No children yet</p>;
  return (
    <ul className="divide-y divide-border">
      {childTodos.map(child => <ChildRow key={child.ref} child={child} nowMs={nowMs} onOpenTodo={onOpenTodo} />)}
    </ul>
  );
}

function ChildrenCard({ canAdd, childTodos, familyStatus, onOpenTodo, onAdd, nowMs }: ChildrenCardProps) {
  const rollup = useMemo(() => childRollup(childTodos, nowMs), [childTodos, nowMs]);
  const cost = formatCost(rollup.costUsd);
  const known = familyStatus.state === 'ready';

  return (
    <section className="overflow-hidden rounded-lg border border-border bg-card shadow-sm">
      <header className="flex flex-wrap items-center gap-2 border-b border-border bg-muted/30 px-3 py-2.5">
        <span className="inline-flex h-6 w-6 shrink-0 items-center justify-center rounded-md border border-border bg-background text-muted-foreground">
          <UiListTree className="text-xs" />
        </span>
        <h2 className="min-w-0 flex-1 truncate text-xs font-semibold uppercase text-muted-foreground">Children</h2>
        {known && rollup.total > 0 && (
          <span className="inline-flex items-center gap-2 text-[11px] tabular-nums text-muted-foreground" title={ROLLUP_TITLE}>
            <span title="Children done of total">{rollup.done}/{rollup.total}</span>
            {rollup.durationMs > 0 && (
              <span className="inline-flex items-center gap-1"><UiTimer className="text-[11px]" />{formatDuration(rollup.durationMs)}</span>
            )}
            {cost && <span className="inline-flex items-center gap-1"><UiCoins className="text-[11px]" />{cost}</span>}
          </span>
        )}
        {canAdd && (
          <Button variant="outline" size="sm" type="button" onClick={onAdd} className="h-7 gap-1 px-2 text-xs">
            <UiAdd className="text-xs" />
            Add child
          </Button>
        )}
      </header>
      <ChildrenBody familyStatus={familyStatus} childTodos={childTodos} nowMs={nowMs} onOpenTodo={onOpenTodo} />
    </section>
  );
}

// Only a live phase's elapsed time changes between polls, so only then does the
// card subscribe to the one-second clock.
function LiveChildrenCard(props: Omit<ChildrenCardProps, 'nowMs'>) {
  return <ChildrenCard {...props} nowMs={useNow()} />;
}

// The children of a todo, with what each one has cost so far. Lives in the
// parent's Overview: the listing hides children, so this is where they are.
// `familyStatus` keeps "still loading" and "could not load" apart from "none".
export function TodoChildren({ todo, dir, childTodos, familyStatus = { state: 'ready' }, onOpenTodo }: {
  todo: TodoItem;
  dir: string;
  childTodos: TodoItem[];
  familyStatus?: TodoFamilyStatus;
  onOpenTodo?: (ref: string) => void;
}) {
  const [adding, setAdding] = useState(false);
  const card = {
    // Only one level: a child has no children of its own, so it cannot be given any.
    canAdd: !todo.parentId,
    childTodos,
    familyStatus,
    onOpenTodo,
    onAdd: () => setAdding(true),
  };
  return (
    <>
      {childTodos.some(childHasLivePhase) ? <LiveChildrenCard {...card} /> : <ChildrenCard {...card} nowMs={Date.now()} />}
      <CreateTodoDialog
        open={adding}
        onClose={() => setAdding(false)}
        parent={{ ref: todo.ref, title: todo.title, dir }}
        onCreated={() => setAdding(false)}
      />
    </>
  );
}
