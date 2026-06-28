import { useEffect, useRef, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import type { TodoGroupBy } from '../../types';
import { GavelIcon } from '../GavelIcon';
import { GROUP_BY_OPTIONS } from './todoGroup';

// TodoGroupByMenu is the tab-strip dropdown that switches how the todo lists are
// grouped: by Workspace (the default, the only mode that supports batch runs),
// Severity (priority), or Age (last activity). It drives the same shared
// group-by preference as the rest of the Todos chrome. The dashboard renders it
// beside the status filter pills, while the menubar renders it beside the compact
// filter dropdown.
export function TodoGroupByMenu({ groupBy, onChange }: {
  groupBy: TodoGroupBy;
  onChange: (groupBy: TodoGroupBy) => void;
}) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);

  // Close on outside click — the menu is anchored to its button, so a stray
  // click should dismiss rather than trap the user (mirrors OrgChooser).
  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener('mousedown', onDown);
    return () => document.removeEventListener('mousedown', onDown);
  }, [open]);

  const active = GROUP_BY_OPTIONS.find(opt => opt.value === groupBy) ?? GROUP_BY_OPTIONS[0];

  return (
    <div className="relative" ref={rootRef}>
      <Button
        variant="ghost"
        type="button"
        onClick={() => setOpen(o => !o)}
        aria-haspopup="menu"
        aria-expanded={open}
        title="Group todos by"
        className="inline-flex h-8 items-center gap-1.5 rounded-md border border-border px-2 text-xs text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
      >
        <GavelIcon name={active.icon} className="text-xs" />
        <span className="font-medium">Group: {active.label}</span>
        <GavelIcon name="codicon:chevron-down" className="text-[10px]" />
      </Button>

      {open && (
        <div
          role="menu"
          aria-label="Group todos by"
          className="absolute top-full left-0 z-50 mt-1 w-44 rounded-lg border border-border bg-popover py-1 text-sm shadow-lg"
        >
          {GROUP_BY_OPTIONS.map(opt => {
            const selected = opt.value === groupBy;
            return (
              <Button
                key={opt.value}
                variant="ghost"
                type="button"
                role="menuitemradio"
                aria-checked={selected}
                onClick={() => { onChange(opt.value); setOpen(false); }}
                className={`flex h-auto w-full items-center justify-start gap-2 px-3 py-1.5 text-left transition-colors ${
                  selected ? 'bg-primary/10 text-primary' : 'text-foreground hover:bg-muted'
                }`}
              >
                <GavelIcon name={opt.icon} className="text-base" />
                <span className="flex-1">{opt.label}</span>
                {selected && <GavelIcon name="codicon:check" className="text-xs" />}
              </Button>
            );
          })}
        </div>
      )}
    </div>
  );
}
