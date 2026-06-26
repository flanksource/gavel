import { useEffect, useRef, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import type { TodoCounts, TodoStatus } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { STATUS_FILTER_DEFS } from './todoFilter';

// TodoFilterMenu is the compact dropdown counterpart to TodoFilterBar: a single
// "Filter" button (badging how many statuses are currently hidden) that opens a
// popover of status toggles. It drives the same shared hidden-status set, but
// folds the inline pill row into one control for tight chrome like the menubar
// todos tab strip. Only statuses that have todos appear, mirroring TodoFilterBar.
export function TodoFilterMenu({ counts, hidden, onToggle }: {
  counts: TodoCounts;
  hidden: Set<TodoStatus>;
  onToggle: (status: TodoStatus) => void;
}) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', onDown);
    return () => document.removeEventListener('mousedown', onDown);
  }, [open]);

  const defs = STATUS_FILTER_DEFS.filter(d => counts[d.countKey] > 0);
  const hiddenCount = defs.filter(d => hidden.has(d.status)).length;

  return (
    <div className="relative" ref={rootRef}>
      <Button
        variant="ghost"
        type="button"
        onClick={() => setOpen(o => !o)}
        aria-haspopup="menu"
        aria-expanded={open}
        title="Filter todos by status"
        className="inline-flex h-8 items-center gap-1.5 rounded-md border border-border px-2 text-xs text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
      >
        <span className="font-medium">Filter</span>
        {hiddenCount > 0 && (
          <span className="rounded-full bg-muted px-1 text-[10px] tabular-nums text-muted-foreground">{hiddenCount} hidden</span>
        )}
        <GavelIcon name="codicon:chevron-down" className="text-[10px]" />
      </Button>

      {open && (
        <div
          role="menu"
          aria-label="Filter todos by status"
          className="absolute top-full left-0 z-50 mt-1 w-44 rounded-lg border border-border bg-popover py-1 text-sm shadow-lg"
        >
          {defs.length === 0 ? (
            <div className="px-3 py-2 text-xs text-muted-foreground">No todos</div>
          ) : (
            defs.map(def => {
              const active = !hidden.has(def.status);
              return (
                <Button
                  key={def.status}
                  variant="ghost"
                  type="button"
                  role="menuitemcheckbox"
                  aria-checked={active}
                  onClick={() => onToggle(def.status)}
                  className="flex h-auto w-full items-center justify-start gap-2 px-3 py-1.5 text-left transition-colors hover:bg-muted"
                >
                  <GavelIcon
                    name={active ? 'codicon:check' : 'codicon:circle-large-outline'}
                    className={`text-xs ${active ? 'text-primary' : 'text-muted-foreground opacity-50'}`}
                  />
                  <span className={`flex-1 ${active ? 'text-foreground' : 'text-muted-foreground line-through opacity-60'}`}>{def.label}</span>
                  <span className="tabular-nums text-[11px] text-muted-foreground">{counts[def.countKey]}</span>
                </Button>
              );
            })
          )}
        </div>
      )}
    </div>
  );
}
