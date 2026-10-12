import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { AgentSession } from '../../types';
import { SessionsDetailPane } from './SessionsView';

vi.mock('@flanksource/clicky-ui/ai', () => ({
  SessionInspector: ({ src, follow }: { src: string; follow: boolean }) => <div data-testid="inspector" data-src={src} data-follow={String(follow)} />,
}));
vi.mock('./SessionDetail', () => ({
  SessionDetail: ({ session }: { session: AgentSession }) => <div>{session.title}</div>,
}));

const selectedId = 'metadata-only';
const actions = { onOpenTodo: vi.fn(), onOpenWorktree: vi.fn() };
const listed: AgentSession = {
  id: selectedId, title: 'Session with activity', source: 'claude', lifecycleStatus: 'running',
  activityState: 'working', processActive: true, lastActivityAt: '2026-10-11T12:00:00Z',
  tokens: { input: 0, output: 0, cacheRead: 0, total: 0 }, costUsd: 0,
};

describe('SessionsDetailPane direct links', () => {
  it('loads an unlisted session following by its lifecycle and switches to its listed detail when it appears', () => {
    const { rerender } = render(<SessionsDetailPane selectedId={selectedId} query={{ data: { sessions: [], errors: [] }, error: null }} {...actions} />);
    expect(screen.getByTestId('inspector').getAttribute('data-src')).toBe(`/api/captain/sessions/${selectedId}`);
    expect(screen.getByTestId('inspector').getAttribute('data-follow')).toBe('undefined');

    rerender(<SessionsDetailPane selectedId={selectedId} query={{ data: { sessions: [listed], errors: [] }, error: null }} {...actions} />);
    expect(screen.getByText(listed.title)).toBeTruthy();
    expect(screen.queryByTestId('inspector')).toBeNull();
  });

  it('loads a direct link before the list query resolves', () => {
    render(<SessionsDetailPane selectedId={selectedId} query={{ data: undefined, error: null }} {...actions} />);
    expect(screen.getByTestId('inspector').getAttribute('data-src')).toBe(`/api/captain/sessions/${selectedId}`);
  });
});
