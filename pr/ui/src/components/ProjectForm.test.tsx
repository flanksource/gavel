import type React from 'react';
import { act, render, renderHook, screen, waitFor } from '@testing-library/react';
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
  it('does not persist the retained legacy todoProvider field', async () => {
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => (
      { ok: true, text: async () => '' }
    ) as Response);
    vi.stubGlobal('fetch', fetchMock);
    const { result } = renderHook(() => useProjectRegistration(true, null));

    act(() => {
      result.current.setName('gavel');
      result.current.setDir('/work/gavel');
      result.current.setRepos(['flanksource/gavel']);
      result.current.setTodoProvider('grite');
    });
    await waitFor(() => expect(result.current.todoProvider).toBe('grite'));

    let saved = false;
    await act(async () => { saved = await result.current.save(); });

    expect(saved).toBe(true);
    expect(fetchMock).toHaveBeenCalledOnce();
    const init = fetchMock.mock.calls[0]?.[1];
    expect(init).toBeDefined();
    expect(JSON.parse(String(init?.body))).toEqual({
      name: 'gavel',
      dir: '/work/gavel',
      repos: ['flanksource/gavel'],
    });
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
      todoProvider: 'grite',
      setTodoProvider: noop,
      error: '',
      saving: false,
      deleting: false,
      editing: true,
      save: async () => true,
      remove: async () => true,
    };

    render(<ProjectFields reg={reg} repoOptions={[]} />);

    expect(screen.getByLabelText('Todo persistence').textContent).toBe('PostgreSQL');
    expect(screen.queryByText('Grite')).toBeNull();
    expect(screen.queryByText('.todos files')).toBeNull();
  });
});
