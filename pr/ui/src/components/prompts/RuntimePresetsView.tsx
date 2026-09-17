import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  RuntimeProfilesWorkspace,
  type RuntimePreset,
  type RuntimeProfilesPersistence,
  type RuntimeProfilesStore,
  type RuntimeRecordMeta,
  type SpecRuntimeFamily,
  type ToolMeta,
} from '@flanksource/clicky-ui/ai';
import {
  createRuntimePreset,
  deleteRuntimePreset,
  runtimePresetClient,
  runtimePresetLibraryQuery,
  updateRuntimePreset,
  type RuntimeRecordSource,
  type StoredRuntimePreset,
} from './runtimePresetCatalog';

interface Props {
  scopeQuery: string;
  families: SpecRuntimeFamily[];
  tools: ToolMeta[];
  runtimeError?: string;
  selectedId?: string;
  onSelect: (id: string | undefined) => void;
}

type Drafts = Record<string, RuntimePreset>;

export function RuntimePresetsView({ scopeQuery, families, tools, runtimeError, selectedId, onSelect }: Props) {
  const queryClient = useQueryClient();
  const library = useQuery(runtimePresetLibraryQuery(scopeQuery));
  const [drafts, setDrafts] = useState<Drafts>({});
  const [error, setError] = useState<string>();
  const [saving, setSaving] = useState(false);
  const queryKey = runtimePresetLibraryQuery(scopeQuery).queryKey;
  const records = useMemo(
    () => (library.data?.presets ?? []).map(preset => drafts[preset.id] ?? preset),
    [drafts, library.data?.presets],
  );
  const client = useMemo(() => runtimePresetClient(scopeQuery, tools), [scopeQuery, tools]);
  const create = useMutation({ mutationFn: ({ target, preset }: { target: string; preset: RuntimePreset }) => createRuntimePreset(scopeQuery, target, preset) });
  const update = useMutation({ mutationFn: (preset: RuntimePreset) => updateRuntimePreset(scopeQuery, preset) });
  const remove = useMutation({ mutationFn: (preset: StoredRuntimePreset) => deleteRuntimePreset(scopeQuery, preset) });

  useEffect(() => {
    setDrafts({});
    setError(undefined);
  }, [scopeQuery]);

  const run = (action: () => Promise<void>) => {
    setError(undefined);
    void action().catch(cause => setError(messageOf(cause)));
  };
  const loaded = (id: string) => {
    const preset = library.data?.presets.find(item => item.id === id);
    if (!preset) throw new Error(`Runtime preset ${id} is not loaded`);
    return preset;
  };
  const createTarget = () => {
    const eligible = (library.data?.sources ?? []).filter(source => source.writable && source.records.includes('preset'));
    const target = eligible.find(source => source.kind === 'db') ?? eligible.at(-1);
    if (!target) throw new Error('No writable runtime preset source is available');
    return target.id;
  };
  const store: RuntimeProfilesStore = {
    createPreset: preset => run(async () => {
      const created = await create.mutateAsync({ target: createTarget(), preset });
      await queryClient.invalidateQueries({ queryKey });
      onSelect(created.id);
    }),
    updatePreset: preset => {
      const source = loaded(preset.id).source;
      if (!source.writable) {
        setError(`${source.label} is read-only`);
        return;
      }
      setDrafts(current => ({ ...current, [preset.id]: preset }));
    },
    deletePreset: id => run(async () => {
      await remove.mutateAsync(loaded(id));
      setDrafts(current => without(current, id));
      await queryClient.invalidateQueries({ queryKey });
    }),
    createProfile: () => { throw new Error('Runtime profiles are not editable here'); },
    updateProfile: () => { throw new Error('Runtime profiles are not editable here'); },
    deleteProfile: () => { throw new Error('Runtime profiles are not editable here'); },
  };
  const persistence: RuntimeProfilesPersistence = {
    dirty: Object.keys(drafts).length > 0,
    saving,
    ...(error ? { error } : {}),
    onSave: () => run(async () => {
      setSaving(true);
      try {
        for (const preset of Object.values(drafts)) {
          await update.mutateAsync(preset);
          setDrafts(current => without(current, preset.id));
        }
        await queryClient.invalidateQueries({ queryKey });
      } finally {
        setSaving(false);
      }
    }),
    onDiscard: () => {
      setDrafts({});
      setError(undefined);
    },
  };

  if (library.isLoading) return <p className="p-6 text-sm text-muted-foreground">Loading runtime presets…</p>;
  if (library.error) return <p role="alert" className="p-6 text-sm text-destructive">{messageOf(library.error)}</p>;
  if (!library.data) return <p role="alert" className="p-6 text-sm text-destructive">The runtime preset library returned no result.</p>;

  return (
    <div className="h-full min-h-0 overflow-y-auto p-4">
      {runtimeError && <p role="alert" className="mb-4 rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">{runtimeError}</p>}
      <RuntimeProfilesWorkspace
        presets={records}
        profiles={library.data.profiles}
        view="presets"
        onViewChange={() => undefined}
        selectedPresetId={selectedId}
        selectedProfileId={undefined}
        onSelectPreset={onSelect}
        onSelectProfile={() => undefined}
        store={store}
        client={client}
        families={families}
        persistence={persistence}
        recordMeta={id => recordMeta(loaded(id).source)}
      />
    </div>
  );
}

function recordMeta(source: RuntimeRecordSource): RuntimeRecordMeta {
  return { sourceLabel: source.label, writable: source.writable };
}

function without(drafts: Drafts, id: string): Drafts {
  const next = { ...drafts };
  delete next[id];
  return next;
}

function messageOf(value: unknown): string {
  return value instanceof Error ? value.message : String(value);
}
