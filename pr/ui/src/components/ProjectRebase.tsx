import { useState, type ReactNode } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { SessionTail } from '@flanksource/clicky-ui/ai';
import { Button, Combobox, Modal, type DropdownMenuItem } from '@flanksource/clicky-ui/components';
import { AnsiHtml } from '@flanksource/clicky-ui/data';
import { UiGitMerge, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import { fetchJSON, HttpError, mutationJSON, queryKeys } from '../query';
import type { RebaseContext, RebaseOperation } from '../types/branchRebase';
import { projectDiffQueryKey } from './projectMutations';
import { usePromptRun } from './promptRun';

export const BRANCH_REBASE_PROMPT = 'branch-rebase';

interface RebaseParts {
  // The Rebase… entry point, as an option for the PR split buttons' menus.
  action: DropdownMenuItem;
  // Notices, failures, recovery actions and the Rebase dialog; a block of its own.
  status: ReactNode;
}

interface Props {
  projectName: string;
  worktree?: string;
  branch?: string;
  generation?: number;
  children: (parts: RebaseParts) => ReactNode;
}

export function ProjectRebase({ projectName, worktree, branch, generation, children }: Props) {
  const client = useQueryClient();
  const [open, setOpen] = useState(false);
  const [notice, setNotice] = useState('');
  const [target, setTarget] = useState('refs/remotes/origin/main');
  const url = `/api/projects/${encodeURIComponent(projectName)}/branch/rebase`;
  const context = useQuery<RebaseContext>({
    queryKey: ['projects', projectName, 'rebase', worktree, generation],
    queryFn: ({ signal }) => fetchJSON({ url: `${url}?worktree=${encodeURIComponent(worktree ?? '')}`, signal, context: 'Load rebase options' }),
    enabled: !!worktree, retry: false,
  });
  const refresh = () => Promise.all([
    client.invalidateQueries({ queryKey: ['projects', projectName, 'rebase'] }),
    client.invalidateQueries({ queryKey: queryKeys.projectGit(projectName) }),
    client.invalidateQueries({ queryKey: queryKeys.projectGitSummary() }),
    client.invalidateQueries({ queryKey: queryKeys.projectStatusScope(projectName) }),
    client.invalidateQueries({ queryKey: projectDiffQueryKey(projectName) }),
  ]);
  const mutation = useMutation({
    onMutate: () => setNotice(''),
    mutationFn: ({ abort = false, operation }: { abort?: boolean; operation?: RebaseOperation }) => mutationJSON<RebaseOperation>({
      url: abort ? `${url}/abort` : url, method: 'POST', context: abort ? 'Abort rebase' : 'Rebase branch',
      body: { worktree, branch: context.data?.branch, target, expectedHead: context.data?.head, ...(operation && { operation: operation.id }) },
    }),
    onSuccess: result => setNotice(result.phase === 'aborted' ? 'Rebase aborted; saved edits restored.' : 'Rebase complete; saved edits restored.'),
    onSettled: refresh,
  });
  const ai = usePromptRun({ prompt: BRANCH_REBASE_PROMPT, label: 'rebase run', onSuccess: () => { mutation.reset(); setNotice('Rebase complete; saved edits restored.'); }, onSettled: () => { void refresh(); } });
  const body = mutation.error instanceof HttpError ? mutation.error.body as { operation?: RebaseOperation } | null : null;
  const operation = context.data?.operation ?? body?.operation;
  const busy = mutation.isPending || ai.active;
  const reason = !worktree ? 'Select a checked-out branch to rebase.' : !branch && !operation ? 'Detached worktrees cannot start a rebase.' : '';
  const error = mutation.error?.message || context.error?.message || operation?.error;
  const active = operation && operation.phase !== 'done' && operation.phase !== 'aborted';
  const send = () => {
    if (!operation || !context.data?.executable) return;
    void ai.start({ gavelExecutable: context.data.executable, worktree: operation.worktree, operation: operation.id, branch: operation.branch, target: operation.target, targetSha: operation.targetSha, phase: operation.phase, error: operation.error, conflicts: operation.conflicts, stashSha: operation.stashSha });
  };
  const openDialog = () => { mutation.reset(); setOpen(true); void context.refetch(); };
  const disabled = !!reason || busy || !!active;

  const action: DropdownMenuItem = { label: <span className="flex items-center gap-2"><UiGitMerge />Rebase…</span>, title: reason || undefined, disabled, onSelect: openDialog };

  const showStatus = (notice && !active) || error || active || ai.active || ai.state.status === 'failed';
  const status = <>
    {showStatus && <div className="flex min-w-0 shrink-0 flex-col gap-1.5 border-b border-border px-3 py-1.5 text-xs md:px-4">
      {notice && !active && <span role="status">{notice}</span>}
      {error && <RebaseFailure error={error} operation={operation} />}
      {active && <div className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1.5">
        <span className="min-w-0 break-words">Rebase of {operation.branch}: {operation.phase}{operation.stashSha && ` · saved edits: ${operation.stashSha.slice(0, 8)}`}</span>
        {!error && <ConflictPaths paths={operation.conflicts} />}
        <Button variant="outline" size="sm" disabled={busy || !context.data?.executable} onClick={send}>Send to rebase prompt</Button>
        <Button variant="outline" size="sm" disabled={busy || operation.phase.startsWith('restore')} onClick={() => mutation.mutate({ abort: true, operation })}>Abort and restore edits</Button>
      </div>}
      {ai.active && <span className="flex min-w-0 flex-1 items-center gap-2">
        {ai.state.status === 'running' && ai.state.sessionId ? <SessionTail src={`/api/captain/sessions/${encodeURIComponent(ai.state.sessionId)}`} className="min-w-0 flex-1" /> : 'Starting the rebase prompt…'}
        {ai.state.status === 'running' && <Button variant="ghost" size="sm" onClick={() => { void ai.stop(); }}>Stop</Button>}
      </span>}
      {ai.state.status === 'failed' && <span role="alert" className="text-destructive">{ai.state.error}</span>}
    </div>}
    <Modal open={open} onClose={() => { if (!mutation.isPending) setOpen(false); }} title="Rebase branch" size="md" footer={<>
      <Button variant="ghost" disabled={busy} onClick={() => setOpen(false)}>Cancel</Button>
      <Button disabled={busy || !context.data?.branch || context.isFetching || !!context.error} loading={mutation.isPending} onClick={() => { mutation.mutate({}, { onSettled: () => setOpen(false) }); }}>Rebase</Button>
    </>}>
      <div className="flex flex-col gap-3 text-sm">
        <p>Rebase <strong>{context.data?.branch}</strong> in <span className="break-all font-mono">{worktree}</span>.</p>
        <p>{context.data?.dirtyCount ?? 0} uncommitted files will be stashed and restored, including untracked files.</p>
        <label className="flex flex-col gap-1">Rebase from
          <Combobox ariaLabel="Rebase from" value={target} onChange={setTarget} options={(context.data?.targets ?? []).map(item => ({ value: item.ref, label: item.label }))} allowCustomValue={false} required />
        </label>
        <p className="text-muted-foreground">origin/main is fetched before rebasing. Local branches use their current local commit.</p>
        {context.error && <p role="alert" className="text-destructive">{context.error.message}</p>}
      </div>
    </Modal>
  </>;

  return <>{children({ action, status })}</>;
}

function ConflictPaths({ paths }: { paths: string[] }) {
  if (paths.length === 0) return null;
  return (
    <ul className="flex min-w-0 flex-wrap gap-x-3 gap-y-0.5 font-mono">
      {paths.map(path => <li key={path} className="max-w-full truncate" title={path}>{path}</li>)}
    </ul>
  );
}

// Git's stderr runs to dozens of lines; the first names the failure and the
// rest is evidence, so it sits behind a toggle in a bounded terminal panel.
function RebaseFailure({ error, operation }: { error: string; operation?: RebaseOperation }) {
  const [details, setDetails] = useState(false);
  const summary = error.split('\n')[0];
  return (
    <div role="alert" className="flex min-w-0 flex-col gap-1 rounded-md border border-destructive/30 bg-destructive/5 px-2 py-1.5">
      <div className="flex min-w-0 items-center gap-1.5 text-destructive">
        <UiWarningTriangle className="shrink-0" />
        <span data-testid="rebase-error-summary" className="min-w-0 flex-1 truncate" title={summary}>{summary}</span>
        {error !== summary && <Button variant="ghost" size="sm" className="h-6 shrink-0 px-1.5 text-xs" aria-expanded={details} onClick={() => setDetails(open => !open)}>Details</Button>}
      </div>
      {operation?.phase === 'restore-conflict' && <p className="text-muted-foreground">Your saved edits conflict with the rebased branch; stash {operation.stashSha.slice(0, 8)} is kept</p>}
      {operation && <ConflictPaths paths={operation.conflicts} />}
      {details && <div data-testid="rebase-error-details" className="max-h-48 overflow-auto rounded bg-black">
        <AnsiHtml text={error} className="whitespace-pre-wrap break-words px-2 py-1.5 text-[11px] leading-relaxed text-gray-100" />
      </div>}
    </div>
  );
}
