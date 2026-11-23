import { useMemo, useRef, useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { compactAISpecRuntime, SpecRuntimeBar, type AISpecRuntimeValue } from '@flanksource/clicky-ui/ai';
import { Button, Combobox, Modal } from '@flanksource/clicky-ui/components';
import { ActionError } from '@flanksource/clicky-ui/data';
import { CommentAgentSessions, DEFAULT_RUN_PROMPT, type CommentEdits, type DeliverTarget, type InboxItem } from '@flanksource/clicky-ui/comments/gavel';
import { randomUUID } from '@flanksource/clicky-ui/utils';
import { chatModelsQuery, chatRuntimesQuery } from '../oneShotQueries';
import { promptCatalogOptions, promptCatalogQuery, scopeQueryFor } from '../prompts/promptCatalog';
import { commentClient, commentSessionsQuery } from './queries';

type Destination = 'session' | 'run';
export type SendCommentChoice = { kind: 'session'; sessionId: string } | { kind: 'run'; fresh: boolean; advanced?: boolean };

export function SendCommentDialog({ item, choice, onClose, refresh }: { item: InboxItem; choice: SendCommentChoice; onClose: () => void; refresh: () => Promise<unknown> }) {
  const [destination, setDestination] = useState<Destination>(choice.kind);
  const [sessionId, setSessionId] = useState(choice.kind === 'session' ? choice.sessionId : undefined);
  const [prompt, setPrompt] = useState(DEFAULT_RUN_PROMPT);
  const [edits, setEdits] = useState<CommentEdits | undefined>();
  const [fresh, setFresh] = useState(choice.kind === 'run' && choice.fresh);
  const [runtime, setRuntime] = useState<AISpecRuntimeValue>({ mode: 'agent' });
  const advanced = choice.kind === 'run' && !!choice.advanced && destination === 'run';
  const { id, version } = item.thread;
  const client = commentClient(item.scope);
  const sessions = useQuery({ ...commentSessionsQuery(item.scope), enabled: destination === 'session' });
  const models = useQuery({ ...chatModelsQuery(), enabled: advanced });
  const families = useQuery({ ...chatRuntimesQuery(), enabled: advanced });
  const catalogError = models.error ?? families.error;
  const catalog = useQuery({ ...promptCatalogQuery(scopeQueryFor(item.project)), enabled: destination === 'run' });
  const promptOptions = useMemo(() => promptCatalogOptions(catalog.data ?? []), [catalog.data]);
  const preview = useMutation({ mutationFn: ({ target }: { target: DeliverTarget; key: string }) => client.previewDelivery({ id, target }) });
  const request = useRef<{ key: string; id: string } | undefined>(undefined);
  const send = useMutation({
    mutationFn: async ({ target, key }: { target: DeliverTarget; key: string }) => {
      if (request.current?.key !== key) request.current = { key, id: randomUUID() };
      try { await client.deliver({ id, version, requestId: request.current.id, target }); }
      finally { await refresh(); }
    },
    onSuccess: onClose,
  });
  const policy = edits ? { edits } : {};
  let target: DeliverTarget | undefined;
  if (destination === 'session' && sessionId && sessions.data?.some(session => session.id === sessionId)) {
    target = { kind: 'session', sessionId, ...policy };
  } else if (destination === 'run' && prompt.trim()) {
    target = { kind: 'run', prompt: prompt.trim(), ...(fresh ? { fresh } : {}), ...policy,
      ...(advanced && runtime.model ? { runOptions: { spec: { ...compactAISpecRuntime(runtime) } } } : {}) };
  }
  const key = JSON.stringify([version, target]);
  const currentPreview = preview.variables?.key === key ? preview.data : undefined;
  const previewError = preview.variables?.key === key ? preview.error : null;
  const canSend = item.thread.phase === 'active' && target && (target.kind === 'session' || (currentPreview && !currentPreview.validationError));

  return <Modal open title="Send comment to agent" onClose={() => { if (!send.isPending) onClose(); }}
    footer={<div className="flex justify-end gap-2">
      <Button variant="outline" disabled={send.isPending} onClick={onClose}>Cancel</Button>
      <Button disabled={!canSend || send.isPending} loading={send.isPending} onClick={() => { if (target && canSend) send.mutate({ target, key }); }}>Send comment</Button>
    </div>}>
    <div className="space-y-4 text-sm">
      <div role="group" aria-label="Agent destination" className="flex flex-wrap gap-2">
        <Button size="sm" variant="outline" aria-pressed={destination === 'session'} disabled={send.isPending} onClick={() => setDestination('session')}>Existing session</Button>
        <Button size="sm" variant="outline" aria-pressed={destination === 'run'} disabled={send.isPending} onClick={() => setDestination('run')}>Managed run</Button>
      </div>
      <fieldset disabled={send.isPending} className="space-y-3">
        <legend className="font-medium">Edit policy</legend>
        <div className="flex flex-wrap gap-3">
          <label className="flex items-center gap-1.5"><input type="radio" name="comment-edits" checked={!edits} onChange={() => setEdits(undefined)} />Use project settings</label>
          <label className="flex items-center gap-1.5"><input type="radio" name="comment-edits" checked={edits === 'direct'} onChange={() => setEdits('direct')} />Direct edits</label>
          <label className="flex items-center gap-1.5"><input type="radio" name="comment-edits" checked={edits === 'propose'} onChange={() => setEdits('propose')} />Proposal only</label>
        </div>
      </fieldset>
      {destination === 'session' ? <CommentAgentSessions busy={send.isPending} sessions={{
        items: sessions.data ?? [], state: sessions.isError ? 'error' : sessions.isPending ? 'loading' : 'ready', error: sessions.error?.message,
        selectedId: sessionId, onSelect: session => setSessionId(session.id),
      }} /> : <div className="space-y-3">
        <div className="space-y-1">
          <span className="font-medium">Prompt</span>
          <Combobox ariaLabel="Prompt" required options={promptOptions} loading={catalog.isPending} value={prompt} disabled={send.isPending} onChange={setPrompt} />
          {catalog.error && <ActionError error={catalog.error} title="Could not load prompts" />}
        </div>
        {advanced && (catalogError ? <ActionError error={catalogError} title="Could not load agent runtimes" /> : models.data && families.data ? (
          <SpecRuntimeBar ariaLabel="Agent runtime" sections={['model']} models={models.data} families={families.data} value={runtime} onChange={setRuntime} readOnly={send.isPending} />
        ) : <p role="status" className="text-muted-foreground">Loading agent runtimes…</p>)}
        <label className="flex items-center gap-1.5"><input type="checkbox" checked={fresh} disabled={send.isPending} onChange={event => setFresh(event.target.checked)} />Start a fresh session</label>
        <Button variant="outline" size="sm" loading={preview.isPending} disabled={!target || preview.isPending || send.isPending}
          onClick={() => { if (target?.kind === 'run') preview.mutate({ target, key }); }}>Preview run</Button>
        {previewError && <ActionError error={previewError} />}
        {currentPreview && <section aria-label="Run preview" className="space-y-2">
          <p className="font-medium">{currentPreview.model} · {currentPreview.mode}</p>
          {currentPreview.validationError && <ActionError error={currentPreview.validationError} />}
          {currentPreview.warnings?.map(warning => <p key={warning}>{warning}</p>)}
          <h3 className="font-medium">Tools</h3>
          <ul className="list-inside list-disc">{currentPreview.tools.map(tool => <li key={`${tool.source}/${tool.name}`} title={tool.description}>{tool.name}</li>)}</ul>
        </section>}
      </div>}
      {send.error && <ActionError error={send.error} />}
    </div>
  </Modal>;
}
