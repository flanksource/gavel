import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { TodoEvent } from '../../types';
import { TodoTimeline } from './TodoTimeline';

vi.mock('@flanksource/clicky-ui/data', async importOriginal => ({
  ...(await importOriginal<typeof import('@flanksource/clicky-ui/data')>()),
  // Streamdown pulls its own React copy under vitest; the body text is all these tests read.
  Markdown: ({ text }: { text: string }) => <div>{text}</div>,
}));

const anchored: TodoEvent = {
  id: 'c1', kind: 'comment', actor: 'reviewer', timestamp: '2026-10-01T10:00:00Z', body: 'handle the error',
  payload: { anchor: { path: 'cmd/main.go', side: 'new', line: 12, lineText: 'return nil', commit: 'abc1234' } },
};

describe('TodoTimeline line comments', () => {
  it('shows an anchored comment with its path:line chip and no resolved badge while it is open', () => {
    render(<TodoTimeline events={[anchored]} />);

    expect(screen.getByText('cmd/main.go:12')).toBeTruthy();
    expect(screen.getByText('handle the error')).toBeTruthy();
    expect(screen.queryByText('Resolved')).toBeNull();
  });

  it('marks the anchored comment resolved and logs the resolution as its own entry', () => {
    const resolution: TodoEvent = { id: 'r1', kind: 'comment_resolved', actor: 'reviewer', timestamp: '2026-10-01T11:00:00Z', payload: { commentId: 'c1', resolved: true } };

    render(<TodoTimeline events={[anchored, resolution]} />);

    expect(screen.getByText('Resolved')).toBeTruthy();
    expect(screen.getByText('resolved a comment')).toBeTruthy();
    expect(screen.getAllByText('cmd/main.go:12')).toHaveLength(2);
  });

  it('labels a reopening and reverts the comment to open', () => {
    const resolve: TodoEvent = { id: 'r1', kind: 'comment_resolved', timestamp: '2026-10-01T11:00:00Z', payload: { commentId: 'c1', resolved: true } };
    const reopen: TodoEvent = { id: 'r2', kind: 'comment_resolved', timestamp: '2026-10-01T12:00:00Z', payload: { commentId: 'c1', resolved: false } };

    render(<TodoTimeline events={[anchored, resolve, reopen]} />);

    expect(screen.getByText('reopened a comment')).toBeTruthy();
    expect(screen.queryByText('Resolved')).toBeNull();
  });

  it('leaves a plain comment without a path chip', () => {
    render(<TodoTimeline events={[{ id: 'p1', kind: 'comment', body: 'looks good', timestamp: '2026-10-01T10:00:00Z' }]} />);

    expect(screen.getByText('looks good')).toBeTruthy();
    expect(screen.queryByText(/:\d+$/)).toBeNull();
  });
});
