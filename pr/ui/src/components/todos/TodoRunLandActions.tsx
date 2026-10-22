import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button, InputField, SplitButton } from '@flanksource/clicky-ui/components';
import { UiGitMerge, UiGitPr } from '@flanksource/clicky-ui/icons';
import type { TodoCommitFile, TodoCommitFilesResponse } from '../../types';
import type { LatestRunWorkspace } from './runWorkspace';
import { TodoRunLandingSummary } from './TodoRunLandingSummary';
import { changeParams } from './TodoRunChanges';
import { todoQueryKeys } from './todoQueries';
import { useTodoLandMutation, type TodoLandRequest } from './todoMutations';
import { currentAuthMode, fetchJSON } from '../../query';
import { hostControlsVisible } from '../../hostedUI';

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

/** A file's label in the landing's file list: a rename names both of its paths. */
function fileLabel(file: TodoCommitFile): string {
  return file.previousPath ? `${file.previousPath} → ${file.path}` : file.path;
}

export interface TodoRunLandActionsProps {
  dir: string;
  todoRef: string;
  /** The PR title a hosted landing starts from. */
  todoTitle: string;
  latest: LatestRunWorkspace;
}

/**
 * Lands the todo's latest run branch — "Merge into current branch" or "Create
 * PR" (draft from the menu) — and, once landed, shows the landing in place of
 * the actions. A hosted run lands only as a PR, whose title starts as the
 * todo's and can be edited, carrying the files left ticked in the run's file
 * list. Server refusals (409) and a recorded landing whose cleanup failed
 * (500) render inline.
 */
export function TodoRunLandActions({ dir, todoRef, todoTitle, latest }: TodoRunLandActionsProps) {
  const land = useTodoLandMutation(dir, todoRef);
  const landing = latest.attempt.landing ?? land.data?.landing;
  const hosted = !hostControlsVisible(currentAuthMode());
  const [title, setTitle] = useState(todoTitle);
  const [unticked, setUnticked] = useState<ReadonlySet<string>>(new Set());
  const worktree = latest.workspace.worktree;
  const range = { base: worktree?.setup ?? '', hash: worktree?.head ?? '' };
  const files = useQuery({
    queryKey: todoQueryKeys.commitFiles(dir, range.hash, range.base),
    queryFn: ({ signal }) => fetchJSON<TodoCommitFilesResponse>({
      url: `/api/todos/commits/files?${changeParams(dir, range).toString()}`,
      signal,
      context: 'Could not list the files of the run',
    }),
    enabled: hosted && !landing && !landBlockedReason(latest),
  });
  const runFiles = files.data?.files ?? [];
  const blocked = landBlockedReason(latest) || (hosted ? hostedLandBlockedReason() : '');
  const disabled = !!blocked || land.isPending;
  const submit = (request: TodoLandRequest) => land.mutate(hosted
    ? {
      ...request,
      title: title.trim(),
      leaveOut: runFiles.filter(file => unticked.has(file.path))
        .flatMap(file => file.previousPath ? [file.previousPath, file.path] : [file.path]),
    }
    : request);
  const toggle = (path: string) => setUnticked(current => {
    const next = new Set(current);
    if (!next.delete(path)) next.add(path);
    return next;
  });

  function hostedLandBlockedReason(): string {
    if (files.isPending) return 'Loading the run\'s files…';
    if (files.isError) return files.error.message;
    if (!title.trim()) return 'A PR needs a title.';
    if (runFiles.every(file => unticked.has(file.path))) return 'Tick at least one file to land.';
    return '';
  }

  return (
    <span className="flex min-w-0 flex-col items-end gap-1">
      {landing ? (
        <TodoRunLandingSummary landing={landing} />
      ) : (
        <span className="inline-flex items-center gap-1.5" title={blocked || undefined}>
          {hosted && (
            <InputField
              aria-label="PR title"
              placeholder="PR title"
              value={title}
              disabled={land.isPending}
              onChange={setTitle}
              className="h-8 w-64 text-xs"
            />
          )}
          {/* A merge cherry-picks onto the host's checkout; a hosted run lands as a pull request. */}
          {!hosted && (
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
          )}
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
      {hosted && !landing && runFiles.length > 0 && (
        <fieldset className="flex max-w-md flex-col items-start gap-0.5 text-[11px]">
          <legend className="text-muted-foreground">Files in the PR</legend>
          {runFiles.map(file => (
            <label key={file.path} className="flex items-center gap-1.5 font-mono">
              <input
                type="checkbox"
                aria-label={fileLabel(file)}
                checked={!unticked.has(file.path)}
                disabled={land.isPending}
                onChange={() => toggle(file.path)}
              />
              <span>{fileLabel(file)}</span>
            </label>
          ))}
        </fieldset>
      )}
      {hosted && files.isError && !landing && (
        <span role="alert" className="max-w-md whitespace-pre-wrap text-right text-[11px] text-red-600">
          {files.error.message}
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
