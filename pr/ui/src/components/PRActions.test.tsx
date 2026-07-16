import type { ComponentType, ReactNode } from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { PRActions, type ExtraAction } from './PRActions';
import type { PRItem } from '../types';
import { UiAdd, type IconProps } from '@flanksource/clicky-ui/icons';

// clicky-ui's SplitButton / DropdownMenu pull @floating-ui/react, which resolves
// a duplicate React 18 under vitest and crashes on render. Stub the components so
// the test exercises PRActions' own collapse branching, not clicky internals. A
// closed DropdownMenu renders only its trigger — mirror that so hidden menu items
// stay out of the DOM.
vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, onClick, title, disabled, ...rest }: any) => (
    <button onClick={onClick} title={title} disabled={disabled} aria-label={rest['aria-label']}>{children}</button>
  ),
  SplitButton: ({ label, onClick, title, disabled }: { label: ReactNode; onClick?: () => void; title?: string; disabled?: boolean }) => (
    <button onClick={onClick} title={title} disabled={disabled}>{label}</button>
  ),
  DropdownMenu: ({ trigger }: { trigger: ReactNode }) => <div>{trigger}</div>,
  Modal: ({ children }: { children: ReactNode }) => <div>{children}</div>,
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

const newTodo: ExtraAction = { label: 'New todo', icon: UiAdd as ComponentType<IconProps>, onClick: () => {} };

describe('PRActions responsive collapse', () => {
  it('renders every action inline when not collapsed', () => {
    render(<PRActions pr={makePR()} detail={null} collapsed={false} extras={[newTodo]} />);

    expect(screen.getByText('Approve')).toBeTruthy();
    expect(screen.getByText('Merge')).toBeTruthy();
    expect(screen.getByText('New todo')).toBeTruthy();
    expect(screen.queryByLabelText('Pull request actions')).toBeNull();
  });

  it('collapses the actions behind an overflow trigger when narrow', () => {
    render(<PRActions pr={makePR()} detail={null} collapsed extras={[newTodo]} />);

    expect(screen.getByLabelText('Pull request actions')).toBeTruthy();
    // The closed menu keeps the individual actions out of the DOM until opened.
    expect(screen.queryByText('Approve')).toBeNull();
    expect(screen.queryByText('New todo')).toBeNull();
  });

  it('keeps a lone action inline even when collapsed', () => {
    // A merged PR has no GitHub actions, so only the New todo extra remains.
    render(<PRActions pr={makePR({ state: 'MERGED' })} detail={null} collapsed extras={[newTodo]} />);

    expect(screen.getByText('New todo')).toBeTruthy();
    expect(screen.queryByLabelText('Pull request actions')).toBeNull();
  });

  it('renders nothing for a merged PR with no extras', () => {
    const { container } = render(<PRActions pr={makePR({ state: 'MERGED' })} detail={null} collapsed={false} />);
    expect(container.firstChild).toBeNull();
  });
});
