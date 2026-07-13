import type React from 'react';
import { act, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem } from '../../types';
import { TodoPlan } from './TodoPlan';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children }: { children: React.ReactNode }) => (
    // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky Button itself.
    <button type="button">{children}</button>
  ),
}));

const todo: TodoItem = {
  ref: 'native-id',
  version: 3,
  title: 'Native issue',
  status: 'pending',
  priority: 'medium',
};

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('TodoPlan', () => {
  it('renders the actionable database error when plan loading fails', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => ({
      ok: false,
      status: 503,
      json: async () => ({ error: 'native TODO storage requires PostgreSQL' }),
    }) as Response));

    render(<TodoPlan dir="/repo" todo={todo} active />);
    await act(async () => {});

    expect(screen.getByRole('alert').textContent).toContain('native TODO storage requires PostgreSQL');
    expect(screen.queryByText('No plan yet. Run this todo in Plan mode to produce one.')).toBeNull();
  });
});
