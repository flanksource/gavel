// Sets the name and email the signed-in user's hosted runs commit as, over their GitHub
// identity. Shown as a page in Clerk's user profile in clerk mode.
import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@flanksource/clicky-ui/components';
import { fetchJSON, mutationJSON, queryKeys } from '../query';

export interface CommitAuthor {
  name: string;
  email: string;
}

// CommitIdentity is what the user set, empty fields falling back to `default`.
export interface CommitIdentity extends CommitAuthor {
  default: CommitAuthor;
}

const inputClass = 'w-full rounded border border-border bg-background px-2 py-1 text-sm';

export function CommitIdentityPanel() {
  const queryClient = useQueryClient();
  const identity = useQuery({
    queryKey: queryKeys.commitIdentity(),
    queryFn: ({ signal }) => fetchJSON<CommitIdentity>({ url: '/api/commit-identity', signal, context: 'load commit identity' }),
  });
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  useEffect(() => {
    if (!identity.data) return;
    setName(identity.data.name);
    setEmail(identity.data.email);
  }, [identity.data]);
  const save = useMutation({
    mutationFn: () => mutationJSON<CommitIdentity>({ url: '/api/commit-identity', method: 'PUT', body: { name, email }, context: 'save commit identity' }),
    onSuccess: saved => queryClient.setQueryData(queryKeys.commitIdentity(), saved),
  });
  const fallback = identity.data?.default;
  const saved = identity.data;
  const author = saved && {
    name: saved.name || saved.default.name || saved.email || saved.default.email,
    email: saved.email || saved.default.email,
  };
  const error = identity.error ?? save.error;

  return (
    <div className="flex flex-col gap-4 text-sm">
      <p className="text-muted-foreground">
        Commits your runs make are by you, as GitHub knows you unless you set your own name or email here. gavel is
        their committer.
      </p>
      {author && (
        <p>
          Your runs commit as{' '}
          <span className="font-mono">{author.email ? `${author.name} <${author.email}>` : 'gavel, until an email is known'}</span>
        </p>
      )}
      <form
        className="flex flex-col gap-2"
        onSubmit={event => {
          event.preventDefault();
          save.mutate();
        }}
      >
        <label className="flex flex-col gap-1">
          <span>Name</span>
          <input className={inputClass} value={name} onChange={event => setName(event.target.value)} placeholder={fallback?.name} />
        </label>
        <label className="flex flex-col gap-1">
          <span>Email</span>
          <input className={inputClass} value={email} onChange={event => setEmail(event.target.value)} placeholder={fallback?.email} />
        </label>
        <Button type="submit" aria-label="Save commit identity" disabled={save.isPending || !identity.data}>
          Save
        </Button>
      </form>
      {error && <p className="text-destructive">{error instanceof Error ? error.message : String(error)}</p>}
    </div>
  );
}
