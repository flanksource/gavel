import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { TodoEvent } from '../../types';
import { TodoTimeline } from './TodoTimeline';

function landedEvent(payload: Record<string, unknown>): TodoEvent {
  return { id: 'event-1', kind: 'run_landed', actor: 'moshe', timestamp: '2026-09-14T11:00:00Z', title: 'run_landed', payload };
}

describe('TodoTimeline run_landed', () => {
  it('shows a PR landing with its link, target branch and landed sha', () => {
    render(<TodoTimeline events={[landedEvent({
      promptRunId: 'run-2', via: 'pr', targetBranch: 'main',
      landedSha: '1a2b3c4d5e6f1a2b3c4d5e6f1a2b3c4d5e6f1a2b', prNumber: 42, prUrl: 'https://github.com/acme/widgets/pull/42',
    })]} />);

    expect(screen.getByText('landed the run')).toBeTruthy();
    expect(screen.getByRole('link', { name: /#42/ }).getAttribute('href')).toBe('https://github.com/acme/widgets/pull/42');
    expect(screen.getByText('main')).toBeTruthy();
    expect(screen.getByText('1a2b3c4')).toBeTruthy();
  });

  it('shows a merge landing with its target branch', () => {
    render(<TodoTimeline events={[landedEvent({ via: 'merge', targetBranch: 'feature/x', landedSha: 'abcdef0123' })]} />);

    expect(screen.getByText('Merged into')).toBeTruthy();
    expect(screen.getByText('feature/x')).toBeTruthy();
    expect(screen.getByText('abcdef0')).toBeTruthy();
    expect(screen.queryByRole('link')).toBeNull();
  });

  it('says what is wrong with a malformed landing payload instead of hiding it', () => {
    render(<TodoTimeline events={[landedEvent({ via: 'merge', landedSha: 'abcdef0123' })]} />);

    expect(screen.getByText(/no target branch/)).toBeTruthy();
  });
});
