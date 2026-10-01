import { Button, SplitButton } from '@flanksource/clicky-ui/components';
import { UiGitMerge, UiGitPr } from '@flanksource/clicky-ui/icons';
import type { LatestRunWorkspace } from './runWorkspace';
import { TodoRunLandingSummary } from './TodoRunLandingSummary';
import { useTodoLandMutation, type TodoLandRequest } from './todoMutations';

/**
 * Why the run's branch cannot be landed yet, mirroring the server's refusals
 * (todos/land): no run branch, a worktree kept with uncommitted paths, or
 * nothing past setup. Empty when landing may be attempted.
 */
export function landBlockedReason({ workspace }: LatestRunWorkspace): string {
  const worktree = workspace.worktree;
  if (!worktree?.branch) return 'The run worked in the checkout itself, so there is no run branch to land.';
  const dirty = worktree.dirty?.length ?? 0;
  if (worktree.kept && dirty > 0) {
    return `The run's worktree was kept with ${dirty} uncommitted path${dirty === 1 ? '' : 's'}: commit or discard them in ${worktree.path || 'the worktree'} before landing.`;
  }
  if (!worktree.setup || !worktree.head || worktree.setup === worktree.head) return 'The run made no commits to land.';
  return '';
}

export interface TodoRunLandActionsProps {
  dir: string;
  todoRef: string;
  latest: LatestRunWorkspace;
}

/**
 * Lands the todo's latest run branch — "Merge into current branch" or "Create
 * PR" (draft from the menu) — and, once landed, shows the landing in place of
 * the actions. Server refusals (409) and a recorded landing whose cleanup
 * failed (500) render inline.
 */
export function TodoRunLandActions({ dir, todoRef, latest }: TodoRunLandActionsProps) {
  const land = useTodoLandMutation(dir, todoRef);
  const landing = latest.attempt.landing ?? land.data?.landing;
  const blocked = landBlockedReason(latest);
  const disabled = !!blocked || land.isPending;
  const submit = (request: TodoLandRequest) => land.mutate(request);

  return (
    <span className="flex min-w-0 flex-col items-end gap-1">
      {landing ? (
        <TodoRunLandingSummary landing={landing} />
      ) : (
        <span className="inline-flex items-center gap-1.5" title={blocked || undefined}>
          <Button
            type="button"
            variant="outline"
            size="sm"
            disabled={disabled}
            loading={land.isPending && land.variables?.via === 'merge'}
            title={blocked || 'Cherry-pick the run\'s commits onto the checked-out branch'}
            onClick={() => submit({ via: 'merge' })}
          >
            <span className="flex items-center gap-1"><UiGitMerge />Merge into current branch</span>
          </Button>
          <SplitButton
            variant="outline"
            size="sm"
            disabled={disabled}
            loading={land.isPending && land.variables?.via === 'pr'}
            title={blocked || 'Pull request options'}
            label={<span className="flex items-center gap-1"><UiGitPr />Create PR</span>}
            onClick={() => submit({ via: 'pr' })}
            items={[{
              label: <span className="flex items-center gap-2"><UiGitPr />Create draft PR</span>,
              onSelect: () => submit({ via: 'pr', draft: true }),
            }]}
          />
        </span>
      )}
      {land.error && (
        <span role="alert" className="max-w-md whitespace-pre-wrap text-right text-[11px] text-red-600">
          {land.error.message}
        </span>
      )}
    </span>
  );
}
