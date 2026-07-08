import { afterEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { PromptOverrideField } from './PromptOverrideField';
import type { PromptSpecDetail, PromptSpecSavePayload } from '@flanksource/clicky-ui/ai';

vi.mock('@flanksource/clicky-ui/ai', async () => {
  const React = await import('react');
  return {
    PromptPickerField: (props: {
      title: string;
      loadDetail: () => Promise<PromptSpecDetail>;
      saveDetail: (payload: PromptSpecSavePayload) => Promise<PromptSpecDetail>;
      onChange: (value: unknown) => void;
    }) => {
      const [label, setLabel] = React.useState('Loading prompt...');
      React.useEffect(() => {
        void props.loadDetail()
          .then((detail) => setLabel(`${String(detail.spec.model ?? 'Default model')} · ${detail.body}`))
          .catch((error: Error) => setLabel(error.message));
      }, [props.loadDetail]);
      return (
        <div>
          {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky prompt picker itself. */}
          <button type="button" aria-label={`Edit prompt ${props.title}`}>
            {label}
          </button>
          {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the Clicky prompt picker itself. */}
          <button
            type="button"
            onClick={async () => {
              const next = await props.saveDetail({
                source: 'inline',
                spec: { model: 'new' },
                body: 'body',
                baseRaw: 'base',
              });
              props.onChange(next.source === 'file' ? { file: next.path ?? '' } : { inline: next.raw });
            }}
          >
            Save prompt
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

function mockFetch(impl: (url: string, init?: RequestInit) => Promise<{ ok: boolean; json: () => Promise<unknown>; text?: () => Promise<string> }>) {
  vi.stubGlobal('fetch', vi.fn(impl));
}

function renderRow(onChange = vi.fn()) {
  render(
    <PromptOverrideField
      value={undefined}
      onChange={onChange}
      id="verify"
      title="Verify"
      scopeQuery="scope=global"
    />,
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
    const onChange = renderRow();

    fireEvent.click(await screen.findByRole('button', { name: 'Save prompt' }));

    await waitFor(() => expect(onChange).toHaveBeenCalledWith({ inline: saved.raw }));
    expect(putUrl).toContain('/api/settings/prompts/verify?scope=global');
  });

  it('surfaces a load error instead of badges', async () => {
    mockFetch(async () => ({ ok: false, json: async () => ({}), text: async () => 'boom' }));
    renderRow();

    expect(await screen.findByText('boom')).toBeTruthy();
  });
});
