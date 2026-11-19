import { useState, useSyncExternalStore } from 'react';
import { createPortal } from 'react-dom';
import { useIsMutating, useMutation, useQueryClient } from '@tanstack/react-query';
import { SessionTail } from '@flanksource/clicky-ui/ai';
import { Button, Modal, SegmentedControl, Switch, zIndex, type HostedModalProps, type ModalFormDefinition, type SegmentedOption } from '@flanksource/clicky-ui/components';
import { UiGitGraph, UiGitMerge, UiGitPr, UiRocket, UiSparkles } from '@flanksource/clicky-ui/icons';
import { HttpError, mutationJSON, queryKeys } from '../query';
import { usePromptRun } from './promptRun';
import { prActionFormKey, prActionMutationKey, readPRActionDraft, type MergeMethod, type PendingPRAction, type PRActionAvailability } from './prActionDraft';

type Draft = ReturnType<typeof readPRActionDraft>;
type UpdateBranchConflict = { error: string; rebase?: Record<string, unknown>; rebaseBlocked?: string };
const titles = { merge: 'Merge pull request', approve: 'Approve pull request', 'update-branch': 'Update branch' } as const;
const methods: SegmentedOption<MergeMethod>[] = [
  { id: 'rebase', label: 'Rebase', icon: UiGitMerge, title: 'Rebase and merge' },
  { id: 'squash', label: 'Squash', icon: UiGitPr, title: 'Squash and merge' },
  { id: 'merge', label: 'Merge commit', icon: UiGitGraph, title: 'Create a merge commit' },
];
const effects: Record<MergeMethod, string> = {
  rebase: 'replays its commits onto',
  squash: 'squashes its commits into one commit on',
  merge: 'adds a merge commit to',
};

export const prActionForms = Object.fromEntries(Object.entries(titles).map(([type, title]) => [
  prActionFormKey(type as PendingPRAction['type']),
  { title, Form: PRActionDialog, recovery: {
    serialize: readPRActionDraft,
    restore: (saved: unknown) => {
      const draft = readPRActionDraft(saved);
      if (draft.pending.type !== type) throw new Error('Saved PR action does not match its form');
      return draft;
    },
  } } satisfies ModalFormDefinition,
]));

function PRActionDialog({ initialDraft, modalProps, submit }: HostedModalProps) {
  const [draft, setDraft] = useState(() => readPRActionDraft(initialDraft));
  const [conflict, setConflict] = useState<UpdateBranchConflict | null>(null);
  const [rebaseHidden, setRebaseHidden] = useState(false);
  const availability = initialDraft.availability as PRActionAvailability | undefined;
  const onChanged = initialDraft.onChanged as (() => void) | undefined;
  function subscribe(listener: () => void) {
    availability?.changes.addEventListener('change', listener);
    return () => availability?.changes.removeEventListener('change', listener);
  }
  const disabled = useSyncExternalStore(subscribe, () => availability?.disabled ?? false);
  const anchor = useSyncExternalStore(subscribe, () => availability?.anchor ?? null);
  const queryClient = useQueryClient();
  const mutationKey = prActionMutationKey(draft.repo, draft.number);
  const mutations = useIsMutating({ mutationKey, exact: true });
  const action = useMutation({ mutationKey, mutationFn: sendAction, retry: false });
  const refresh = () => Promise.all([
    queryClient.invalidateQueries({ queryKey: queryKeys.prSnapshot(), exact: true }),
    queryClient.invalidateQueries({ queryKey: queryKeys.prDetail(draft.repo, draft.number), exact: true }),
  ]);
  const notify = () => { if (availability?.anchor?.isConnected) onChanged?.(); };
  const rebase = usePromptRun({
    prompt: 'pr-rebase',
    label: 'AI rebase',
    onSuccess: () => { modalProps.onClose(); notify(); },
    onSettled: () => { void refresh(); },
  });
  const busy = disabled || mutations > 0 || rebase.active;
  const { pending } = draft;
  const copy = actionCopy(draft);

  function confirm() {
    if (busy || availability?.disabled || queryClient.isMutating({ mutationKey, exact: true }) > 0) return;
    void submit({
      draft,
      request: async () => {
        try {
          await action.mutateAsync(draft);
          return { status: 'success', value: undefined };
        } catch (error) {
          const next = draft.pending.type === 'update-branch' ? updateBranchConflict(error) : null;
          setConflict(next);
          if (next) return { status: 'rejected', error: next.error };
          if (error instanceof HttpError && error.status < 500) return { status: 'rejected', error: error.message };
          return { status: 'unknown', error: error instanceof Error ? error.message : 'Could not update pull request' };
        }
      },
      onSuccess: async () => { await refresh(); notify(); },
    });
  }

  function close() {
    if (rebase.active) setRebaseHidden(true);
    else modalProps.onClose();
  }

  const rebaseIndicator = rebaseHidden && (
    <Button variant="outline" size="sm" className="h-7 shrink-0 gap-1 px-2 text-xs"
      onClick={() => setRebaseHidden(false)} title={`Show the AI rebase for ${draft.repo}#${draft.number}`}>
      <UiSparkles />{rebase.active ? 'Rebasing…' : 'AI rebase failed'}
    </Button>
  );

  return (
    <>
      <Modal
        {...modalProps}
        onClose={close}
        minimized={modalProps.minimized || rebaseHidden}
        error={conflict ? (
          <div className="space-y-1">
            <p>{conflict.error}</p>
            <p className="text-muted-foreground">
              {conflict.rebase
                ? 'An agent can rebase the branch in a fresh worktree, resolve the conflicts, verify there are no merge commits, then force-push it.'
                : `AI rebase unavailable: ${conflict.rebaseBlocked || 'the server did not say why'}`}
            </p>
          </div>
        ) : modalProps.error}
        size="lg"
        footer={
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={close}>{rebase.active ? 'Close' : 'Cancel'}</Button>
            {conflict?.rebase && !rebase.active ? (
              <Button disabled={busy} onClick={() => { void rebase.start(conflict.rebase!); }}>
                <UiSparkles />Rebase with AI
              </Button>
            ) : (
              <Button onClick={confirm} disabled={busy || !!conflict}>{copy.label}</Button>
            )}
          </div>
        }
      >
        <div className="space-y-3 text-sm text-foreground">
          <p>{copy.description}</p>
          {rebase.active && (
            <div className="flex items-center gap-2 text-xs">
              {rebase.state.status === 'running' && rebase.state.sessionId
                ? <SessionTail src={`/api/captain/sessions/${encodeURIComponent(rebase.state.sessionId)}`} className="min-w-0 flex-1" />
                : <span className="min-w-0 flex-1 text-muted-foreground">Starting the AI rebase…</span>}
              {rebase.state.status === 'running' && (
                <Button variant="ghost" size="sm" className="h-6 px-2 text-xs" onClick={() => { void rebase.stop(); }}>Stop</Button>
              )}
            </div>
          )}
          {rebase.state.status === 'failed' && (
            <div role="alert" className="whitespace-pre-wrap text-sm text-destructive">AI rebase failed: {rebase.state.error}</div>
          )}
          {pending.type === 'merge' && (
            <div className="space-y-2">
              <SegmentedControl aria-label="Merge method" size="sm" value={pending.method} options={methods}
                onChange={method => setDraft({ ...draft, pending: { ...pending, method } })} />
              {!draft.azure && (
                <Switch checked={pending.auto} disabled={busy} onChange={auto => setDraft({ ...draft, pending: { ...pending, auto } })}
                  label={<span className="inline-flex items-center gap-1.5 text-sm"><UiRocket className="text-xs" />Auto-merge</span>} />
              )}
            </div>
          )}
          {pending.type === 'approve' && (
            <textarea className="h-20 w-full resize-none rounded-md border border-border bg-background px-2 py-1.5 text-sm"
              value={draft.comment} placeholder="Optional approval comment" aria-label="Approval comment"
              onChange={event => setDraft({ ...draft, comment: event.currentTarget.value })} autoFocus />
          )}
        </div>
      </Modal>
      {rebaseIndicator && (anchor ? createPortal(rebaseIndicator, anchor) : createPortal(
        <div className="fixed bottom-4 right-4" style={{ zIndex: zIndex.popover }}>{rebaseIndicator}</div>, document.body,
      ))}
    </>
  );
}

function updateBranchConflict(error: unknown): UpdateBranchConflict | null {
  if (!(error instanceof HttpError) || error.status !== 409 || typeof error.body !== 'object' || error.body === null) return null;
  const body = error.body as Partial<UpdateBranchConflict> & { conflict?: boolean };
  return body.conflict ? { error: body.error ?? error.message, rebase: body.rebase, rebaseBlocked: body.rebaseBlocked } : null;
}

function sendAction({ pending, repo, number, nodeId, headSha, url, comment }: Draft) {
  const method = 'POST';
  switch (pending.type) {
    case 'approve':
      return mutationJSON({ url: '/api/prs/approve', method, body: { repo, number, nodeId, body: comment.trim() }, context: titles.approve });
    case 'update-branch':
      return mutationJSON({ url: '/api/prs/update-branch', method, body: { repo, number, nodeId, headSha: headSha || undefined }, context: 'Update pull request branch' });
    case 'merge':
      return mutationJSON({ url: '/api/prs/merge', method, body: { url, method: pending.method, auto: pending.auto, headSha: headSha || undefined }, context: titles.merge });
  }
}

function actionCopy({ pending, repo, number, base }: Draft) {
  const ref = `${repo}#${number}`;
  switch (pending.type) {
    case 'merge':
      return {
        label: pending.auto ? 'Enable auto-merge' : 'Merge',
        description: pending.auto
          ? `Once all required checks pass, ${ref} merges: it ${effects[pending.method]} ${base}.`
          : `Merge ${ref}: it ${effects[pending.method]} ${base}. This cannot be undone.`,
      };
    case 'approve':
      return { label: 'Approve', description: `Submit an approving review on ${ref}.` };
    case 'update-branch':
      return { label: 'Update branch', description: `Rebase the head branch of ${ref} onto ${base} to bring it up to date. This rewrites the PR's commits without adding a merge commit.` };
  }
}
