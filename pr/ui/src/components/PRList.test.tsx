// Tests the sidebar list of pull requests: collapsing a repository, and what it says while the
// first fetch is loading or once it finds none.
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { PRItem } from '../types';
import { PRList } from './PRList';

function pr(repo: string, number: number, title: string): PRItem {
  return {
    number, title, repo, author: 'ada', source: `feature-${number}`, target: 'main', state: 'OPEN',
    isDraft: false, url: `https://github.com/${repo}/pull/${number}`, updatedAt: '2026-10-15T09:00:00Z',
  };
}

function renderList(prs: PRItem[], loading?: boolean) {
  render(
    <QueryClientProvider client={new QueryClient()}>
      <PRList prs={prs} selected={null} onSelect={() => {}} {...(loading === undefined ? {} : { loading })} />
    </QueryClientProvider>,
  );
}

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

  it('says it is loading, not that there are none, until the first fetch finishes', () => {
    renderList([], true);
    expect(screen.getByRole('status').textContent).toContain('Loading pull requests');
    expect(screen.queryByText('No pull requests found')).toBeNull();
  });

  it('says there are none once a fetch found none', () => {
    renderList([], false);
    expect(screen.getByText('No pull requests found')).toBeTruthy();
    expect(screen.queryByRole('status')).toBeNull();
  });
});
