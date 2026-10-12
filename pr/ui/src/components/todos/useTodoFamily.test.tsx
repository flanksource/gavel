import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { TodoItem, TodoListResponse } from '../../types';
import { useParentTitles, useTodoFamily } from './useTodoFamily';

const DIR = '/repos/billing';
const PARENT_ID = '3c8e1f5a-7b2d-4960-a1e4-5d7f9b0c2e83';

function todo(ref: string, overrides: Partial<TodoItem> = {}): TodoItem {
  return { ref, id: `id-${ref}`, title: `Todo ${ref}`, status: 'pending', priority: 'medium', ...overrides };
}

const parent = todo('parent', { id: PARENT_ID });
const child = todo('child', { parentId: PARENT_ID });
const sibling = todo('sibling', { parentId: PARENT_ID });
const loner = todo('loner');
const byDir: Record<string, TodoListResponse> = {
  [DIR]: { dir: DIR, counts: {} as TodoListResponse['counts'], items: [parent, child, sibling, loner] },
  '/repos/other': { dir: '/repos/other', counts: {} as TodoListResponse['counts'], items: [todo('foreign')] },
};

type Listing = { listReady: boolean; error: string; errorsByDir: Record<string, { code: string; message: string }> };
const loaded: Listing = { listReady: true, error: '', errorsByDir: {} };

function family(detail: TodoItem | null, { select = vi.fn(), listing = loaded, lists = byDir }: {
  select?: ReturnType<typeof vi.fn>;
  listing?: Listing;
  lists?: Record<string, TodoListResponse>;
} = {}) {
  return renderHook(() => useTodoFamily({
    byDir: lists,
    selected: detail ? { dir: DIR, ref: detail.ref } : null,
    detail,
    select,
    ...listing,
  }));
}

describe('useTodoFamily', () => {
  it('gives a parent its children and every other open top-level todo of its workspace as candidates', () => {
    const { result } = family(parent);

    expect(result.current.childTodos.map(item => item.ref)).toEqual(['child', 'sibling']);
    expect(result.current.parentTodo).toBeNull();
    expect(result.current.parentCandidates.map(item => item.ref)).toEqual(['loner']);
    expect(result.current.familyStatus).toEqual({ state: 'ready' });
  });

  it('gives a child its parent, and only top-level todos as candidates', () => {
    const { result } = family(child);

    expect(result.current.childTodos).toEqual([]);
    expect(result.current.parentTodo).toBe(parent);
    expect(result.current.parentCandidates.map(item => item.ref)).toEqual(['parent', 'loner']);
  });

  it('is empty until a todo is open', () => {
    const { result } = family(null);

    expect(result.current).toMatchObject({ childTodos: [], parentTodo: null, parentCandidates: [] });
  });

  it('opens another todo of the selected workspace', () => {
    const select = vi.fn();
    const { result } = family(child, { select });

    result.current.onOpenTodo('parent');
    expect(select).toHaveBeenCalledExactlyOnceWith({ dir: DIR, ref: 'parent' });
  });

  describe('before the listing has arrived', () => {
    const placeholder = { [DIR]: { dir: DIR, counts: {} as TodoListResponse['counts'], items: [] } };

    it('reports loading rather than an empty family', () => {
      const { result } = family(parent, { listing: { listReady: false, error: '', errorsByDir: {} }, lists: placeholder });

      expect(result.current.familyStatus).toEqual({ state: 'loading' });
      expect(result.current.childTodos).toEqual([]);
    });

    it('reports the batch failure when nothing ever arrived', () => {
      const { result } = family(parent, { listing: { listReady: false, error: 'batch unreachable', errorsByDir: {} }, lists: placeholder });

      expect(result.current.familyStatus).toEqual({ state: 'error', message: 'batch unreachable' });
    });
  });

  it('reports the failure of this workspace\'s own list even though the batch answered', () => {
    const listing: Listing = {
      listReady: true,
      error: 'db down',
      errorsByDir: { [DIR]: { code: 'db', message: 'db down' }, '/repos/other': { code: 'x', message: 'elsewhere' } },
    };
    const { result } = family(parent, { listing });

    expect(result.current.familyStatus).toEqual({ state: 'error', message: 'db down' });
  });

  it('ignores a failure in a different workspace', () => {
    const listing: Listing = { listReady: true, error: 'elsewhere', errorsByDir: { '/repos/other': { code: 'x', message: 'elsewhere' } } };
    const { result } = family(parent, { listing });

    expect(result.current.familyStatus).toEqual({ state: 'ready' });
  });
});

describe('useParentTitles', () => {
  it('keeps one map across a refresh that changes no parent title', () => {
    const { result, rerender } = renderHook(({ lists }) => useParentTitles(lists), { initialProps: { lists: byDir } });
    const first = result.current;

    const refreshed = {
      ...byDir,
      [DIR]: { ...byDir[DIR], items: byDir[DIR].items.map(item => (item.ref === 'loner' ? { ...item, title: 'Retitled' } : { ...item })) },
    };
    rerender({ lists: refreshed });
    expect(result.current).toBe(first);
  });

  it('hands out a new map once a parent is retitled', () => {
    const { result, rerender } = renderHook(({ lists }) => useParentTitles(lists), { initialProps: { lists: byDir } });
    const first = result.current;

    rerender({ lists: { ...byDir, [DIR]: { ...byDir[DIR], items: byDir[DIR].items.map(item => (item.ref === 'parent' ? { ...item, title: 'Renamed' } : item)) } } });
    expect(result.current).not.toBe(first);
    expect(result.current.get(PARENT_ID)).toBe('Renamed');
  });
});
