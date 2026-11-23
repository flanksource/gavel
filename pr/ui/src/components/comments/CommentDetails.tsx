import { useState } from 'react';
import { keepPreviousData, useQuery } from '@tanstack/react-query';
import { AgentIdentityProvider } from '@flanksource/clicky-ui/chat';
import { CommentCapture, CommentThread, DEFAULT_COMMENT_STATUSES, type CommentConfig } from '@flanksource/clicky-ui/comments';
import { commentAgentAssignmentItems, commentCaptureOf, projectNative, type InboxItem } from '@flanksource/clicky-ui/comments/gavel';
import { DropdownMenu } from '@flanksource/clicky-ui/components';
import { ActionError } from '@flanksource/clicky-ui/data';
import { capabilitiesQuery, commentClient, commentSessionsQuery } from './queries';
import { SendCommentDialog, type SendCommentChoice } from './SendCommentDialog';
import { CommentSession } from './CommentSession';

const config: CommentConfig = { statuses: DEFAULT_COMMENT_STATUSES };

export function CommentDetails({ item, local, refresh }: { item: InboxItem; local: boolean; refresh: () => Promise<unknown> }) {
  const [sending, setSending] = useState<SendCommentChoice>();
  const [maximized, setMaximized] = useState(false);
  const { id, version, phase, content, messages } = item.thread;
  const capabilities = useQuery({ ...capabilitiesQuery(id, version), placeholderData: keepPreviousData });
  const has = (action: string) => capabilities.data?.includes(action) ?? false;
  const client = commentClient(item.scope);
  const thread = { id, version };
  const settled = async (request: Promise<unknown>) => {
    try { await request; }
    finally { await refresh(); }
  };
  const transition = phase === 'active' ? 'resolve' : 'reopen';
  const run = content?.agentRun;
  const canSend = local && phase === 'active' && run?.state !== 'queued';
  const sessions = useQuery({ ...commentSessionsQuery(item.scope), enabled: local && (canSend || Boolean(run?.sessionId)) });
  const capture = commentCaptureOf(item.thread, client.attachmentUrl);
  const identities = Object.fromEntries((sessions.data ?? []).map(session => [session.id, { sessionName: session.title, ...(session.model ? { model: session.model } : { provider: session.source }) }]));
  const deliveryItems = commentAgentAssignmentItems({
    sessions: sessions.data ?? [],
    currentSessionId: run?.state === 'error' ? undefined : run?.sessionId,
    disabled: !canSend,
    onSession: session => setSending({ kind: 'session', sessionId: session.id }),
    onNew: () => setSending({ kind: 'run', fresh: true }),
    onAdvanced: () => setSending({ kind: 'run', fresh: true, advanced: true }),
  });
  const sendToAgent = canSend
    ? <DropdownMenu size="sm" variant="outline" label={run?.state === 'error' ? 'Retry agent delivery' : 'Send to agent'} items={deliveryItems} />
    : undefined;

  return <div className="space-y-2">
    {capabilities.isPending && <p role="status" className="text-sm text-muted-foreground">Loading comment actions…</p>}
    {capabilities.error && <ActionError error={capabilities.error} />}
    <AgentIdentityProvider identities={identities}>
    <CommentThread
      comments={projectNative(item.thread)}
      config={config}
      compact
      defaultExpanded
      hideComposer
      onMaximizedChange={setMaximized}
      renderMaximizedActions={() => sendToAgent}
      detailTabs={[
        { id: 'capture', label: 'Capture', render: () => <CommentCapture capture={capture} expanded /> },
        { id: 'session', label: 'Session', render: () => <CommentSession run={run} session={sessions.data?.find(entry => entry.id === run?.sessionId)} local={local} /> },
      ]}
      renderRootDetails={() => <div className="space-y-2">
        {run && <p className="text-sm text-muted-foreground" role="status">Agent delivery {run.state}{run.target?.label ? ` · ${run.target.label}` : ''}</p>}
        {run?.error && <ActionError error={run.error} />}
        {!maximized && sendToAgent}
        {canSend && sessions.error && <ActionError error={sessions.error} title="Could not load agent sessions" />}
      </div>}
      {...(has('reply') ? { onReply: ({ body }) => settled(client.reply({ ...thread, body })) } : {})}
      {...(has('update') ? { onEdit: (_id, body) => settled(client.update({ ...thread, body })) } : {})}
      {...(has('edit-reply') ? { onEditReply: (_id, messageId, body) => settled(client.editReply({ ...thread, messageId, body })) } : {})}
      {...(has('acknowledge') ? { onAcknowledge: () => settled(client.status({ ...thread, action: 'acknowledge' })) } : {})}
      {...(has('close') ? { onClose: () => settled(client.status({ ...thread, action: 'close' })) } : {})}
      {...(has(transition) ? { onUpdateStatus: () => settled(client.status({ ...thread, action: transition })) } : {})}
      {...(local && messages?.some(message => message.content?.proposal) ? { loadProposalDiff: (messageId) => client.proposalDiff(id, messageId) } : {})}
      {...(local && has('accept-proposal') ? { onApplyProposal: (messageId, mode) => settled(client.applyProposal({ ...thread, messageId, mode })) } : {})}
      {...(local && has('reject-proposal') ? { onRejectProposal: (messageId, reason) => settled(client.rejectProposal({ ...thread, messageId, ...(reason ? { reason } : {}) })) } : {})}
      {...(local && has('request-revision') ? { onRequestRevision: (messageId, feedback) => settled(client.requestRevision({ ...thread, messageId, feedback })) } : {})}
    />
    </AgentIdentityProvider>
    {sending && <SendCommentDialog item={item} choice={sending} onClose={() => setSending(undefined)} refresh={refresh} />}
  </div>;
}
