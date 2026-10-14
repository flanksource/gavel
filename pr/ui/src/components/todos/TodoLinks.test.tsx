import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TodoLink, TodoPRLinkItem } from '../../types';
import { queryTestWrapper } from './queryTestWrapper';
import { TodoLinks } from './TodoLinks';

const DIR = '/repos/billing';
const REF = 'abc123';
const PR_URL = 'https://github.com/acme/ledger/pull/42';

const prLinks: TodoPRLinkItem[] = [
  { provider: 'github', repo: 'acme/ledger', number: 42, url: PR_URL, relation: 'implements', source: 'triage', createdAt: '2026-10-01T00:00:00Z', title: 'Rework ledger rounding', state: 'OPEN' },
  { provider: 'github', repo: 'acme/ledger', number: 7, url: 'https://github.com/acme/ledger/pull/7', relation: 'related', source: 'manual', createdAt: '2026-10-02T00:00:00Z' },
];

const links: TodoLink[] = [
  { relation: 'duplicates', target_id: 'dup-full-id', target_short_id: 'dup1', target_title: 'Round totals twice', target_status: 'pending' },
  { relation: 'related_to', target_id: 'rel-full-id', target_short_id: 'rel1', target_title: 'Audit ledger precision', target_status: 'completed' },
  { relation: 'depends_on', target_id: 'dep-full-id', target_short_id: 'dep1', target_title: 'Upgrade decimal library', target_status: 'pending' },
];

interface Backend {
  links: TodoLink[] | unknown;
  prLinks: TodoPRLinkItem[] | unknown;
}

let backend: Backend;
let fetchMock: ReturnType<typeof vi.fn>;

function respond(body: unknown, status = 200) {
  return Promise.resolve({
    ok: status < 400,
    status,
    json: async () => body,
    text: async () => (body === undefined ? '' : JSON.stringify(body)),
    clone() { return this; },
  } as unknown as Response);
}

beforeEach(() => {
  backend = { links, prLinks };
  fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (init?.method === 'DELETE') return respond(url.startsWith('/api/todos/pr-links') ? undefined : { links: [] }, url.startsWith('/api/todos/pr-links') ? 204 : 200);
    if (url.startsWith('/api/todos/pr-links')) return respond(backend.prLinks);
    if (url.startsWith('/api/todos/links')) return respond({ links: backend.links });
    throw new Error(`unexpected fetch ${url}`);
  });
  vi.stubGlobal('fetch', fetchMock);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

async function renderLinks(props: Partial<React.ComponentProps<typeof TodoLinks>> = {}) {
  const view = render(<TodoLinks todoRef={REF} dir={DIR} {...props} />, { wrapper: queryTestWrapper() });
  await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
  await act(async () => { await new Promise(resolve => setTimeout(resolve, 0)); });
  return view;
}

const group = (name: string) => screen.getByRole('heading', { name }).closest('section')!;

describe('TodoLinks', () => {
  it('renders Pull requests, Duplicates and Related as separate groups, in that order', async () => {
    await renderLinks();

    const headings = screen.getAllByRole('heading').map(heading => heading.textContent);
    expect(headings).toEqual(['Pull requests', 'Duplicates', 'Related']);

    const prs = within(group('Pull requests'));
    expect(prs.getByText('implements')).toBeTruthy();
    expect(prs.getByText('PR #42').getAttribute('href')).toBe(PR_URL);
    expect(prs.getByText('Rework ledger rounding')).toBeTruthy();
    expect(prs.getByText('OPEN')).toBeTruthy();
    expect(prs.getByText('triage')).toBeTruthy();
    expect(prs.getByText('manual')).toBeTruthy();

    expect(within(group('Duplicates')).getByRole('link', { name: 'Round totals twice' }).getAttribute('href')).toBe('/todos/dup1');
    const related = within(group('Related'));
    expect(related.getByRole('link', { name: 'Audit ledger precision' }).getAttribute('href')).toBe('/todos/rel1');
    expect(related.queryByText('Upgrade decimal library')).toBeNull();
  });

  it.each([
    ['Pull requests', { prLinks: [] }],
    ['Duplicates', { links: links.filter(link => link.relation !== 'duplicates') }],
    ['Related', { links: links.filter(link => link.relation !== 'related_to') }],
  ] as const)('omits the %s group when it has no entries', async (name, override) => {
    Object.assign(backend, override);
    await renderLinks();

    expect(screen.queryByRole('heading', { name })).toBeNull();
    expect(screen.getAllByRole('heading')).toHaveLength(2);
  });

  it('renders nothing when the todo has no links at all', async () => {
    backend = { links: [], prLinks: [] };
    const { container } = await renderLinks();

    expect(container.textContent).toBe('');
  });

  it('opens a linked todo in place through onOpenTodo', async () => {
    const onOpenTodo = vi.fn();
    await renderLinks({ onOpenTodo });

    fireEvent.click(screen.getByRole('link', { name: 'Round totals twice' }));

    expect(onOpenTodo).toHaveBeenCalledExactlyOnceWith('dup1');
  });

  it('unlinks a pull request with DELETE on /api/todos/pr-links and refetches the list', async () => {
    await renderLinks();

    fireEvent.click(screen.getByRole('button', { name: 'Unlink PR #42' }));

    await waitFor(() => {
      const call = fetchMock.mock.calls.find(([, init]) => init?.method === 'DELETE');
      expect(call).toBeTruthy();
      const url = new URL(String(call![0]), 'http://x');
      expect(url.pathname).toBe('/api/todos/pr-links');
      expect(Object.fromEntries(url.searchParams)).toEqual({ dir: DIR, id: REF, url: PR_URL });
    });
    await waitFor(() => expect(fetchMock.mock.calls.filter(([url]) => String(url).startsWith('/api/todos/pr-links') && !String(url).includes('url=')).length).toBeGreaterThan(1));
  });

  it.each([
    ['Duplicates', 'Unlink Round totals twice', { target: 'dup-full-id', relation: 'duplicates' }],
    ['Related', 'Unlink Audit ledger precision', { target: 'rel-full-id', relation: 'related_to' }],
  ] as const)('unlinks a %s entry with DELETE on /api/todos/links', async (_group, buttonName, expected) => {
    await renderLinks();

    fireEvent.click(screen.getByRole('button', { name: buttonName }));

    await waitFor(() => {
      const call = fetchMock.mock.calls.find(([url, init]) => init?.method === 'DELETE' && String(url) === '/api/todos/links');
      expect(call).toBeTruthy();
      expect(JSON.parse(String(call![1].body))).toEqual({ dir: DIR, ref: REF, ...expected });
    });
  });

  it('disables every unlink button while the todo is busy', async () => {
    await renderLinks({ disabled: true });

    for (const button of screen.getAllByRole('button')) expect((button as HTMLButtonElement).disabled).toBe(true);
  });

  it('reports a load failure instead of hiding the links', async () => {
    backend.prLinks = { not: 'a list' };
    await renderLinks();

    expect(screen.getByText(/Failed to load pull request links/)).toBeTruthy();
    expect(screen.getByRole('heading', { name: 'Duplicates' })).toBeTruthy();
  });
});
