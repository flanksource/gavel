import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import type { PromptIndicatorProps } from '@flanksource/clicky-ui/data';
import { RunPrompts } from './RunPrompts';
import { connectionLookupFetcher } from '../prompts';

// PromptIndicator's own behaviour (count, panel, dialog, autoOpen) is covered by
// clicky-ui's PromptIndicator spec; vitest here resolves the published
// clicky-ui, so this spec checks only how RunPrompts wires it to the run.
const indicatorProps: PromptIndicatorProps[] = [];
vi.mock('@flanksource/clicky-ui/data', () => ({
  PromptIndicator: (props: PromptIndicatorProps) => {
    indicatorProps.push(props);
    return <div data-testid="prompt-indicator" />;
  },
}));

afterEach(() => {
  cleanup();
  indicatorProps.length = 0;
  vi.unstubAllGlobals();
});

describe('RunPrompts', () => {
  it('mounts one prompt indicator on the run\'s own prompts, opening each pending form on its own', () => {
    render(<RunPrompts owner="run-2" />);

    expect(screen.getAllByTestId('prompt-indicator')).toHaveLength(1);
    expect(indicatorProps.at(-1)).toMatchObject({ basePath: '/api', owner: 'run-2', autoOpen: true });
  });

  it('resolves a form\'s connection picker against the run\'s server', async () => {
    const fetch = vi.fn(async (_url: string) => new Response(JSON.stringify({
      filters: { connection: { options: { lab: { plain: 'lab (sql_server)' } } } },
    }), { status: 200 }));
    vi.stubGlobal('fetch', fetch);
    render(<RunPrompts owner="run-2" />);

    const lookup = indicatorProps.at(-1)?.lookupFetcher;
    expect(lookup).toBeDefined();
    await expect(lookup!({ descriptor: { url: '/api/v1/connection', filter: 'connection' }, query: 'la' }))
      .resolves.toEqual([{ value: 'lab', label: 'lab (sql_server)' }]);
    expect(fetch.mock.calls[0]?.[0]).toBe('/api/v1/connection?__lookup_q=la');
  });

  it('shows nothing for a run with no prompt owner', () => {
    const { container } = render(<RunPrompts owner={undefined} />);
    expect(container.innerHTML).toBe('');
    expect(indicatorProps).toHaveLength(0);
  });
});

describe('connectionLookupFetcher', () => {
  it('keeps the picker\'s ?types= scope, adds the search text and reads the connection options', async () => {
    const fetch = vi.fn(async (_url: string) => new Response(JSON.stringify({
      filters: { connection: { options: { lab: { plain: 'lab (sql_server)' }, 'connection://ops/warehouse': { text: 'ops/warehouse' } } } },
    }), { status: 200 }));
    vi.stubGlobal('fetch', fetch);

    const options = await connectionLookupFetcher()({
      descriptor: { url: '/api/v1/connection?types=sql_server', filter: 'connection', searchParam: '__lookup_q' },
      query: 'la b',
    });

    expect(fetch.mock.calls[0]?.[0]).toBe('/api/v1/connection?types=sql_server&__lookup_q=la+b');
    expect(options).toEqual([
      { value: 'lab', label: 'lab (sql_server)' },
      { value: 'connection://ops/warehouse', label: 'ops/warehouse' },
    ]);
  });

  it('fails loudly when the server refuses the lookup', async () => {
    vi.stubGlobal('fetch', vi.fn(async () => new Response('this server has no connection catalog', { status: 503 })));
    await expect(connectionLookupFetcher()({ descriptor: { url: '/api/v1/connection', filter: 'connection' }, query: '' }))
      .rejects.toThrow('lookup connection: 503 this server has no connection catalog');
  });
});
