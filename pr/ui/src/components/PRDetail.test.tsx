import type { ReactNode } from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { PRDetailPanel } from './PRDetail';
import type { PRItem } from '../types';

// clicky-ui's SplitButton / DropdownMenu pull @floating-ui/react, which resolves
// a duplicate React 18 under vitest and crashes on render (see PRActions.test.tsx).
// Stub the components so the test exercises PRDetailPanel's own close-button
// wiring, not clicky internals. useListMenuSelection is called unconditionally
// by CreateTodoFromPRDialog (which PRDetailPanel always mounts, closed) before
// its own `if (!open) return null` — its return value is unused while closed,
// so the stub only needs to avoid throwing.
vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, onClick, title, disabled, ...rest }: any) => (
    <button onClick={onClick} title={title} disabled={disabled} aria-label={rest['aria-label']}>{children}</button>
  ),
  SplitButton: ({ label, onClick, title, disabled }: { label: ReactNode; onClick?: () => void; title?: string; disabled?: boolean }) => (
    <button onClick={onClick} title={title} disabled={disabled}>{label}</button>
  ),
  DropdownMenu: ({ trigger }: { trigger: ReactNode }) => <div>{trigger}</div>,
  Modal: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  useListMenuSelection: () => ({ selectedKeys: [], toggle: () => {}, selectAll: () => {}, clear: () => {} }),
}));

function makePR(overrides: Partial<PRItem> = {}): PRItem {
  return {
    number: 7,
    title: 'Test PR',
    author: 'octocat',
    repo: 'acme/widget',
    source: 'feature',
    target: 'main',
    state: 'OPEN',
    isDraft: false,
    url: 'https://github.com/acme/widget/pull/7',
    updatedAt: '2026-01-01T00:00:00Z',
    ...overrides,
  };
}

describe('PRDetailPanel close button', () => {
  it('renders a close button and fires onClose when clicked', () => {
    const onClose = vi.fn();
    render(<PRDetailPanel pr={makePR()} detail={null} loading={false} onClose={onClose} />);

    const closeButton = screen.getByLabelText('Close pull request details');
    fireEvent.click(closeButton);

    expect(onClose).toHaveBeenCalledOnce();
  });

  it('omits the close button when onClose is not provided', () => {
    render(<PRDetailPanel pr={makePR()} detail={null} loading={false} />);

    expect(screen.queryByLabelText('Close pull request details')).toBeNull();
  });
});
