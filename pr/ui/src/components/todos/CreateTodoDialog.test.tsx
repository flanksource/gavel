import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
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

const createMutation = vi.hoisted(() => ({
  mutateAsync: vi.fn(),
  dirs: [] as string[],
}));

vi.mock('./todoMutations', () => ({
  useCreateTodoMutation: (dir: string) => {
    createMutation.dirs.push(dir);
    return { isPending: false, mutateAsync: createMutation.mutateAsync };
  },
}));

describe('CreateTodoDialog', () => {
	 it('keeps a saved todo open for triage recovery without creating it again', async () => {
	   const onCreated = vi.fn();
	   createMutation.mutateAsync.mockReset();
	   createMutation.mutateAsync.mockResolvedValue({ todo: { ref: 'saved' }, triage: { status: 'failed', error: 'Admission unavailable' } });
	   render(<CreateTodoDialog open onClose={vi.fn()} workspaces={[{ name: 'acme', dir: '/work/acme', repos: [] }]} onCreated={onCreated} />);
	   fireEvent.change(screen.getByPlaceholderText('What needs doing?'), { target: { value: 'Repair parser' } });
	   fireEvent.click(screen.getByRole('button', { name: 'Add todo' }));
	   expect(await screen.findByRole('button', { name: 'Retry triage' })).toBeTruthy();
	   expect(screen.queryByRole('button', { name: 'Add todo' })).toBeNull();
	   expect(onCreated).not.toHaveBeenCalled();
	   fireEvent.click(screen.getByRole('button', { name: 'Continue without triage' }));
	   expect(onCreated).toHaveBeenCalledOnce();
	   expect(createMutation.mutateAsync).toHaveBeenCalledOnce();
	 });
	 it('enables triage by default and allows opting out', () => {
	   render(<CreateTodoDialog open onClose={vi.fn()} workspaces={[{ name: 'acme', dir: '/work/acme', repos: [] }]} onCreated={vi.fn()} />);
	   const checkbox = screen.getByRole<HTMLInputElement>('checkbox', { name: 'Triage after creation' });
	   expect(checkbox.checked).toBe(true);
	   fireEvent.click(checkbox);
	   expect(checkbox.checked).toBe(false);
	 });
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

describe('CreateTodoDialog as a child of a todo', () => {
  const parent = { ref: 'parent-ref', title: 'Migrate ledger', dir: '/work/billing' };
  const createdTodo = { ref: 'new-child', title: 'Backfill rows', status: 'pending', priority: 'medium' };

  beforeEach(() => {
    createMutation.dirs.length = 0;
    createMutation.mutateAsync.mockReset();
    createMutation.mutateAsync.mockResolvedValue({ todo: createdTodo });
  });

  it('files the todo in the parent workspace and names the parent instead of offering a workspace choice', () => {
    render(<CreateTodoDialog open onClose={vi.fn()} parent={parent} onCreated={vi.fn()} />);

    expect(screen.getByText('Migrate ledger')).toBeTruthy();
    expect(screen.queryByLabelText('Workspace')).toBeNull();
    expect(new Set(createMutation.dirs)).toEqual(new Set([parent.dir]));
  });

  it('posts the parent ref with the new todo and reports it created in the parent workspace', async () => {
    const onCreated = vi.fn();
    render(<CreateTodoDialog open onClose={vi.fn()} parent={parent} onCreated={onCreated} />);

    fireEvent.change(screen.getByPlaceholderText('What needs doing?'), { target: { value: 'Backfill rows' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add todo' }));

    await waitFor(() => expect(onCreated).toHaveBeenCalledWith(parent.dir, createdTodo));
    const request = createMutation.mutateAsync.mock.calls[0][0] as { body: string };
    expect(JSON.parse(request.body)).toEqual({
      title: 'Backfill rows', body: '', priority: 'medium', status: 'pending', parent: parent.ref, triage: true,
    });
  });

  it('sends no parent for an ordinary new todo', async () => {
    render(<CreateTodoDialog open onClose={vi.fn()} workspaces={[{ name: 'gavel', dir: '/work/gavel', repos: [] }]} onCreated={vi.fn()} />);

    fireEvent.change(screen.getByPlaceholderText('What needs doing?'), { target: { value: 'Standalone' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add todo' }));

    await waitFor(() => expect(createMutation.mutateAsync).toHaveBeenCalled());
    const request = createMutation.mutateAsync.mock.calls[0][0] as { body: string };
    expect('parent' in JSON.parse(request.body)).toBe(false);
  });
});
