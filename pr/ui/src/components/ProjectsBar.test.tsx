import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { ProjectsBar } from './ProjectsBar';
import type { ProjectGitSummaryView } from '../projectGitQueries';
import type { Project } from '../types';
import type { ProjectRuns } from './tests/types';

vi.mock('./ProcControl', () => ({ ProcControl: () => <span>process</span> }));
vi.mock('./TodoBadge', () => ({ TodoBadge: () => <span>todo-badge</span> }));
vi.mock('./GitChangesBadge', () => ({ GitChangesBadge: () => null }));
vi.mock('./RelativeTime', () => ({ RelativeTime: ({ iso }: { iso: string }) => <span>{iso}</span> }));

const runHistory: ProjectRuns[] = [{
  name: 'gavel',
  dir: '/work/gavel',
  runs: [{
    runId: 'run-2026-07-21T10-59-33Z',
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
}];

const noGitSummary: ProjectGitSummaryView = { byProject: new Map(), error: '', loading: true };

describe('ProjectsBar', () => {
  it('shows the git work summary of each project in place of the todo badge', () => {
    const projects: Project[] = [
      { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] },
      { name: 'clicky', dir: '/work/clicky', repos: [] },
    ];
    const gitSummary: ProjectGitSummaryView = {
      byProject: new Map([['gavel', { name: 'gavel', base: 'main', adds: 40, dels: 7, worktrees: 2, branches: 1 }]]),
      error: '',
      loading: false,
    };

    render(
      <ProjectsBar
        projects={projects}
        runs={[]}
        procStatus={{}}
        gitSummary={gitSummary}
        selected=""
        selectedRunId=""
        historyEnabled={false}
        onSelect={() => {}}
        onSelectRun={() => {}}
        onHistoryChange={() => {}}
        onChanged={() => {}}
        onAdd={() => {}}
        onSettings={() => {}}
      />,
    );

    expect(screen.getByText('+40')).toBeTruthy();
    expect(screen.getByText('−7')).toBeTruthy();
    expect(screen.getByLabelText('2 linked worktrees')).toBeTruthy();
    expect(screen.queryByText('todo-badge')).toBeNull();
  });

  it('selects a project from the dedicated projects list without treating settings as selection', () => {
    const project: Project = { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] };
    const onSelect = vi.fn();
    const onSettings = vi.fn();

    render(
      <ProjectsBar
        projects={[project]}
        runs={runHistory}
        procStatus={{}}
        gitSummary={noGitSummary}
        selected="gavel"
        selectedRunId=""
        historyEnabled
        onSelect={onSelect}
        onSelectRun={() => {}}
        onHistoryChange={() => {}}
        onChanged={() => {}}
        onAdd={() => {}}
        onSettings={onSettings}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Open gavel project' }));
    expect(onSelect).toHaveBeenCalledWith(project);

    fireEvent.click(screen.getByRole('button', { name: 'Edit gavel settings' }));
    expect(onSettings).toHaveBeenCalledWith(project);
    expect(onSelect).toHaveBeenCalledOnce();
  });

  it('auto-expands the selected project and selects its nested run', () => {
    const project: Project = { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] };
    const onSelectRun = vi.fn();

    render(
      <ProjectsBar
        projects={[project]}
        runs={runHistory}
        procStatus={{}}
        gitSummary={noGitSummary}
        selected="gavel"
        selectedRunId="run-2026-07-21T10-59-33Z"
        historyEnabled
        onSelect={() => {}}
        onSelectRun={onSelectRun}
        onHistoryChange={() => {}}
        onChanged={() => {}}
        onAdd={() => {}}
        onSettings={() => {}}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Open lint run run-2026-07-21T10-59-33Z for gavel' }));
    expect(onSelectRun).toHaveBeenCalledWith('gavel', 'run-2026-07-21T10-59-33Z');
  });

  it('expands project branches independently', () => {
    const projects: Project[] = [
      { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] },
      { name: 'clicky', dir: '/work/clicky', repos: ['acme/clicky'] },
    ];
    const histories: ProjectRuns[] = [
      runHistory[0],
      {
        name: 'clicky',
        dir: '/work/clicky',
        runs: [{ ...runHistory[0].runs[0], runId: 'run-clicky', kind: 'test' }],
      },
    ];

    render(
      <ProjectsBar
        projects={projects}
        runs={histories}
        procStatus={{}}
        gitSummary={noGitSummary}
        selected=""
        selectedRunId=""
        historyEnabled
        onSelect={() => {}}
        onSelectRun={() => {}}
        onHistoryChange={() => {}}
        onChanged={() => {}}
        onAdd={() => {}}
        onSettings={() => {}}
      />,
    );

    expect(screen.queryByRole('button', { name: 'Open lint run run-2026-07-21T10-59-33Z for gavel' })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Expand gavel runs' }));
    fireEvent.click(screen.getByRole('button', { name: 'Expand clicky runs' }));
    expect(screen.getByRole('button', { name: 'Open lint run run-2026-07-21T10-59-33Z for gavel' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Open test run run-clicky for clicky' })).toBeTruthy();
  });

  it('keeps run controls hidden until history is explicitly enabled', () => {
    const project: Project = { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] };
    const onHistoryChange = vi.fn();

    render(
      <ProjectsBar
        projects={[project]}
        runs={runHistory}
        procStatus={{}}
        gitSummary={noGitSummary}
        selected="gavel"
        selectedRunId=""
        historyEnabled={false}
        onSelect={() => {}}
        onSelectRun={() => {}}
        onHistoryChange={onHistoryChange}
        onChanged={() => {}}
        onAdd={() => {}}
        onSettings={() => {}}
      />,
    );

    expect(screen.queryByRole('button', { name: 'Expand gavel runs' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Open lint run run-2026-07-21T10-59-33Z for gavel' })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Show test and lint history' }));
    expect(onHistoryChange).toHaveBeenCalledWith(true);
  });
});
