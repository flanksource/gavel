import { describe, expect, it } from 'vitest';
import { selectionKey, selectionTargets, toggleSelectionKeys, setSelectionKeys, type SelectionState } from './todoSelection';

const alpha = { dir: '/repos/alpha', ref: 'todo-1' };
const beta = { dir: '/repos/beta', ref: 'todo-1' };

function state(...keys: string[]): SelectionState {
  return new Set(keys);
}

describe('selectionKey', () => {
  it('keeps same-ref todos in different workspaces apart', () => {
    expect(selectionKey(alpha)).not.toEqual(selectionKey(beta));
  });

  it('normalizes surrounding whitespace so a padded dir matches its trimmed form', () => {
    expect(selectionKey({ dir: `  ${alpha.dir}  `, ref: ` ${alpha.ref} ` })).toEqual(selectionKey(alpha));
  });
});

describe('toggleSelectionKeys', () => {
  it('adds a key that is absent', () => {
    expect(toggleSelectionKeys(state(), selectionKey(alpha))).toEqual(state(selectionKey(alpha)));
  });

  it('removes a key that is present, leaving the rest untouched', () => {
    const before = state(selectionKey(alpha), selectionKey(beta));
    expect(toggleSelectionKeys(before, selectionKey(alpha))).toEqual(state(selectionKey(beta)));
  });

  it('does not mutate the state it was given', () => {
    const before = state(selectionKey(alpha));
    toggleSelectionKeys(before, selectionKey(beta));
    expect(before).toEqual(state(selectionKey(alpha)));
  });
});

describe('setSelectionKeys', () => {
  it('adds every key when selecting a whole group', () => {
    expect(setSelectionKeys(state(), [selectionKey(alpha), selectionKey(beta)], true))
      .toEqual(state(selectionKey(alpha), selectionKey(beta)));
  });

  it('removes only the named keys when deselecting a group', () => {
    const before = state(selectionKey(alpha), selectionKey(beta));
    expect(setSelectionKeys(before, [selectionKey(alpha)], false)).toEqual(state(selectionKey(beta)));
  });

  it('is idempotent when the keys are already in the requested state', () => {
    const before = state(selectionKey(alpha));
    expect(setSelectionKeys(before, [selectionKey(alpha)], true)).toEqual(before);
  });
});

describe('selectionTargets', () => {
  it('round-trips keys back into the dir/ref pairs the bulk API takes', () => {
    const keys = state(selectionKey(alpha), selectionKey(beta));
    expect(selectionTargets(keys)).toEqual(expect.arrayContaining([alpha, beta]));
    expect(selectionTargets(keys)).toHaveLength(2);
  });

  it('preserves a todo with no workspace dir as an empty dir rather than dropping it', () => {
    expect(selectionTargets(state(selectionKey({ dir: '', ref: 'todo-9' })))).toEqual([{ dir: '', ref: 'todo-9' }]);
  });
});
