import type React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { CommandPalette } from './CommandPalette';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Button: ({ children, onClick, className, 'aria-label': ariaLabel }: {
    children?: React.ReactNode;
    onClick?: React.MouseEventHandler<HTMLSpanElement>;
    className?: string;
    'aria-label'?: string;
  }) => <span role="button" tabIndex={0} className={className} aria-label={ariaLabel} onClick={onClick}>{children}</span>,
  Modal: ({ open, children, expandable, scrollBody }: {
    open: boolean;
    children?: React.ReactNode;
    expandable?: boolean;
    scrollBody?: boolean;
  }) => open ? (
    <div data-testid="command-palette-modal" data-expandable={String(expandable)} data-scroll-body={String(scrollBody)}>{children}</div>
  ) : null,
}));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  UiGitPr: () => null,
  UiListDashes: () => null,
  UiSearch: () => null,
  UiArrowLeft: () => null,
}));

beforeEach(() => {
  HTMLElement.prototype.scrollIntoView = vi.fn();
});

describe('CommandPalette UUID entry', () => {
  it('provides back navigation and a pinned search header for the mobile page', () => {
    const onClose = vi.fn();
    render(
      <CommandPalette
        open
        onClose={onClose}
        prs={[]}
        todos={[]}
        todosLoading={false}
        onSelectPR={vi.fn()}
        onSelectTodo={vi.fn()}
        onOpenUUID={vi.fn()}
      />,
    );

    expect(screen.getByTestId('command-palette-modal').getAttribute('data-expandable')).toBe('false');
    expect(screen.getByTestId('command-palette-modal').getAttribute('data-scroll-body')).toBe('false');
    const back = screen.getByRole('button', { name: 'Back' });
    expect(back.className).toContain('sm:!hidden');
    expect(screen.getByText('esc').className).toContain('sm:!inline-flex');
    expect(screen.getByRole('textbox').parentElement?.className).toContain('shrink-0');
    expect(screen.getByRole('textbox').parentElement?.className).toContain('max-sm:-mx-density-4');
    fireEvent.click(back);
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('opens a canonical Todo or session UUID directly on Enter', () => {
    const onClose = vi.fn();
    const onOpenUUID = vi.fn();
    render(
      <CommandPalette
        open
        onClose={onClose}
        prs={[]}
        todos={[]}
        todosLoading={false}
        onSelectPR={vi.fn()}
        onSelectTodo={vi.fn()}
        onOpenUUID={onOpenUUID}
      />,
    );

    const input = screen.getByRole('textbox', { name: 'Search pull requests, todos, and sessions' });
    fireEvent.change(input, { target: { value: '019F5B29-7890-7C11-8E7A-838E5D373E39' } });
    expect(screen.getByText('Open Todo or session UUID')).toBeTruthy();

    fireEvent.keyDown(input, { key: 'Enter' });
    expect(onOpenUUID).toHaveBeenCalledWith('019f5b29-7890-7c11-8e7a-838e5d373e39');
    expect(onClose).toHaveBeenCalledTimes(1);
  });
});

describe('CommandPalette child todos', () => {
  const workspace = { name: 'billing', dir: '/repos/billing', repos: [] };
  const parentId = '6b1f3c9a-2d4e-4f50-8a71-93c0d5e2b418';
  const entries = [
    { workspace, todo: { ref: 'parent', id: parentId, title: 'Migrate ledger', status: 'pending' as const, priority: 'medium' as const } },
    { workspace, todo: { ref: 'child', id: 'child-id', parentId, title: 'Backfill ledger rows', status: 'pending' as const, priority: 'medium' as const } },
  ];

  it('finds a child by its own title and shows which parent it hangs under', () => {
    render(
      <CommandPalette
        open
        onClose={vi.fn()}
        prs={[]}
        todos={entries}
        todosLoading={false}
        onSelectPR={vi.fn()}
        onSelectTodo={vi.fn()}
        onOpenUUID={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'backfill' } });
    expect(screen.getByText('Backfill ledger rows')).toBeTruthy();
    expect(screen.getByTitle('Child of Migrate ledger').textContent).toBe('↳ Migrate ledger');
  });

  it('shows no parent hint on a top-level todo', () => {
    render(
      <CommandPalette
        open
        onClose={vi.fn()}
        prs={[]}
        todos={entries}
        todosLoading={false}
        onSelectPR={vi.fn()}
        onSelectTodo={vi.fn()}
        onOpenUUID={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByRole('textbox'), { target: { value: 'migrate' } });
    expect(screen.getByText('Migrate ledger')).toBeTruthy();
    expect(screen.queryByText(/↳/)).toBeNull();
  });
});
