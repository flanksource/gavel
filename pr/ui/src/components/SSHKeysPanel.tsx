// Manages the signed-in user's SSH public keys, which identify them to the hosted push
// server. Shown as a page in Clerk's user profile in clerk mode.
import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@flanksource/clicky-ui/components';
import { UiTrash } from '@flanksource/clicky-ui/icons';
import { fetchJSON, mutationJSON, queryKeys } from '../query';

export interface SSHKey {
  id: string;
  name: string;
  fingerprint: string;
  createdAt: string;
  lastUsedAt: string | null;
}

const inputClass = 'w-full rounded border border-border bg-background px-2 py-1 text-sm';

export function SSHKeysPanel() {
  const queryClient = useQueryClient();
  const keys = useQuery({
    queryKey: queryKeys.sshKeys(),
    queryFn: ({ signal }) => fetchJSON<SSHKey[]>({ url: '/api/ssh-keys', signal, context: 'load SSH keys' }),
  });
  const [name, setName] = useState('');
  const [publicKey, setPublicKey] = useState('');
  const refresh = () => queryClient.invalidateQueries({ queryKey: queryKeys.sshKeys() });
  const add = useMutation({
    mutationFn: () => mutationJSON<SSHKey>({ url: '/api/ssh-keys', method: 'POST', body: { name, publicKey }, context: 'add SSH key' }),
    onSuccess: () => {
      setName('');
      setPublicKey('');
      return refresh();
    },
  });
  const remove = useMutation({
    mutationFn: (id: string) => mutationJSON<void>({ url: `/api/ssh-keys/${encodeURIComponent(id)}`, method: 'DELETE', context: 'delete SSH key' }),
    onSuccess: refresh,
  });
  const error = keys.error ?? add.error ?? remove.error;

  return (
    <div className="flex flex-col gap-4 text-sm">
      <p className="text-muted-foreground">
        Keys you add here let you push to this organization's repositories over SSH.
      </p>
      <ul className="flex flex-col divide-y divide-border rounded border border-border">
        {(keys.data ?? []).map(key => (
          <li key={key.id} className="flex items-center justify-between gap-2 px-3 py-2">
            <div className="min-w-0">
              <div className="font-medium">{key.name || 'Unnamed key'}</div>
              <div className="truncate font-mono text-xs text-muted-foreground">{key.fingerprint}</div>
              <div className="text-xs text-muted-foreground">
                {key.lastUsedAt ? `Last used ${new Date(key.lastUsedAt).toLocaleString()}` : 'Never used'}
              </div>
            </div>
            <Button
              variant="ghost"
              size="icon"
              type="button"
              aria-label={`Delete SSH key ${key.name}`}
              disabled={remove.isPending}
              onClick={() => remove.mutate(key.id)}
            >
              <UiTrash />
            </Button>
          </li>
        ))}
        {keys.data?.length === 0 && <li className="px-3 py-2 text-muted-foreground">No SSH keys yet.</li>}
      </ul>
      <form
        className="flex flex-col gap-2"
        onSubmit={event => {
          event.preventDefault();
          add.mutate();
        }}
      >
        <label className="flex flex-col gap-1">
          <span>Key name</span>
          <input className={inputClass} value={name} onChange={event => setName(event.target.value)} placeholder="laptop" />
        </label>
        <label className="flex flex-col gap-1">
          <span>Public key</span>
          <textarea
            className={`${inputClass} font-mono`}
            rows={3}
            value={publicKey}
            onChange={event => setPublicKey(event.target.value)}
            placeholder="ssh-ed25519 AAAA… (the contents of ~/.ssh/id_ed25519.pub)"
          />
        </label>
        <Button type="submit" aria-label="Add SSH key" disabled={add.isPending || !publicKey.trim()}>
          Add key
        </Button>
      </form>
      {error && <p className="text-destructive">{error instanceof Error ? error.message : String(error)}</p>}
    </div>
  );
}
