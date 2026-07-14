import type React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { CommandPalette } from './CommandPalette';

vi.mock('@flanksource/clicky-ui/components', () => ({
  Modal: ({ open, children }: { open: boolean; children?: React.ReactNode }) => open ? <div>{children}</div> : null,
}));

vi.mock('@flanksource/clicky-ui/icons', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  UiGitPr: () => null,
  UiListDashes: () => null,
  UiSearch: () => null,
}));

beforeEach(() => {
  HTMLElement.prototype.scrollIntoView = vi.fn();
});

describe('CommandPalette UUID entry', () => {
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
