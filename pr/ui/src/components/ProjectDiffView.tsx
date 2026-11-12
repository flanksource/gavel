import { useMemo } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { GitDiffPanel, type DiffLineSelection, type GitDiffPayload } from '@flanksource/clicky-ui/data';
import { UiDiff } from '@flanksource/clicky-ui/icons';
import { fetchJSON, mutationJSON, queryKeys } from '../query';
import { projectApiUrl } from './projectApiUrl';
import { invalidateProjectData } from './projectMutations';

interface Props {
  projectName: string;
  path: string;
  refreshKey?: number;
  // Absolute path of a linked worktree whose working tree is diffed; omitted
  // for the main checkout.
  worktree?: string;
}

// ProjectDiffPayload is the diff endpoint's answer: the combined diff, plus the
// HEAD→index (staged) and index→worktree (unstaged) halves of a single file.
// Both halves are empty for a folder target.
interface ProjectDiffPayload extends GitDiffPayload {
  staged: string;
  unstaged: string;
}

type LineTarget = { side: 'old' | 'new'; line: number };

export function ProjectDiffView({ projectName, path, refreshKey = 0, worktree }: Props) {
  const queryClient = useQueryClient();
  const diff = useQuery({
    queryKey: queryKeys.projectDiff(projectName, path, refreshKey, worktree),
    queryFn: ({ signal }) => fetchJSON<ProjectDiffPayload>({
      url: projectApiUrl({ projectName, resource: 'diff', query: [['path', path]], worktree }),
      signal,
      context: `Failed to load diff for ${path}`,
    }),
    enabled: path !== '',
    staleTime: Infinity,
  });
  const stageLines = useMutation({
    mutationFn: ({ targets, unstage }: { targets: LineTarget[]; unstage: boolean }) => mutationJSON({
      url: projectApiUrl({ projectName, resource: 'stage/lines', worktree }),
      method: 'POST',
      body: { path, ...(unstage ? { unstage: true } : {}), lines: targets.map(({ side, line }) => ({ side, line })) },
      context: `Failed to ${unstage ? 'unstage' : 'stage'} lines of ${path}`,
    }),
    onSuccess: () => invalidateProjectData(queryClient, projectName),
  });
  const mutate = stageLines.mutate;
  const stagedSelection = useMemo(() => lineSelection({ staged: true, disabled: stageLines.isPending, onChange: (targets, unstage) => mutate({ targets, unstage }) }), [mutate, stageLines.isPending]);
  const unstagedSelection = useMemo(() => lineSelection({ staged: false, disabled: stageLines.isPending, onChange: (targets, unstage) => mutate({ targets, unstage }) }), [mutate, stageLines.isPending]);

  if (!path) {
    return (
      <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground">
        <div className="text-center"><UiDiff className="mx-auto mb-2 text-3xl" />Select a file or folder to view its diff</div>
      </div>
    );
  }

  const data = diff.data;
  const error = diff.error instanceof Error ? diff.error.message : '';
  const gutter = data !== undefined && (data.staged !== '' || data.unstaged !== '');
  return (
    <div className="flex h-full min-h-0 flex-col">
      <div className="flex shrink-0 items-center gap-2 border-b border-border px-4 py-3">
        <UiDiff className="shrink-0 text-muted-foreground" />
        <h3 className="min-w-0 flex-1 truncate font-mono text-sm font-semibold" title={path}>{path}</h3>
      </div>
      {stageLines.error && <div role="alert" className="shrink-0 border-b border-border px-4 py-2 text-xs text-red-600 dark:text-red-400">{stageLines.error.message}</div>}
      {gutter ? (
        <div className="min-h-0 flex-1 overflow-auto">
          {data.unstaged !== '' && <DiffSection title="Unstaged changes" payload={{ ...data, diff: data.unstaged }} selection={unstagedSelection} />}
          {data.staged !== '' && <DiffSection title="Staged changes" payload={{ ...data, diff: data.staged }} selection={stagedSelection} />}
        </div>
      ) : (
        <GitDiffPanel
          loading={diff.isPending}
          payload={data ?? null}
          error={error}
          className="min-h-0 flex-1 overflow-auto border-t-0"
          maxHeightClassName="max-h-none"
        />
      )}
    </div>
  );
}

function DiffSection({ title, payload, selection }: { title: string; payload: GitDiffPayload; selection: DiffLineSelection }) {
  return (
    <section aria-label={title} className="border-b border-border last:border-b-0">
      <h4 className="sticky top-0 z-10 border-b border-border bg-muted px-4 py-1.5 text-xs font-semibold text-muted-foreground">{title}</h4>
      <GitDiffPanel payload={payload} lineSelection={selection} className="border-t-0" maxHeightClassName="max-h-none" />
    </section>
  );
}

// lineSelection is the gutter of one section: every box is on in the staged
// section and off in the unstaged one, so a click always moves lines across the
// index. onChange receives whether the lines should end up unstaged.
function lineSelection({ staged, disabled, onChange }: { staged: boolean; disabled: boolean; onChange: (targets: LineTarget[], unstage: boolean) => void }): DiffLineSelection {
  const verb = (selected: boolean) => selected ? 'Unstage' : 'Stage';
  return {
    isSelected: () => staged,
    onChange: (targets, selected) => onChange(targets, !selected),
    disabled,
    label: (target, selected) => `${verb(selected)} ${target.side === 'new' ? 'added' : 'removed'} line ${target.line}`,
    hunkLabel: (hunk, selected) => `${verb(selected)} hunk ${hunk}`,
  };
}
