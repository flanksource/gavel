import { queryOptions } from '@tanstack/react-query';
import type {
  ResolvedRuntimeProfile,
  RuntimePreset,
  RuntimeProfile,
  RuntimeProfileResolveRequest,
  RuntimeProfileScope,
  RuntimeProfilesClient,
  RuntimePresetSpec,
  ToolMeta,
} from '@flanksource/clicky-ui/ai';
import { fetchJSON, responseError } from '../oneShotQueries';

export type RuntimeRecordSource = {
  kind: 'db' | 'file' | 'builtin';
  id: string;
  label: string;
  root?: string;
  writable: boolean;
  implicit?: boolean;
  records: Array<'preset' | 'profile'>;
};

type StoredRecord = {
  key: string;
  source: RuntimeRecordSource;
  updatedAt: string;
};

export type RuntimePresetRecord = RuntimePreset & { presets?: string[] };
export type StoredRuntimePreset = RuntimePresetRecord & StoredRecord;
export type StoredRuntimeProfile = RuntimeProfile & StoredRecord;

export type RuntimePresetLibrary = {
  presets: StoredRuntimePreset[];
  profiles: StoredRuntimeProfile[];
  sources: RuntimeRecordSource[];
};

export type RuntimePresetWrite = {
  name: string;
  description?: string;
  scope: RuntimeProfileScope;
  spec: RuntimePresetSpec;
  presets: string[];
};

type RuntimePresetResolveRequestCompat = {
  selected: string[];
  presets: RuntimePresetRecord[];
};

type RuntimeProfilesClientCompat = RuntimeProfilesClient & {
  resolvePresets: (
    request: RuntimePresetResolveRequestCompat,
    signal?: AbortSignal,
  ) => Promise<ResolvedRuntimeProfile>;
};

export function runtimePresetLibraryQuery(scopeQuery: string) {
  return queryOptions({
    queryKey: ['settings', 'runtime-presets', scopeQuery] as const,
    queryFn: ({ signal }) => fetchJSON<RuntimePresetLibrary>(
      `/api/settings/runtime-presets?${scopeQuery}`,
      signal,
      'Failed to load runtime presets',
    ),
    staleTime: 30_000,
  });
}

export function presetWrite(preset: RuntimePresetRecord): RuntimePresetWrite {
  return {
    name: preset.name,
    ...(preset.description !== undefined ? { description: preset.description } : {}),
    scope: preset.scope,
    spec: preset.spec,
    presets: preset.presets ?? [],
  };
}

export function createRuntimePreset(scopeQuery: string, target: string, preset: RuntimePreset) {
  return writeRuntimePreset(
    'POST',
    `/api/settings/runtime-presets?${scopeQuery}`,
    { target, ...presetWrite(preset) },
    `Failed to create runtime preset ${preset.name}`,
  );
}

export function updateRuntimePreset(scopeQuery: string, preset: RuntimePreset) {
  return writeRuntimePreset(
    'PUT',
    runtimePresetURL(scopeQuery, preset.id),
    presetWrite(preset),
    `Failed to update runtime preset ${preset.name}`,
  );
}

export async function deleteRuntimePreset(scopeQuery: string, preset: RuntimePreset): Promise<void> {
  const url = runtimePresetURL(scopeQuery, preset.id);
  const response = await fetch(url, { method: 'DELETE', headers: { Accept: 'application/json' } });
  if (!response.ok) throw await responseError(response, `Failed to delete runtime preset ${preset.name}`);
}

export function runtimePresetClient(scopeQuery: string, tools: ToolMeta[]): RuntimeProfilesClient {
  const client: RuntimeProfilesClientCompat = {
    resolvePresets: (request, signal) => resolveRuntimePresets(scopeQuery, request, signal),
    resolve: (request, signal) => resolveRuntimePresets(scopeQuery, request, signal),
    loadPermissionCatalog: async () => ({
      tools: tools.map(tool => ({ id: tool.name, label: tool.label, group: tool.group })),
    }),
  };
  return client;
}

async function resolveRuntimePresets(
  scopeQuery: string,
  request: RuntimePresetResolveRequestCompat | RuntimeProfileResolveRequest,
  signal?: AbortSignal,
): Promise<ResolvedRuntimeProfile> {
  const resolvedRequest = 'selected' in request
    ? request
    : { selected: request.profile.presets, presets: request.presets };
  const response = await fetch(`/api/settings/runtime-presets/resolve?${scopeQuery}`, {
    method: 'POST',
    headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
    body: JSON.stringify({
      selected: resolvedRequest.selected,
      presets: resolvedRequest.presets.map(preset => ({ id: preset.id, ...presetWrite(preset) })),
    }),
    ...(signal ? { signal } : {}),
  });
  if (!response.ok) throw await responseError(response, 'Failed to resolve runtime presets');
  return await response.json() as ResolvedRuntimeProfile;
}

async function writeRuntimePreset(
  method: 'POST' | 'PUT',
  url: string,
  body: object,
  context: string,
): Promise<StoredRuntimePreset> {
  const response = await fetch(url, {
    method,
    headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!response.ok) throw await responseError(response, context);
  return await response.json() as StoredRuntimePreset;
}

function runtimePresetURL(scopeQuery: string, id: string): string {
  return `/api/settings/runtime-presets/${encodeURIComponent(id)}?${scopeQuery}`;
}
