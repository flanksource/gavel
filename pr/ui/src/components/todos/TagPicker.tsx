import { useEffect, useMemo, useRef, useState } from 'react';

import { TodoTag } from './TodoTag';
import { normalizeTag, tagHash } from './tagPalette';
import type { ResolvedTag, TagIndex } from './tagResolve';
import type { TodoTagDef } from '../../types';

/**
 * TagPicker is the one "choose a tag" surface: a filter box over the labels
 * that already exist, ordered by how much this project actually uses them.
 *
 * Typing a name that nothing matches is not a dead end — the picker offers to
 * create it — but picking is the default path, which is the whole point.
 * Free-text tagging is how a backlog ends up with `flaky`, `Flaky` and `flakey`
 * as three separate colours; a list of what already exists, sorted by usage,
 * makes reusing the existing vocabulary the least effort.
 *
 * It is a plain absolutely-positioned panel rather than a clicky-ui
 * DropdownMenu: that one mounts floating-ui, which crashes under vitest and is
 * unusable inside list rows.
 */
export function TagPicker({
  candidates,
  counts = {},
  index,
  onPick,
  onClose,
  allowCreate = true,
  placeholder = 'Filter tags…',
  emptyLabel = 'No tags left to add.',
}: {
  /** The definitions to offer, already filtered to what makes sense here. */
  candidates: TodoTagDef[];
  /** Per-label todo counts, used only for ordering and the hint text. */
  counts?: Record<string, number>;
  index: TagIndex;
  onPick: (name: string) => void;
  onClose: () => void;
  /** When false, a name nothing matches cannot be created from here. */
  allowCreate?: boolean;
  placeholder?: string;
  emptyLabel?: string;
}) {
  const [query, setQuery] = useState('');
  const panel = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function onPointerDown(event: MouseEvent) {
      if (!panel.current?.contains(event.target as Node)) onClose();
    }
    document.addEventListener('mousedown', onPointerDown);
    return () => document.removeEventListener('mousedown', onPointerDown);
  }, [onClose]);

  // Most-used first, then alphabetical. A project's own vocabulary rises to the
  // top on its own, so "common" needs no curation to stay accurate.
  const matches = useMemo(() => {
    const needle = normalizeTag(query);
    return candidates
      .filter(def => !needle || def.name.includes(needle))
      .sort((a, b) => ((counts[b.name] ?? 0) - (counts[a.name] ?? 0)) || a.name.localeCompare(b.name));
  }, [candidates, counts, query]);

  const typed = normalizeTag(query);
  const canCreate = allowCreate && typed !== '' && !matches.some(def => def.name === typed);

  function commit(name: string) {
    const normalized = normalizeTag(name);
    if (!normalized) return;
    onPick(normalized);
    setQuery('');
  }

  // A draft tag has no definition yet, so preview it with the colour it would
  // actually get — the same hash the server would derive.
  const draft: ResolvedTag = { token: typed, value: typed, color: tagHash(typed), defined: false };

  return (
    <div
      ref={panel}
      role="dialog"
      aria-label="Choose a tag"
      className="absolute left-0 top-full z-30 mt-1 w-64 rounded-md border border-border bg-background p-1.5 shadow-lg"
    >
      <input
        autoFocus
        value={query}
        onChange={event => setQuery(event.target.value)}
        onKeyDown={event => {
          if (event.key === 'Escape') { event.stopPropagation(); onClose(); }
          if (event.key !== 'Enter') return;
          event.preventDefault();
          if (matches.length > 0) commit(matches[0].name);
          else if (canCreate) commit(typed);
        }}
        placeholder={placeholder}
        aria-label="Filter tags"
        className="mb-1 h-7 w-full rounded border border-border bg-background px-2 text-xs text-foreground"
      />

      <div role="listbox" aria-label="Tags" className="max-h-56 overflow-y-auto">
        {matches.map(def => (
          <button
            key={def.name}
            type="button"
            role="option"
            aria-selected={false}
            onClick={() => commit(def.name)}
            title={def.description}
            className="flex w-full items-center gap-2 rounded px-1.5 py-1 text-left hover:bg-muted"
          >
            <TodoTag tag={index.resolve(def.name)} />
            {(counts[def.name] ?? 0) > 0 && (
              <span className="ml-auto shrink-0 text-[10px] text-muted-foreground">
                {counts[def.name]}
              </span>
            )}
          </button>
        ))}

        {matches.length === 0 && !canCreate && (
          <p className="px-1.5 py-3 text-center text-[11px] text-muted-foreground">{emptyLabel}</p>
        )}
      </div>

      {canCreate && (
        <button
          type="button"
          onClick={() => commit(typed)}
          className="mt-1 flex w-full items-center gap-2 rounded border-t border-border px-1.5 py-1 pt-1.5 text-left hover:bg-muted"
        >
          <span className="shrink-0 text-[11px] text-muted-foreground">Create</span>
          <TodoTag tag={draft} />
        </button>
      )}
    </div>
  );
}
