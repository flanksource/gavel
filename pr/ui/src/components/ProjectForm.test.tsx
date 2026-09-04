import type React from 'react';
import { act, render, renderHook, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { queryKeys } from '../query';
import { ProjectFields, useProjectRegistration, type ProjectRegistration } from './ProjectForm';
import { projectDiffQueryKey } from './projectMutations';

let queryClient: QueryClient;

vi.mock('@flanksource/clicky-ui/components', () => ({
  Field: ({ label, helper, children }: { label: React.ReactNode; helper?: React.ReactNode; children: React.ReactNode }) => (
    <div>
      <div>{label}</div>
      {helper && <div>{helper}</div>}
      {children}
    </div>
  ),
  Combobox: () => <div data-testid="repos-combobox" />,
}));

beforeEach(() => {
  queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false }, queries: { retry: false } } });
});

function wrapper({ children }: { children: React.ReactNode }) {
  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('ProjectForm PostgreSQL cutover', () => {
  it('initializes add defaults and persists normalized native project identity fields', async () => {
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => (
      { ok: true, text: async () => '' }
    ) as Response);
    vi.stubGlobal('fetch', fetchMock);
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
    const { result } = renderHook(() => useProjectRegistration({
      open: true,
      project: null,
      defaults: {
        name: 'widget',
        repos: ['acme/widget'],
      },
    }), { wrapper });

    expect(result.current.name).toBe('widget');
    expect(result.current.dir).toBe('');
    expect(result.current.repos).toEqual(['acme/widget']);

    act(() => {
      result.current.setName(' widget ');
      result.current.setDir(' /work/widget ');
    });

    let saved = false;
    await act(async () => { saved = await result.current.save(); });

    expect(saved).toBe(true);
    expect(fetchMock).toHaveBeenCalledOnce();
    const init = fetchMock.mock.calls[0]?.[1];
    expect(init).toBeDefined();
    expect(JSON.parse(String(init?.body))).toEqual({
      name: 'widget',
      dir: '/work/widget',
      repos: ['acme/widget'],
    });
    expect(invalidateQueries).toHaveBeenCalledOnce();
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: queryKeys.projects(), exact: true });
  });

  it('keeps an edited project authoritative over add defaults', () => {
    const project = {
      name: 'gavel',
      dir: '/work/gavel',
      repos: ['flanksource/gavel'],
    };
    const { result } = renderHook(() => useProjectRegistration({
      open: true,
      project,
      defaults: {
        name: 'ignored',
        dir: '/work/ignored',
        repos: ['acme/ignored'],
      },
    }), { wrapper });

    expect(result.current.name).toBe(project.name);
    expect(result.current.dir).toBe(project.dir);
    expect(result.current.repos).toEqual(project.repos);
    expect(result.current.editing).toBe(true);
  });

  it('updates a project and invalidates its catalog, status, and diff caches', async () => {
    const project = { name: 'gavel', dir: '/work/gavel', repos: ['flanksource/gavel'] };
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
    vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ ...project, dir: '/work/gavel-next' }), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    })));
    const { result } = renderHook(() => useProjectRegistration({ open: true, project }), { wrapper });

    act(() => result.current.setDir('/work/gavel-next'));
    let saved = false;
    await act(async () => { saved = await result.current.save(); });

    expect(saved).toBe(true);
    expect(fetch).toHaveBeenCalledWith('/api/projects/gavel', expect.objectContaining({ method: 'PUT' }));
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: queryKeys.projects(), exact: true });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: queryKeys.projectStatusScope('gavel') });
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: projectDiffQueryKey('gavel') });
    expect(invalidateQueries).toHaveBeenCalledTimes(3);
  });

  it('deletes a project, refreshes the catalog, and evicts its scoped caches', async () => {
    const project = { name: 'gavel', dir: '/work/gavel', repos: ['flanksource/gavel'] };
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries');
    const removeQueries = vi.spyOn(queryClient, 'removeQueries');
    vi.stubGlobal('confirm', vi.fn(() => true));
    vi.stubGlobal('fetch', vi.fn(async () => new Response(null, { status: 204 })));
    const { result } = renderHook(() => useProjectRegistration({ open: true, project }), { wrapper });

    let removed = false;
    await act(async () => { removed = await result.current.remove(); });

    expect(removed).toBe(true);
    expect(fetch).toHaveBeenCalledWith('/api/projects/gavel', { method: 'DELETE' });
    expect(invalidateQueries).toHaveBeenCalledOnce();
    expect(invalidateQueries).toHaveBeenCalledWith({ queryKey: queryKeys.projects(), exact: true });
    expect(removeQueries).toHaveBeenCalledWith({ queryKey: queryKeys.projectStatusScope('gavel') });
    expect(removeQueries).toHaveBeenCalledWith({ queryKey: projectDiffQueryKey('gavel') });
  });

  it('keeps failed project mutations visible with operation context', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('already registered', { status: 409 })));
    const { result } = renderHook(() => useProjectRegistration({
      open: true,
      project: null,
      defaults: { name: 'widget', dir: '/work/widget' },
    }), { wrapper });

    let saved = true;
    await act(async () => { saved = await result.current.save(); });

    expect(saved).toBe(false);
    expect(result.current.error).toBe('Failed to create project widget: already registered');
    render(<ProjectFields reg={result.current} repoOptions={[]} />);
    expect(screen.getByRole('alert').textContent).toBe('Failed to create project widget: already registered');
  });

  it('shows PostgreSQL as read-only persistence and offers no provider choices', () => {
    const noop = () => {};
    const reg: ProjectRegistration = {
      name: 'gavel',
      setName: noop,
      dir: '/work/gavel',
      setDir: noop,
      repos: [],
      setRepos: noop,
      error: '',
      saving: false,
      deleting: false,
      editing: true,
      save: async () => true,
      remove: async () => true,
    };

    render(<ProjectFields reg={reg} repoOptions={[]} />);

    expect(screen.getByLabelText('Todo persistence').textContent).toBe('PostgreSQL');
    expect(screen.queryByRole('combobox', { name: /provider/i })).toBeNull();
  });
});
