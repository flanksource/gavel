import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { SessionTail } from '@flanksource/clicky-ui/ai';
import { Button, Combobox, Modal } from '@flanksource/clicky-ui/components';
import { fetchJSON, HttpError, mutationJSON, queryKeys } from '../query';
import type { RebaseContext, RebaseOperation } from '../types/branchRebase';
import { projectDiffQueryKey } from './projectMutations';
import { usePromptRun } from './promptRun';

export const BRANCH_REBASE_PROMPT = 'branch-rebase';

export function ProjectRebase({ projectName, worktree, branch, generation }: { projectName: string; worktree?: string; branch?: string; generation?: number }) {
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

  return <div className="flex shrink-0 flex-wrap items-center gap-2 border-b border-border px-4 py-1.5 text-xs">
    <Button variant="outline" size="sm" disabled={!!reason || busy || !!active} title={reason || undefined} onClick={() => { mutation.reset(); setOpen(true); void context.refetch(); }}>Rebase…</Button>
    {notice && !active && <span role="status">{notice}</span>}
    {error && <span role="alert" className="max-w-xl whitespace-pre-wrap break-words text-destructive">{error}</span>}
    {active && <>
      <span>Rebase of {operation.branch}: {operation.phase}{operation.stashSha && ` · saved edits: ${operation.stashSha.slice(0, 8)}`}</span>
      {operation.conflicts.length > 0 && <span className="max-w-full break-words font-mono">{operation.conflicts.join(', ')}</span>}
      <Button variant="outline" size="sm" disabled={busy || !context.data?.executable} onClick={send}>Send to rebase prompt</Button>
      <Button variant="outline" size="sm" disabled={busy || operation.phase.startsWith('restore')} onClick={() => mutation.mutate({ abort: true, operation })}>Abort and restore edits</Button>
    </>}
    {ai.active && <span className="flex min-w-0 flex-1 items-center gap-2">
      {ai.state.status === 'running' && ai.state.sessionId ? <SessionTail src={`/api/captain/sessions/${encodeURIComponent(ai.state.sessionId)}`} className="min-w-0 flex-1" /> : 'Starting the rebase prompt…'}
      {ai.state.status === 'running' && <Button variant="ghost" size="sm" onClick={() => { void ai.stop(); }}>Stop</Button>}
    </span>}
    {ai.state.status === 'failed' && <span role="alert" className="text-destructive">{ai.state.error}</span>}
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
  </div>;
}
