import { useState } from 'react';
import { Button, SplitButton, DropdownMenu, Modal, type DropdownMenuItem } from '@flanksource/clicky-ui/components';
import { UiCheck, UiEllipsis, UiGitGraph, UiGitMerge, UiGitPr, UiRocket, UiSync } from '@flanksource/clicky-ui/icons';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import type { ComponentType } from 'react';
import type { PRItem, PRDetail } from '../types';

type MergeMethod = 'rebase' | 'squash' | 'merge';

// ExtraAction is a caller-supplied button (e.g. "New todo") folded into the PR
// action cluster so it collapses together with the GitHub actions: an outline
// button when the bar is expanded, a menu item once it collapses into the
// overflow dropdown.
export interface ExtraAction {
  label: string;
  icon: ComponentType<IconProps>;
  onClick: () => void;
  title?: string;
  disabled?: boolean;
}

// Pending is the action awaiting confirmation in the modal. method is unused for
// approvals; for auto-merge it is the method GitHub applies once checks pass.
type Pending =
  | { type: 'merge'; method: MergeMethod }
  | { type: 'automerge'; method: MergeMethod }
  | { type: 'approve' }
  | { type: 'update-branch' };

// menuLabel pairs an icon component with text, rendered inside the label of a
// Button/SplitButton/DropdownMenuItem (rather than via their `icon` prop) so the
// label controls the icon's sizing and spacing.
function menuLabel(icon: ComponentType<IconProps>, text: string) {
  const Icon = icon;
  return (
    <span className="inline-flex items-center gap-1.5">
      <Icon className="text-xs" />
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

// PRActions renders merge / auto-merge / approve / update-branch controls for an
// OPEN PR. Each action is confirmed in a modal before it hits GitHub. On success
// it calls onChanged so the host can re-fetch the detail and reflect the new state
// (merged / approved), at which point these controls disappear or update.
export function PRActions({
  pr,
  detail,
  onChanged,
  collapsed = false,
  extras = [],
}: {
  pr: PRItem;
  detail: PRDetail | null;
  onChanged?: () => void;
  // collapsed folds every action into an overflow dropdown — set by the detail
  // panel when its container is too narrow to lay the buttons out inline.
  collapsed?: boolean;
  extras?: ExtraAction[];
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
  const behind = pr.behind ?? 0;
  // Only OPEN PRs are actionable; merged/closed ones have nothing to merge or
  // approve, but callers may still pass extras (e.g. New todo) to render.
  const isOpen = state === 'OPEN';

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

  // mergeMethods list every merge option; the first (rebase) mirrors the
  // SplitButton's default click, so it stays reachable once Merge collapses into
  // a submenu. mergeItems are the SplitButton's dropdown (everything but the
  // default rebase click).
  const mergeMethods: DropdownMenuItem[] = [
    {
      label: menuLabel(UiGitMerge, 'Rebase and merge'),
      onSelect: () => open({ type: 'merge', method: 'rebase' }),
    },
    {
      label: menuLabel(UiGitPr, 'Squash and merge'),
      onSelect: () => open({ type: 'merge', method: 'squash' }),
    },
    {
      label: menuLabel(UiGitGraph, 'Create a merge commit'),
      onSelect: () => open({ type: 'merge', method: 'merge' }),
    },
    {
      label: menuLabel(UiRocket, 'Enable auto-merge (rebase)'),
      title: 'Merge automatically once all required checks pass',
      onSelect: () => open({ type: 'automerge', method: 'rebase' }),
    },
    {
      label: menuLabel(UiSync, 'Update branch'),
      title: `Merge ${base} into this PR's branch to bring it up to date`,
      onSelect: () => open({ type: 'update-branch' }),
    },
  ];
  const mergeItems = mergeMethods.slice(1);

  const extraItems: DropdownMenuItem[] = extras.map(x => ({
    label: menuLabel(x.icon, x.label),
    title: x.title,
    disabled: x.disabled,
    onSelect: x.onClick,
  }));

  // menuItems is the flat overflow list: PR actions (only when OPEN) then extras.
  const menuItems: DropdownMenuItem[] = [
    ...(isOpen
      ? [
          {
            label: menuLabel(UiCheck, 'Approve'),
            title: 'Submit an approving review',
            disabled: loading,
            onSelect: () => open({ type: 'approve' as const }),
          },
          {
            label: menuLabel(UiGitMerge, 'Merge'),
            title: mergeDisabledReason ?? 'Merge options',
            disabled: mergeDisabled,
            onSelect: () => {},
            children: mergeMethods,
          },
          ...(behind > 0
            ? [
                {
                  label: menuLabel(UiSync, 'Update branch'),
                  title: `Branch is ${behind} commit${behind !== 1 ? 's' : ''} behind ${base}`,
                  disabled: loading,
                  onSelect: () => open({ type: 'update-branch' as const }),
                },
              ]
            : []),
        ]
      : []),
    ...extraItems,
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
    if (!pending) return;
    setBusy(true);
    setError('');
    try {
      if (pending.type === 'approve') {
        if (!nodeId) return;
        await postAction('/api/prs/approve', { repo: pr.repo, number: pr.number, nodeId, body: comment.trim() });
      } else if (pending.type === 'update-branch') {
        await postAction('/api/prs/update-branch', { repo: pr.repo, number: pr.number });
      } else {
        if (!nodeId) return;
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

  // Nothing to render for a merged/closed PR with no extra actions.
  if (menuItems.length === 0) return null;

  const modal = pending && (
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
  );

  // A lone action reads better as its own button than hidden behind a kebab, so
  // only collapse once there are at least two.
  if (collapsed && menuItems.length > 1) {
    return (
      <>
        <DropdownMenu
          align="right"
          title="Actions"
          menuLabel="Pull request actions"
          trigger={
            <Button
              variant="outline"
              size="sm"
              className="h-7 shrink-0 px-2 text-xs"
              aria-label="Pull request actions"
            >
              <UiEllipsis className="text-sm" />
            </Button>
          }
          items={menuItems}
        />
        {modal}
      </>
    );
  }

  return (
    <>
      {isOpen && (
        <>
          <Button
            variant="outline"
            size="sm"
            className="h-7 shrink-0 gap-1 px-2 text-xs"
            onClick={() => open({ type: 'approve' })}
            disabled={loading}
            title="Submit an approving review"
          >
            <UiCheck className="text-xs" />
            Approve
          </Button>

          <span title={mergeDisabledReason}>
            <SplitButton
              variant="outline"
              size="sm"
              label={menuLabel(UiGitMerge, 'Merge')}
              title="Merge options"
              disabled={mergeDisabled}
              onClick={() => open({ type: 'merge', method: 'rebase' })}
              items={mergeItems}
            />
          </span>

          {behind > 0 && (
            <Button
              variant="outline"
              size="sm"
              className="h-7 shrink-0 gap-1 px-2 text-xs"
              onClick={() => open({ type: 'update-branch' })}
              disabled={loading}
              title={`Branch is ${behind} commit${behind !== 1 ? 's' : ''} behind ${base} — update to bring it up to date`}
            >
              <UiSync className="text-xs" />
              Update
            </Button>
          )}
        </>
      )}

      {extras.map((x, i) => {
        const Icon = x.icon;
        return (
          <Button
            key={i}
            variant="outline"
            size="sm"
            className="h-7 shrink-0 gap-1 px-2 text-xs"
            onClick={x.onClick}
            disabled={x.disabled}
            title={x.title}
          >
            <Icon className="text-xs" />
            {x.label}
          </Button>
        );
      })}

      {modal}
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
    case 'update-branch':
      return {
        title: 'Update branch',
        confirmLabel: 'Update branch',
        description: `Merge ${base} into the head branch of ${ref} to bring it up to date. This adds a merge commit to the PR.`,
      };
  }
}
