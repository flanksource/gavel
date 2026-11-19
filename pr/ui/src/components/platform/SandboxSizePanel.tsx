// The platform dashboard's sandbox size of one organization: the CPU and memory each of its
// substrate environments gets, which a platform admin sets for it or returns to serve's default.
import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button } from '@flanksource/clicky-ui/components';
import { fetchJSON, mutationJSON, queryKeys } from '../../query';
import { PanelHeading } from './PanelHeading';

export interface PlatformSize {
  cpuMillicores: number;
  memoryBytes: number;
}

// PlatformSandboxSize is the organization's own size, null without one, and the default it gets
// then; defaultError says why the default is missing.
export interface PlatformSandboxSize {
  size: PlatformSize | null;
  default?: PlatformSize;
  defaultError?: string;
}

const gib = 2 ** 30;
const inputClass = 'w-28 rounded border border-border bg-background px-2 py-1 text-sm';

// figure shows n with at most two decimals.
function figure(n: number): string {
  return String(Number(n.toFixed(2)));
}

function sizeName(size: PlatformSize): string {
  return `${figure(size.cpuMillicores / 1000)} CPU · ${figure(size.memoryBytes / gib)} GiB`;
}

function errorText(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

export function SandboxSizePanel({ org }: { org: string }) {
  const url = `/api/platform/orgs/${encodeURIComponent(org)}/sandbox-size`;
  const queryKey = queryKeys.platform(`orgs/${org}/sandbox-size`);
  const queryClient = useQueryClient();
  const sandboxSize = useQuery({
    queryKey,
    queryFn: ({ signal }) => fetchJSON<PlatformSandboxSize>({ url, signal, context: 'load the organization\'s sandbox size' }),
  });
  const [cpu, setCPU] = useState('');
  const [memory, setMemory] = useState('');
  const stored = (data: PlatformSandboxSize) => queryClient.setQueryData(queryKey, data);
  const save = useMutation({
    mutationFn: () => mutationJSON<PlatformSandboxSize>({
      url, method: 'PUT', context: 'save the sandbox size',
      body: { cpuMillicores: Math.round(Number(cpu) * 1000), memoryBytes: Math.round(Number(memory) * gib) },
    }),
    onSuccess: stored,
  });
  const reset = useMutation({
    mutationFn: () => mutationJSON<PlatformSandboxSize>({ url, method: 'DELETE', context: 'return to the default sandbox size' }),
    onSuccess: stored,
  });
  const data = sandboxSize.data;
  const error = sandboxSize.error ?? save.error ?? reset.error;

  return (
    <section className="flex flex-col gap-3 text-sm">
      <PanelHeading>Sandbox size</PanelHeading>
      <p className="text-muted-foreground">
        The CPU and memory substrate gives each of this organization's sandbox environments, and holds it to.
        Runs started from now on get the new size; runs already going keep theirs.
      </p>
      {data && (
        <>
          {data.size
            ? <p>{`${sizeName(data.size)} per environment, set for this organization`}</p>
            : data.default && <p>{`${sizeName(data.default)} per environment, the default`}</p>}
          {data.size && data.default && <p className="text-muted-foreground">{`The default is ${sizeName(data.default)}.`}</p>}
          {data.defaultError && <p className="text-muted-foreground">{`gavel couldn't read the default: ${data.defaultError}`}</p>}
          <form
            className="flex flex-wrap items-end gap-3"
            onSubmit={event => {
              event.preventDefault();
              save.mutate();
            }}
          >
            <label className="flex flex-col gap-1">
              <span>CPU (cores)</span>
              <input className={inputClass} type="number" min="0.1" step="0.1" value={cpu} onChange={event => setCPU(event.target.value)} />
            </label>
            <label className="flex flex-col gap-1">
              <span>Memory (GiB)</span>
              <input className={inputClass} type="number" min="0.25" step="0.25" value={memory} onChange={event => setMemory(event.target.value)} />
            </label>
            <Button type="submit" disabled={save.isPending || !cpu || !memory}>Save size</Button>
            {data.size && (
              <Button type="button" variant="ghost" disabled={reset.isPending} onClick={() => reset.mutate()}>Use the default</Button>
            )}
          </form>
        </>
      )}
      {error && <p role="alert" className="text-red-600">{errorText(error)}</p>}
    </section>
  );
}
