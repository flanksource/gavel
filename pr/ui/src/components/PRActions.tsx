import { useCallback, useLayoutEffect, useMemo } from 'react';
import { useIsMutating } from '@tanstack/react-query';
import { Button, DropdownMenu, useModal, type DropdownMenuItem } from '@flanksource/clicky-ui/components';
import { UiCheck, UiEllipsis, UiGitMerge, UiSync } from '@flanksource/clicky-ui/icons';
import type { IconProps } from '@flanksource/clicky-ui/icons';
import type { ComponentType } from 'react';
import type { PRItem, PRDetail } from '../types';
import { prActionFormKey, prActionMutationKey, type PendingPRAction, type PRActionAvailability } from './prActionDraft';

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

// PRActions renders merge / auto-merge / approve / update-branch controls for an
// OPEN PR. Each action is confirmed in a modal before it hits the PR host. On success
// it calls onChanged so the host can re-fetch the detail and reflect the new state
// (merged / approved), at which point these controls disappear or update.
export function PRActions({
  pr,
  detail,
  onChanged,
  collapsed = false,
  extras = [],
  disabled = false,
}: {
  pr: PRItem;
  detail: PRDetail | null;
  onChanged?: () => void;
  // collapsed folds every action into an overflow dropdown — set by the detail
  // panel when its container is too narrow to lay the buttons out inline.
  collapsed?: boolean;
  extras?: ExtraAction[];
  disabled?: boolean;
}) {
  const { openModal } = useModal();
  const actionBusy = useIsMutating({ mutationKey: prActionMutationKey(pr.repo, pr.number), exact: true }) > 0;
  const availability = useMemo<PRActionAvailability>(() => ({ disabled: false, changes: new EventTarget(), anchor: null }), [pr.repo, pr.number]);
  const setRebaseAnchor = useCallback((node: HTMLSpanElement | null) => {
    availability.anchor = node;
    availability.changes.dispatchEvent(new Event('change'));
  }, [availability]);
  const info = detail?.pr;
  const azure = info?.provider === 'azuredevops';
  const nodeId = info?.nodeId;
  const state = (info?.state || pr.state || '').toUpperCase();
  const isDraft = info?.isDraft ?? pr.isDraft ?? false;
  const mergeable = (info?.mergeable || pr.mergeable || '').toUpperCase();
  const base = info?.baseRefName || pr.target;
  const behind = pr.behind ?? 0;
  // Only OPEN PRs are actionable; merged/closed ones have nothing to merge or
  // approve, but callers may still pass extras (e.g. New todo) to render.
  const isOpen = state === 'OPEN';
  const loading = disabled || actionBusy || !info || (!azure && !nodeId) || (azure && !info.headRefOid);
  useLayoutEffect(() => {
    availability.disabled = disabled;
    availability.changes.dispatchEvent(new Event('change'));
  }, [availability, disabled]);
  const conflicting = mergeable === 'CONFLICTING' || info?.mergeReadiness?.state === 'conflicting';
  const mergeDisabled = loading || isDraft || conflicting;
  const mergeDisabledReason = disabled
    ? 'Wait for the TODO turn to finish'
    : loading
    ? 'Loading PR details…'
    : isDraft
      ? 'PR is a draft — mark it ready for review first'
      : conflicting
        ? 'Resolve merge conflicts before merging'
        : undefined;

  // Merging opens one confirmation that picks the method; rebase is the default.
  const openMerge = () => open({ type: 'merge', method: 'rebase', auto: false });

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
          ...(!azure ? [{
            label: menuLabel(UiCheck, 'Approve'),
            title: 'Submit an approving review',
            disabled: loading,
            onSelect: () => open({ type: 'approve' as const }),
          }] : []),
          {
            label: menuLabel(UiGitMerge, 'Merge'),
            title: mergeDisabledReason ?? 'Merge options',
            disabled: mergeDisabled,
            onSelect: openMerge,
          },
          ...(!azure && behind > 0
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

  function open(pending: PendingPRAction) {
    if (loading || (pending.type === 'merge' && mergeDisabled)) return;
    openModal(prActionFormKey(pending.type), {
      repo: pr.repo, number: pr.number, url: pr.url, base,
      nodeId: nodeId ?? '', headSha: info?.headRefOid ?? '', azure,
      pending, comment: '', availability, onChanged,
    });
  }

  if (menuItems.length === 0) return null;

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
        <span ref={setRebaseAnchor} className="contents" />
      </>
    );
  }

  return (
    <>
      {isOpen && (
        <>
          {!azure && <Button
            variant="outline"
            size="sm"
            className="h-7 shrink-0 gap-1 px-2 text-xs"
            onClick={() => open({ type: 'approve' })}
            disabled={loading}
            title="Submit an approving review"
          >
            <UiCheck className="text-xs" />
            Approve
          </Button>}

          <span title={mergeDisabledReason}>
            <Button
              variant="outline"
              size="sm"
              className="h-7 shrink-0 gap-1 px-2 text-xs"
              disabled={mergeDisabled}
              onClick={openMerge}
            >
              <UiGitMerge className="text-xs" />
              Merge
            </Button>
          </span>

          {!azure && behind > 0 && (
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

      {extras.map((x) => {
        const Icon = x.icon;
        return (
          <Button
            key={x.label}
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

      <span ref={setRebaseAnchor} className="contents" />
    </>
  );
}
