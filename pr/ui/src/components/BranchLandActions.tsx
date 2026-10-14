import { useState } from 'react';
import { SessionTail } from '@flanksource/clicky-ui/ai';
import { Button, InputField, SplitButton } from '@flanksource/clicky-ui/components';
import { UiGitMerge, UiGitPr, UiSparkles } from '@flanksource/clicky-ui/icons';
import { useProjectGit } from '../projectGitQueries';
import type { BranchMergeMode, BranchMergeResult } from '../types';
import { useBranchAIMerge } from './branchAIMerge';
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

/**
 * Why the AI merge cannot start: everything that refuses a merge, plus the
 * prompt's own needs — a linked worktree for the agent to rebase in, and the
 * primary checkout to land into.
 */
export function branchAIMergeBlockedReason({ merge, worktreePath, repo }: { merge: string; worktreePath?: string; repo?: string }): string {
  if (merge) return merge;
  if (!worktreePath) return 'The AI merge rebases in the branch\'s worktree, and this branch has none.';
  if (!repo) return 'Loading the project\'s primary checkout…';
  return '';
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
 * "Merge with AI" (menu, or beside a conflict) runs the branch-merge prompt,
 * which rebases in the worktree, resolves conflicts and lands; its latest
 * message or tool call tails inline until it finishes.
 */
export function BranchLandActions({ projectName, branch, onMerged, ...reasonInput }: BranchLandActionsProps) {
  const land = useBranchLandMutation({ projectName, onMerged });
  const ai = useBranchAIMerge({ projectName, onMerged });
  const repo = useProjectGit(projectName).data?.worktrees.find(worktree => worktree.primary)?.path;
  const [message, setMessage] = useState('');
  const reasons = branchLandBlockedReasons(reasonInput);
  const aiReason = branchAIMergeBlockedReason({ merge: reasons.merge, worktreePath: reasonInput.worktreePath, repo });
  const aiActive = ai.state.status === 'starting' || ai.state.status === 'running';
  const busy = land.isPending || aiActive;
  const submit = (request: BranchLandRequest) => land.mutate(request);
  const merge = (mode: BranchMergeMode) => {
    const text = message.trim();
    submit({ via: 'merge', branch, mode, ...(mode === 'squash' && text ? { message: text } : {}) });
  };
  const conflicts = landConflicts(land.error);
  const mergeWithAI = (mode: BranchMergeMode) => {
    if (!repo || !reasonInput.worktreePath) return;
    land.reset();
    void ai.start({
      repo, worktree: reasonInput.worktreePath, branch, base: reasonInput.base, mode,
      message, conflicts, ahead: reasonInput.ahead,
    });
  };
  const failedMode: BranchMergeMode = land.variables?.via === 'merge' ? land.variables.mode : 'squash';
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
          disabled={busy}
          onChange={setMessage}
          className="h-8 w-56 text-xs"
        />
        <span className="inline-flex" title={reasons.merge || undefined}>
          <SplitButton
            variant="outline"
            size="sm"
            disabled={!!reasons.merge || busy}
            loading={merging || aiActive}
            title={reasons.merge || 'Merge options'}
            label={<span className="flex items-center gap-1"><UiGitMerge />Squash &amp; merge into {reasonInput.base}</span>}
            onClick={() => merge('squash')}
            items={[{
              label: <span className="flex items-center gap-2"><UiGitMerge />Merge each commit (cherry-pick)</span>,
              onSelect: () => merge('incremental'),
            }, {
              label: <span className="flex items-center gap-2"><UiSparkles />Merge with AI (rebase &amp; resolve conflicts)</span>,
              title: aiReason || `Rebase onto ${reasonInput.base} in the worktree, resolve conflicts with an agent, then squash into ${reasonInput.base}`,
              disabled: !!aiReason,
              onSelect: () => mergeWithAI('squash'),
            }]}
          />
        </span>
        <span className="inline-flex" title={reasons.pr || undefined}>
          <SplitButton
            variant="outline"
            size="sm"
            disabled={!!reasons.pr || busy}
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
      {conflicts.length > 0 && !aiActive && (
        <span className="inline-flex" title={aiReason || undefined}>
          <Button variant="outline" size="sm" className="h-7 text-xs" disabled={!!aiReason} onClick={() => mergeWithAI(failedMode)}>
            <UiSparkles />Resolve with AI &amp; {failedMode === 'squash' ? 'squash' : 'merge each commit'}
          </Button>
        </span>
      )}
      {aiActive && (
        <span className="flex w-full max-w-md items-center justify-end gap-2 text-[11px]">
          {ai.state.status === 'running' && ai.state.sessionId
            ? <SessionTail src={`/api/captain/sessions/${encodeURIComponent(ai.state.sessionId)}`} className="min-w-0 flex-1" />
            : <span className="min-w-0 flex-1 text-right text-muted-foreground">Starting the AI merge…</span>}
          {ai.state.status === 'running' && (
            <Button variant="ghost" size="sm" className="h-6 px-2 text-xs" onClick={() => { void ai.stop(); }}>Stop</Button>
          )}
        </span>
      )}
      {ai.state.status === 'failed' && (
        <span role="alert" className="max-w-md whitespace-pre-wrap text-right text-[11px] text-red-600">
          AI merge failed: {ai.state.error}
        </span>
      )}
    </span>
  );
}
