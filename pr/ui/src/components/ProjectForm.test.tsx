import type React from 'react';
import { act, render, renderHook, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ProjectFields, useProjectRegistration, type ProjectRegistration } from './ProjectForm';

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
    const { result } = renderHook(() => useProjectRegistration({
      open: true,
      project: null,
      defaults: {
        name: 'widget',
        repos: ['acme/widget'],
      },
    }));

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
    }));

    expect(result.current.name).toBe(project.name);
    expect(result.current.dir).toBe(project.dir);
    expect(result.current.repos).toEqual(project.repos);
    expect(result.current.editing).toBe(true);
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
