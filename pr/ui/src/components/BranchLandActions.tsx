import { useState } from 'react';
import { InputField, SplitButton } from '@flanksource/clicky-ui/components';
import { UiGitMerge, UiGitPr } from '@flanksource/clicky-ui/icons';
import type { BranchMergeResult } from '../types';
import { landConflicts, useBranchLandMutation, type BranchLandRequest } from './projectGitMutations';

export interface BranchLandReasonInput {
  base: string;
  ahead: number;
  dirtyCount: number;
  baseCheckedOut: boolean;
  worktreePath?: string;
}

/**
 * Why the branch cannot be merged into the base branch or turned into a PR
 * yet, mirroring the server's refusals: nothing ahead of the base, a worktree
 * holding uncommitted changes, and (merge only) a main checkout that is not on
 * the base branch. An empty string means the action may be attempted.
 */
export function branchLandBlockedReasons({ base, ahead, dirtyCount, baseCheckedOut, worktreePath }: BranchLandReasonInput): { merge: string; pr: string } {
  const common = ahead === 0
    ? `The branch has no commits ahead of ${base}.`
    : dirtyCount > 0
      ? `${dirtyCount} uncommitted change${dirtyCount === 1 ? '' : 's'} in ${worktreePath || 'the worktree'}: commit or discard them before landing.`
      : '';
  const merge = common || (baseCheckedOut ? '' : `The main checkout is not on ${base}: check out ${base} there before merging.`);
  return { merge, pr: common };
}

export interface BranchLandActionsProps extends BranchLandReasonInput {
  projectName: string;
  branch: string;
  /** Called once a merge landed; the branch (and its worktree) are gone by then. */
  onMerged?: (result: BranchMergeResult) => void;
}

/**
 * Lands a project branch: "Squash & merge into <base>" (or each commit from the
 * menu, an optional message for the squash commit), and "Create PR" (draft from
 * the menu). Disabled with the reason as a tooltip while landing is refused;
 * server refusals render inline, listing conflicting paths when it names them.
 */
export function BranchLandActions({ projectName, branch, onMerged, ...reasonInput }: BranchLandActionsProps) {
  const land = useBranchLandMutation({ projectName, onMerged });
  const [message, setMessage] = useState('');
  const reasons = branchLandBlockedReasons(reasonInput);
  const submit = (request: BranchLandRequest) => land.mutate(request);
  const merge = (mode: 'squash' | 'incremental') => {
    const text = message.trim();
    submit({ via: 'merge', branch, mode, ...(mode === 'squash' && text ? { message: text } : {}) });
  };
  const conflicts = landConflicts(land.error);
  const merging = land.isPending && land.variables?.via === 'merge';
  const opening = land.isPending && land.variables?.via === 'pr';
  const opened = land.data?.via === 'pr' ? land.data : null;

  return (
    <span className="flex min-w-0 flex-col items-end gap-1">
      <span className="flex flex-wrap items-center justify-end gap-1.5">
        <InputField
          aria-label="Squash commit message"
          placeholder="Squash message (optional)"
          value={message}
          disabled={land.isPending}
          onChange={setMessage}
          className="h-8 w-56 text-xs"
        />
        <span className="inline-flex" title={reasons.merge || undefined}>
          <SplitButton
            variant="outline"
            size="sm"
            disabled={!!reasons.merge || land.isPending}
            loading={merging}
            title={reasons.merge || 'Merge options'}
            label={<span className="flex items-center gap-1"><UiGitMerge />Squash &amp; merge into {reasonInput.base}</span>}
            onClick={() => merge('squash')}
            items={[{
              label: <span className="flex items-center gap-2"><UiGitMerge />Merge each commit (cherry-pick)</span>,
              onSelect: () => merge('incremental'),
            }]}
          />
        </span>
        <span className="inline-flex" title={reasons.pr || undefined}>
          <SplitButton
            variant="outline"
            size="sm"
            disabled={!!reasons.pr || land.isPending}
            loading={opening}
            title={reasons.pr || 'Pull request options'}
            label={<span className="flex items-center gap-1"><UiGitPr />Create PR</span>}
            onClick={() => submit({ via: 'pr', branch, draft: false })}
            items={[{
              label: <span className="flex items-center gap-2"><UiGitPr />Create draft PR</span>,
              onSelect: () => submit({ via: 'pr', branch, draft: true }),
            }]}
          />
        </span>
      </span>
      {opened && (
        <span role="status" className="text-[11px] text-muted-foreground">
          Opened <a href={opened.url} target="_blank" rel="noreferrer" className="text-primary underline">#{opened.number}</a> from {opened.topicBranch}
        </span>
      )}
      {land.error && (
        <span role="alert" className="max-w-md whitespace-pre-wrap text-right text-[11px] text-red-600">
          {land.error.message}
          {conflicts.length > 0 && <ul className="mt-0.5 list-none font-mono">{conflicts.map(path => <li key={path}>{path}</li>)}</ul>}
        </span>
      )}
    </span>
  );
}
