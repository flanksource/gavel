import { useState } from 'react';
import { Button, SplitButton, Modal, type DropdownMenuItem } from '@flanksource/clicky-ui/components';
import type { PRItem, PRDetail } from '../types';
import { GavelIcon } from './GavelIcon';

type MergeMethod = 'rebase' | 'squash' | 'merge';

// Pending is the action awaiting confirmation in the modal. method is unused for
// approvals; for auto-merge it is the method GitHub applies once checks pass.
type Pending =
  | { type: 'merge'; method: MergeMethod }
  | { type: 'automerge'; method: MergeMethod }
  | { type: 'approve' };

// menuLabel pairs a (working) GavelIcon with text. clicky-ui's own Icon renders
// string names as broken boxes here (no fallback provider), so action icons must
// go through GavelIcon's <iconify-icon> inside the label rather than the `icon`
// prop of Button/SplitButton/DropdownMenuItem.
function menuLabel(icon: string, text: string) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <GavelIcon name={icon} className="text-xs" />
      {text}
    </span>
  );
}

async function postAction(path: string, body: Record<string, unknown>): Promise<void> {
  const res = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    let msg = `Request failed (${res.status})`;
    try {
      const data = await res.json();
      if (data?.error) msg = data.error;
    } catch {
      // non-JSON error body; keep the status-based message
    }
    throw new Error(msg);
  }
}

// PRActions renders merge / auto-merge / approve controls for an OPEN PR. Each
// action is confirmed in a modal before it hits GitHub. On success it calls
// onChanged so the host can re-fetch the detail and reflect the new state
// (merged / approved), at which point these controls disappear or update.
export function PRActions({
  pr,
  detail,
  onChanged,
}: {
  pr: PRItem;
  detail: PRDetail | null;
  onChanged?: () => void;
}) {
  const [pending, setPending] = useState<Pending | null>(null);
  const [comment, setComment] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const info = detail?.pr;
  const nodeId = info?.nodeId;
  const state = (info?.state || pr.state || '').toUpperCase();
  const isDraft = info?.isDraft ?? pr.isDraft ?? false;
  const mergeable = (info?.mergeable || pr.mergeable || '').toUpperCase();
  const base = info?.baseRefName || pr.target;

  // Only OPEN PRs are actionable; merged/closed ones have nothing to merge or approve.
  if (state !== 'OPEN') return null;

  // The node ID arrives with the first detail SSE frame; gate actions until then.
  const loading = !nodeId;
  const conflicting = mergeable === 'CONFLICTING';
  const mergeDisabled = loading || isDraft || conflicting;
  const mergeDisabledReason = loading
    ? 'Loading PR details…'
    : isDraft
      ? 'PR is a draft — mark it ready for review first'
      : conflicting
        ? 'Resolve merge conflicts before merging'
        : undefined;

  const mergeItems: DropdownMenuItem[] = [
    {
      label: menuLabel('codicon:git-pull-request', 'Squash and merge'),
      onSelect: () => open({ type: 'merge', method: 'squash' }),
    },
    {
      label: menuLabel('codicon:git-commit', 'Create a merge commit'),
      onSelect: () => open({ type: 'merge', method: 'merge' }),
    },
    {
      label: menuLabel('codicon:rocket', 'Enable auto-merge (rebase)'),
      title: 'Merge automatically once all required checks pass',
      onSelect: () => open({ type: 'automerge', method: 'rebase' }),
    },
  ];

  function open(p: Pending) {
    setError('');
    setComment('');
    setPending(p);
  }

  function cancel() {
    if (busy) return;
    setPending(null);
  }

  async function confirm() {
    if (!pending || !nodeId) return;
    setBusy(true);
    setError('');
    try {
      if (pending.type === 'approve') {
        await postAction('/api/prs/approve', { repo: pr.repo, number: pr.number, nodeId, body: comment.trim() });
      } else {
        await postAction('/api/prs/merge', {
          repo: pr.repo,
          number: pr.number,
          nodeId,
          method: pending.method,
          auto: pending.type === 'automerge',
        });
      }
      setPending(null);
      onChanged?.();
    } catch (err: any) {
      setError(err?.message || 'Action failed');
    } finally {
      setBusy(false);
    }
  }

  const copy = modalCopy(pending, pr, base);

  return (
    <>
      <Button
        variant="outline"
        size="sm"
        className="h-7 shrink-0 gap-1 px-2 text-xs"
        onClick={() => open({ type: 'approve' })}
        disabled={loading}
        title="Submit an approving review"
      >
        <GavelIcon name="codicon:check" className="text-xs" />
        Approve
      </Button>

      <span title={mergeDisabledReason}>
        <SplitButton
          variant="outline"
          size="sm"
          label={menuLabel('codicon:git-merge', 'Merge')}
          title="Merge options"
          disabled={mergeDisabled}
          onClick={() => open({ type: 'merge', method: 'rebase' })}
          items={mergeItems}
        />
      </span>

      {pending && (
        <Modal
          open
          onClose={cancel}
          title={copy.title}
          size="sm"
          footer={
            <div className="flex justify-end gap-2">
              <Button variant="outline" onClick={cancel} disabled={busy}>Cancel</Button>
              <Button onClick={confirm} loading={busy}>{copy.confirmLabel}</Button>
            </div>
          }
        >
          <div className="space-y-3 text-sm text-foreground">
            <p>{copy.description}</p>
            {pending.type === 'approve' && (
              <textarea
                className="w-full rounded-md border border-border bg-background px-2 py-1.5 text-sm h-20 resize-none"
                value={comment}
                placeholder="Optional approval comment"
                onChange={e => setComment(e.currentTarget.value)}
                autoFocus
              />
            )}
            {error && <div className="text-sm text-destructive">{error}</div>}
          </div>
        </Modal>
      )}
    </>
  );
}

function modalCopy(pending: Pending | null, pr: PRItem, base: string): { title: string; confirmLabel: string; description: string } {
  const ref = `${pr.repo}#${pr.number}`;
  if (!pending) return { title: '', confirmLabel: '', description: '' };
  switch (pending.type) {
    case 'merge':
      return {
        title: 'Merge pull request',
        confirmLabel: 'Merge',
        description: `Merge ${ref} into ${base} using a ${pending.method} merge. This cannot be undone.`,
      };
    case 'automerge':
      return {
        title: 'Enable auto-merge',
        confirmLabel: 'Enable auto-merge',
        description: `Auto-merge ${ref} into ${base} with a ${pending.method} merge once all required checks pass.`,
      };
    case 'approve':
      return {
        title: 'Approve pull request',
        confirmLabel: 'Approve',
        description: `Submit an approving review on ${ref}.`,
      };
  }
}
