import type React from 'react';
import { act, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ProjectsView } from './ProjectsView';
import type { Project } from '../types';

const streamListeners = new Map<string, EventListener>();

vi.mock('@flanksource/clicky-ui/components', async importOriginal => ({
  ...await importOriginal<typeof import('@flanksource/clicky-ui/components')>(),
  SplitPane: ({ left, right }: { left: React.ReactNode; right: React.ReactNode }) => <div>{left}{right}</div>,
}));
vi.mock('./ProcControl', () => ({ ProcControl: () => null }));
vi.mock('./TodoBadge', () => ({ TodoBadge: () => null }));
vi.mock('./GitChangesBadge', () => ({ GitChangesBadge: () => null }));
vi.mock('./RelativeTime', () => ({ RelativeTime: ({ iso }: { iso: string }) => <span>{iso}</span> }));
vi.mock('./ProjectStatusView', () => ({ ProjectStatusView: ({ project }: { project: Project }) => <div>Status {project.name}</div> }));
vi.mock('./tests/TestRunDetail', () => ({ TestRunDetail: ({ project, runId }: { project: string; runId: string }) => <div>Run {project}/{runId}</div> }));

const project: Project = { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] };
const response = {
  projects: [{
    name: 'gavel',
    dir: '/work/gavel',
    runs: [{
      runId: 'run-1',
      kind: 'lint',
      started: '2026-07-21T10:59:33Z',
      passed: 0,
      failed: 0,
      skipped: 0,
      warned: 0,
      total: 0,
      lintViolations: 3,
      lintLinters: 2,
    }],
  }],
};

function stubRunHistory(payload: unknown) {
  vi.stubGlobal('fetch', vi.fn(async () => ({ ok: true, json: async () => payload }) as Response));
  vi.stubGlobal('EventSource', vi.fn(() => ({
    addEventListener: (type: string, listener: EventListener) => streamListeners.set(type, listener),
    close: vi.fn(),
    onerror: null,
  })));
}

afterEach(() => {
  streamListeners.clear();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('ProjectsView', () => {
  it('lists history-only projects when the configured project catalog is empty', async () => {
    stubRunHistory(response);

    render(
      <ProjectsView
        projects={[]}
        procStatus={{}}
        selectedName="gavel"
        selectedRunId="run-1"
        diffPath=""
        onSelect={() => {}}
        onSelectRun={() => {}}
        onDiffPathChange={() => {}}
        onChanged={() => {}}
        onAdd={() => {}}
        onSettings={() => {}}
      />,
    );

    expect(await screen.findByRole('button', { name: 'Open lint run run-1 for gavel' })).toBeTruthy();
    expect(screen.getByText('Run gavel/run-1')).toBeTruthy();
  });

  it('loads project run history and renders the selected run in the detail pane', async () => {
    stubRunHistory(response);

    render(
      <ProjectsView
        projects={[project]}
        procStatus={{}}
        selectedName="gavel"
        selectedRunId="run-1"
        diffPath=""
        onSelect={() => {}}
        onSelectRun={() => {}}
        onDiffPathChange={() => {}}
        onChanged={() => {}}
        onAdd={() => {}}
        onSettings={() => {}}
      />,
    );

    expect(await screen.findByRole('button', { name: 'Open lint run run-1 for gavel' })).toBeTruthy();
    expect(screen.getByText('Run gavel/run-1')).toBeTruthy();
  });

  it('applies valid SSE updates and surfaces malformed frames', async () => {
    stubRunHistory({ projects: [{ name: 'gavel', dir: '/work/gavel', runs: [] }] });

    render(
      <ProjectsView
        projects={[project]}
        procStatus={{}}
        selectedName="gavel"
        selectedRunId=""
        diffPath=""
        onSelect={() => {}}
        onSelectRun={() => {}}
        onDiffPathChange={() => {}}
        onChanged={() => {}}
        onAdd={() => {}}
        onSettings={() => {}}
      />,
    );

    await screen.findByText('No test or lint runs');
    act(() => streamListeners.get('message')?.(new MessageEvent('message', { data: JSON.stringify(response) })));
    expect(await screen.findByRole('button', { name: 'Open lint run run-1 for gavel' })).toBeTruthy();

    act(() => streamListeners.get('message')?.(new MessageEvent('message', { data: '{' })));
    expect((await screen.findByRole('alert')).textContent).toContain('invalid update');
  });
});
