import type React from 'react';
import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ProjectStatusView } from './ProjectStatusView';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement>) => <button {...props}>{children}</button>,
  SplitButton: ({ label, onClick, items, disabled }: {
    label: ReactNode;
    onClick: () => void;
    items: { label: ReactNode; onSelect?: () => void }[];
    disabled?: boolean;
  }) => (
    <div>
      <button type="button" onClick={onClick} disabled={disabled}>{label}</button>
      {items.map((item, index) => <button key={index} type="button" onClick={item.onSelect} disabled={disabled}>{item.label}</button>)}
    </div>
  ),
  SplitPane: ({ left, right }: { left: ReactNode; right: ReactNode }) => <div>{left}{right}</div>,
}));

vi.mock('./ProjectActionDialog', () => ({ ProjectActionDialog: () => null }));
vi.mock('./ProjectActionRunDialog', () => ({ ProjectActionRunDialog: () => null }));
vi.mock('./ProjectCommitTasks', () => ({ ProjectCommitTasks: () => null }));
vi.mock('./ProjectDiffView', () => ({ ProjectDiffView: () => null }));
vi.mock('./ProjectFileTree', () => ({
  ProjectFileTree: ({ files, selected, onToggleFile }: {
    files: { path: string }[];
    selected: Set<string>;
    onToggleFile: (path: string) => void;
  }) => <>{files.map(file => <input key={file.path} type="checkbox" aria-label={`Select ${file.path}`} checked={selected.has(file.path)} onChange={() => onToggleFile(file.path)} />)}</>,
}));

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('Project Open PR action', () => {
  it('queues the selected files from the Commit split button', async () => {
    const project = { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] };
    const status = {
      project,
      workDir: project.dir,
      branch: 'feature/project-actions',
      files: [{
        path: 'one.go',
        state: 'unstaged',
        adds: 1,
        dels: 0,
        testStatus: { passed: 1, failed: 0, skipped: 0 },
        lintStatus: { errors: 0, warnings: 0, infos: 0 },
        resultsStale: false,
      }],
      resultsStale: false,
      action: { running: false },
    };
    const fetchMock = vi.fn(async (_input: RequestInfo | URL, init?: RequestInit) => (
      new Response(JSON.stringify(
        init?.method === 'POST' ? { runId: 'pr-run-1' } : status,
      ), {
        headers: { 'content-type': 'application/json' },
        status: 200,
      })
    ));
    vi.stubGlobal('fetch', fetchMock);

    render(
      <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
        <ProjectStatusView project={project} />
      </QueryClientProvider>,
    );
    fireEvent.click(await screen.findByRole('checkbox', { name: 'Select one.go' }));
    fireEvent.click(screen.getByRole('button', { name: 'Open PR' }));

    await waitFor(() => expect(fetchMock.mock.calls.some(([, init]) => init?.method === 'POST')).toBe(true));
    const request = fetchMock.mock.calls.find(([, init]) => init?.method === 'POST');
    expect(request?.[0]).toBe('/api/projects/gavel/commit-queue');
    expect(JSON.parse(String(request?.[1]?.body))).toEqual({ action: 'open-pr', files: ['one.go'] });
    expect(await screen.findByRole('button', { name: 'Commit selected (0)' })).toBeTruthy();
  });
});
