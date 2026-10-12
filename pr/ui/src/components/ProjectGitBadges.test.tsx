import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { ProjectGitSummaryView } from '../projectGitQueries';
import type { ProjectGitSummary } from '../types';
import { ProjectGitBadges, projectGitBadgesState } from './ProjectGitBadges';

const summary: ProjectGitSummary = { name: 'gavel', base: 'main', adds: 120, dels: 34, worktrees: 3, branches: 2 };

const view = (overrides: Partial<ProjectGitSummaryView> = {}, summaries: ProjectGitSummary[] = []): ProjectGitSummaryView => ({
  byProject: new Map(summaries.map(entry => [entry.name, entry])),
  error: '',
  loading: false,
  ...overrides,
});

describe('projectGitBadgesState', () => {
  it.each([
    ['still loading', view({ loading: true }), { status: 'loading' }],
    ['loaded without this project yet', view({}, [{ ...summary, name: 'other' }]), { status: 'loading' }],
    ['failed to load', view({ error: 'Load project git summary: HTTP 500' }), { status: 'error', message: 'Load project git summary: HTTP 500' }],
    ['loaded', view({}, [summary]), { status: 'ready', summary }],
  ])('maps a summary query that is %s', (_label, summaries, expected) => {
    expect(projectGitBadgesState(summaries, 'gavel')).toEqual(expected);
  });
});

describe('ProjectGitBadges', () => {
  it('renders nothing while the summary is loading', () => {
    const { container } = render(<ProjectGitBadges state={{ status: 'loading' }} />);

    expect(container.firstChild).toBeNull();
  });

  it('shows a warning with the failure reason instead of numbers when the summary failed', () => {
    render(<ProjectGitBadges state={{ status: 'error', message: 'Load project git summary: HTTP 500' }} />);

    const warning = screen.getByRole('img', { name: 'Git summary unavailable' });
    expect(warning.getAttribute('title')).toBe('Load project git summary: HTTP 500');
    expect(screen.queryByText(/\+\d/)).toBeNull();
  });

  it('shows the per-project failure the same way and never zero work', () => {
    render(<ProjectGitBadges state={{ status: 'ready', summary: { ...summary, adds: 0, dels: 0, error: 'not a git repository' } }} />);

    expect(screen.getByRole('img', { name: 'Git summary unavailable' }).getAttribute('title')).toBe('not a git repository');
    expect(screen.queryByText('+0')).toBeNull();
  });

  it('shows additions, deletions, worktree and branch counts with tooltips', () => {
    render(<ProjectGitBadges state={{ status: 'ready', summary }} />);

    expect(screen.getByText('+120').className).toContain('text-green');
    expect(screen.getByText('+120').getAttribute('title')).toBe('120 lines added vs main');
    expect(screen.getByText('−34').className).toContain('text-red');
    expect(screen.getByText('−34').getAttribute('title')).toBe('34 lines removed vs main');
    expect(screen.getByLabelText('3 linked worktrees').getAttribute('title')).toBe('3 linked worktrees');
    expect(screen.getByLabelText('2 unmerged branches').getAttribute('title')).toBe('2 branches with commits not in main');
  });

  it('omits the line numbers when there is no unmerged work but keeps the counts', () => {
    render(<ProjectGitBadges state={{ status: 'ready', summary: { ...summary, adds: 0, dels: 0 } }} />);

    expect(screen.queryByText(/^\+/)).toBeNull();
    expect(screen.queryByText(/^−/)).toBeNull();
    expect(screen.getByLabelText('3 linked worktrees')).toBeTruthy();
  });

  it('renders nothing for a project with no work, worktrees or branches', () => {
    const { container } = render(
      <ProjectGitBadges state={{ status: 'ready', summary: { ...summary, adds: 0, dels: 0, worktrees: 0, branches: 0 } }} />,
    );

    expect(container.firstChild).toBeNull();
  });

  it('pluralises singular counts', () => {
    render(<ProjectGitBadges state={{ status: 'ready', summary: { ...summary, worktrees: 1, branches: 1 } }} />);

    expect(screen.getByLabelText('1 linked worktree')).toBeTruthy();
    expect(screen.getByLabelText('1 unmerged branch').getAttribute('title')).toBe('1 branch with commits not in main');
  });
});
