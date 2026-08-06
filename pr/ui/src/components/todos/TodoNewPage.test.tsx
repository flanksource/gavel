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
});
