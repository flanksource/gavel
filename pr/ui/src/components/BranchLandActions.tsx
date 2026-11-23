import { useState } from 'react';
import { SessionTail } from '@flanksource/clicky-ui/ai';
import { Button, InputField, Modal, SegmentedControl, SplitButton, type SegmentedOption } from '@flanksource/clicky-ui/components';
import { UiGitCommit, UiGitMerge, UiGitPr, UiSparkles } from '@flanksource/clicky-ui/icons';
import { useProjectGit } from '../projectGitQueries';
import type { BranchMergeMode, BranchMergeResult, GitWorktree } from '../types';
import { useBranchAIMerge } from './branchAIMerge';
import { landConflicts, useBranchLandMutation, type BranchLandRequest } from './projectGitMutations';
import { useRebaseAction } from './rebaseAction';

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

/**
 * The branches besides the base a branch can be merged into: each one checked
 * out in a project checkout, since the merge lands in that checkout's working
 * tree. The branch itself, detached and prunable checkouts are left out.
 */
export function mergeTargets({ worktrees, base, branch }: { worktrees: readonly GitWorktree[]; base: string; branch: string }): { branch: string; path: string }[] {
  return worktrees
    .filter(worktree => !worktree.detached && !worktree.prunable && worktree.branch && worktree.branch !== base && worktree.branch !== branch)
    .map(worktree => ({ branch: worktree.branch, path: worktree.path }))
    .sort((a, b) => a.branch.localeCompare(b.branch));
}

export interface BranchLandActionsProps extends BranchLandReasonInput {
  projectName: string;
  branch: string;
  /** Called once a merge landed; the branch (and its worktree) are gone by then. */
  onMerged?: (result: BranchMergeResult) => void;
}

/** How the merge dialog lands the branch: a plain squash or cherry-pick, or the AI merge prompt. */
type LandMethod = BranchMergeMode | 'ai';

const CONFIRM_LABEL: Record<LandMethod, string> = {
  squash: 'Squash & merge',
  incremental: 'Merge each commit',
  ai: 'Merge with AI',
};

function landMethods(base: string): SegmentedOption<LandMethod>[] {
  return [
    { id: 'squash', label: 'Squash', icon: UiGitMerge, title: `Squash every commit into one commit on ${base}` },
    { id: 'incremental', label: 'Each commit', icon: UiGitCommit, title: `Cherry-pick each commit onto ${base}` },
    { id: 'ai', label: 'With AI', icon: UiSparkles, title: `Rebase onto ${base} in the worktree, resolve conflicts with an agent, then squash into ${base}` },
  ];
}

function landEffect(method: LandMethod, base: string): string {
  if (method === 'squash') return `Squashes every commit into one commit on ${base}.`;
  if (method === 'incremental') return `Cherry-picks each commit onto ${base}.`;
  return `An agent rebases onto ${base} in the worktree, resolves conflicts, then squashes into ${base}.`;
}

/**
 * Lands a project branch. "Merge into <base>" opens a confirmation that picks
 * how — squash (with an optional commit message), each commit, or the AI merge
 * prompt, which rebases in the worktree, resolves conflicts and lands — and
 * says there why landing is refused, if it is. Server refusals render in the
 * dialog, listing conflicting paths with an offer to resolve them with AI; the
 * AI merge's latest message or tool call tails there until it finishes.
 * "Create PR" (draft from the menu) lands through a pull request instead.
 */
export function BranchLandActions({ projectName, branch, onMerged, ...reasonInput }: BranchLandActionsProps) {
  const { base } = reasonInput;
  const land = useBranchLandMutation({ projectName, onMerged });
  const ai = useBranchAIMerge({ projectName, onMerged });
  const worktrees = useProjectGit(projectName).data?.worktrees ?? [];
  const targets = mergeTargets({ worktrees, base, branch });
  const rebaseAction = useRebaseAction();
  const [open, setOpen] = useState(false);
  const [target, setTarget] = useState(base);
  const [method, setMethod] = useState<LandMethod>('squash');
  const [message, setMessage] = useState('');
  const reasons = branchLandBlockedReasons(reasonInput);
  // The base needs the primary checkout on it; any other target is offered only
  // because a checkout already has it, so only the branch's own state can block it.
  const mergeReason = target === base ? reasons.merge : reasons.pr;
  const repo = target === base ? worktrees.find(worktree => worktree.primary)?.path : targets.find(candidate => candidate.branch === target)?.path;
  const aiReason = branchAIMergeBlockedReason({ merge: mergeReason, worktreePath: reasonInput.worktreePath, repo });
  const reason = method === 'ai' ? aiReason : mergeReason;
  const openMerge = (next: string) => { setTarget(next); land.reset(); setOpen(true); };
  const aiActive = ai.state.status === 'starting' || ai.state.status === 'running';
  const busy = land.isPending || aiActive;
  const submit = (request: BranchLandRequest) => land.mutate(request);
  const conflicts = landConflicts(land.error);
  const mergeWithAI = (mode: BranchMergeMode) => {
    if (!repo || !reasonInput.worktreePath) return;
    land.reset();
    void ai.start({
      repo, worktree: reasonInput.worktreePath, branch, base: target, mode,
      message, conflicts, ahead: reasonInput.ahead,
    });
  };
  const confirm = () => {
    if (reason || busy) return;
    if (method === 'ai') return mergeWithAI('squash');
    const text = message.trim();
    land.mutate(
      { via: 'merge', branch, mode: method, ...(method === 'squash' && text ? { message: text } : {}), ...(target !== base ? { target } : {}) },
      { onSuccess: () => setOpen(false) },
    );
  };
  const mergeFailed = land.variables?.via === 'merge' && land.error;
  const failedMode: BranchMergeMode = land.variables?.via === 'merge' ? land.variables.mode : 'squash';
  const merging = land.isPending && land.variables?.via === 'merge';
  const opening = land.isPending && land.variables?.via === 'pr';
  const opened = land.data?.via === 'pr' ? land.data : null;

  const dialog = open && (
    <Modal
      open
      onClose={() => setOpen(false)}
      title={`Merge ${branch} into ${target}`}
      size="lg"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => setOpen(false)}>{aiActive ? 'Close' : 'Cancel'}</Button>
          <Button onClick={confirm} disabled={!!reason || busy} loading={merging || aiActive}>
            {method === 'ai' ? <UiSparkles /> : <UiGitMerge />}{CONFIRM_LABEL[method]}
          </Button>
        </div>
      }
    >
      <div className="space-y-3 text-sm text-foreground">
        <SegmentedControl aria-label="Merge method" size="sm" value={method} options={landMethods(target)} onChange={setMethod} />
        <p className="text-muted-foreground">{landEffect(method, target)}</p>
        {method !== 'incremental' && (
          <InputField
            aria-label="Squash commit message"
            placeholder="Squash message (optional)"
            value={message}
            disabled={busy}
            onChange={setMessage}
            className="h-8 w-full text-xs"
          />
        )}
        {reason && <p className="text-amber-700 dark:text-amber-400">{reason}</p>}
        {mergeFailed && (
          <div role="alert" className="whitespace-pre-wrap text-destructive">
            {land.error!.message}
            {conflicts.length > 0 && <ul className="mt-0.5 list-none font-mono text-xs">{conflicts.map(path => <li key={path}>{path}</li>)}</ul>}
          </div>
        )}
        {conflicts.length > 0 && !aiActive && (
          <span className="inline-flex" title={aiReason || undefined}>
            <Button variant="outline" size="sm" className="h-7 text-xs" disabled={!!aiReason} onClick={() => mergeWithAI(failedMode)}>
              <UiSparkles />Resolve with AI &amp; {failedMode === 'squash' ? 'squash' : 'merge each commit'}
            </Button>
          </span>
        )}
        {aiActive && (
          <div className="flex items-center gap-2 text-xs">
            {ai.state.status === 'running' && ai.state.sessionId
              ? <SessionTail src={`/api/captain/sessions/${encodeURIComponent(ai.state.sessionId)}`} className="min-w-0 flex-1" />
              : <span className="min-w-0 flex-1 text-muted-foreground">Starting the AI merge…</span>}
            {ai.state.status === 'running' && (
              <Button variant="ghost" size="sm" className="h-6 px-2 text-xs" onClick={() => { void ai.stop(); }}>Stop</Button>
            )}
          </div>
        )}
        {ai.state.status === 'failed' && (
          <div role="alert" className="whitespace-pre-wrap text-destructive">AI merge failed: {ai.state.error}</div>
        )}
      </div>
    </Modal>
  );

  return (
    <span className="flex min-w-0 flex-col items-end gap-1 max-sm:!items-stretch">
      <span className="flex flex-wrap items-center justify-end gap-1.5 max-sm:!justify-start">
        <span className="inline-flex" title={reasons.merge || undefined}>
          <SplitButton
            variant="outline"
            size="sm"
            title="Merge target"
            label={aiActive
              ? <span className="flex items-center gap-1"><UiSparkles />Merging with AI…</span>
              : <span className="flex items-center gap-1"><UiGitMerge />Merge into {base}</span>}
            onClick={() => (aiActive ? setOpen(true) : openMerge(base))}
            items={aiActive ? [] : targets.map(candidate => ({
              label: <span className="flex items-center gap-2"><UiGitMerge />Merge into {candidate.branch}</span>,
              title: candidate.path,
              onSelect: () => openMerge(candidate.branch),
            }))}
          />
        </span>
        <span className="inline-flex" title={reasons.pr || undefined}>
          <SplitButton
            variant="outline"
            size="sm"
            primaryDisabled={!!reasons.pr || busy}
            loading={opening}
            title={reasons.pr || 'Pull request options'}
            label={<span className="flex items-center gap-1"><UiGitPr />Create PR</span>}
            onClick={() => submit({ via: 'pr', branch, draft: false })}
            items={[{
              label: <span className="flex items-center gap-2"><UiGitPr />Create draft PR</span>,
              disabled: !!reasons.pr || busy,
              onSelect: () => submit({ via: 'pr', branch, draft: true }),
            }, ...(rebaseAction ? [rebaseAction] : [])]}
          />
        </span>
      </span>
      {opened && (
        <span role="status" className="text-[11px] text-muted-foreground">
          Opened <a href={opened.url} target="_blank" rel="noreferrer" className="text-primary underline">#{opened.number}</a> from {opened.topicBranch}
        </span>
      )}
      {land.variables?.via === 'pr' && land.error && (
        <span role="alert" className="max-w-md whitespace-pre-wrap text-right text-[11px] text-red-600 max-sm:!text-left">
          {land.error.message}
        </span>
      )}
      {!open && ai.state.status === 'failed' && (
        <span role="alert" className="max-w-md whitespace-pre-wrap text-right text-[11px] text-red-600 max-sm:!text-left">
          AI merge failed: {ai.state.error}
        </span>
      )}
      {dialog}
    </span>
  );
}
