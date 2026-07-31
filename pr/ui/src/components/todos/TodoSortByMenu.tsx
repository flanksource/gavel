import { useEffect, useRef, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import { UiCheck, UiChevronDown, UiChevronUp } from '@flanksource/clicky-ui/icons';
import type { TodoSort } from './todoSort';
import { TODO_SORT_COLUMN_OPTIONS } from './todoSort';

export function TodoSortByMenu({ sortBy, onChange }: {
  sortBy: TodoSort;
  onChange: (sortBy: TodoSort) => void;
}) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);

  // Close on outside click — the menu is anchored to its button, so a stray
  // click should dismiss rather than trap the user (mirrors TodoGroupByMenu).
  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', onDown);
    return () => document.removeEventListener('mousedown', onDown);
  }, [open]);

  const active = TODO_SORT_COLUMN_OPTIONS.find(opt => opt.value === sortBy.column)
    ?? TODO_SORT_COLUMN_OPTIONS[0];
  const ActiveIcon = active.icon;
  const DirectionIcon = sortBy.dir === 'desc' ? UiChevronDown : UiChevronUp;
  const directionLabel = sortBy.dir === 'desc' ? 'Sort descending' : 'Sort ascending';

  return (
    <div className="relative inline-flex items-center" ref={rootRef}>
      <div className="inline-flex overflow-hidden rounded-md border border-border">
        <Button
          variant="ghost"
          type="button"
          onClick={() => setOpen(o => !o)}
          aria-haspopup="menu"
          aria-expanded={open}
          className="inline-flex h-8 items-center gap-1.5 rounded-none border-r border-border px-2 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        >
          <ActiveIcon className="text-xs" />
          <span className="font-medium">Sort: {active.label}</span>
          <UiChevronDown className="text-[10px]" />
        </Button>
        <Button
          variant="ghost"
          type="button"
          onClick={() => onChange({ ...sortBy, dir: sortBy.dir === 'desc' ? 'asc' : 'desc' })}
          aria-label={directionLabel}
          title={directionLabel}
          className="h-8 w-8 rounded-none p-0 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        >
          <DirectionIcon className="text-xs" />
        </Button>
      </div>

      {open && (
        <div
          role="menu"
          aria-label="Sort todos by"
          className="absolute top-full left-0 z-50 mt-1 w-44 rounded-lg border border-border bg-popover py-1 text-sm shadow-lg"
        >
          {TODO_SORT_COLUMN_OPTIONS.map(opt => {
            const selected = opt.value === sortBy.column;
            const OptIcon = opt.icon;
            return (
              <Button
                key={opt.value}
                variant="ghost"
                type="button"
                role="menuitemradio"
                aria-checked={selected}
                onClick={() => {
                  onChange({ ...sortBy, column: opt.value });
                  setOpen(false);
                }}
                className={`flex h-auto w-full items-center justify-start gap-2 px-3 py-1.5 text-left transition-colors ${
                  selected ? 'bg-primary/10 text-primary' : 'text-foreground hover:bg-muted'
                }`}
              >
                <OptIcon className="text-base" />
                <span className="flex-1">{opt.label}</span>
                {selected && <UiCheck className="text-xs" />}
              </Button>
            );
          })}
        </div>
      )}
    </div>
  );
}
