import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ComponentProps } from 'react';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { PRItem, TodoCounts } from '../types';
import { PRList } from './PRList';

function pr(repo: string, number: number, title: string): PRItem {
  return {
    number, title, repo, author: 'ada', source: `feature-${number}`, target: 'main', state: 'OPEN',
    isDraft: false, url: `https://github.com/${repo}/pull/${number}`, updatedAt: '2026-10-15T09:00:00Z',
  };
}

function renderList(prs: PRItem[], extra: Partial<ComponentProps<typeof PRList>> = {}) {
  render(
    <QueryClientProvider client={new QueryClient()}>
      <PRList prs={prs} selected={null} onSelect={() => {}} {...extra} />
    </QueryClientProvider>,
  );
}

const failingChecks = { passed: 1, failed: 1, running: 0, pending: 0 };
const failing = (p: PRItem): PRItem => ({ ...p, checkStatus: failingChecks });
const conflicting = (p: PRItem): PRItem => ({ ...p, mergeable: 'CONFLICTING' });
const todos: TodoCounts = {
  total: 7, open: 3, draft: 0, pending: 0, inProgress: 4, review: 0, ask: 0,
  failed: 0, unverified: 0, verified: 0, completed: 0, skipped: 0,
};

function repoHeader(repo: string): HTMLElement {
  const header = screen.getByRole('button', { name: `Collapse ${repo}` }).parentElement;
  if (!header) throw new Error(`no header for ${repo}`);
  return header;
}

describe('PRList mobile group headers', () => {
  it.each([
    ['failing checks win over conflicts', [failing(pr('acme/api', 1, 'a')), failing(pr('acme/api', 2, 'b')), conflicting(pr('acme/api', 3, 'c'))], '2 failing'],
    ['conflicts show when nothing fails', [conflicting(pr('acme/api', 1, 'a')), pr('acme/api', 2, 'b')], '1 conflicting'],
  ])('keeps one action count per header: %s', (_name, prs, label) => {
    renderList(prs, { mobile: true, projectsByRepo: { 'acme/api': { name: 'api', dir: '/w/api', repos: ['acme/api'], todoCounts: todos } } });

    const header = repoHeader('acme/api');
    expect(within(header).getAllByLabelText(/failing|conflicting/).map(el => el.getAttribute('aria-label'))).toEqual([label]);
    expect(within(header).queryByTitle('3 open todos, 4 in progress')).toBeNull();
    expect(within(header).queryByTitle(`${prs.length} total`)).toBeNull();
  });

  it('shows no action count on a header with nothing failing or conflicting', () => {
    renderList([pr('acme/api', 1, 'a')], { mobile: true });

    expect(within(repoHeader('acme/api')).queryByLabelText(/failing|conflicting/)).toBeNull();
  });

  it('moves the full breakdown into the org and repo header titles', () => {
    renderList(
      [failing(pr('acme/api', 1, 'a')), conflicting(pr('acme/api', 2, 'b'))],
      { mobile: true, projectsByRepo: { 'acme/api': { name: 'api', dir: '/w/api', repos: ['acme/api'], todoCounts: todos } } },
    );

    expect(repoHeader('acme/api').getAttribute('title')).toBe('2 pull requests · 2 open · 1 failing · 1 conflicting · 3 open todos, 4 in progress');
    expect(screen.getByText('acme').parentElement?.getAttribute('title')).toBe('2 pull requests · 2 open · 1 failing · 1 conflicting');
  });

  it('keeps the full counter set on desktop', () => {
    renderList(
      [failing(pr('acme/api', 1, 'a'))],
      { projectsByRepo: { 'acme/api': { name: 'api', dir: '/w/api', repos: ['acme/api'], todoCounts: todos } } },
    );

    const header = repoHeader('acme/api');
    expect(within(header).getByTitle('1 failing')).toBeTruthy();
    expect(within(header).getByTitle('3 open todos, 4 in progress')).toBeTruthy();
  });
});

describe('PRList author avatars', () => {
  it("hides the viewer's own avatar and keeps other authors'", () => {
    renderList([pr('acme/api', 1, 'mine'), { ...pr('acme/api', 2, 'theirs'), author: 'grace' }], { viewer: 'ada' });

    expect(screen.queryByTitle('@ada')).toBeNull();
    expect(screen.getByTitle('@grace')).toBeTruthy();
  });
});

describe('PRList', () => {
  it("collapses and expands a repository's pull requests from the arrow beside it", () => {
    renderList([pr('acme/api', 1, 'Add login'), pr('acme/web', 2, 'Fix header')]);
    const collapse = screen.getByRole('button', { name: 'Collapse acme/api' });
    expect(collapse.getAttribute('aria-expanded')).toBe('true');

    fireEvent.click(collapse);

    expect(screen.queryByText('Add login')).toBeNull();
    expect(screen.getByText('Fix header')).toBeTruthy();
    expect(screen.getByText('api')).toBeTruthy();

    const expand = screen.getByRole('button', { name: 'Expand acme/api' });
    expect(expand.getAttribute('aria-expanded')).toBe('false');
    fireEvent.click(expand);
    expect(screen.getByText('Add login')).toBeTruthy();
  });
});
