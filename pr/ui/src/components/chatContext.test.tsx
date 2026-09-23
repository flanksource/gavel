import { fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { OpenAPISpec, OperationsApiClient } from '@flanksource/clicky-ui/rpc';
import { describe, expect, it, vi } from 'vitest';
import {
  chatContextSurfaceFilter,
  chatContextTypeConfig,
  createGavelEntityClient,
  GAVEL_ENTITY_OPENAPI_PATH,
  GavelContextPicker,
} from './chatContext';

const listOperation = (surface: string) => ({
  operationId: surface,
  'x-clicky': { surface, verb: 'list' as const, scope: 'collection' as const },
  responses: {},
});

// The generated document carries every registered entity; only TODOs and
// projects are things a chat thread is about.
const entitySpec: OpenAPISpec = {
  openapi: '3.0.0',
  info: { title: 'gavel', version: '1' },
  'x-clicky': {
    surfaces: [
      { key: 'todo', entity: 'todo', title: 'todo' },
      { key: 'project', entity: 'project', title: 'project' },
      { key: 'label', entity: 'label', title: 'Labels' },
    ],
  },
  paths: {
    '/api/v1/todo': { get: listOperation('todo') },
    '/api/v1/project': { get: listOperation('project') },
    '/api/v1/label': { get: listOperation('label') },
  },
};

function renderPicker(client: OperationsApiClient) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={queryClient}>
      <GavelContextPicker client={client} items={[]} onAdd={vi.fn()} onAddMany={vi.fn()} />
    </QueryClientProvider>,
  );
}

describe('chat context surfaces', () => {
  it.each([
    ['todo', true],
    ['project', true],
    ['label', false],
    ['toString', false],
  ])('attaching %s is allowed: %s', (key, allowed) => {
    expect(chatContextSurfaceFilter({ key, entity: key, title: key }, {} as never)).toBe(allowed);
  });

  it('gives each attachable surface an icon for its context chips', () => {
    expect(Object.keys(chatContextTypeConfig).sort()).toEqual(['project', 'todo']);
    expect(chatContextTypeConfig.todo?.icon).toBeDefined();
    expect(chatContextTypeConfig.project?.icon).toBeDefined();
  });
});

describe('createGavelEntityClient', () => {
  it('reads the generated entity document rather than the hand-built projects one', async () => {
    const fetchSpy = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => new Response(JSON.stringify(entitySpec), {
      headers: { 'Content-Type': 'application/json' },
    }));

    await createGavelEntityClient(fetchSpy as typeof fetch).getOpenAPISpec();

    expect(fetchSpy).toHaveBeenCalledTimes(1);
    expect(fetchSpy.mock.calls[0]?.[0]).toBe(GAVEL_ENTITY_OPENAPI_PATH);
  });
});

describe('GavelContextPicker', () => {
  it('offers TODOs and projects, and nothing else the document lists', async () => {
    renderPicker({ getOpenAPISpec: async () => entitySpec, executeCommand: vi.fn() });

    fireEvent.click(screen.getByRole('button', { name: 'Add context' }));

    expect(await screen.findByRole('menuitem', { name: 'TODOs' })).toBeTruthy();
    expect(screen.getByRole('menuitem', { name: 'Projects' })).toBeTruthy();
    expect(screen.queryByRole('menuitem', { name: 'Labels' })).toBeNull();
  });
});
