import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { CreateTodoDialog } from './CreateTodoDialog';

const resetMutation = vi.fn();

vi.mock('./attachments', () => ({
  ScreenshotPicker: () => <div />,
  todoFormData: vi.fn(),
  useAttachments: () => ({
    attachments: [],
    previews: [],
    add: vi.fn(),
    remove: vi.fn(),
    clear: vi.fn(),
  }),
}));

vi.mock('./todoMutations', () => ({
  useCreateTodoMutation: () => ({
    isPending: false,
    mutateAsync: vi.fn(),
    reset: resetMutation,
  }),
}));

describe('CreateTodoDialog', () => {
  it('uses the extra-large modal and the shared tall body editor', () => {
    render(
      <CreateTodoDialog
        open
        onClose={vi.fn()}
        workspaces={[{ name: 'gavel', dir: '/work/gavel', repos: [] }]}
        onCreated={vi.fn()}
      />,
    );

    expect(screen.getByRole('dialog', { name: 'New todo' }).className).toContain('max-w-4xl');
    const body = screen.getByLabelText('Body');
    expect(body.className).toContain('h-64');
    expect(body.className).toContain('resize-y');
    expect(screen.getByText(/HTML.*Markdown/i)).toBeTruthy();
  });
});
