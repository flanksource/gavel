import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor, within } from '@testing-library/react';

import type { TodoTagDef } from '../../../types';
import { TagDefinitionsPanel } from './TagDefinitionsPanel';

const definitions: TodoTagDef[] = [
  { name: 'flaky', color: 'amber', scope: 'workspace' },
  { name: 'bug', color: 'red', scope: 'builtin' },
  { name: 'docs', color: 'sky', scope: 'builtin' },
];
const counts: Record<string, number> = { flaky: 12, bug: 3 };

let requests: { url: string; method: string; body?: string }[] = [];

function mockFetch(
  removal = { name: 'flaky', definition: true, todos: 12 },
  loadedDefinitions = definitions,
) {
  return vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    const method = init?.method ?? 'GET';
    requests.push({ url, method, body: init?.body as string | undefined });
    const payload = method === 'DELETE'
      ? { definitions: loadedDefinitions, counts, removed: removal }
      : { definitions: loadedDefinitions, counts };
    return new Response(JSON.stringify(payload), {
      status: 200,
      headers: { 'Content-Type': 'application/json' },
    });
  });
}

function renderPanel() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <TagDefinitionsPanel dir="/work/gavel" />
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  requests = [];
  vi.stubGlobal('fetch', mockFetch());
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe('TagDefinitionsPanel', () => {
  it('lists the project tags it loaded', async () => {
    renderPanel();
    expect(await screen.findByText('flaky')).toBeTruthy();
  });

  it('updates a tag icon through an icon picker button', async () => {
    renderPanel();

    const trigger = await screen.findByRole('button', { name: 'Tag icon for flaky' });
    expect(screen.queryByRole('combobox', { name: 'Tag icon for flaky' })).toBeNull();

    fireEvent.click(trigger);
    const picker = within(screen.getByRole('menu', { name: 'Tag icon for flaky' }));
    expect(picker.getByRole('menuitemradio', { name: /No icon/i }).getAttribute('aria-checked')).toBe('true');
    fireEvent.click(picker.getByRole('menuitemradio', { name: /debug/i }));
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(requests.some(request => request.method === 'POST')).toBe(true));
    const save = requests.find(request => request.method === 'POST');
    expect(JSON.parse(save?.body ?? '{}')).toMatchObject({ name: 'flaky', icon: 'debug' });
  });

  it('keeps an existing custom icon available until it is explicitly replaced', async () => {
    const customDefinitions: TodoTagDef[] = [
      { ...definitions[0], icon: 'custom-registry-key', iconify: 'ph:alien' },
      ...definitions.slice(1),
    ];
    vi.stubGlobal('fetch', mockFetch(undefined, customDefinitions));
    renderPanel();

    fireEvent.click(await screen.findByRole('button', { name: 'Tag icon for flaky' }));
    const custom = within(screen.getByRole('menu', { name: 'Tag icon for flaky' }))
      .getByRole('menuitemradio', { name: /custom-registry-key/i });
    expect(custom.getAttribute('aria-checked')).toBe('true');

    fireEvent.change(screen.getAllByLabelText('Tag description')[0], { target: { value: 'Still custom' } });
    fireEvent.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => expect(requests.some(request => request.method === 'POST')).toBe(true));
    const save = requests.find(request => request.method === 'POST');
    expect(JSON.parse(save?.body ?? '{}')).toMatchObject({
      name: 'flaky',
      icon: 'custom-registry-key',
      description: 'Still custom',
    });
  });

  // Removing a project tag rewrites todos, so it must never happen on a stray
  // click — the count in the dialog is the whole point of asking.
  it('asks before removing a tag that todos are using', async () => {
    renderPanel();
    await screen.findByLabelText('Remove flaky');

    fireEvent.click(screen.getByLabelText('Remove flaky'));

    const dialog = await screen.findByRole('dialog');
    expect(dialog.textContent).toContain('12');
    expect(dialog.textContent).toContain('todos');
    expect(requests.some(request => request.method === 'DELETE')).toBe(false);
  });

  it('does not remove the tag when the dialog is cancelled', async () => {
    renderPanel();
    await screen.findByLabelText('Remove flaky');

    fireEvent.click(screen.getByLabelText('Remove flaky'));
    fireEvent.click(within(await screen.findByRole('dialog')).getByText('Cancel'));

    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    expect(requests.some(request => request.method === 'DELETE')).toBe(false);
  });

  it('removes the tag once the dialog is confirmed', async () => {
    renderPanel();
    await screen.findByLabelText('Remove flaky');

    fireEvent.click(screen.getByLabelText('Remove flaky'));
    fireEvent.click(within(await screen.findByRole('dialog')).getByText('Remove'));

    await waitFor(() => expect(requests.some(request => request.method === 'DELETE')).toBe(true));
    const deletion = requests.find(request => request.method === 'DELETE');
    expect(deletion?.url).toContain('name=flaky');
  });

  it('reports how many todos the removal touched', async () => {
    renderPanel();
    await screen.findByLabelText('Remove flaky');

    fireEvent.click(screen.getByLabelText('Remove flaky'));
    fireEvent.click(within(await screen.findByRole('dialog')).getByText('Remove'));

    expect((await screen.findByRole('status')).textContent).toContain('12 todos');
  });

  // A tag no todo carries has nothing to strip, so the dialog would be noise.
  it('removes an unused tag without asking', async () => {
    renderPanel();
    fireEvent.click(await screen.findByLabelText('Remove docs'));

    expect(screen.queryByRole('dialog')).toBeNull();
    await waitFor(() => expect(requests.some(request => request.method === 'DELETE')).toBe(true));
  });

  it('adopts a common tag for this project in one click', async () => {
    renderPanel();
    await screen.findByLabelText('Remove flaky');

    fireEvent.click(screen.getByText('Add common tag'));
    const picker = within(await screen.findByRole('dialog', { name: 'Choose a tag' }));
    fireEvent.click(picker.getByRole('option', { name: /bug/ }));

    await waitFor(() => expect(requests.some(request => request.method === 'POST')).toBe(true));
    const save = requests.find(request => request.method === 'POST');
    const body = JSON.parse(save?.body ?? '{}');
    expect(body.name).toBe('bug');
    // Adoption must not repaint: the stored row keeps the colour the built-in
    // was already rendering with.
    expect(body.color).toBe('red');
  });

  // The backfill adopts what a project already used; anything typed since is
  // adopted the same way, from here.
  it('offers an in-use tag that nothing has defined yet', async () => {
    vi.stubGlobal('fetch', vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      requests.push({ url: String(input), method: init?.method ?? 'GET', body: init?.body as string });
      return new Response(
        JSON.stringify({ definitions, counts: { ...counts, 'typed-by-hand': 5 } }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      );
    }));
    renderPanel();
    await screen.findByLabelText('Remove flaky');

    fireEvent.click(screen.getByText('Add common tag'));

    const picker = within(await screen.findByRole('dialog', { name: 'Choose a tag' }));
    const offered = picker.getAllByRole('option')
      .map(option => option.querySelector('[data-tag]')?.getAttribute('data-tag'));
    expect(offered).toContain('typed-by-hand');
  });

  it('offers only tags this project has not defined itself', async () => {
    renderPanel();
    await screen.findByLabelText('Remove flaky');

    fireEvent.click(screen.getByText('Add common tag'));

    const picker = within(await screen.findByRole('dialog', { name: 'Choose a tag' }));
    const offered = picker.getAllByRole('option')
      .map(option => option.querySelector('[data-tag]')?.getAttribute('data-tag'));
    expect(offered).toContain('bug');
    expect(offered).not.toContain('flaky');
  });
});
