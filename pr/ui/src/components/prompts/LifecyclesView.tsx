import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button, JsonSchemaForm, Tabs, type JsonSchemaObject } from '@flanksource/clicky-ui/components';
import { settingsSchemaQuery } from '../settings/queries';

interface Step { name: string; prompt: string; [key: string]: unknown }
interface Definition { name?: string; subject?: Record<string, string>; steps?: Step[]; order?: string[]; disabled?: string[]; file?: string }
interface View { path: string; revision: string; source: Definition; fileSource?: Definition; inherited: Definition; effective: Definition }
interface Props { scopeQuery: string; kind: 'todos' | 'pr'; onSelect: (kind: 'todos' | 'pr') => void }

async function request(url: string, method = 'GET', body?: unknown): Promise<View> {
  const response = await fetch(url, { method, headers: { 'Content-Type': 'application/json' }, ...(body ? { body: JSON.stringify(body) } : {}) });
  if (!response.ok) throw new Error(await response.text());
  return response.json() as Promise<View>;
}

export function LifecyclesView({ scopeQuery, kind, onSelect }: Props) {
  const url = `/api/settings/lifecycles?${scopeQuery}&kind=${kind}`;
  const query = useQuery({ queryKey: ['lifecycles', scopeQuery, kind], queryFn: () => request(url) });
  const schema = useQuery(settingsSchemaQuery());
  const [source, setSource] = useState<Definition>({});
  const [preview, setPreview] = useState<View>();
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [saved, setSaved] = useState(false);
  useEffect(() => {
    setSource(query.data?.fileSource ?? query.data?.source ?? {});
    setPreview(undefined); setError(''); setSaved(false);
  }, [query.data, url]);
  const view = query.data;
  const lifecycleSchema = schema.data?.properties?.[kind]?.properties?.lifecycle as JsonSchemaObject | undefined;
  const change = (next: Definition) => { setSource(next); setPreview(undefined); setSaved(false); };
  const payload = { revision: view?.revision, source: view?.fileSource ? view.source : source, ...(view?.fileSource ? { fileSource: source } : {}) };
  async function submit(save: boolean) {
    setBusy(true); setError('');
    try {
      const result = await request(url, save ? 'PUT' : 'POST', payload);
      if (save) { await query.refetch(); setSaved(true); } else { setPreview(result); }
    } catch (cause) { setError(cause instanceof Error ? cause.message : String(cause)); }
    finally { setBusy(false); }
  }
  const inherited = view?.inherited.steps ?? [];
  const steps = [
    ...inherited.map(step => source.steps?.find(item => item.name === step.name) ?? step),
    ...(source.steps ?? []).filter(step => !inherited.some(item => item.name === step.name)),
  ];
  const order = source.order ?? view?.inherited.order ?? steps.map(step => step.name);
  steps.sort((a, b) => {
    const rank = (name: string) => order.includes(name) ? order.indexOf(name) : order.length;
    return rank(a.name) - rank(b.name);
  });
  function move(index: number, delta: number) {
    const next = steps.map(step => step.name);
    [next[index], next[index + delta]] = [next[index + delta], next[index]];
    change({ ...source, order: next });
  }
  return <div className="min-h-0 flex-1 overflow-y-auto p-4">
    <Tabs tabs={[{ id: 'todos', label: 'TODO lifecycle' }, { id: 'pr', label: 'PR lifecycle' }]} value={kind} onChange={next => onSelect(next as Props['kind'])} />
    <p className="my-3 text-sm text-muted-foreground">Steps run in priority order. PR actions start manually from PR details. Preview validates conditions and shows the resolved workflow before saving.</p>
    {(error || query.error || schema.error) && <p role="alert" className="my-3 text-sm text-destructive">{error || query.error?.message || schema.error?.message}</p>}
    {view && lifecycleSchema && <>
      <p className="my-3 break-all text-xs text-muted-foreground">{view.path}{view.source.file ? ` → ${view.source.file}` : ''}</p>
      <div className="mb-4 space-y-2">{steps.map((step, index) => {
        const overridden = (source.steps ?? []).some(item => item.name === step.name);
        const disabled = (source.disabled ?? view.inherited.disabled ?? []).includes(step.name);
        return <div key={step.name} className="flex flex-wrap items-center gap-2 rounded border border-border p-2">
          <span className="min-w-32 flex-1 text-sm">{step.name} <span className="text-muted-foreground">{overridden ? 'Override' : 'Inherited'}</span></span>
          <Button size="sm" variant="ghost" disabled={index === 0} onClick={() => move(index, -1)} aria-label={`Move ${step.name} up`}>Up</Button>
          <Button size="sm" variant="ghost" disabled={index === steps.length - 1} onClick={() => move(index, 1)} aria-label={`Move ${step.name} down`}>Down</Button>
          <Button size="sm" variant="ghost" disabled={kind === 'todos' && step.name === 'verify'} onClick={() => change({ ...source, disabled: disabled ? (source.disabled ?? view.inherited.disabled ?? []).filter(name => name !== step.name) : [...(source.disabled ?? view.inherited.disabled ?? []), step.name] })}>{disabled ? 'Enable' : 'Disable'} {step.name}</Button>
          <Button size="sm" variant="outline" onClick={() => change({ ...source, steps: overridden ? source.steps?.filter(item => item.name !== step.name) : [...(source.steps ?? []), step] })}>{overridden ? 'Reset' : 'Edit'} {step.name}</Button>
        </div>;
      })}</div>
      <JsonSchemaForm schema={{ ...lifecycleSchema, $defs: schema.data?.$defs }} value={source as Record<string, unknown>} onChange={next => change(next as Definition)} showPreferencesMenu={false} persistPreferences={false} />
      <div className="my-4 flex flex-wrap gap-2">
        <Button variant="outline" onClick={() => change({})}>Reset scope overrides</Button>
        <Button variant="outline" onClick={() => change(view.fileSource ?? view.source)}>Discard changes</Button>
        <Button variant="outline" loading={busy} onClick={() => void submit(false)}>Preview lifecycle</Button>
        <Button disabled={!preview || busy} onClick={() => void submit(true)}>Save lifecycle</Button>
      </div>
      {saved && <p role="status" className="text-sm">Lifecycle saved.</p>}
      {preview && <section aria-label="Resolved lifecycle"><h3 className="font-medium">Resolved lifecycle</h3><pre className="mt-2 overflow-x-auto rounded bg-muted p-3 text-xs">{JSON.stringify(preview.effective, null, 2)}</pre></section>}
    </>}
    {(query.isLoading || schema.isLoading) && <p role="status">Loading lifecycle…</p>}
  </div>;
}
