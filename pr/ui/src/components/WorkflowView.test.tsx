import type { ReactNode } from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { WorkflowRun } from '../types';
import { WorkflowRunView } from './WorkflowView';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, onClick, className }: { children: ReactNode; onClick: () => void; className?: string }) => (
    <button className={className} onClick={onClick}>{children}</button>
  ),
}));

vi.mock('@flanksource/clicky-ui/icons', () => ({
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

    const { container } = render(<WorkflowRunView run={run} repo="acme/widget" />);
    fireEvent.click(screen.getByText('failed build'));
    await waitFor(() => expect(container.querySelector('pre')).not.toBeNull());

    expect(fetchMock).toHaveBeenCalledWith('/api/prs/job-logs?repo=acme%2Fwidget&runId=11&jobId=22&tail=100');
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
