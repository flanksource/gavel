import { useEffect, useMemo, useRef, useState, type ComponentType, type KeyboardEvent, type ReactNode } from 'react';
import { Button, Modal } from '@flanksource/clicky-ui/components';
import { UiArrowLeft, UiGitPr, UiListDashes, UiSearch } from '@flanksource/clicky-ui/icons';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import type { PRItem } from '../types';
import type { TodoEntry } from './todos/todoGroup';
import { todoMatchesQuery } from './todos/todoFilter';
import { Spinner } from '../icons/Spinner';

// The ⌘K palette is the dashboard's single global text search: it spans pull
// requests and todos regardless of the active tab and jumps to the chosen item
// (switching tabs as needed) rather than filtering a list in place. The FilterBar
// still owns structured PR facet filtering; this owns "find that one thing".

const isMac = typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent || '');
export const paletteShortcutLabel = isMac ? '⌘K' : 'Ctrl K';

// GROUP_CAP keeps each section short so the list stays scannable; overflow is
// surfaced as a "+N more" hint rather than silently dropped.
const GROUP_CAP = 8;

interface Row {
  key: string;
  icon: ComponentType<IconProps>;
  title: string;
  subtitle: string;
  meta: string;
  onSelect: () => void;
}

function prMatchesQuery(pr: PRItem, q: string): boolean {
  return (
    pr.title.toLowerCase().includes(q) ||
    String(pr.number).includes(q) ||
    pr.repo.toLowerCase().includes(q) ||
    (pr.source || '').toLowerCase().includes(q) ||
    (pr.target || '').toLowerCase().includes(q)
  );
}

const UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

// uuidQuery deliberately accepts every UUID version: native Todo IDs, Captain
// IDs, and provider-issued Claude/Codex IDs do not all use the same version.
export function uuidQuery(value: string): string | null {
  const trimmed = value.trim();
  return UUID_PATTERN.test(trimmed) ? trimmed.toLowerCase() : null;
}

// SearchTrigger is the top-bar affordance that stands in for the old inline
// search box: a click (or the ⌘K/Ctrl+K shortcut) opens the palette. Present on
// every tab so global search is always one key away.
export function SearchTrigger({ onOpen }: { onOpen: () => void }) {
  return (
    <button
      type="button"
      onClick={onOpen}
      aria-label="Search pull requests, todos, and sessions"
      className="flex w-full items-center gap-2 rounded-md border border-border bg-muted px-3 py-1.5 text-sm text-muted-foreground hover:bg-muted/70"
    >
      <UiSearch className="shrink-0 text-sm" />
      <span className="flex-1 truncate text-left">Search PRs, todos, or UUID…</span>
      <kbd className="shrink-0 rounded border border-border bg-background px-1.5 py-0.5 text-[10px] font-medium tabular-nums text-muted-foreground">
        {paletteShortcutLabel}
      </kbd>
    </button>
  );
}

export function CommandPalette({ open, onClose, prs, todos, todosLoading, onSelectPR, onSelectTodo, onOpenUUID }: {
  open: boolean;
  onClose: () => void;
  prs: PRItem[];
  todos: TodoEntry[];
  todosLoading: boolean;
  onSelectPR: (pr: PRItem) => void;
  onSelectTodo: (entry: TodoEntry) => void;
  onOpenUUID: (uuid: string) => void;
}) {
  const [query, setQuery] = useState('');
  const [active, setActive] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  // Reset to a clean slate every time the palette opens, then move focus to the
  // input so the user can type immediately. The Modal focuses its own panel in a
  // passive effect whose flush time varies (opening cascades state updates), so a
  // single deferred focus keeps losing the race. Instead retry briefly until the
  // input holds focus, then stop — there is no focus trap, so once it lands it
  // stays. This self-corrects regardless of when the Modal grabs focus.
  useEffect(() => {
    if (!open) return;
    setQuery('');
    setActive(0);
    let tries = 0;
    const timer = setInterval(() => {
      const el = inputRef.current;
      if (el && document.activeElement !== el) el.focus();
      if (++tries >= 10 || (el && document.activeElement === el)) clearInterval(timer);
    }, 30);
    return () => clearInterval(timer);
  }, [open]);

  const q = query.trim().toLowerCase();
  const directUUID = uuidQuery(query);
  const prMatches = useMemo(() => (q ? prs.filter(pr => prMatchesQuery(pr, q)) : []), [prs, q]);
  const todoMatches = useMemo(() => (q ? todos.filter(e => todoMatchesQuery(e.todo, q)) : []), [todos, q]);

  // One flat, ordered list (PRs then todos) backs keyboard navigation; the two
  // rendered groups index into it so the highlight and Enter stay in sync.
  const rows = useMemo<Row[]>(() => {
    const out: Row[] = [];
    if (directUUID) {
      out.push({
        key: `uuid:${directUUID}`,
        icon: UiSearch,
        title: 'Open Todo or session UUID',
        subtitle: directUUID,
        meta: 'UUID',
        onSelect: () => { onClose(); onOpenUUID(directUUID); },
      });
    }
    for (const pr of prMatches.slice(0, GROUP_CAP)) {
      out.push({
        key: `pr:${pr.repo}#${pr.number}`,
        icon: UiGitPr,
        title: pr.title,
        subtitle: `${pr.repo} #${pr.number}`,
        meta: pr.source || '',
        onSelect: () => { onClose(); onSelectPR(pr); },
      });
    }
    for (const entry of todoMatches.slice(0, GROUP_CAP)) {
      out.push({
        key: `todo:${entry.workspace.dir}\t${entry.todo.ref}`,
        icon: UiListDashes,
        title: entry.todo.title,
        subtitle: entry.workspace.name,
        meta: entry.todo.status.replace('_', ' '),
        onSelect: () => { onClose(); onSelectTodo(entry); },
      });
    }
    return out;
  }, [directUUID, prMatches, todoMatches, onClose, onOpenUUID, onSelectPR, onSelectTodo]);

  // Keep the active index in range as results change while typing.
  useEffect(() => { setActive(a => (rows.length === 0 ? 0 : Math.min(a, rows.length - 1))); }, [rows.length]);

  // Scroll the highlighted row into view as the selection moves.
  useEffect(() => {
    listRef.current?.querySelector<HTMLElement>(`[data-row="${active}"]`)?.scrollIntoView({ block: 'nearest' });
  }, [active]);

  function onKeyDown(e: KeyboardEvent) {
    if (rows.length === 0) return;
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setActive(a => (a + 1) % rows.length);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setActive(a => (a - 1 + rows.length) % rows.length);
    } else if (e.key === 'Enter') {
      e.preventDefault();
      rows[active]?.onSelect();
    }
  }

  const directBase = directUUID ? 1 : 0;
  const prBase = directBase;
  const todoBase = directBase + Math.min(prMatches.length, GROUP_CAP);

  return (
    <Modal open={open} onClose={onClose} size="lg" hideClose expandable={false} closeOnBackdrop closeOnEsc scrollBody={false}>
      <div className="flex shrink-0 items-center gap-2 border-b border-border pb-3 max-sm:-mx-density-4 max-sm:-mt-density-4 max-sm:min-h-14 max-sm:bg-background max-sm:px-density-4 max-sm:py-density-3">
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="-ml-2 sm:!hidden"
          aria-label="Back"
          onClick={onClose}
        >
          <UiArrowLeft />
        </Button>
        <UiSearch className="shrink-0 text-base text-muted-foreground max-sm:hidden" />
        <input
          ref={inputRef}
          value={query}
          onChange={e => { setQuery((e.target as HTMLInputElement).value); setActive(0); }}
          onKeyDown={onKeyDown}
          placeholder="Search pull requests, todos, or paste a UUID…"
          aria-label="Search pull requests, todos, and sessions"
          className="flex-1 min-w-0 bg-transparent text-sm outline-none placeholder:text-muted-foreground"
        />
        <kbd className="hidden shrink-0 rounded border border-border bg-muted px-1.5 py-0.5 text-[10px] text-muted-foreground sm:!inline-flex">esc</kbd>
      </div>

      <div ref={listRef} className="min-h-0 flex-1 overflow-auto pt-2">
        {!q ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            Type to search pull requests and todos, or paste a Todo or session UUID.
          </div>
        ) : rows.length === 0 && !todosLoading ? (
          <div className="py-8 text-center text-sm text-muted-foreground">
            No pull requests, todos, or sessions match “{query.trim()}”.
          </div>
        ) : (
          <>
            {directUUID && (
              <Group label="UUID" overflow={0}>
                <PaletteRow row={rows[0]} index={0} active={active} onHover={setActive} />
              </Group>
            )}
            {prMatches.length > 0 && (
              <Group label="Pull requests" overflow={prMatches.length - GROUP_CAP}>
                {prMatches.slice(0, GROUP_CAP).map((_, i) => (
                  <PaletteRow key={rows[prBase + i].key} row={rows[prBase + i]} index={prBase + i} active={active} onHover={setActive} />
                ))}
              </Group>
            )}
            {todoMatches.length > 0 && (
              <Group label="Todos" overflow={todoMatches.length - GROUP_CAP}>
                {todoMatches.slice(0, GROUP_CAP).map((_, i) => (
                  <PaletteRow key={rows[todoBase + i].key} row={rows[todoBase + i]} index={todoBase + i} active={active} onHover={setActive} />
                ))}
              </Group>
            )}
            {todosLoading && (
              <div className="px-2 py-2 text-xs text-muted-foreground">
                <Spinner className="mr-1" />
                Loading todos…
              </div>
            )}
          </>
        )}
      </div>
    </Modal>
  );
}

function Group({ label, overflow, children }: { label: string; overflow: number; children: ReactNode }) {
  return (
    <div className="mb-1">
      <div className="flex items-center justify-between px-2 py-1 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
        <span>{label}</span>
        {overflow > 0 && <span className="normal-case tracking-normal">+{overflow} more</span>}
      </div>
      {children}
    </div>
  );
}

function PaletteRow({ row, index, active, onHover }: {
  row: Row;
  index: number;
  active: number;
  onHover: (i: number) => void;
}) {
  const isActive = index === active;
  const Icon = row.icon;
  return (
    <button
      type="button"
      data-row={index}
      onMouseMove={() => onHover(index)}
      onClick={row.onSelect}
      className={`flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left ${isActive ? 'bg-primary/10' : 'hover:bg-muted'}`}
    >
      <Icon className="shrink-0 text-sm text-muted-foreground" />
      <span className="min-w-0 flex-1 truncate text-sm text-foreground">{row.title}</span>
      {row.meta && <span className="shrink-0 truncate text-[11px] capitalize text-muted-foreground">{row.meta}</span>}
      <span className="shrink-0 max-w-[12rem] truncate text-[11px] tabular-nums text-muted-foreground">{row.subtitle}</span>
    </button>
  );
}
