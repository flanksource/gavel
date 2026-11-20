import type { ReactNode } from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { WorkflowRun } from '../types';
import { WorkflowRunView } from './WorkflowView';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, onClick, className }: { children: ReactNode; onClick: () => void; className?: string }) => (
    <button className={className} onClick={onClick}>{children}</button>
  ),
}));

vi.mock('@flanksource/clicky-ui/icons', async importOriginal => ({
  ...await importOriginal<typeof import('@flanksource/clicky-ui/icons')>(),
  UiChevronDown: () => <span>down</span>,
  UiChevronRight: () => <span>right</span>,
  UiLinkExternal: () => <span>external</span>,
}));

vi.mock('../icons/Spinner', () => ({ Spinner: () => <span>loading</span> }));

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('WorkflowRunView job log fallback', () => {
  it('requests the 100-line tail and renders ANSI in a neutral shell while preserving failure status', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        jobId: 22,
        logs: '\x1b[31mtail failure\x1b[0m\nplain detail',
      }),
    });
    vi.stubGlobal('fetch', fetchMock);
    const run: WorkflowRun = {
      databaseId: 11,
      name: 'CI',
      status: 'completed',
      conclusion: 'failure',
      jobs: [{
        databaseId: 22,
        name: 'failed build',
        status: 'completed',
        conclusion: 'failure',
      }],
    };

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const { container } = render(
      <QueryClientProvider client={queryClient}>
        <WorkflowRunView run={run} repo="acme/widget" />
      </QueryClientProvider>,
    );
    fireEvent.click(screen.getByText('failed build'));
    await waitFor(() => expect(container.querySelector('pre')).not.toBeNull());

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/prs/job-logs?repo=acme%2Fwidget&runId=11&jobId=22&tail=100',
      { signal: expect.any(AbortSignal) },
    );
    const pre = container.querySelector('pre');
    expect(pre).not.toBeNull();
    expect(pre!.className).toContain('bg-muted');
    expect(pre!.className).toContain('border-border');
    expect(pre!.className).not.toMatch(/bg-red-|border-red-/);
    expect(Array.from(pre!.querySelectorAll('span')).some(span => (span as HTMLElement).style.color !== '')).toBe(true);
    expect(screen.getByText('failed build').className).toContain('text-red-700');
    expect(screen.getAllByText('✗').some(icon => icon.className.includes('text-red-600'))).toBe(true);
  });
});

describe('WorkflowRunView Azure DevOps jobs', () => {
  it('shows the log tails a job carries without fetching, keeping jobs apart by their own IDs', async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal('fetch', fetchMock);
    const run: WorkflowRun = {
      databaseId: 31,
      name: 'ci',
      status: 'COMPLETED',
      conclusion: 'FAILURE',
      jobs: [
        {
          databaseId: 0, nativeId: 'job-1', name: 'build', status: 'COMPLETED', conclusion: 'FAILURE',
          steps: [{ name: 'compile', number: 2, status: 'COMPLETED', conclusion: 'FAILURE', logs: 'undefined: rule' }],
        },
        { databaseId: 0, nativeId: 'job-2', name: 'package', status: 'COMPLETED', conclusion: 'FAILURE', logs: 'no space left' },
      ],
    };

    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={queryClient}>
        <WorkflowRunView run={run} repo="acme/Platform/rules" />
      </QueryClientProvider>,
    );
    fireEvent.click(screen.getByText('compile'));
    fireEvent.click(screen.getByText('package'));

    await waitFor(() => expect(screen.getByText('undefined: rule')).toBeTruthy());
    expect(screen.getByText('no space left')).toBeTruthy();
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
