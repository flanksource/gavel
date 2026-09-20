import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { CreateTodoDialog } from './CreateTodoDialog';

vi.mock('./attachments', () => ({
  ScreenshotPicker: () => <div />,
  todoFormData: vi.fn(),
  useAttachments: () => ({
    attachments: [],
    previews: [],
    add: vi.fn(),
    remove: vi.fn(),
  }),
}));

vi.mock('./todoMutations', () => ({
  useCreateTodoMutation: () => ({
    isPending: false,
    mutateAsync: vi.fn(),
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

  // The dashboard re-renders on every live stream frame and hands the dialog a
  // freshly-filtered workspaces array each time, so a draft must survive new
  // prop identities and only reset when the dialog is reopened.
  it('keeps the draft across re-renders and resets it only on reopen', () => {
    const workspaces = [{ name: 'acme', dir: '/work/acme', repos: [] }, { name: 'widgets', dir: '/work/widgets', repos: [] }];
    const dialog = (open: boolean, defaultDir: string) => (
      <CreateTodoDialog open={open} onClose={vi.fn()} workspaces={workspaces.map(w => ({ ...w }))} onCreated={vi.fn()} defaultDir={defaultDir} />
    );
    const view = render(dialog(true, '/work/acme'));
    fireEvent.change(screen.getByPlaceholderText('What needs doing?'), { target: { value: 'Draft title' } });
    fireEvent.change(screen.getByLabelText('Body'), { target: { value: 'Draft body' } });

    view.rerender(dialog(true, '/work/widgets'));
    expect(screen.getByPlaceholderText<HTMLInputElement>('What needs doing?').value).toBe('Draft title');
    expect(screen.getByLabelText<HTMLTextAreaElement>('Body').value).toBe('Draft body');
    expect(screen.getByLabelText<HTMLSelectElement>('Workspace').value).toBe('/work/acme');

    view.rerender(dialog(false, '/work/widgets'));
    view.rerender(dialog(true, '/work/widgets'));
    expect(screen.getByPlaceholderText<HTMLInputElement>('What needs doing?').value).toBe('');
    expect(screen.getByLabelText<HTMLTextAreaElement>('Body').value).toBe('');
    expect(screen.getByLabelText<HTMLSelectElement>('Workspace').value).toBe('/work/widgets');
  });
});
