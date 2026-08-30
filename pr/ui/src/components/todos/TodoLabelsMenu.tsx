import { useMemo, useState } from 'react';
import { Button } from '@flanksource/clicky-ui/components';

import type { TodoItem } from '../../types';
import { TodoTag } from './TodoTag';
import { normalizeTag } from './tagPalette';
import { tagCandidates, todoVisibleLabels, type TagIndex } from './tagResolve';

/**
 * How a label sits across a selection: on every todo, on some of them, or on
 * none. "mixed" is the state a bulk editor exists to resolve, and the reason
 * this cannot be a plain checkbox list — a two-state control would have to lie
 * about the half of the selection it does not describe.
 */
export type LabelState = 'on' | 'mixed' | 'off';

/** The net edit, not a replay of clicks: a label cycled back to where it
 *  started appears in neither list. */
export interface LabelEdit {
  add: string[];
  remove: string[];
}

/**
 * labelStates reports each label's standing across the selected todos.
 *
 * `selected` is the todos the ids actually resolved to. A filter-scoped
 * selection resolves to fewer todos than it covers, which is why the caller
 * tells us whether the sample is complete rather than us inferring it from a
 * count.
 */
export function labelStates(selected: TodoItem[]): Map<string, LabelState> {
  const counts = new Map<string, number>();
  for (const todo of selected) {
    for (const label of new Set(todoVisibleLabels(todo).map(normalizeTag))) {
      counts.set(label, (counts.get(label) ?? 0) + 1);
    }
  }
  const states = new Map<string, LabelState>();
  for (const [label, count] of counts) {
    states.set(label, count === selected.length ? 'on' : 'mixed');
  }
  return states;
}

/** The next state in the cycle. A mixed label gets three stops so "leave these
 *  as they are" stays reachable without closing the menu. */
function cycle(current: LabelState, baseline: LabelState): LabelState {
  if (current === 'on') return 'off';
  if (current === 'off') return baseline === 'mixed' ? 'mixed' : 'on';
  return 'on';
}

/** The net of a whole editing session against what the labels started as. */
export function labelEdit(
  baseline: Map<string, LabelState>,
  draft: Map<string, LabelState>,
): LabelEdit {
  const add: string[] = [];
  const remove: string[] = [];
  for (const [label, state] of draft) {
    const was = baseline.get(label) ?? 'off';
    if (state === was) continue;
    if (state === 'on') add.push(label);
    if (state === 'off') remove.push(label);
  }
  return { add, remove };
}

/**
 * The body of the bar's "Labels ▾" dropdown: every label this workspace knows,
 * each showing whether it is on all, some, or none of the selection.
 *
 * Edits accumulate and apply once, when the menu closes. Applying per click
 * would fire a bulk write per keystroke of intent and leave the list fighting
 * its own refetches.
 */
export function TodoLabelsMenu({
  index,
  counts,
  selected,
  complete,
  busy,
  onApply,
}: {
  index: TagIndex;
  counts: Record<string, number>;
  /** The todos the selection resolved to. */
  selected: TodoItem[];
  /** False when the selection covers todos not in `selected` — a filter scope. */
  complete: boolean;
  busy: boolean;
  onApply: (edit: LabelEdit) => void;
}) {
  const [query, setQuery] = useState('');
  const baseline = useMemo(
    // Without the full set of todos, nothing can be claimed about agreement, so
    // every label starts off and the menu becomes plain add/remove.
    () => (complete ? labelStates(selected) : new Map<string, LabelState>()),
    [complete, selected],
  );
  const [draft, setDraft] = useState<Map<string, LabelState>>(() => new Map(baseline));

  const options = useMemo(() => {
    const needle = normalizeTag(query);
    return tagCandidates(index, counts)
      .filter(def => !needle || def.name.includes(needle))
      .sort((a, b) => {
        const used = (counts[b.name] ?? 0) - (counts[a.name] ?? 0);
        return used || a.name.localeCompare(b.name);
      });
  }, [index, counts, query]);

  const edit = labelEdit(baseline, draft);
  const dirty = edit.add.length > 0 || edit.remove.length > 0;

  return (
    <div className="w-64 p-1.5">
      <input
        autoFocus
        value={query}
        onChange={event => setQuery(event.target.value)}
        placeholder="Filter labels…"
        aria-label="Filter labels"
        className="mb-1 h-7 w-full rounded border border-border bg-background px-2 text-xs text-foreground"
      />

      <div role="group" aria-label="Labels" className="max-h-56 overflow-y-auto">
        {options.map(def => {
          const state = draft.get(def.name) ?? 'off';
          return (
            <label
              key={def.name}
              className="flex cursor-pointer items-center gap-2 rounded px-1.5 py-1 hover:bg-muted"
            >
              <input
                type="checkbox"
                disabled={busy}
                checked={state === 'on'}
                ref={node => {
                  // `indeterminate` is a DOM property with no attribute, so it
                  // has to be written rather than rendered.
                  if (node) node.indeterminate = state === 'mixed';
                }}
                onChange={() =>
                  setDraft(current => {
                    const next = new Map(current);
                    next.set(def.name, cycle(state, baseline.get(def.name) ?? 'off'));
                    return next;
                  })
                }
                className="size-3.5 rounded border-border accent-primary"
              />
              <TodoTag tag={index.resolve(def.name)} />
              {(counts[def.name] ?? 0) > 0 && (
                <span className="ml-auto shrink-0 text-[10px] text-muted-foreground">
                  {counts[def.name]}
                </span>
              )}
            </label>
          );
        })}

        {options.length === 0 && (
          <p className="px-1.5 py-3 text-center text-[11px] text-muted-foreground">
            No labels match.
          </p>
        )}
      </div>

      <div className="mt-1 flex items-center justify-between border-t border-border pt-1.5">
        <span className="text-[11px] text-muted-foreground">
          {dirty
            ? `${edit.add.length} to add · ${edit.remove.length} to remove`
            : 'No changes'}
        </span>
        <Button size="sm" disabled={!dirty || busy} onClick={() => onApply(edit)}>
          Apply
        </Button>
      </div>
    </div>
  );
}
