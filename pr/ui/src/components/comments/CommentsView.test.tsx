import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { EMPTY_INBOX_FILTERS, projectScope, type NativeListItem } from '@flanksource/clicky-ui/comments/gavel';
import { CommentsView } from './CommentsView';
import { CommentSession } from './CommentSession';

const { openEventStream } = vi.hoisted(() => ({ openEventStream: vi.fn() }));
vi.mock('../../eventHub', () => ({ openEventStream }));
vi.mock('@flanksource/clicky-ui/ai', async importOriginal => ({
  ...await importOriginal<typeof import('@flanksource/clicky-ui/ai')>(),
  SessionInspector: ({ src }: { src: string }) => <div data-testid="comment-session" data-src={src} />,
}));

const projects = [{ name: 'Acme UI', dir: '/work/ui', repos: [] }, { name: 'API', dir: '/work/api', repos: [] }];
const uiScope = projectScope('Acme UI');
const apiScope = projectScope('API');

function thread(id: string, scope: string): NativeListItem {
  return {
    id, bodyMarkdown: `Review ${id}`, author: { kind: 'user', id: 'ada' },
    target: { scopeUri: scope, uri: `${scope}/files/README.md`, kind: 'file' },
    status: 'open', phase: 'active', version: 7, unread: false, readVersion: 7,
    createdAt: '2026-11-01T10:00:00Z', updatedAt: '2026-11-01T10:00:00Z',
  };
}

function show({ local = false, selectedId = '', visible = false } = {}) {
  const query = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={query}>
    <CommentsView projects={projects} scopedProjects={projects} projectsLoaded projectError="" local={local} visible={visible}
      filters={EMPTY_INBOX_FILTERS} selectedId={selectedId} onSelect={vi.fn()} onFiltersChange={vi.fn()} />
  </QueryClientProvider>);
}

afterEach(() => vi.unstubAllGlobals());

describe('dashboard comments', () => {
  it.each([
    { capability: 'update', kind: 'root', original: 'Review ui' },
    { capability: 'edit-reply', kind: 'reply', original: 'Reply to review' },
  ])('offers to edit the $kind when Gavel reports $capability', async ({ capability, kind, original }) => {
    const root = thread('ui', uiScope);
    const item = { ...root, messages: [{ id: 'message', bodyMarkdown: 'Reply to review', author: root.author, createdAt: root.createdAt }] };
    vi.stubGlobal('fetch', (input: string) => {
      const url = new URL(input, 'http://localhost');
      if (url.pathname.endsWith('/capabilities')) return Response.json({ actions: [capability] });
      return Response.json({ data: url.searchParams.get('scope') === uiScope ? [item] : [] });
    });
    show({ selectedId: 'ui' });

    const card = within((await screen.findAllByTestId('comment-card')).find(element => element.dataset.commentKind === kind)!);
    await waitFor(() => expect(screen.queryByText('Loading comment actions…')).toBeNull());
    fireEvent.click(card.getByRole('button', { name: 'Comment actions' }));
    fireEvent.click(await screen.findByRole('menuitem', { name: 'Edit' }));
    expect(card.getByTestId('comment-edit-input').textContent).toBe(original);
  });

  it('links an Amp delivery to Amp without mounting the transcript', () => {
    const ampRun = { state: 'sent', sessionId: 'T-review', requestedAt: '2026-11-01T10:00:00Z', target: { label: 'Review', channel: 'amp-http' } };
    render(<CommentSession run={ampRun} local />);
    expect(screen.getByRole('link', { name: 'Open in Amp' }).getAttribute('href')).toBe('https://ampcode.com/threads/T-review');
    expect(screen.queryByTestId('comment-session')).toBeNull();
  });

  it('loads each project scope and continues through later pages', async () => {
    const fetcher = vi.fn((input: string) => {
      const url = new URL(input, 'http://localhost');
      if (url.searchParams.get('scope') === apiScope) return Response.json({ data: [thread('api', apiScope)] });
      const size = Number(url.searchParams.get('limit'));
      if (url.searchParams.get('offset') === '0') return Response.json({ data: Array.from({ length: size }, (_, index) => thread(`first-${index}`, uiScope)) });
      return Response.json({ data: [thread('later-page', uiScope)] });
    });
    vi.stubGlobal('fetch', fetcher);
    show();

    await screen.findByText('Review later-page');
    expect(screen.getByText('Review api')).toBeTruthy();
    expect(fetcher).toHaveBeenCalledTimes(3);
  });

  it('uses the dashboard event hub to reload only the changed scope', async () => {
    const events = new Map<string, EventTarget>();
    openEventStream.mockImplementation((input: string) => {
      const scope = new URL(input, 'http://localhost').searchParams.get('scope')!;
      const target = new EventTarget();
      events.set(scope, target);
      return Object.assign(target, { close: vi.fn() });
    });
    const listCalls = vi.fn();
    let changed = false;
    vi.stubGlobal('fetch', (input: string) => {
      const scope = new URL(input, 'http://localhost').searchParams.get('scope')!;
      listCalls(scope);
      return Response.json({ data: [thread(changed && scope === uiScope ? 'updated' : scope, scope)] });
    });
    show({ visible: true });
    await screen.findByText(`Review ${uiScope}`);
    expect(listCalls.mock.calls).toEqual([[uiScope], [apiScope]]);

    changed = true;
    events.get(uiScope)!.dispatchEvent(new MessageEvent('comment', { lastEventId: '42' }));
    await screen.findByText('Review updated');
    expect(listCalls.mock.calls).toEqual([[uiScope], [apiScope], [uiScope]]);
  });

  it('uses the root version to acknowledge an expanded hosted thread and hides the host actions', async () => {
    const acknowledge = vi.fn();
    vi.stubGlobal('fetch', (input: string, init?: RequestInit) => {
      const url = new URL(input, 'http://localhost');
      if (url.pathname.endsWith('/capabilities')) return Response.json({ actions: ['acknowledge'] });
      if (url.pathname.endsWith('/acknowledge')) {
        acknowledge({ path: url.pathname, body: JSON.parse(String(init?.body)) });
        return Response.json(thread('ui', uiScope));
      }
      return Response.json({ data: url.searchParams.get('scope') === uiScope ? [thread('ui', uiScope)] : [] });
    });
    show({ selectedId: 'ui' });

    await waitFor(() => expect(screen.queryByText('Loading comment actions…')).toBeNull());
    fireEvent.click(await screen.findByRole('button', { name: 'Comment actions' }));
    fireEvent.click(await screen.findByRole('menuitem', { name: 'Acknowledge' }));
    await waitFor(() => expect(acknowledge).toHaveBeenCalledWith({ path: '/api/v1/comments/ui/acknowledge', body: { expectedVersion: 7 } }));
    expect(screen.queryByRole('button', { name: 'Send to agent' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Open README.md' })).toBeNull();
  });

  it('previews a proposal-only run before delivering the displayed version', async () => {
    const operations = vi.fn();
    vi.stubGlobal('fetch', (input: string, init?: RequestInit) => {
      const url = new URL(input, 'http://localhost');
      if (url.pathname.endsWith('/capabilities')) return Response.json({ actions: [] });
      if (url.pathname.endsWith('/agent-sessions')) return Response.json([]);
      if (url.pathname === '/api/settings/prompts/catalog') return Response.json([{ id: 'editor-comments' }]);
      if (url.pathname.endsWith('/preview') || url.pathname.endsWith('/deliver')) {
        operations({ path: url.pathname, body: JSON.parse(String(init?.body)) });
        return url.pathname.endsWith('/preview')
          ? Response.json({ model: 'Claude', mode: 'managed', tools: [{ name: 'reply_to_thread', description: '', source: 'thread' }] })
          : Response.json(thread('ui', uiScope));
      }
      return Response.json({ data: url.searchParams.get('scope') === uiScope ? [thread('ui', uiScope)] : [] });
    });
    show({ selectedId: 'ui', local: true });

    fireEvent.click(await screen.findByRole('button', { name: 'Send to agent' }));
    fireEvent.click(await screen.findByRole('menuitem', { name: "New session (Gavel's default)" }));
    fireEvent.click(screen.getByRole('radio', { name: 'Proposal only' }));
    const send = screen.getByRole('button', { name: 'Send comment' }) as HTMLButtonElement;
    expect(send.disabled).toBe(true);
    fireEvent.click(screen.getByRole('button', { name: 'Preview run' }));
    await screen.findByText('reply_to_thread');
    fireEvent.click(send);

    const target = { kind: 'run', prompt: 'editor-comments', fresh: true, edits: 'propose' };
    await waitFor(() => expect(operations.mock.calls).toEqual([
      [{ path: '/api/v1/comments/ui/deliver/preview', body: { target } }],
      [{ path: '/api/v1/comments/ui/deliver', body: { expectedVersion: 7, requestId: expect.any(String), target } }],
    ]));
  });

  it("picks the run's prompt from the thread project's prompt catalog", async () => {
    const previews = vi.fn();
    const catalogScopes = vi.fn();
    vi.stubGlobal('fetch', (input: string, init?: RequestInit) => {
      const url = new URL(input, 'http://localhost');
      if (url.pathname === '/api/settings/prompts/catalog') {
        catalogScopes(url.searchParams.get('project'));
        return Response.json([{ id: 'editor-comments', title: 'Editor comments' }, { id: 'review-docs', title: 'Review docs' }]);
      }
      if (url.pathname.endsWith('/capabilities')) return Response.json({ actions: [] });
      if (url.pathname.endsWith('/preview')) {
        previews(JSON.parse(String(init?.body)).target.prompt);
        return Response.json({ model: 'Claude', mode: 'managed', tools: [] });
      }
      return Response.json({ data: url.searchParams.get('scope') === uiScope ? [thread('ui', uiScope)] : [] });
    });
    show({ selectedId: 'ui', local: true });

    fireEvent.click(await screen.findByRole('button', { name: 'Send to agent' }));
    fireEvent.click(await screen.findByRole('menuitem', { name: "New session (Gavel's default)" }));
    await waitFor(() => expect(catalogScopes).toHaveBeenCalledWith('Acme UI'));
    const picker = screen.getByRole('combobox', { name: 'Prompt' });
    fireEvent.focus(picker);
    fireEvent.change(picker, { target: { value: 'review' } });
    fireEvent.mouseDown(await screen.findByRole('option', { name: 'review-docs — Review docs' }));
    fireEvent.click(screen.getByRole('button', { name: 'Preview run' }));

    await waitFor(() => expect(previews.mock.calls).toEqual([['review-docs']]));
  });

  it('opens attachment-backed capture context and the recorded agent session from the maximized thread', async () => {
    const captured = {
      ...thread('ui', uiScope),
      attachments: [{ id: 'image-1', name: 'screenshot.png' }],
      content: {
        playground: { page: 'example.com/docs', element: { pageUrl: 'https://example.com/docs', componentName: 'HelpLink', source: 'help.tsx:17', html: '<a>Help</a>' } },
        agentRun: { state: 'sent', sessionId: 'captain-123', requestedAt: '2026-11-01T10:00:00Z' },
      },
    };
    vi.stubGlobal('fetch', (input: string) => {
      const url = new URL(input, 'http://localhost');
      if (url.pathname.endsWith('/capabilities')) return Response.json({ actions: [] });
      if (url.pathname.endsWith('/agent-sessions')) return Response.json([{ id: 'alternate-session', title: 'Docs review', source: 'claude', inCmux: false, status: 'ended', resumable: true }]);
      return Response.json({ data: url.searchParams.get('scope') === uiScope ? [captured] : [] });
    });
    show({ selectedId: 'ui', local: true });

    fireEvent.click(await screen.findByRole('button', { name: 'Comment actions' }));
    fireEvent.click(await screen.findByRole('menuitem', { name: 'Maximise' }));
    const dialog = within(await screen.findByRole('dialog', { name: 'Comment thread' }));
    fireEvent.click(dialog.getByText('Capture'));
    expect(dialog.getByText('help.tsx:17')).toBeTruthy();
    expect(dialog.getByRole('img', { name: 'Captured screenshot' }).getAttribute('src')).toBe('/api/v1/comments/ui/attachments/image-1');
    fireEvent.click(dialog.getByText('Session'));
    expect(dialog.getByTestId('comment-session').getAttribute('data-src')).toBe('/api/captain/sessions/captain-123');
    fireEvent.click(dialog.getByRole('button', { name: 'Send to agent' }));
    fireEvent.click(await screen.findByRole('menuitem', { name: 'Docs review · resumes' }));
    const send = within(await screen.findByRole('dialog', { name: 'Send comment to agent' }));
    expect(send.getByRole('button', { name: 'Use session Docs review' }).getAttribute('aria-pressed')).toBe('true');
  });
});
