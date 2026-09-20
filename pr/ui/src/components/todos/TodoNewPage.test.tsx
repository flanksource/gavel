import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoItem } from '../../types';
import { loadRecentTodos, rememberRecentTodo } from './recentTodos';
import { TodoNewPage } from './TodoNewPage';

const mocks = vi.hoisted(() => ({
  items: [] as unknown[],
  create: vi.fn(),
  update: vi.fn(),
}));

vi.mock('@tanstack/react-query', async importOriginal => ({
  ...await importOriginal<typeof import('@tanstack/react-query')>(),
  useQuery: () => ({ data: { items: mocks.items }, error: null, isFetching: false, isPending: false }),
}));

vi.mock('../../procStatusQuery', () => ({ useProcStatus: () => ({}) }));

vi.mock('@flanksource/clicky-ui/icons', async importOriginal => ({
  ...await importOriginal<typeof import('@flanksource/clicky-ui/icons')>(),
  UiClose: () => <span />,
}));

vi.mock('./attachments', () => ({
  ScreenshotPicker: () => <div />,
  todoCommentFormData: vi.fn(),
  todoFormData: vi.fn(),
  useAttachments: () => ({
    attachments: [],
    previews: [],
    add: vi.fn(),
    remove: vi.fn(),
  }),
}));

vi.mock('./todoMutations', () => ({
  useCreateTodoMutation: () => ({ isPending: false, mutateAsync: mocks.create }),
  useUpdateTodoMutation: () => ({ isPending: false, mutateAsync: mocks.update }),
}));

const workspace = { name: 'acme', dir: '/work/acme', repos: [] };
const sourceHost = 'localhost:5173';
const sourceUrl = `http://${sourceHost}/checkout`;

function chooseWorkspace() {
  fireEvent.change(screen.getByLabelText('Workspace'), { target: { value: workspace.dir } });
}

function todo(ref: string, title: string, status: TodoItem['status'] = 'pending'): TodoItem {
  return { ref, shortId: ref.slice(0, 4), title, status } as TodoItem;
}

beforeEach(() => {
  window.history.replaceState({}, '', '/todos/new?embed=1');
  localStorage.clear();
  mocks.items = [];
  mocks.create.mockReset();
  mocks.update.mockReset();
});
afterEach(() => vi.restoreAllMocks());

describe('TodoNewPage', () => {
  it('uses the wider page and shared tall editor for new bodies and existing comments', () => {
    const { container } = render(
      <TodoNewPage projects={[{ name: 'gavel', dir: '/work/gavel', repos: [] }]} />,
    );

    expect(container.querySelector('main')?.className).toContain('max-w-4xl');
    const body = screen.getByLabelText('Body');
    expect(body.className).toContain('h-64');
    expect(body.className).toContain('resize-y');

    fireEvent.click(screen.getByRole('button', { name: 'Existing issue' }));

    const comment = screen.getByLabelText('Comment');
    expect(comment.className).toContain('h-64');
    expect(comment.className).toContain('resize-y');
    expect(screen.getByText(/multiline.*code block/i)).toBeTruthy();
  });

  // A rejected projects payload leaves this page with an empty catalog, which is
  // indistinguishable from "no workspaces are configured" unless the failure is
  // shown: without it the form silently offers the server's own work dir and
  // files the todo against the wrong project.
  it('reports a failed project catalog instead of offering the default workspace', () => {
    render(<TodoNewPage projects={[]} projectError="Load projects: invalid project" />);

    expect(screen.getByText('Load projects: invalid project')).toBeTruthy();
    const workspace = screen.getByLabelText('Workspace') as HTMLSelectElement;
    expect(workspace.disabled).toBe(true);
    expect(workspace.options[0].text).toBe('Workspaces unavailable');
  });

  it('creates the todo with the query-seeded labels plus one picked in the form', async () => {
    window.history.replaceState({}, '', '/todos/new?embed=1&labels=ui,bug');
    mocks.create.mockResolvedValue({ todo: todo('t-new', 'Fix checkout') });
    render(<TodoNewPage projects={[workspace]} />);
    chooseWorkspace();

    fireEvent.change(screen.getByPlaceholderText('What needs doing?'), { target: { value: 'Fix checkout' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add tag' }));
    const filter = screen.getByLabelText('Filter tags');
    fireEvent.change(filter, { target: { value: 'flaky' } });
    fireEvent.keyDown(filter, { key: 'Enter' });
    fireEvent.click(screen.getByRole('button', { name: 'Add todo' }));

    await waitFor(() => expect(mocks.create).toHaveBeenCalledTimes(1));
    expect(JSON.parse(mocks.create.mock.calls[0][0].body).labels).toEqual(['ui', 'bug', 'flaky']);
  });

  it('remembers a created todo under the source page host', async () => {
    window.history.replaceState({}, '', `/todos/new?embed=1&sourceUrl=${encodeURIComponent(sourceUrl)}`);
    mocks.create.mockResolvedValue({ todo: todo('t-new', 'Fix checkout') });
    render(<TodoNewPage projects={[workspace]} />);
    chooseWorkspace();

    fireEvent.change(screen.getByPlaceholderText('What needs doing?'), { target: { value: 'Fix checkout' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add todo' }));

    await waitFor(() => expect(loadRecentTodos(sourceHost)).toHaveLength(1));
    expect(loadRecentTodos(sourceHost)[0]).toMatchObject({ ref: 't-new', title: 'Fix checkout', dir: workspace.dir });
  });

  it('offers recent open todos from the same host as chips that select the existing issue', () => {
    rememberRecentTodo(sourceHost, { ref: 't-done', title: 'Closed since', dir: workspace.dir, at: 2 });
    rememberRecentTodo(sourceHost, { ref: 't-open', title: 'Stored title', dir: workspace.dir, at: 1 });
    rememberRecentTodo('other.example.com', { ref: 't-other', title: 'Other app', dir: workspace.dir, at: 3 });
    mocks.items = [todo('t-open', 'Current title'), todo('t-done', 'Closed since', 'completed'), todo('t-other', 'Other app')];
    window.history.replaceState({}, '', `/todos/new?embed=1&sourceUrl=${encodeURIComponent(sourceUrl)}`);
    render(<TodoNewPage projects={[workspace]} />);

    const chips = screen.getByRole('group', { name: 'Add to recent' });
    expect(Array.from(chips.querySelectorAll('button')).map(b => b.title)).toEqual(['Current title']);

    fireEvent.click(screen.getByTitle('Current title'));

    expect((screen.getByLabelText('Issue') as HTMLSelectElement).value).toBe('t-open');
    expect(screen.getByTitle('Current title').getAttribute('aria-pressed')).toBe('true');
  });

  // Nothing identifies the project, so guessing one (the first in the list) would
  // file the todo against the wrong workspace: the user has to pick.
  it('leaves the workspace unselected and blocks submit until one is picked', () => {
    const first = { name: 'alpha', dir: '/work/alpha', repos: [] };
    window.history.replaceState({}, '', `/todos/new?embed=1&sourceUrl=${encodeURIComponent(sourceUrl)}`);
    render(<TodoNewPage projects={[first, workspace]} />);

    const select = screen.getByLabelText('Workspace') as HTMLSelectElement;
    expect(select.value).toBe('');
    expect(select.selectedOptions[0].text).toBe('Select workspace');
    fireEvent.change(screen.getByPlaceholderText('What needs doing?'), { target: { value: 'Fix checkout' } });
    expect((screen.getByRole('button', { name: 'Add todo' }) as HTMLButtonElement).disabled).toBe(true);

    chooseWorkspace();
    expect((screen.getByRole('button', { name: 'Add todo' }) as HTMLButtonElement).disabled).toBe(false);
  });

  // Port detection waits on /api/proc/status; the host's last-used workspace is
  // known synchronously, so its chips render on first paint instead of after it.
  it('opens on the workspace last used from this host so its chips show on first render', () => {
    const first = { name: 'alpha', dir: '/work/alpha', repos: [] };
    rememberRecentTodo(sourceHost, { ref: 't-open', title: 'Checkout bug', dir: workspace.dir, at: 1 });
    mocks.items = [todo('t-open', 'Checkout bug')];
    window.history.replaceState({}, '', `/todos/new?embed=1&sourceUrl=${encodeURIComponent(sourceUrl)}`);
    render(<TodoNewPage projects={[first, workspace]} />);

    expect((screen.getByLabelText('Workspace') as HTMLSelectElement).value).toBe(workspace.dir);
    expect(screen.getByTitle('Checkout bug')).toBeTruthy();
  });

  it('remembers a todo commented on from the existing-issue mode', async () => {
    mocks.items = [todo('t-open', 'Checkout bug')];
    mocks.update.mockResolvedValue(todo('t-open', 'Checkout bug'));
    window.history.replaceState({}, '', `/todos/new?embed=1&ref=t-open&sourceUrl=${encodeURIComponent(sourceUrl)}`);
    render(<TodoNewPage projects={[workspace]} />);
    chooseWorkspace();

    fireEvent.change(screen.getByLabelText('Comment'), { target: { value: 'Also on mobile' } });
    fireEvent.click(screen.getByRole('button', { name: 'Add comment' }));

    await waitFor(() => expect(loadRecentTodos(sourceHost).map(e => e.ref)).toEqual(['t-open']));
  });

  it('offers the default workspace when no projects are configured', () => {
    render(<TodoNewPage projects={[]} />);

    const workspace = screen.getByLabelText('Workspace') as HTMLSelectElement;
    expect(workspace.disabled).toBe(false);
    expect(workspace.options[0].text).toBe('Default workspace');
  });
});
