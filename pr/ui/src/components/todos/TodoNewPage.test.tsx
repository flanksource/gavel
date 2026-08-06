import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { TodoNewPage } from './TodoNewPage';

vi.mock('@tanstack/react-query', async importOriginal => ({
  ...await importOriginal<typeof import('@tanstack/react-query')>(),
  useQuery: () => ({ data: { items: [] }, error: null, isFetching: false }),
}));

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
  useCreateTodoMutation: () => ({ isPending: false, mutateAsync: vi.fn() }),
  useUpdateTodoMutation: () => ({ isPending: false, mutateAsync: vi.fn() }),
}));

beforeEach(() => window.history.replaceState({}, '', '/todos/new?embed=1'));
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

  it('offers the default workspace when no projects are configured', () => {
    render(<TodoNewPage projects={[]} />);

    const workspace = screen.getByLabelText('Workspace') as HTMLSelectElement;
    expect(workspace.disabled).toBe(false);
    expect(workspace.options[0].text).toBe('Default workspace');
  });
});
