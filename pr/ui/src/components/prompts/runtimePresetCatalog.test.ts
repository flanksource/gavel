import { afterEach, describe, expect, it, vi } from 'vitest';
import type { ResolvedRuntimeProfile, RuntimePreset } from '@flanksource/clicky-ui/ai';
import { runtimePresetClient, type RuntimeRecordSource, type StoredRuntimePreset } from './runtimePresetCatalog';

const preset: RuntimePreset = {
  id: 'review',
  name: 'Review',
  scope: 'context',
  spec: {},
};

const resolvedPreset = { ...preset, presets: [] };

const resolved: ResolvedRuntimeProfile = {
  resolved: { spec: {}, constraints: {}, trace: [] },
  tools: [],
  permissions: {},
  permissionSupport: {},
  effectivePolicy: [],
};

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('runtimePresetClient', () => {
  it('supports both published profile resolution and linked preset resolution', async () => {
    const bodies: unknown[] = [];
    vi.stubGlobal('fetch', vi.fn(async (_url: string, init?: RequestInit) => {
      bodies.push(JSON.parse(String(init?.body)) as unknown);
      return { ok: true, json: async () => resolved };
    }));
    const client = runtimePresetClient('scope=global', []);

    await client.resolve({
      profile: { id: 'profile', name: 'Profile', spec: {}, presets: [preset.id] },
      presets: [preset],
    });
    expect('resolvePresets' in client).toBe(true);
    if (!('resolvePresets' in client) || typeof client.resolvePresets !== 'function') {
      throw new Error('runtime preset client does not support linked preset resolution');
    }
    await client.resolvePresets({
      selected: [preset.id],
      presets: [{
        ...preset,
        key: 'stored-key',
        source: { kind: 'db', id: 'db', label: 'Database', writable: true, records: ['preset'] },
        updatedAt: '2026-09-15T10:00:00Z',
      }],
    });

    expect(bodies).toEqual([
      { selected: [preset.id], presets: [resolvedPreset] },
      { selected: [preset.id], presets: [resolvedPreset] },
    ]);
  });

  it('resolves a builtin-source preset and keeps its read-only label', async () => {
    const builtinSource: RuntimeRecordSource = {
      kind: 'builtin',
      id: 'builtin',
      label: 'Built-in',
      writable: false,
      records: ['preset'],
    };
    const builtinPreset: StoredRuntimePreset = {
      ...preset,
      id: 'plan',
      name: 'Plan',
      key: 'plan',
      source: builtinSource,
      updatedAt: '2026-09-15T10:00:00Z',
    };
    const bodies: unknown[] = [];
    vi.stubGlobal('fetch', vi.fn(async (_url: string, init?: RequestInit) => {
      bodies.push(JSON.parse(String(init?.body)) as unknown);
      return { ok: true, json: async () => resolved };
    }));
    const client = runtimePresetClient('scope=global', []);

    if (!('resolvePresets' in client) || typeof client.resolvePresets !== 'function') {
      throw new Error('runtime preset client does not support linked preset resolution');
    }
    await client.resolvePresets({
      selected: [builtinPreset.id],
      presets: [builtinPreset],
    });

    expect(builtinPreset.source.kind).toBe('builtin');
    expect(builtinPreset.source.label).toBe('Built-in');
    expect(builtinPreset.source.writable).toBe(false);
    expect(bodies).toEqual([
      { selected: [builtinPreset.id], presets: [{ ...preset, id: 'plan', name: 'Plan', presets: [] }] },
    ]);
  });
});
