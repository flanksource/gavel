import type React from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { CreateTodoFromRunDialog } from './CreateTodoFromRunDialog';
import { queryTestWrapper } from '../todos/queryTestWrapper';
import type { RunFailureCandidate } from './RunFailureCandidates';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, loading: _loading, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { loading?: boolean }) => <button {...props}>{children}</button>,
  Field: ({ label, children }: { label: string; children: React.ReactNode }) => <label>{label}{children}</label>,
  Modal: ({ open, title, children, footer }: { open: boolean; title: string; children: React.ReactNode; footer?: React.ReactNode }) => open ? <section aria-label={title}>{children}{footer}</section> : null,
}));

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

const candidates: RunFailureCandidate[] = [
  {
    key: 'test:0',
    group: 'tests',
    title: 'TestSave',
    location: 'save_test.go:31',
    criterion: 'Test `TestSave` passes at `save_test.go:31`.',
    detail: 'expected success',
  },
  {
    key: 'linter:0',
    group: 'lint',
    title: 'golangci-lint failed',
    criterion: 'Linter `golangci-lint` completes successfully.',
    detail: 'invalid config',
  },
];

describe('CreateTodoFromRunDialog', () => {
  it('creates one project-scoped todo from the selected failures', async () => {
    const onCreated = vi.fn();
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, _init?: RequestInit) => ({
      ok: true,
      json: async () => ({
        todo: {
          ref: 'todo/123',
          title: 'Fix test and lint failures in gavel',
          status: 'pending',
          priority: 'medium',
          criteria: [{ text: candidates[0].criterion }],
        },
      }),
    }) as Response);
    vi.stubGlobal('fetch', fetchMock);

    render(
      <CreateTodoFromRunDialog
        open
        projectName="gavel"
        projectDir="/work/gavel"
        runId="run-123"
        candidates={candidates}
        onClose={vi.fn()}
        onCreated={onCreated}
      />,
      { wrapper: queryTestWrapper() },
    );

    expect((screen.getByLabelText('Include TestSave') as HTMLInputElement).checked).toBe(true);
    expect((screen.getByLabelText('Include golangci-lint failed') as HTMLInputElement).checked).toBe(true);
    fireEvent.click(screen.getByLabelText('Include golangci-lint failed'));
    fireEvent.change(screen.getByLabelText('Notes'), { target: { value: 'Keep the public API stable.' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add todo' }));

    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
    expect(fetchMock.mock.calls[0]?.[0]).toBe('/api/todos/new?dir=%2Fwork%2Fgavel');
    expect(JSON.parse(String(fetchMock.mock.calls[0]?.[1]?.body))).toEqual({
      title: 'Fix test and lint failures in gavel',
      body: 'Keep the public API stable.\n\n_From [Project run `run-123`](/projects/gavel/runs/run-123)._',
      priority: 'medium',
      status: 'pending',
      criteria: [{ text: 'Test `TestSave` passes at `save_test.go:31`.' }],
    });
    expect(onCreated).toHaveBeenCalledTimes(1);
    expect((await screen.findByRole('link', { name: 'Open todo' })).getAttribute('href')).toBe('/todos/todo/123');
  });

  it('surfaces the server error without reporting a created todo', async () => {
    const onCreated = vi.fn();
    vi.stubGlobal('fetch', vi.fn(async () => ({
      ok: false,
      status: 500,
      json: async () => ({ error: 'native todo provider is unavailable' }),
    }) as Response));

    render(
      <CreateTodoFromRunDialog
        open
        projectName="gavel"
        projectDir="/work/gavel"
        runId="run-123"
        candidates={candidates}
        onClose={vi.fn()}
        onCreated={onCreated}
      />,
      { wrapper: queryTestWrapper() },
    );
    fireEvent.click(screen.getByRole('button', { name: 'Add todo' }));

    expect((await screen.findByRole('alert')).textContent).toContain('native todo provider is unavailable');
    expect(onCreated).not.toHaveBeenCalled();
  });
});
