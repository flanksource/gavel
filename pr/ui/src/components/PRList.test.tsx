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

function renderList(prs: PRItem[]) {
  render(
    <QueryClientProvider client={new QueryClient()}>
      <PRList prs={prs} selected={null} onSelect={() => {}} />
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
});
