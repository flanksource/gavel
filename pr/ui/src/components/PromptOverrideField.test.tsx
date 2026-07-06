import { afterEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { PromptOverrideField } from './PromptOverrideField';
import type { PromptDetail } from './settings/promptSpec';

// The real SpecRuntimeEditor pulls in observers jsdom lacks; stub only it while
// keeping the adapter helpers (compactAISpecRuntime) that promptSpec imports.
vi.mock('@flanksource/clicky-ui/ai', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  SpecRuntimeEditor: () => <div data-testid="spec-editor" />,
}));

const detail: PromptDetail = {
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

describe('PromptOverrideField summary row', () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it('renders source and model badges from the resolved detail', async () => {
    mockFetch(async () => ({ ok: true, json: async () => detail }));
    renderRow();

    expect(await screen.findByText('claude-sonnet')).toBeTruthy();
    expect(screen.getByText('default')).toBeTruthy();
  });

  it('opens the nested editor dialog on Edit', async () => {
    mockFetch(async () => ({ ok: true, json: async () => detail }));
    renderRow();

    fireEvent.click(await screen.findByRole('button', { name: 'Edit' }));
    expect(await screen.findByText(/Edit prompt/)).toBeTruthy();
    expect(screen.getByTestId('spec-editor')).toBeTruthy();
  });

  it('PUTs on save and syncs the in-form value to the persisted override', async () => {
    const saved: PromptDetail = {
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

    fireEvent.click(await screen.findByRole('button', { name: 'Edit' }));
    fireEvent.click(await screen.findByRole('button', { name: 'Save' }));

    await waitFor(() => expect(onChange).toHaveBeenCalledWith({ inline: saved.raw }));
    expect(putUrl).toContain('/api/settings/prompts/verify?scope=global');
  });

  it('surfaces a load error instead of badges', async () => {
    mockFetch(async () => ({ ok: false, json: async () => ({}), text: async () => 'boom' }));
    renderRow();

    expect(await screen.findByText('boom')).toBeTruthy();
  });
});
