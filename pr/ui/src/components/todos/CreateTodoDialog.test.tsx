import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { Button, ModalProvider } from '@flanksource/clicky-ui/components';
import { setAuthMode } from '../../query';
import { todoForms, useOpenCreateTodo, type CreateTodoDraft } from './CreateTodoDialog';
import { prTodoForms } from './HostedPRTodoDialog';
import { queryToastTestWrapper } from './queryTestWrapper';
import { TodoMutationError } from './todoMutations';
import { CREATE_RUNTIME, CREATE_RUNTIME_CONTEXT } from './createRuntimeTestHelpers';

vi.mock('./run', async importOriginal => ({
  ...await importOriginal<typeof import('./run')>(),
  useTodoRunContext: () => ({ context: CREATE_RUNTIME_CONTEXT, loading: false, error: '' }),
}));
vi.mock('./TodoRunActionButton', async () => ({ TodoRunRuntimeBar: (await import('./createRuntimeTestHelpers')).CreateRuntimeBarStub }));

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

vi.mock('./todoMutations', async importOriginal => ({
  ...await importOriginal<typeof import('./todoMutations')>(),
  useCreateTodoMutation: (dir: string) => {
    createMutation.dirs.push(dir);
    return { isPending: false, mutateAsync: createMutation.mutateAsync };
  },
}));

const stepsByStatus = vi.hoisted(() => ({
  statuses: [] as string[],
  byStatus: {
    pending: [{ name: 'run', label: 'Run', prompt: 'run', auxiliary: false }],
    draft: [{ name: 'triage.new', label: 'Triage new', prompt: 'triage-new', auxiliary: true }],
  } as Record<string, { name: string; label: string; prompt: string; auxiliary: boolean }[]>,
}));

vi.mock('./newTodoSteps', () => ({
  useNewTodoSteps: ({ status }: { status: string }) => {
    stepsByStatus.statuses.push(status);
    return { data: stepsByStatus.byStatus[status] ?? [], isPending: false, error: null };
  },
}));

const forms = { ...prTodoForms, ...todoForms };

function Opener({ draft }: { draft: CreateTodoDraft }) {
  const openCreateTodo = useOpenCreateTodo();
  return <Button onClick={() => openCreateTodo(draft)}>Open new todo</Button>;
}

// Renders the form the way the app does: hosted by the modal provider, opened
// through useOpenCreateTodo.
function openDialog(draft: CreateTodoDraft) {
  const view = render(<ModalProvider scope="test" forms={forms}><Opener draft={draft} /></ModalProvider>, { wrapper: queryToastTestWrapper() });
  fireEvent.click(screen.getByRole('button', { name: 'Open new todo' }));
  return view;
}

function addTodo(title: string) {
  fireEvent.change(screen.getByPlaceholderText('What needs doing?'), { target: { value: title } });
  fireEvent.click(screen.getByRole('button', { name: 'Add todo' }));
}

const acme = [{ name: 'acme', dir: '/work/acme', repos: [] }];

beforeEach(() => {
  localStorage.clear();
  sessionStorage.clear();
  stepsByStatus.statuses.length = 0;
  createMutation.dirs.length = 0;
  createMutation.mutateAsync.mockReset();
});

describe('CreateTodoDialog', () => {
  afterEach(() => setAuthMode('none', vi.fn()));

  it('lets a hosted org member share a todo, private by default', async () => {
    setAuthMode('clerk', vi.fn());
    createMutation.mutateAsync.mockResolvedValue({ todo: { ref: 'shared' } });
    openDialog({ workspaces: [{ name: 'app', dir: 'repo:1', repos: [] }] });
    const visibility = screen.getByLabelText<HTMLSelectElement>('Visibility');
    expect(visibility.value).toBe('private');
    fireEvent.change(visibility, { target: { value: 'public' } });
    addTodo('Shared work');
    await waitFor(() => expect(createMutation.mutateAsync).toHaveBeenCalledOnce());
    expect(JSON.parse(createMutation.mutateAsync.mock.calls[0][0].body)).toMatchObject({ title: 'Shared work', visibility: 'public' });
  });

  it('offers no visibility on a local dashboard, which shows every todo', async () => {
    createMutation.mutateAsync.mockResolvedValue({ todo: { ref: 'local' } });
    openDialog({ workspaces: acme });
    expect(screen.queryByLabelText('Visibility')).toBeNull();
    addTodo('Local work');
    await waitFor(() => expect(createMutation.mutateAsync).toHaveBeenCalledOnce());
    expect(JSON.parse(createMutation.mutateAsync.mock.calls[0][0].body)).not.toHaveProperty('visibility');
  });

  it('submits in the background: the form makes way for a Submitting card, then a created toast opens the todo', async () => {
    const onOpen = vi.fn();
    const created = { ref: 'background', title: 'Repair parser' };
    let respond!: (value: { todo: typeof created }) => void;
    createMutation.mutateAsync.mockReturnValue(new Promise(resolve => { respond = resolve; }));
    openDialog({ workspaces: acme, onOpen });

    addTodo('Repair parser');
    expect(screen.queryByRole('dialog')).toBeNull();
    expect(screen.getByRole('status').textContent).toContain('Submitting');

    respond({ todo: created });
    fireEvent.click(await screen.findByRole('button', { name: 'Open todo' }));
    expect(screen.getByText(/Todo created/)).toBeTruthy();
    expect(onOpen).toHaveBeenCalledWith('/work/acme', created);
    expect(screen.queryByText('Submitting…')).toBeNull();
  });

  it('reopens the draft with the server rejection when the todo is refused', async () => {
    createMutation.mutateAsync.mockRejectedValue(new TodoMutationError('Failed to create todo: title already used', 400));
    openDialog({ workspaces: acme });

    addTodo('Repair parser');

    expect(await screen.findByText('Failed to create todo: title already used')).toBeTruthy();
    expect(screen.getByRole('dialog', { name: 'New todo' })).toBeTruthy();
    expect(screen.getByPlaceholderText<HTMLInputElement>('What needs doing?').value).toBe('Repair parser');
    expect(createMutation.mutateAsync).toHaveBeenCalledOnce();
  });

  it('opens run recovery for a saved todo whose step could not start, without creating it again', async () => {
    createMutation.mutateAsync.mockResolvedValue({ todo: { ref: 'saved' }, run: { status: 'failed', error: 'Admission unavailable' } });
    openDialog({ workspaces: acme });
    fireEvent.change(screen.getByPlaceholderText('What needs doing?'), { target: { value: 'Repair parser' } });
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: 'run' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add todo' }));
    expect(await screen.findByRole('button', { name: 'Retry run' })).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Add todo' })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Continue without running' }));
    expect(screen.queryByRole('dialog')).toBeNull();
    expect(createMutation.mutateAsync).toHaveBeenCalledOnce();
  });

  it('offers None plus the steps for the selected status, defaulting to None', () => {
    openDialog({ workspaces: acme });
    const select = screen.getByLabelText<HTMLSelectElement>('Run after creation');
    expect(select.value).toBe('');
    expect(Array.from(select.options, option => option.textContent)).toEqual(['None', 'Run']);
    fireEvent.change(screen.getByLabelText('Status'), { target: { value: 'draft' } });
    expect(Array.from(screen.getByLabelText<HTMLSelectElement>('Run after creation').options, option => option.textContent)).toEqual(['None', 'Triage new']);
    expect(stepsByStatus.statuses).toContain('draft');
  });

  it('uses the extra-large modal and the shared tall body editor', () => {
    openDialog({ workspaces: [{ name: 'gavel', dir: '/work/gavel', repos: [] }] });

    expect(screen.getByRole('dialog', { name: 'New todo' }).className).toContain('max-w-4xl');
    const body = screen.getByLabelText('Body');
    expect(body.className).toContain('h-64');
    expect(body.className).toContain('resize-y');
    expect(screen.getByText(/HTML.*Markdown/i)).toBeTruthy();
  });

  it('preselects the default workspace and starts each opening from a fresh draft', () => {
    openDialog({ workspaces: [...acme, { name: 'widgets', dir: '/work/widgets', repos: [] }], defaultDir: '/work/widgets' });
    expect(screen.getByLabelText<HTMLSelectElement>('Workspace').value).toBe('/work/widgets');
    fireEvent.change(screen.getByPlaceholderText('What needs doing?'), { target: { value: 'Draft title' } });
    fireEvent.change(screen.getByLabelText('Body'), { target: { value: 'Draft body' } });

    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    expect(screen.queryByRole('dialog')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'Open new todo' }));
    expect(screen.getByPlaceholderText<HTMLInputElement>('What needs doing?').value).toBe('');
    expect(screen.getByLabelText<HTMLTextAreaElement>('Body').value).toBe('');
  });
});

describe('CreateTodoDialog as a child of a todo', () => {
  const parent = { ref: 'parent-ref', title: 'Migrate ledger', dir: '/work/billing' };
  const createdTodo = { ref: 'new-child', title: 'Backfill rows', status: 'pending', priority: 'medium' };

  beforeEach(() => {
    createMutation.mutateAsync.mockResolvedValue({ todo: createdTodo });
  });

  it('files the todo in the parent workspace and names the parent instead of offering a workspace choice', () => {
    openDialog({ parent });

    expect(screen.getByRole('dialog', { name: 'New child todo' })).toBeTruthy();
    expect(screen.getByText('Migrate ledger')).toBeTruthy();
    expect(screen.queryByLabelText('Workspace')).toBeNull();
    expect(new Set(createMutation.dirs)).toEqual(new Set([parent.dir]));
  });

  it('posts the parent ref with the new todo and opens it in the parent workspace', async () => {
    const onOpen = vi.fn();
    openDialog({ parent, onOpen });

    addTodo('Backfill rows');

    fireEvent.click(await screen.findByRole('button', { name: 'Open todo' }));
    expect(onOpen).toHaveBeenCalledWith(parent.dir, createdTodo);
    const request = createMutation.mutateAsync.mock.calls[0][0] as { body: string };
    expect(JSON.parse(request.body)).toEqual({
      title: 'Backfill rows', body: '', priority: 'medium', status: 'pending', parent: parent.ref,
    });
  });

  it('posts the chosen step, and no triage field, with the new todo', async () => {
    openDialog({ parent });

    fireEvent.change(screen.getByPlaceholderText('What needs doing?'), { target: { value: 'Backfill rows' } });
    fireEvent.change(screen.getByLabelText('Run after creation'), { target: { value: 'run' } });
    fireEvent.click(screen.getByRole('button', { name: 'run runtime' }));
    fireEvent.click(screen.getByRole('button', { name: 'Add todo' }));

    await waitFor(() => expect(createMutation.mutateAsync).toHaveBeenCalled());
    const request = createMutation.mutateAsync.mock.calls[0][0] as { body: string };
    const payload = JSON.parse(request.body);
    expect(payload.step).toBe('run');
    expect(payload.spec).toEqual(CREATE_RUNTIME);
    expect('triage' in payload).toBe(false);
  });

  it('sends no parent for an ordinary new todo', async () => {
    openDialog({ workspaces: [{ name: 'gavel', dir: '/work/gavel', repos: [] }] });

    addTodo('Standalone');

    await waitFor(() => expect(createMutation.mutateAsync).toHaveBeenCalled());
    const request = createMutation.mutateAsync.mock.calls[0][0] as { body: string };
    expect('parent' in JSON.parse(request.body)).toBe(false);
  });
});
