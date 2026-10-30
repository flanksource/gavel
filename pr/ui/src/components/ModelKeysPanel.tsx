// Manages the organization's model API keys, which its sandbox runs use through substrate's
// egress gateway. Shown as a page in Clerk's organization profile; only admins change keys.
import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@flanksource/clicky-ui/components';
import { UiTrash } from '@flanksource/clicky-ui/icons';
import { fetchJSON, mutationJSON, queryKeys } from '../query';

export interface ModelKeyKind {
  name: string;
  provider: string;
  label: string;
}

export interface ModelKeySummary {
  provider: string;
  kind: string;
  hint: string;
  updatedAt: string;
  updatedBy: string | null;
}

interface ModelKeys {
  kinds: ModelKeyKind[];
  keys: ModelKeySummary[];
  canEdit: boolean;
}

const inputClass = 'w-full rounded border border-border bg-background px-2 py-1 text-sm';

const providerName = (provider: string) => provider.charAt(0).toUpperCase() + provider.slice(1);

export function ModelKeysPanel() {
  const modelKeys = useQuery({
    queryKey: queryKeys.modelKeys(),
    queryFn: ({ signal }) => fetchJSON<ModelKeys>({ url: '/api/model-keys', signal, context: 'load model keys' }),
  });
  const data = modelKeys.data;
  const providers = [...new Set((data?.kinds ?? []).map(kind => kind.provider))];

  return (
    <div className="flex flex-col gap-4 text-sm">
      <p className="text-muted-foreground">
        Agent runs in this organization call model APIs with these keys. A key never enters a run's
        sandbox: substrate's egress gateway adds it to the run's requests.
      </p>
      {data && !data.canEdit && (
        <p className="text-muted-foreground">Only an organization admin can change model keys; ask one to set them.</p>
      )}
      {data && providers.map(provider => (
        <ProviderKey
          key={provider}
          provider={provider}
          kinds={data.kinds.filter(kind => kind.provider === provider)}
          current={data.keys.find(key => key.provider === provider)}
          canEdit={data.canEdit}
        />
      ))}
      {modelKeys.error && <p className="text-destructive">{modelKeys.error instanceof Error ? modelKeys.error.message : String(modelKeys.error)}</p>}
    </div>
  );
}

function ProviderKey({ provider, kinds, current, canEdit }: {
  provider: string;
  kinds: ModelKeyKind[];
  current?: ModelKeySummary;
  canEdit: boolean;
}) {
  const queryClient = useQueryClient();
  const name = providerName(provider);
  const [kind, setKind] = useState(current?.kind ?? kinds[0]?.name ?? '');
  const [value, setValue] = useState('');
  const refresh = () => queryClient.invalidateQueries({ queryKey: queryKeys.modelKeys() });
  const save = useMutation({
    mutationFn: () => mutationJSON<ModelKeySummary>({
      url: `/api/model-keys/${encodeURIComponent(provider)}`, method: 'PUT', body: { kind, key: value }, context: `save ${name} key`,
    }),
    onSuccess: () => {
      setValue('');
      return refresh();
    },
  });
  const remove = useMutation({
    mutationFn: () => mutationJSON<void>({ url: `/api/model-keys/${encodeURIComponent(provider)}`, method: 'DELETE', context: `remove ${name} key` }),
    onSuccess: refresh,
  });
  const error = save.error ?? remove.error;
  const currentLabel = kinds.find(k => k.name === current?.kind)?.label ?? current?.kind;

  return (
    <section className="flex flex-col gap-2 rounded border border-border px-3 py-2">
      <div className="flex items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="font-medium">{name}</div>
          {current ? (
            <>
              <div className="text-xs">{`${currentLabel} ending in ${current.hint}`}</div>
              <div className="text-xs text-muted-foreground">Updated {new Date(current.updatedAt).toLocaleString()}</div>
            </>
          ) : (
            <div className="text-xs text-muted-foreground">Not set</div>
          )}
        </div>
        {canEdit && current && (
          <Button
            variant="ghost"
            size="icon"
            type="button"
            aria-label={`Remove ${name} key`}
            disabled={remove.isPending}
            onClick={() => remove.mutate()}
          >
            <UiTrash />
          </Button>
        )}
      </div>
      {canEdit && (
        <form
          className="flex flex-col gap-2"
          onSubmit={event => {
            event.preventDefault();
            save.mutate();
          }}
        >
          <label className="flex flex-col gap-1">
            <span>Kind</span>
            <select className={inputClass} aria-label={`${name} key kind`} value={kind} onChange={event => setKind(event.target.value)}>
              {kinds.map(k => <option key={k.name} value={k.name}>{k.label}</option>)}
            </select>
          </label>
          <label className="flex flex-col gap-1">
            <span>Key</span>
            <input
              className={`${inputClass} font-mono`}
              type="password"
              autoComplete="off"
              aria-label={`${name} key`}
              value={value}
              onChange={event => setValue(event.target.value)}
              placeholder={current ? 'Paste a key to replace it' : 'Paste the key'}
            />
          </label>
          <Button type="submit" aria-label={`Save ${name} key`} disabled={save.isPending || !value.trim()}>
            Save key
          </Button>
        </form>
      )}
      {error && <p className="text-destructive">{error instanceof Error ? error.message : String(error)}</p>}
    </section>
  );
}
