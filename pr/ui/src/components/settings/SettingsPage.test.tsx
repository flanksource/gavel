import type { ReactNode } from 'react';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { SettingsPage } from './SettingsPage';
import {
  settingsConfigQuery,
  settingsPromptsQuery,
  settingsRunContextQuery,
  settingsSchemaQuery,
  settingsTraceQuery,
} from './queries';

vi.mock('@flanksource/clicky-ui/components', async () => {
  return {
    Button: ({ children, onClick, disabled, ...props }: {
      children: ReactNode;
      onClick?: () => void;
      disabled?: boolean;
    }) => (
      // oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for Clicky's Button.
      <button type="button" onClick={onClick} disabled={disabled} {...props}>{children}</button>
    ),
    Tabs: () => null,
    JsonSchemaForm: ({ value, onChange }: {
      value: Record<string, unknown>;
      onChange: (value: Record<string, unknown>) => void;
    }) => (
      <div>
        <span data-testid="settings-value">{JSON.stringify(value)}</span>
        {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- deterministic control for the schema form mock. */}
        <button type="button" onClick={() => onChange({ model: 'edited' })}>Edit settings</button>
      </div>
    ),
  };
});

vi.mock('@flanksource/clicky-ui/icons', () => ({
  UiFolder: () => null,
  UiGavel: () => null,
  UiChevronRight: () => null,
  UiClose: () => null,
  UiGitBranch: () => null,
}));

vi.mock('../../icons/Spinner', () => ({ Spinner: () => null }));
vi.mock('../../icons/settings', () => ({ sectionIcon: { ai: () => null } }));
vi.mock('../ProjectForm', () => ({
  ProjectFields: () => null,
  useProjectRegistration: () => ({ save: vi.fn(), remove: vi.fn(), saving: false, deleting: false }),
}));
vi.mock('../todos/providers', () => ({ buildRunFamilies: () => [] }));
vi.mock('./models', () => ({ promptModelCatalog: () => [] }));
vi.mock('./extensions', () => ({ buildPre: () => [], buildPost: () => [] }));
vi.mock('./LayerSwitch', () => ({ LayerSwitch: () => null }));
vi.mock('./SettingsSectionCard', () => ({
  SettingsSectionCard: ({ children }: { children: ReactNode }) => <section>{children}</section>,
}));
vi.mock('./SaveBar', () => ({
  SaveBar: ({ dirty, saving, onDiscard, onSave }: {
    dirty: boolean;
    saving: boolean;
    onDiscard: () => void;
    onSave: () => void;
  }) => (
    <footer>
      <span>{dirty ? 'Unsaved changes' : 'Up to date'}</span>
      {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the settings save bar. */}
      <button type="button" onClick={onDiscard} disabled={!dirty || saving}>Discard</button>
      {/* oxlint-disable-next-line clicky-ui/prefer-clicky-components -- test mock for the settings save bar. */}
      <button type="button" onClick={onSave} disabled={!dirty || saving}>Save changes</button>
    </footer>
  ),
}));

function renderSettings(fetchImpl: (url: string, init?: RequestInit) => Promise<unknown>) {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
  client.setQueryData(settingsSchemaQuery().queryKey, {
    type: 'object',
    properties: {
      ai: { type: 'object', properties: { model: { type: 'string' } } },
    },
  });
  client.setQueryData(settingsPromptsQuery().queryKey, {});
  client.setQueryData(settingsRunContextQuery().queryKey, { modes: [], runtimes: [], models: [], efforts: [], tools: [], lifecycle: { steps: [] } });
  client.setQueryData(settingsTraceQuery('scope=global').queryKey, { sources: [], merged: {} });
  client.setQueryData(settingsConfigQuery('scope=global').queryKey, {
    config: { ai: { model: 'initial' } },
    path: '/home/user/.gavel.yaml',
  });
  vi.stubGlobal('fetch', vi.fn(fetchImpl));
  render(
    <QueryClientProvider client={client}>
      <SettingsPage scope={{ kind: 'global' }} repoOptions={[]} onClose={vi.fn()} onSaved={vi.fn()} />
    </QueryClientProvider>,
  );
  return client;
}

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

describe('SettingsPage persistence', () => {
  it('PUTs edits and replaces the form and exact config cache with the server response', async () => {
    const saved = {
      config: { ai: { model: 'server-normalized' } },
      path: '/home/user/.gavel.yaml',
    };
    const client = renderSettings(async (url, init) => {
      if (init?.method === 'PUT') return { ok: true, json: async () => saved };
      if (url === '/api/settings/gavel/trace?scope=global') {
        return { ok: true, json: async () => ({ sources: [], merged: saved.config }) };
      }
      throw new Error(`unexpected request ${url}`);
    });
    expect(await screen.findByText('{"model":"initial"}')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Edit settings' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

    expect(await screen.findByText('{"model":"server-normalized"}')).toBeTruthy();
    expect(screen.getByText('Up to date')).toBeTruthy();
    expect(fetch).toHaveBeenCalledWith('/api/settings/gavel?scope=global', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ai: { model: 'edited' } }),
    });
    expect(client.getQueryData(settingsConfigQuery('scope=global').queryKey)).toEqual(saved);
  });

  it('reloads the current layer before clearing dirty state on discard', async () => {
    renderSettings(async (url, init) => {
      expect(init?.method).toBeUndefined();
      expect(url).toBe('/api/settings/gavel?scope=global');
      return {
        ok: true,
        json: async () => ({
          config: { ai: { model: 'reloaded' } },
          path: '/home/user/.gavel.yaml',
        }),
      };
    });
    expect(await screen.findByText('{"model":"initial"}')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Edit settings' }));
    expect(screen.getByText('Unsaved changes')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Discard' }));

    expect(await screen.findByText('{"model":"reloaded"}')).toBeTruthy();
    expect(screen.getByText('Up to date')).toBeTruthy();
  });

  it('keeps edits dirty and surfaces a contextual PUT error when save fails', async () => {
    const client = renderSettings(async (_url, init) => {
      expect(init?.method).toBe('PUT');
      return {
        ok: false,
        status: 422,
        text: async () => 'invalid model policy',
      };
    });
    expect(await screen.findByText('{"model":"initial"}')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Edit settings' }));

    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

    expect((await screen.findByRole('alert')).textContent).toContain('Failed to save global settings: invalid model policy');
    expect(screen.getByText('Unsaved changes')).toBeTruthy();
    expect(client.getQueryData(settingsConfigQuery('scope=global').queryKey)).toEqual({
      config: { ai: { model: 'initial' } },
      path: '/home/user/.gavel.yaml',
    });
  });

  it('keeps edits dirty and surfaces the scoped reload error when discard fails', async () => {
    renderSettings(async () => ({
      ok: false,
      status: 500,
      text: async () => 'disk unavailable',
    }));
    expect(await screen.findByText('{"model":"initial"}')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Edit settings' }));

    fireEvent.click(screen.getByRole('button', { name: 'Discard' }));

    expect((await screen.findByRole('alert')).textContent).toContain('Failed to load global settings: disk unavailable');
    expect(screen.getByText('Unsaved changes')).toBeTruthy();
    await waitFor(() => expect((screen.getByRole('button', { name: 'Discard' }) as HTMLButtonElement).disabled).toBe(false));
  });
});
