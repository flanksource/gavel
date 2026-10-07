import type { ReactNode } from 'react';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { Button } from '@flanksource/clicky-ui/components';
import type { ComboboxOption } from '@flanksource/clicky-ui/components';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { BranchMergeResult, GitBranchInfo, GitWorktree, Project, ProjectGit } from '../types';
import type { ProjectBranchChangesProps } from './ProjectBranchChanges';
import { ProjectRefPane } from './ProjectRefPane';
import { queryTestWrapper } from './todos/queryTestWrapper';

vi.mock('./ProjectStatusView', () => ({
  ProjectStatusView: ({ project, worktree, diffPath }: { project: Project; worktree?: string; diffPath?: string }) => (
    <div>Status {project.name} in {worktree ?? 'main checkout'} diff={diffPath}</div>
  ),
}));

/* oxlint-disable clicky-ui/prefer-clicky-components --
   A test double for clicky-ui's Combobox: the real one positions a floating
   menu jsdom cannot lay out. Every option renders as a row carrying its group,
   description and trailing node, and a button that selects it. */
vi.mock('@flanksource/clicky-ui/components', async importOriginal => ({
  ...await importOriginal<typeof import('@flanksource/clicky-ui/components')>(),
  Combobox: ({ ariaLabel, value, options, onChange }: { ariaLabel: string; value: string; options: ComboboxOption[]; onChange: (value: string) => void }) => (
    <div role="listbox" aria-label={ariaLabel} data-value={value}>
      {options.map(option => (
        <div key={option.value} role="option" aria-selected={option.value === value} aria-label={option.value} data-group={option.group}>
          <button type="button" disabled={option.disabled} onClick={() => onChange(option.value)}>{option.selectedLabel ?? option.label}</button>
          {option.trailing as ReactNode}
        </div>
      ))}
    </div>
  ),
}));
/* oxlint-enable clicky-ui/prefer-clicky-components */

const NOW = new Date('2026-10-04T12:00:00Z');

const branchChanges = vi.hoisted(() => ({ props: null as unknown }));
vi.mock('./ProjectBranchChanges', () => ({
  ProjectBranchChanges: (props: ProjectBranchChangesProps) => {
    branchChanges.props = props;
    return <Button type="button" onClick={() => props.onMerged(merged)}>Branch view {props.branch.name}</Button>;
  },
}));

const project: Project = { name: 'gavel', dir: '/work/gavel', repos: ['acme/gavel'] };
const merged: BranchMergeResult = {
  targetBranch: 'main', landedSha: '1a2b3c4d5e6f', mode: 'squash', commits: 2, worktreeRemoved: true, branchDeleted: true,
};

const changes = { staged: 0, unstaged: 0, both: 0, untracked: 0, conflict: 0, adds: 0, dels: 0 };
const primary: GitWorktree = {
  path: '/work/gavel', branch: 'main', head: 'a1', primary: true, detached: false, prunable: false, changes, ahead: 0,
  lastCommitAt: '2026-10-02T12:00:00Z',
};
const feat: GitWorktree = {
  path: '/work/gavel-feat', branch: 'feat/x', head: 'b2', primary: false, detached: false, prunable: false,
  changes: { ...changes, unstaged: 2, adds: 3, dels: 1 }, ahead: 2,
  lastCommitAt: '2026-10-04T09:00:00Z', touchedAt: '2026-10-04T11:55:00Z',
};
const aheadClean: GitWorktree = { ...feat, path: '/work/gavel-ahead', branch: 'feat/y', changes, ahead: 1, touchedAt: undefined };
const branch = (name: string, worktree: string, ahead: number): GitBranchInfo => ({
  name, head: 'e5', ahead, behind: 0, worktree, diff: { commits: ahead, files: 1, adds: 5, dels: 2 }, lastCommitAt: '2026-10-04T11:00:00Z',
});

const git: ProjectGit = {
  base: 'main',
  currentBranch: 'main',
  baseCheckedOut: true,
  worktrees: [primary, feat, aheadClean],
  branches: [branch('feat/x', feat.path, 2), branch('feat/y', aheadClean.path, 1), branch('spike', '', 3)],
  computedAt: '2026-10-04T11:59:52Z',
};

function stubGit(response: () => Response = () => Response.json(git)) {
  vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL) => {
    if (String(input) !== '/api/projects/gavel/git') throw new Error(`unexpected request ${String(input)}`);
    return response();
  }));
}

function renderPane(projectRef = '', overrides: Partial<React.ComponentProps<typeof ProjectRefPane>> = {}) {
  const onRefChange = vi.fn();
  render(
    <ProjectRefPane
      project={project}
      projectRef={projectRef}
      diffPath=""
      showResults={false}
      onRefChange={onRefChange}
      onDiffPathChange={() => {}}
      onChanged={() => {}}
      {...overrides}
    />,
    { wrapper: queryTestWrapper() },
  );
  return { onRefChange };
}

beforeEach(() => {
  vi.useFakeTimers({ toFake: ['Date'] });
  vi.setSystemTime(NOW);
});

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

const pickerValue = () => screen.getByRole('listbox', { name: 'Worktree or branch' }).getAttribute('data-value');

describe('ProjectRefPane', () => {
  it('shows the main checkout straight away and the grouped picker once the git state loads', async () => {
    stubGit();

    renderPane();

    expect(screen.getByText('Status gavel in main checkout diff=')).toBeTruthy();
    const picker = await screen.findByRole('listbox', { name: 'Worktree or branch' });
    expect(pickerValue()).toBe('@main');
    expect(within(picker).getAllByRole('option').map(option => [option.getAttribute('aria-label'), option.getAttribute('data-group')])).toEqual([
      ['@main', 'Checkout'],
      ['wt:/work/gavel-feat', 'Worktrees'],
      ['wt:/work/gavel-ahead', 'Worktrees'],
      ['br:spike', 'Branches'],
    ]);
  });

  const summaryParts = (option: HTMLElement) =>
    Array.from(option.querySelectorAll(':scope > span:last-child > span')).map(part => part.textContent);

  it('shows each entry with its commits, uncommitted files, lines, last commit and last touched ages', async () => {
    stubGit();
    renderPane();

    const feature = await screen.findByRole('option', { name: 'wt:/work/gavel-feat' });

    expect(summaryParts(feature)).toEqual(['2 commits', '2 uncommitted', '+3−1', '3h', '5m']);
    expect(summaryParts(screen.getByRole('option', { name: 'br:spike' }))).toEqual(['3 commits', '+5−2', '1h']);
    expect(summaryParts(screen.getByRole('option', { name: '@main' }))).toEqual(['2d']);
  });

  it('colours added lines green and removed lines red', async () => {
    stubGit();
    renderPane();

    const feature = await screen.findByRole('option', { name: 'wt:/work/gavel-feat' });

    expect(within(feature).getByText('+3').className).toContain('text-green-600');
    expect(within(feature).getByText('−1').className).toContain('text-red-600');
  });

  it('asks to switch ref with the raw ref value, mapping the main checkout back to ""', async () => {
    stubGit();
    const { onRefChange } = renderPane('br:spike');

    fireEvent.click(await screen.findByRole('button', { name: 'Checkout main' }));
    fireEvent.click(screen.getByRole('button', { name: 'Worktree feat/x' }));

    expect(onRefChange.mock.calls).toEqual([[''], ['wt:/work/gavel-feat']]);
  });

  it('summarises the selected entry beside the picker', async () => {
    stubGit();

    renderPane('wt:/work/gavel-feat');

    const picker = await screen.findByRole('listbox', { name: 'Worktree or branch' });
    expect(Array.from(picker.nextElementSibling?.children ?? []).map(part => part.textContent)).toEqual(['2 commits', '2 uncommitted', '+3−1', '3h', '5m']);
  });

  it('shows how old the memoized git state is', async () => {
    stubGit();

    renderPane('');

    expect(await screen.findByText('git state 8s ago')).toBeTruthy();
  });

  it('hides the picker for a project with only its main checkout', async () => {
    stubGit(() => Response.json({ ...git, worktrees: [primary], branches: [] }));

    renderPane();

    await waitFor(() => expect(screen.queryByText('Loading git state…')).toBeNull());
    expect(screen.queryByRole('listbox')).toBeNull();
    expect(screen.getByText(/Status gavel/)).toBeTruthy();
  });

  it('keeps the main checkout usable and reports the failure when the git state cannot load', async () => {
    stubGit(() => Response.json({ error: 'git exploded' }, { status: 500 }));

    renderPane();

    expect((await screen.findByRole('alert')).textContent).toContain('git exploded');
    expect(screen.getByText(/Status gavel in main checkout/)).toBeTruthy();
  });

  it('reviews a linked worktree through the status view scoped to it', async () => {
    stubGit();

    renderPane('wt:/work/gavel-feat', { diffPath: 'one.go' });

    expect(await screen.findByText('Status gavel in /work/gavel-feat diff=one.go')).toBeTruthy();
    expect(pickerValue()).toBe('wt:/work/gavel-feat');
  });

  it('offers a dirty worktree that is ahead its working changes and its committed changes', async () => {
    stubGit();
    renderPane('wt:/work/gavel-feat');

    await screen.findByText(/Status gavel in \/work\/gavel-feat/);
    fireEvent.click(screen.getByRole('radio', { name: /Committed changes \(2 ahead\)/ }));

    expect(screen.getByRole('button', { name: 'Branch view feat/x' })).toBeTruthy();
    expect(branchChanges.props).toMatchObject({
      projectName: 'gavel', base: 'main', baseCheckedOut: true, dirtyCount: 2, worktreePath: '/work/gavel-feat',
    });
  });

  it('opens a clean worktree that is ahead on its committed changes', async () => {
    stubGit();

    renderPane('wt:/work/gavel-ahead');

    expect(await screen.findByRole('button', { name: 'Branch view feat/y' })).toBeTruthy();
    expect(branchChanges.props).toMatchObject({ dirtyCount: 0, worktreePath: '/work/gavel-ahead' });
    fireEvent.click(screen.getByRole('radio', { name: 'Working changes' }));
    expect(screen.getByText(/Status gavel in \/work\/gavel-ahead/)).toBeTruthy();
  });

  it('shows a branch without a worktree as its changes against the base', async () => {
    stubGit();

    renderPane('br:spike');

    expect(await screen.findByRole('button', { name: 'Branch view spike' })).toBeTruthy();
    expect(branchChanges.props).toMatchObject({ projectName: 'gavel', base: 'main', dirtyCount: 0 });
    expect(branchChanges.props).not.toHaveProperty('worktreePath');
    expect(screen.queryByText(/Status gavel/)).toBeNull();
  });

  it.each([
    ['worktree', 'wt:/work/gone', /Worktree \/work\/gone no longer exists/],
    ['branch', 'br:gone', /Branch gone no longer has commits to review/],
  ])('says so and offers the main checkout when the selected %s is gone', async (_kind, ref, message) => {
    stubGit();
    const { onRefChange } = renderPane(ref);

    expect((await screen.findByText(message))).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Back to main checkout' }));
    expect(onRefChange).toHaveBeenCalledWith('');
  });

  it('returns to the main checkout after a merge and reports what landed', async () => {
    stubGit();
    const { onRefChange } = renderPane('br:spike');

    fireEvent.click(await screen.findByRole('button', { name: 'Branch view spike' }));

    expect(onRefChange).toHaveBeenCalledWith('');
    const status = await screen.findByRole('status');
    expect(status.textContent).toContain('Merged spike into main');
    expect(status.textContent).toContain('1a2b3c4');
    expect(status.textContent).toContain('2 commits');
    expect(status.textContent).toContain('worktree removed');
    expect(status.textContent).toContain('branch deleted');
  });
});
