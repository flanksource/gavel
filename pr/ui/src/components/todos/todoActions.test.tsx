import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { ReactNode } from 'react';
import { useTodoSelectionActions } from './todoActions';
import { todoBulkActionURL, todoBulkResultMessage, type TodoBulkAction } from './todoEntity';
import { selectionKey } from './todoSelection';
import { buildTagIndex } from './tagResolve';

// The whole point of deriving the toolbar from the catalog is that a bulk
// action registered on the server appears in the UI without a React change.
// These fixtures are therefore the *server's* shape, not a UI-local invention.
const statusAction: TodoBulkAction = {
  name: 'status',
  short: 'Set the status of many TODOs',
  method: 'POST',
  path: '/api/v1/todo/{id}/status',
  supports_filter_mode: true,
  tool_hints: { icon: 'check-circle', group: 'Status' },
  param_schema: {
    type: 'object',
    properties: { to: { type: 'string', enum: ['pending', 'completed'] }, comment: { type: 'string' } },
    required: ['to'],
  },
};

const deleteAction: TodoBulkAction = {
  name: 'delete',
  short: 'Delete many TODOs',
  // Not a POST, and not a /{action} path: clicky infers both from the name.
  method: 'DELETE',
  path: '/api/v1/todo/{id}',
  tool_hints: { icon: 'trash', group: 'Danger', destructiveHint: true },
  param_schema: { type: 'object', properties: { confirm: { type: 'boolean' } }, required: ['confirm'] },
};

// Both parameters are optional arrays: a set toggled across the selection
// rather than one value assigned to it.
const labelsAction: TodoBulkAction = {
  name: 'labels',
  short: 'Add or remove labels across many TODOs',
  method: 'POST',
  path: '/api/v1/todo/{id}/labels',
  tool_hints: { icon: 'tag', group: 'Labels' },
  param_schema: {
    type: 'object',
    properties: {
      add: { type: 'array', items: { type: 'string' } },
      remove: { type: 'array', items: { type: 'string' } },
    },
  },
};

const triageAction: TodoBulkAction = {
  name: 'triage',
  short: 'Triage many TODOs',
  method: 'POST',
  path: '/api/v1/todo/{id}/triage',
  tool_hints: { icon: 'play', group: 'Run' },
};

function catalogResponse(actions: TodoBulkAction[]) {
  return {
    ok: true,
    status: 200,
    json: async () => [{ name: 'todo', bulk_actions: actions }],
  } as Response;
}

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

const ALPHA = selectionKey({ dir: '/repos/alpha', ref: 'todo-1' });
const BETA = selectionKey({ dir: '/repos/beta', ref: 'todo-2' });

describe('todoBulkActionURL', () => {
  // The route is the server's, not a convention rebuilt here — guessing POST
  // /{id}/delete would 404 on exactly the destructive action.
  it('substitutes ids into the published path and keeps the published method', () => {
    expect(todoBulkActionURL({ action: deleteAction, refs: ['a', 'b'] }))
      .toBe('/api/v1/todo/a,b');
    expect(todoBulkActionURL({ action: statusAction, refs: ['a'], params: { to: 'completed' } }))
      .toBe('/api/v1/todo/a/status?to=completed');
  });

  it('escapes each ref so a comma stays the separator', () => {
    expect(todoBulkActionURL({ action: statusAction, refs: ['a/b', 'c'] }))
      .toBe('/api/v1/todo/a%2Fb,c/status');
  });

  it('omits empty parameters rather than sending blanks the server must reject', () => {
    expect(todoBulkActionURL({ action: statusAction, refs: ['a'], params: { to: 'completed', comment: '' } }))
      .toBe('/api/v1/todo/a/status?to=completed');
  });
});

describe('todoBulkResultMessage', () => {
  it('reports a clean batch as a count', () => {
    expect(todoBulkResultMessage({ action: 'status', applied: 3, failed: 0, results: [] }))
      .toBe('Updated 3 todos');
  });

  // A partial batch is a 200: the failures are per-item, and dropping them
  // would report a success that silently lost work.
  it('names every failure so a partial batch is not read as a success', () => {
    const message = todoBulkResultMessage({
      action: 'status',
      applied: 1,
      failed: 1,
      results: [
        { ref: 'todo-1', title: 'Kept' },
        { ref: 'todo-2', title: 'Gone', error: 'archived in another tab' },
      ],
    });
    expect(message).toBe('Updated 1 todo; 1 failed: Gone (archived in another tab)');
  });

  it('says runs were started rather than finished', () => {
    expect(todoBulkResultMessage({ action: 'triage', applied: 2, failed: 0, results: [] }, 'Started'))
      .toBe('Started 2 todos');
  });
});

describe('useTodoSelectionActions', () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      if (String(input) === '/api/entities') {
        return catalogResponse([statusAction, deleteAction, triageAction]);
      }
      return {
        ok: true,
        status: 200,
        json: async () => ({ action: 'status', applied: 1, failed: 0, results: [] }),
      } as Response;
    });
    vi.stubGlobal('fetch', fetchMock);
  });

  afterEach(() => vi.unstubAllGlobals());

  async function actions(
    selection: Set<string>,
    catalog?: TodoBulkAction[],
    extra?: Partial<Parameters<typeof useTodoSelectionActions>[0]>,
  ) {
    if (catalog) {
      fetchMock.mockImplementation(async (input: RequestInfo | URL) =>
        String(input) === '/api/entities'
          ? catalogResponse(catalog)
          : ({ ok: true, status: 200, json: async () => ({}) } as Response),
      );
    }
    const { result } = renderHook(
      () => useTodoSelectionActions({ selection, ...extra }),
      { wrapper },
    );
    await waitFor(() => expect(result.current.length).toBeGreaterThan(0));
    return result;
  }

  // The drift this replaced: the dashboard hardcoded three of the actions the
  // server had. Every published action must now become a descriptor.
  it('renders one descriptor per published action', async () => {
    const result = await actions(new Set([ALPHA]));
    expect(result.current.map(action => action.id)).toEqual(
      expect.arrayContaining(['status', 'delete', 'triage']),
    );
  });

  it('orders the destructive action last so a click never lands on it by habit', async () => {
    const result = await actions(new Set([ALPHA]));
    expect(result.current[result.current.length - 1].id).toBe('delete');
  });

  // A required parameter with a closed set of values becomes a submenu of those
  // values, so picking a status is one gesture rather than a form.
  it('turns a required enum parameter into a submenu of its values', async () => {
    const result = await actions(new Set([ALPHA]));
    const status = result.current.find(action => action.id === 'status');
    expect(status?.children?.map(child => child.id)).toEqual(['status:pending', 'status:completed']);
  });

  it('leaves an action with no closed-set parameter as a single button', async () => {
    const result = await actions(new Set([ALPHA]));
    expect(result.current.find(action => action.id === 'triage')?.children).toBeUndefined();
  });

  // A field with a closed set of values is what a bulk editor edits, so it goes
  // on the bar as its own named dropdown. Deriving this from the parameter
  // shape rather than a list kept here is what lets a bulk action added to the
  // Go registry reach the bar without touching React.
  it('puts a field with a closed set of values on the bar as a dropdown', async () => {
    const result = await actions(new Set([ALPHA]));
    expect(result.current.find(action => action.id === 'status')?.display).toBe('menu');
  });

  it('puts a set of values toggled across the selection on the bar too', async () => {
    const result = await actions(new Set([ALPHA]), [labelsAction], {
      tags: buildTagIndex([{ name: 'bug', color: 'red', scope: 'builtin' }]),
    });
    expect(result.current.find(action => action.id === 'labels')?.display).toBe('menu');
    expect(result.current.find(action => action.id === 'labels')?.menu).toBeTypeOf('function');
  });

  // A dropdown that opened on nothing would be worse than one more click.
  it('collapses the set editor when there is no taxonomy to offer', async () => {
    const result = await actions(new Set([ALPHA]), [labelsAction]);
    expect(result.current.find(action => action.id === 'labels')?.display).toBe('overflow');
  });

  // The bar is a row of controls the reader scans, so a trigger reads "Status",
  // not "Set the status of many TODOs". The sentence stays as the tooltip.
  it('names a bar control after the field, not the catalog sentence', async () => {
    const result = await actions(new Set([ALPHA]));
    const status = result.current.find(action => action.id === 'status');
    expect(status?.label).toBe('Status');
    expect(status?.description).toBe('Set the status of many TODOs');
  });

  it('collapses a command into the overflow rather than crowding the row', async () => {
    const result = await actions(new Set([ALPHA]));
    for (const id of ['triage', 'delete']) {
      expect(result.current.find(action => action.id === id)?.display).toBe('overflow');
    }
  });

  // The destructive action asks first, and the prompt carries the count — the
  // one number that makes stopping worth it when the selection was matched by
  // a filter rather than enumerated.
  it('gates the destructive action behind a confirmation naming the count', async () => {
    const result = await actions(new Set([ALPHA, BETA]));
    const remove = result.current.find(action => action.id === 'delete');
    const confirm = remove?.confirm;
    expect(typeof confirm === 'object' && typeof confirm.message === 'function').toBe(true);
    if (typeof confirm !== 'object' || typeof confirm.message !== 'function') return;
    expect(confirm.message({ selectedRowIds: [ALPHA, BETA], selectedRows: [], clearSelection: () => {} }))
      .toContain('2 todos');
  });

  // Selection keys carry dir so a row is identifiable across workspaces, but
  // the entity id is the bare ref — a dir cannot survive a URL path segment.
  it('dispatches bare refs, never the dir-qualified selection key', async () => {
    const result = await actions(new Set([ALPHA, BETA]));
    const status = result.current.find(action => action.id === 'status');
    await status!.children![1].onSelect({ selectedRowIds: [], selectedRows: [], clearSelection: () => {} });

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/todo/todo-1,todo-2/status?to=completed',
      expect.objectContaining({ method: 'POST' }),
    ));
  });

  it('sends the confirmation the server refuses to delete without', async () => {
    const result = await actions(new Set([ALPHA]));
    await result.current.find(action => action.id === 'delete')!
      .onSelect({ selectedRowIds: [], selectedRows: [], clearSelection: () => {} });

    await waitFor(() => expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/todo/todo-1?confirm=true',
      expect.objectContaining({ method: 'DELETE' }),
    ));
  });

  it('disables every action while nothing is checked', async () => {
    const { result } = renderHook(() => useTodoSelectionActions({ selection: new Set<string>() }), { wrapper });
    await waitFor(() => expect(result.current.length).toBeGreaterThan(0));
    expect(result.current.every(action => action.disabled)).toBe(true);
  });

  // An empty catalog means the entity failed to register. Rendering an empty
  // toolbar would hide that; the query fails instead.
  it('surfaces a server with no todo entity as an error rather than an empty toolbar', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({
      ok: true, status: 200, json: async () => [{ name: 'project' }],
    } as Response)));
    const { result } = renderHook(() => useTodoSelectionActions({ selection: new Set([ALPHA]) }), { wrapper });
    await waitFor(() => expect(result.current).toEqual([]));
  });
});
