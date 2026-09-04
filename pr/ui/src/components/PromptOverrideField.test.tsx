import { afterEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { PromptOverrideField } from './PromptOverrideField';
import type { PromptSpecDetail, PromptSpecSavePayload, SpecRuntimeFamily } from '@flanksource/clicky-ui/ai';
import type { ChatModel } from '@flanksource/clicky-ui/chat';

vi.mock('@flanksource/clicky-ui/ai', async () => {
  const React = await import('react');
  return {
    PromptPickerField: (props: {
      title: string;
      loadDetail: () => Promise<PromptSpecDetail>;
      saveDetail: (payload: PromptSpecSavePayload) => Promise<PromptSpecDetail>;
      onChange: (value: unknown) => void;
      models?: ChatModel[];
      families?: SpecRuntimeFamily[];
    }) => {
      const [label, setLabel] = React.useState('Loading prompt...');
      React.useEffect(() => {
        void props.loadDetail()
          .then((detail) =>
            setLabel(
              detail.parseError
                ? `parse error: ${detail.parseError}`
                : `${String(detail.spec?.model ?? 'Default model')} · ${detail.body}`,
            ),
          )
          .catch((error: Error) => setLabel(error.message));
      }, [props.loadDetail]);
      return (
        <div>
          {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky prompt picker itself. */}
          <button type="button" aria-label={`Edit prompt ${props.title}`}>
            {label}
          </button>
          <span data-testid="model-count">{props.models?.length ?? 0}</span>
          <span data-testid="family-count">{props.families?.length ?? 0}</span>
          {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky prompt picker itself. */}
          <button
            type="button"
            onClick={async () => {
              try {
                const next = await props.saveDetail({
                  source: 'inline',
                  spec: { model: 'new' },
                  body: 'body',
                  baseRaw: 'base',
                });
                props.onChange(next.source === 'file' ? { file: next.path ?? '' } : { inline: next.raw });
              } catch (error) {
                setLabel(error instanceof Error ? error.message : 'Failed to save prompt');
              }
            }}
          >
            Save prompt
          </button>
          {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky prompt picker itself. */}
          <button
            type="button"
            onClick={async () => {
              const next = await props.saveDetail({ source: 'inline', raw: 'fixed', baseRaw: 'bad' });
              props.onChange(next.source === 'file' ? { file: next.path ?? '' } : { inline: next.raw });
            }}
          >
            Repair prompt
          </button>
        </div>
      );
    },
  };
});

const detail: PromptSpecDetail = {
  id: 'verify',
  scope: 'global',
  source: 'default',
  spec: { model: 'claude-sonnet' },
  body: 'Review {{diff}}.',
  raw: 'Review {{diff}}.',
};

function mockFetch(impl: (url: string, init?: RequestInit) => Promise<{ ok: boolean; status?: number; json: () => Promise<unknown>; text?: () => Promise<string> }>) {
  vi.stubGlobal('fetch', vi.fn(impl));
}

function renderRow(
  onChange = vi.fn(),
  models?: ChatModel[],
  families?: SpecRuntimeFamily[],
  queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } }),
) {
  render(
    <QueryClientProvider client={queryClient}>
      <PromptOverrideField
        value={undefined}
        onChange={onChange}
        id="verify"
        title="Verify"
        scopeQuery="scope=global"
        models={models}
        families={families}
      />
    </QueryClientProvider>,
  );
  return onChange;
}

describe('PromptOverrideField adapter', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it('passes the resolved prompt detail into the shared one-line picker', async () => {
    mockFetch(async () => ({ ok: true, json: async () => detail }));
    renderRow();

    expect(await screen.findByText('claude-sonnet · Review {{diff}}.')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Edit prompt Verify' })).toBeTruthy();
  });

  it('passes the model catalog to the shared picker', async () => {
    mockFetch(async () => ({ ok: true, json: async () => detail }));

    renderRow(vi.fn(), [
      { id: 'claude-sonnet', provider: 'anthropic', label: 'Sonnet', reasoning: true, configured: true },
      { id: 'gpt-5-codex', provider: 'openai', label: 'GPT-5 Codex', reasoning: true, configured: true },
    ]);

    expect((await screen.findByTestId('model-count')).textContent).toBe('2');
  });

  it('passes runtime families to the shared picker', async () => {
    mockFetch(async () => ({ ok: true, json: async () => detail }));

    renderRow(vi.fn(), undefined, [
      {
        id: 'claude',
        label: 'Claude',
        provider: 'anthropic',
        modes: [{ id: 'agent', label: 'Agent' }],
      },
    ]);

    expect((await screen.findByTestId('family-count')).textContent).toBe('1');
  });

  it('PUTs on save and syncs the in-form value to the persisted override', async () => {
    const saved: PromptSpecDetail = {
      ...detail,
      source: 'inline',
      raw: '---\nmodel: new\n---\nbody',
      spec: { model: 'new' },
    };
    let putUrl = '';
    mockFetch(async (url, init) => {
      if (init?.method === 'PUT') {
        putUrl = url;
        return { ok: true, json: async () => saved };
      }
      return { ok: true, json: async () => detail };
    });
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const onChange = renderRow(vi.fn(), undefined, undefined, queryClient);

    fireEvent.click(await screen.findByRole('button', { name: 'Save prompt' }));

    await waitFor(() => expect(onChange).toHaveBeenCalledWith({ inline: saved.raw }));
    expect(putUrl).toContain('/api/settings/prompts/verify?scope=global');
    expect(queryClient.getQueryData(['settings', 'prompt-detail', 'verify', 'scope=global'])).toEqual(saved);
  });

  it('surfaces a contextual PUT error without changing the in-form override', async () => {
    mockFetch(async (_url, init) => {
      if (init?.method === 'PUT') {
        return {
          ok: false,
          status: 400,
          json: async () => ({}),
          text: async () => '{"error":"invalid prompt body"}',
        };
      }
      return { ok: true, json: async () => detail };
    });
    const onChange = renderRow();

    fireEvent.click(await screen.findByRole('button', { name: 'Save prompt' }));

    expect(await screen.findByText('Failed to save prompt verify: invalid prompt body')).toBeTruthy();
    expect(onChange).not.toHaveBeenCalled();
  });

  it('surfaces a load error instead of badges', async () => {
    mockFetch(async () => ({ ok: false, status: 500, json: async () => ({}), text: async () => 'boom' }));
    renderRow();

    expect(await screen.findByText('Failed to load prompt verify: boom')).toBeTruthy();
  });

  it('delivers a 200 parseError detail to the shared picker', async () => {
    const invalid: PromptSpecDetail = {
      id: 'verify',
      scope: 'global',
      source: 'inline',
      parseError: 'parse prompt frontmatter: yaml: line 2: could not find expected \':\'',
      raw: '---\nmodel: [broken\n---\nbody\n',
    };
    mockFetch(async () => ({ ok: true, json: async () => invalid }));
    renderRow();

    expect(await screen.findByText(`parse error: ${invalid.parseError}`)).toBeTruthy();
  });

  it('forwards a raw-repair PUT payload and syncs the persisted override', async () => {
    const saved: PromptSpecDetail = {
      id: 'verify',
      scope: 'global',
      source: 'inline',
      spec: { model: 'fixed' },
      body: 'ok',
      raw: '---\nmodel: fixed\n---\nok',
    };
    let putBody = '';
    mockFetch(async (url, init) => {
      if (init?.method === 'PUT') {
        putBody = String(init.body);
        return { ok: true, json: async () => saved };
      }
      return { ok: true, json: async () => detail };
    });
    const onChange = renderRow();

    fireEvent.click(await screen.findByRole('button', { name: 'Repair prompt' }));

    await waitFor(() => expect(onChange).toHaveBeenCalledWith({ inline: saved.raw }));
    expect(JSON.parse(putBody)).toEqual({ source: 'inline', raw: 'fixed', baseRaw: 'bad' });
  });

  it('normalizes a JSON error body into its message', async () => {
    mockFetch(async () => ({
      ok: false,
      status: 400,
      json: async () => ({}),
      text: async () => '{"error":"parse prompt frontmatter: line 13"}',
    }));
    renderRow();

    expect(await screen.findByText('Failed to load prompt verify: parse prompt frontmatter: line 13')).toBeTruthy();
  });

  it('falls back to a status message for an empty error body', async () => {
    mockFetch(async () => ({ ok: false, status: 500, json: async () => ({}), text: async () => '' }));
    renderRow();

    expect(await screen.findByText('Failed to load prompt verify (500)')).toBeTruthy();
  });

  it('reuses a cached prompt detail across picker remounts', async () => {
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, json: async () => detail });
    vi.stubGlobal('fetch', fetchMock);
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const first = render(
      <QueryClientProvider client={queryClient}>
        <PromptOverrideField value={undefined} onChange={vi.fn()} id="verify" title="Verify" scopeQuery="scope=global" />
      </QueryClientProvider>,
    );
    await screen.findByText('claude-sonnet · Review {{diff}}.');
    first.unmount();

    renderRow(vi.fn(), undefined, undefined, queryClient);
    await screen.findByText('claude-sonnet · Review {{diff}}.');

    expect(fetchMock).toHaveBeenCalledOnce();
  });
});
