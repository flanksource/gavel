import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { PRSearchItem } from '../../types';
import { LinkPRDialog } from './LinkPRDialog';
import { queryTestWrapper } from './queryTestWrapper';

const DIR = '/repos/billing';
const REF = 'abc123';

const searchResults: PRSearchItem[] = [
  { provider: 'github', repo: 'acme/ledger', number: 42, url: 'https://github.com/acme/ledger/pull/42', title: 'Rework ledger rounding', state: 'OPEN', author: 'octocat', isDraft: false, updatedAt: '2026-10-01T00:00:00Z' },
  { provider: 'github', repo: 'acme/ledger', number: 9, url: 'https://github.com/acme/ledger/pull/9', title: 'Ledger index', state: 'MERGED', author: 'hubot', isDraft: false, updatedAt: '2026-09-01T00:00:00Z' },
];

let fetchMock: ReturnType<typeof vi.fn>;
let postStatus: number;
let onClose: ReturnType<typeof vi.fn>;

function respond(body: unknown, status = 200) {
  return Promise.resolve({
    ok: status < 400,
    status,
    json: async () => body,
    text: async () => JSON.stringify(body),
    clone() { return this; },
  } as unknown as Response);
}

const searchCalls = () => fetchMock.mock.calls.map(([url]) => String(url)).filter(url => url.startsWith('/api/prs/search'));
const postCall = () => fetchMock.mock.calls.find(([, init]) => init?.method === 'POST');

beforeEach(() => {
  postStatus = 201;
  onClose = vi.fn();
  fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (init?.method === 'POST') return postStatus === 201 ? respond(searchResults[0]!, 201) : respond({ error: 'not a pull request URL' }, postStatus);
    if (url.startsWith('/api/prs/search')) return respond(searchResults);
    return respond([]);
  });
  vi.stubGlobal('fetch', fetchMock);
});

afterEach(() => {
  vi.unstubAllGlobals();
});

async function renderDialog(open = true) {
  render(<LinkPRDialog open={open} dir={DIR} todoRef={REF} onClose={onClose} />, { wrapper: queryTestWrapper() });
  await act(async () => {});
}

describe('LinkPRDialog', () => {
  it('renders nothing and searches nothing while closed', async () => {
    await renderDialog(false);

    expect(screen.queryByText('Link pull request')).toBeNull();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('lists recent pull requests on open, then searches once the typing settles', async () => {
    await renderDialog();
    await screen.findByText('Rework ledger rounding');
    expect(searchCalls()).toEqual(['/api/prs/search?limit=20']);

    const search = screen.getByLabelText('Search pull requests');
    fireEvent.change(search, { target: { value: 'l' } });
    fireEvent.change(search, { target: { value: 'ledger' } });

    await waitFor(() => expect(searchCalls()).toContain('/api/prs/search?q=ledger&limit=20'));
    expect(searchCalls().some(url => url.includes('q=l&'))).toBe(false);
  });

  it('links the picked pull request as implementing the todo', async () => {
    await renderDialog();
    fireEvent.click(await screen.findByRole('button', { name: /Rework ledger rounding/ }));

    expect((screen.getByLabelText('Pull request URL') as HTMLInputElement).value).toBe(searchResults[0]!.url);
    fireEvent.click(screen.getByRole('button', { name: 'Link PR' }));

    await waitFor(() => expect(onClose).toHaveBeenCalledOnce());
    const [url, init] = postCall()!;
    expect(url).toBe('/api/todos/pr-links');
    expect(JSON.parse(String(init.body))).toEqual({ dir: DIR, id: REF, url: searchResults[0]!.url, relation: 'implements' });
  });

  it('links a pasted URL with the chosen relation', async () => {
    await renderDialog();
    const pasted = 'https://github.com/acme/ledger/pull/77';
    fireEvent.change(screen.getByLabelText('Pull request URL'), { target: { value: ` ${pasted} ` } });
    fireEvent.change(screen.getByLabelText('Relation'), { target: { value: 'related' } });
    fireEvent.click(screen.getByRole('button', { name: 'Link PR' }));

    await waitFor(() => expect(onClose).toHaveBeenCalledOnce());
    expect(JSON.parse(String(postCall()![1].body))).toEqual({ dir: DIR, id: REF, url: pasted, relation: 'related' });
  });

  it('cannot link until a URL is entered', async () => {
    await renderDialog();

    expect((screen.getByRole('button', { name: 'Link PR' }) as HTMLButtonElement).disabled).toBe(true);
  });

  it('keeps the dialog open and shows the server rejection', async () => {
    postStatus = 400;
    await renderDialog();
    fireEvent.change(screen.getByLabelText('Pull request URL'), { target: { value: 'https://example.com/issues/1' } });
    fireEvent.click(screen.getByRole('button', { name: 'Link PR' }));

    expect(await screen.findByText(/not a pull request URL/)).toBeTruthy();
    expect(onClose).not.toHaveBeenCalled();
  });
});
