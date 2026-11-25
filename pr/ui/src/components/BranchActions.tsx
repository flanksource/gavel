import { useState, type ReactNode } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button, Modal, type DropdownMenuItem } from '@flanksource/clicky-ui/components';
import { UiCopy, UiTrash } from '@flanksource/clicky-ui/icons';
import { copyText } from '../clipboard';
import { parseProjectRef } from '../projectRef';
import { fetchJSON, mutationJSON, queryKeys } from '../query';
import type { ProjectGit } from '../types';
import type { BranchDeletePlan, BranchDeleteRequest, BranchDeleteResult } from '../types/branchDelete';
import { invalidateProjectData } from './projectMutations';
import { useRebaseAction } from './rebaseAction';

export interface SelectedBranch {
  name: string;
  // Why the branch cannot be deleted here; "" when it can.
  deleteReason: string;
}

/**
 * The branch the ref picker has selected: a linked worktree's branch or a
 * branch without a worktree. The main checkout's branch and the base cannot be
 * deleted from here; a detached worktree has no branch at all.
 */
export function selectedBranch(git: ProjectGit, value: string): SelectedBranch | null {
  const ref = parseProjectRef(value);
  const name = ref.kind === 'branch' ? ref.name
    : ref.kind === 'worktree' ? git.worktrees.find(wt => wt.path === ref.path && !wt.detached)?.branch
      : git.worktrees.find(wt => wt.primary && !wt.detached)?.branch;
  if (!name) return null;
  const deleteReason = name === git.base ? `${name} is the base branch.`
    : ref.kind === 'main' ? `${name} is checked out in the main checkout; switch it to another branch first.`
      : '';
  return { name, deleteReason };
}

/**
 * The ref bar's operations on the selected branch — copy its name, rebase it,
 * delete it with its worktree — as menu items, plus the delete confirmation.
 */
export function useBranchActions({ projectName, git, value, onDeleted }: {
  projectName: string;
  git: ProjectGit | undefined;
  value: string;
  onDeleted: () => void;
}): { branch: string; items: DropdownMenuItem[]; dialog: ReactNode } {
  const [deleting, setDeleting] = useState('');
  const rebase = useRebaseAction();
  const branch = git ? selectedBranch(git, value) : null;
  if (!branch) return { branch: '', items: [], dialog: null };
  const items: DropdownMenuItem[] = [
    { label: <span className="flex items-center gap-2"><UiCopy />Copy branch name</span>, onSelect: () => void copyText(branch.name) },
    ...(rebase ? [rebase] : []),
    {
      label: <span className="flex items-center gap-2 text-destructive"><UiTrash />Delete {branch.name}…</span>,
      disabled: !!branch.deleteReason,
      title: branch.deleteReason || undefined,
      onSelect: () => setDeleting(branch.name),
    },
  ];
  const dialog = deleting && <BranchDeleteDialog projectName={projectName} branch={deleting} onClose={() => setDeleting('')} onDeleted={() => { setDeleting(''); onDeleted(); }} />;
  return { branch: branch.name, items, dialog };
}

const plural = (n: number, noun: string) => `${n} ${noun}${n === 1 ? '' : 's'}`;

function BranchDeleteDialog({ projectName, branch, onClose, onDeleted }: { projectName: string; branch: string; onClose: () => void; onDeleted: () => void }) {
  const client = useQueryClient();
  const url = `/api/projects/${encodeURIComponent(projectName)}/branch/delete`;
  const [discardUncommitted, setDiscardUncommitted] = useState(false);
  const [deleteUnmerged, setDeleteUnmerged] = useState(false);
  const plan = useQuery<BranchDeletePlan>({
    queryKey: ['projects', projectName, 'branch-delete', branch],
    queryFn: ({ signal }) => fetchJSON({ url: `${url}?branch=${encodeURIComponent(branch)}`, signal, context: `Plan deleting ${branch}` }),
    retry: false, staleTime: 0,
  });
  const mutation = useMutation({
    mutationFn: (body: BranchDeleteRequest) => mutationJSON<BranchDeleteResult>({ url, method: 'POST', body, context: `Delete ${branch}` }),
    onSuccess: onDeleted,
    onSettled: () => Promise.all([
      client.invalidateQueries({ queryKey: queryKeys.projectGit(projectName) }),
      client.invalidateQueries({ queryKey: queryKeys.projectGitSummary() }),
      invalidateProjectData(client, projectName),
    ]),
  });
  const data = plan.data;
  const acknowledged = !!data && (data.uncommitted.length === 0 || discardUncommitted) && (data.unmerged.length === 0 || deleteUnmerged);
  const confirm = () => {
    if (!data || !acknowledged) return;
    mutation.mutate({ branch, expectedHead: data.head, ...(discardUncommitted && { discardUncommitted }), ...(deleteUnmerged && { deleteUnmerged }) });
  };

  return (
    <Modal open onClose={() => { if (!mutation.isPending) onClose(); }} title={`Delete ${branch}`} size="md" footer={
      <div className="flex justify-end gap-2">
        <Button variant="outline" disabled={mutation.isPending} onClick={onClose}>Cancel</Button>
        <Button variant="destructive" disabled={!acknowledged || plan.isFetching} loading={mutation.isPending} onClick={confirm}>Delete {branch}</Button>
      </div>
    }>
      <div className="flex flex-col gap-3 text-sm">
        {plan.isPending && <p className="text-muted-foreground">Checking {branch}…</p>}
        {plan.error && <p role="alert" className="text-destructive">{plan.error.message}</p>}
        {data && <p>{data.worktree ? <>Removes the worktree <span className="break-all font-mono">{data.worktree}</span> and deletes</> : 'Deletes'} the branch <strong>{branch}</strong>. This cannot be undone.</p>}
        {data && data.uncommitted.length > 0 && (
          <PathList title={`${plural(data.uncommitted.length, 'uncommitted file')} will be lost:`} items={data.uncommitted.map(path => ({ key: path, text: path }))}>
            <label className="flex items-center gap-1.5"><input type="checkbox" checked={discardUncommitted} onChange={event => setDiscardUncommitted(event.currentTarget.checked)} />Discard {plural(data.uncommitted.length, 'uncommitted file')}</label>
          </PathList>
        )}
        {data && data.unmerged.length > 0 && (
          <PathList title={`${plural(data.unmerged.length, 'commit')} not in ${data.base} will be lost:`} items={data.unmerged.map(commit => ({ key: commit.sha, text: `${commit.sha.slice(0, 8)} ${commit.subject}` }))}>
            <label className="flex items-center gap-1.5"><input type="checkbox" checked={deleteUnmerged} onChange={event => setDeleteUnmerged(event.currentTarget.checked)} />Delete {plural(data.unmerged.length, 'unmerged commit')}</label>
          </PathList>
        )}
        {mutation.error && <p role="alert" className="whitespace-pre-wrap text-destructive">{mutation.error.message}</p>}
      </div>
    </Modal>
  );
}

function PathList({ title, items, children }: { title: string; items: { key: string; text: string }[]; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-1">
      <p className="text-amber-700 dark:text-amber-400">{title}</p>
      <ul className="max-h-40 list-none overflow-auto font-mono text-xs">
        {items.map(item => <li key={item.key} className="truncate" title={item.text}>{item.text}</li>)}
      </ul>
      {children}
    </div>
  );
}
