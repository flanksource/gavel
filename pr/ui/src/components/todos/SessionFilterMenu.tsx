import { useEffect, useRef, useState } from 'react';
import { GavelIcon } from '../GavelIcon';
import type { SessionCategory, SessionFacets } from './sessionFilter';

// SessionFilterMenu is the session browser's 3-dot overflow menu: it folds the
// category and tool toggles into one popover so the transcript chrome stays
// tight. Only categories/tools present in the stream are offered (see
// computeFacets); a badge surfaces how many toggles are currently hiding events.
export function SessionFilterMenu({
  facets,
  hiddenCategories,
  hiddenTools,
  onToggleCategory,
  onToggleTool,
}: {
  facets: SessionFacets;
  hiddenCategories: Set<SessionCategory>;
  hiddenTools: Set<string>;
  onToggleCategory: (key: SessionCategory) => void;
  onToggleTool: (key: string) => void;
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

  const hiddenCount =
    facets.categories.filter(c => hiddenCategories.has(c.def.key)).length +
    facets.tools.filter(t => hiddenTools.has(t.key)).length;

  return (
    <div className="relative ml-auto" ref={rootRef}>
      <button
        type="button"
        onClick={() => setOpen(o => !o)}
        aria-haspopup="menu"
        aria-expanded={open}
        title="Filter session events"
        className="inline-flex items-center gap-1 rounded px-1 py-0.5 text-gray-500 hover:bg-zinc-800 hover:text-gray-200 transition-colors"
      >
        {hiddenCount > 0 && (
          <span className="rounded-full bg-zinc-800 px-1 text-[9px] tabular-nums text-gray-400">{hiddenCount}</span>
        )}
        <GavelIcon name="codicon:kebab-vertical" className="text-[13px]" />
      </button>

      {open && (
        <div
          role="menu"
          aria-label="Filter session events"
          className="absolute right-0 top-full z-50 mt-1 w-52 rounded-lg border border-zinc-700 bg-zinc-900 py-1 text-[11px] shadow-lg"
        >
          <MenuSection label="Categories" />
          {facets.categories.map(c => (
            <MenuItem
              key={c.def.key}
              label={c.def.label}
              count={c.count}
              active={!hiddenCategories.has(c.def.key)}
              onClick={() => onToggleCategory(c.def.key)}
            />
          ))}
          {facets.tools.length > 0 && (
            <>
              <MenuSection label="Tools" />
              {facets.tools.map(t => (
                <MenuItem
                  key={t.key}
                  label={t.key}
                  count={t.count}
                  active={!hiddenTools.has(t.key)}
                  onClick={() => onToggleTool(t.key)}
                />
              ))}
            </>
          )}
        </div>
      )}
    </div>
  );
}

function MenuSection({ label }: { label: string }) {
  return (
    <div className="px-3 pb-0.5 pt-1.5 text-[9px] font-semibold uppercase tracking-wide text-gray-600">{label}</div>
  );
}

function MenuItem({ label, count, active, onClick }: {
  label: string;
  count: number;
  active: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      role="menuitemcheckbox"
      aria-checked={active}
      onClick={onClick}
      className="flex w-full items-center gap-2 px-3 py-1 text-left transition-colors hover:bg-zinc-800"
    >
      <GavelIcon
        name={active ? 'codicon:check' : 'codicon:circle-large-outline'}
        className={`text-[11px] ${active ? 'text-emerald-400' : 'text-gray-600'}`}
      />
      <span className={`flex-1 truncate ${active ? 'text-gray-200' : 'text-gray-500 line-through opacity-60'}`}>{label}</span>
      <span className="tabular-nums text-[10px] text-gray-500">{count}</span>
    </button>
  );
}
