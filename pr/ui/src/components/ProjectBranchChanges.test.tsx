import { render, screen, waitFor } from '@testing-library/react';
import { Button } from '@flanksource/clicky-ui/components';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { BranchMergeResult, GitBranchInfo } from '../types';
import type { BranchLandActionsProps } from './BranchLandActions';
import { ProjectBranchChanges } from './ProjectBranchChanges';
import { queryTestWrapper } from './todos/queryTestWrapper';

vi.mock('@flanksource/clicky-ui/data', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/clicky-ui/data')>()),
  GitDiffPanel: ({ payload }: { payload: { diff: string } | null }) => <div data-testid="diff-panel">{payload?.diff ?? 'empty'}</div>,
}));

const landProps = vi.hoisted(() => ({ current: null as unknown }));
vi.mock('./BranchLandActions', () => ({
  BranchLandActions: (props: BranchLandActionsProps) => {
    landProps.current = props;
    return <Button type="button" onClick={() => props.onMerged?.(merged)}>land {props.branch}</Button>;
  },
}));

const merged: BranchMergeResult = {
  targetBranch: 'main', landedSha: 'abc', mode: 'squash', commits: 2, worktreeRemoved: true, branchDeleted: true,
};

const branch: GitBranchInfo = {
  name: 'feat/x y', head: 'e5f6', ahead: 2, behind: 1, worktree: '/work/gavel-feat', diff: { commits: 2, files: 3, adds: 10, dels: 4 }, lastCommitAt: '2026-10-04T09:00:00Z',
};

const files = [{ path: 'cmd/main.go', status: 'modified', adds: 3, dels: 1 }];

function stub(overrides: { files?: () => Response } = {}) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
    const url = new URL(String(input), 'http://localhost');
    if (url.pathname === '/api/projects/my%20app/branch/files') return overrides.files?.() ?? Response.json({ files });
    if (url.pathname === '/api/projects/my%20app/branch/diff') return Response.json({ diff: `diff:${url.searchParams.get('file')}` });
    throw new Error(`unexpected request ${url.pathname}`);
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

const baseProps = {
  projectName: 'my app',
  base: 'main',
  branch,
  baseCheckedOut: true,
  dirtyCount: 0,
  onMerged: () => {},
};

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('ProjectBranchChanges', () => {
  it('loads the branch files and the first diff from the branch endpoints', async () => {
    const fetchMock = stub();

    render(<ProjectBranchChanges {...baseProps} />, { wrapper: queryTestWrapper() });

    expect(await screen.findByText('diff:cmd/main.go')).toBeTruthy();
    expect(fetchMock.mock.calls.map(([input]) => String(input))).toEqual([
      '/api/projects/my%20app/branch/files?branch=feat%2Fx%20y',
      '/api/projects/my%20app/branch/diff?branch=feat%2Fx%20y&file=cmd%2Fmain.go',
    ]);
    expect(screen.queryByRole('combobox')).toBeNull();
  });

  it('summarises the branch against the base', async () => {
    stub();

    render(<ProjectBranchChanges {...baseProps} />, { wrapper: queryTestWrapper() });

    expect(await screen.findByText(/2 commits · 3 files · \+10 −4/)).toBeTruthy();
    expect(screen.getByText('1 behind main')).toBeTruthy();
  });

  it('hands the land actions the readiness facts and reports a merge upward', async () => {
    stub();
    const onMerged = vi.fn();

    render(
      <ProjectBranchChanges {...baseProps} dirtyCount={2} baseCheckedOut={false} worktreePath="/work/gavel-feat" onMerged={onMerged} />,
      { wrapper: queryTestWrapper() },
    );

    await screen.findByText('diff:cmd/main.go');
    expect(landProps.current).toMatchObject({
      projectName: 'my app', branch: 'feat/x y', base: 'main', ahead: 2, dirtyCount: 2, baseCheckedOut: false, worktreePath: '/work/gavel-feat',
    });
    screen.getByRole('button', { name: 'land feat/x y' }).click();
    await waitFor(() => expect(onMerged).toHaveBeenCalledWith(merged));
  });

  it('says the branch is gone when the server answers 410', async () => {
    stub({ files: () => Response.json({ error: 'gone' }, { status: 410 }) });

    render(<ProjectBranchChanges {...baseProps} />, { wrapper: queryTestWrapper() });

    expect(await screen.findByText('Branch no longer exists')).toBeTruthy();
  });
});
