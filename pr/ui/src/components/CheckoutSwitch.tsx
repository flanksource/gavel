import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Button, Modal, SegmentedControl, type SegmentedOption } from '@flanksource/clicky-ui/components';
import { UiGitBranch, UiRefresh } from '@flanksource/clicky-ui/icons';
import { fetchJSON, mutationJSON, queryKeys } from '../query';
import type { ProjectGit } from '../types';
import type { CheckoutSwitchPlan, CheckoutSwitchRequest, CheckoutSwitchResult, CheckoutSync } from '../types/checkoutSwitch';
import { invalidateProjectData } from './projectMutations';

export interface CheckoutSyncChoice {
  id: CheckoutSync;
  label: string;
  effect: string;
  // Why this choice is refused, mirroring the server; "" when it may be attempted.
  reason: string;
}

const plural = (n: number, noun: string) => `${n} ${noun}${n === 1 ? '' : 's'}`;

/**
 * The ways the primary checkout can move onto plan.branch, each with the
 * server's refusal for it: uncommitted tracked edits or another worktree
 * holding the branch refuse all of them; a fast-forward needs a branch with no
 * commits its upstream lacks; a reset needs every commit it drops to be held
 * by another ref or already landed upstream.
 */
export function checkoutSyncChoices(plan: CheckoutSwitchPlan): CheckoutSyncChoice[] {
  const { branch, upstream } = plan;
  const blocked = plan.dirtyCount > 0
    ? `${plural(plan.dirtyCount, 'tracked file')} in the checkout ${plan.dirtyCount === 1 ? 'has' : 'have'} uncommitted changes: commit or discard them first.`
    : plan.checkedOutIn ? `${branch} is checked out in ${plan.checkedOutIn}.` : '';
  const upstreamReason = !upstream ? `${branch} tracks no upstream.` : !plan.upstreamHead ? `${upstream} no longer exists.` : '';
  const target = upstream ?? 'its upstream';
  return [
    {
      id: 'fast-forward', label: 'Fast-forward', effect: `Fetches ${target}, moves ${branch} up to it and checks ${branch} out.`,
      reason: blocked || upstreamReason || (plan.ahead > 0 ? `${branch} has ${plural(plan.ahead, 'commit')} ${target} lacks, so it cannot fast-forward.` : ''),
    },
    {
      id: 'reset', label: `Reset to ${target}`, effect: `Fetches ${target}, points ${branch} at it and checks ${branch} out. ${branch}'s own commits stay on the other branches that hold them.`,
      reason: blocked || upstreamReason || (plan.unpreserved.length > 0 ? `Resetting would orphan ${plural(plan.unpreserved.length, 'commit')} no other branch holds.` : ''),
    },
    {
      id: 'none', label: 'Switch only', effect: `Checks ${branch} out as it is, without fetching.`,
      reason: blocked || (plan.current === branch ? `The checkout is already on ${branch}.` : ''),
    },
  ];
}

/** The choice to preselect: the first one the server would accept, else the first. */
export function defaultCheckoutSync(choices: CheckoutSyncChoice[]): CheckoutSync {
  return (choices.find(choice => !choice.reason) ?? choices[0]).id;
}

/**
 * "Switch to <base>…" when the primary checkout is on another branch, or
 * "Sync <base>…" when it is on base but behind its upstream: a dialog that
 * plans the move from the server, offers fast-forward, reset or a plain
 * switch with the reason each is refused, then performs it.
 */
export function CheckoutSwitch({ projectName, git }: { projectName: string; git: ProjectGit }) {
  const client = useQueryClient();
  const [open, setOpen] = useState(false);
  const [picked, setPicked] = useState<CheckoutSync | null>(null);
  const [notice, setNotice] = useState('');
  const branch = git.base;
  const url = `/api/projects/${encodeURIComponent(projectName)}/checkout/switch`;
  const plan = useQuery<CheckoutSwitchPlan>({
    queryKey: ['projects', projectName, 'checkout-switch', branch, git.generation],
    queryFn: ({ signal }) => fetchJSON({ url: `${url}?branch=${encodeURIComponent(branch)}`, signal, context: `Plan switching to ${branch}` }),
    enabled: open, retry: false,
  });
  const mutation = useMutation({
    mutationFn: (body: CheckoutSwitchRequest) => mutationJSON<CheckoutSwitchResult>({ url, method: 'POST', body, context: `Switch to ${branch}` }),
    onSuccess: result => {
      setNotice(`On ${result.branch} at ${result.head.slice(0, 7)}${result.previousBranchHead !== result.head ? ` (was ${result.previousBranchHead.slice(0, 7)})` : ''}; left ${result.previousBranch || 'a detached HEAD'}.`);
      setOpen(false);
    },
    onSettled: () => Promise.all([
      client.invalidateQueries({ queryKey: queryKeys.projectGit(projectName) }),
      client.invalidateQueries({ queryKey: queryKeys.projectGitSummary() }),
      client.invalidateQueries({ queryKey: ['projects', projectName, 'rebase'] }),
      invalidateProjectData(client, projectName),
    ]),
  });
  if (git.baseCheckedOut && git.upstreamBehind === 0) {
    return notice ? <span role="status" className="text-xs text-muted-foreground">{notice}</span> : null;
  }

  const choices = plan.data ? checkoutSyncChoices(plan.data) : [];
  const sync = picked ?? (choices.length > 0 ? defaultCheckoutSync(choices) : 'fast-forward');
  const choice = choices.find(candidate => candidate.id === sync);
  const label = git.baseCheckedOut ? `Sync ${branch}…` : `Switch to ${branch}…`;
  const options: SegmentedOption<CheckoutSync>[] = choices.map(({ id, label: text, reason }) => ({ id, label: text, title: reason || undefined }));
  const confirm = () => {
    if (!plan.data || !choice || choice.reason) return;
    mutation.mutate({ branch, expectedHead: plan.data.head, sync });
  };

  return (
    <>
      <Button variant="outline" size="sm" className="h-7 gap-1 px-2 text-xs" onClick={() => { mutation.reset(); setPicked(null); setOpen(true); }}>
        {git.baseCheckedOut ? <UiRefresh /> : <UiGitBranch />}{label}
      </Button>
      {open && (
        <Modal open onClose={() => { if (!mutation.isPending) setOpen(false); }} title={git.baseCheckedOut ? `Sync ${branch}` : `Switch the main checkout to ${branch}`} size="md" footer={
          <div className="flex justify-end gap-2">
            <Button variant="outline" disabled={mutation.isPending} onClick={() => setOpen(false)}>Cancel</Button>
            <Button disabled={!choice || !!choice.reason || plan.isFetching} loading={mutation.isPending} onClick={confirm}>{choice?.label ?? 'Switch'}</Button>
          </div>
        }>
          <div className="flex flex-col gap-3 text-sm">
            {plan.isPending && <p className="text-muted-foreground">Reading {branch}…</p>}
            {plan.error && <p role="alert" className="text-destructive">{plan.error.message}</p>}
            {plan.data && <CheckoutSwitchSummary plan={plan.data} />}
            {choices.length > 0 && <SegmentedControl aria-label="How to switch" size="sm" value={sync} options={options} onChange={setPicked} />}
            {choice && <p className="text-muted-foreground">{choice.effect}</p>}
            {choice?.reason && <p className="text-amber-700 dark:text-amber-400">{choice.reason}</p>}
            {plan.data && plan.data.unpreserved.length > 0 && (
              <div>
                <p>Commits only {branch} holds:</p>
                <ul className="mt-0.5 list-none font-mono text-xs">
                  {plan.data.unpreserved.map(commit => <li key={commit.sha} className="truncate" title={commit.subject}>{commit.sha.slice(0, 8)} {commit.subject}</li>)}
                </ul>
              </div>
            )}
            {mutation.error && <p role="alert" className="whitespace-pre-wrap text-destructive">{mutation.error.message}</p>}
          </div>
        </Modal>
      )}
      {notice && <span role="status" className="text-xs text-muted-foreground">{notice}</span>}
    </>
  );
}

function CheckoutSwitchSummary({ plan }: { plan: CheckoutSwitchPlan }) {
  const where = plan.current ? <>on <strong>{plan.current}</strong></> : 'detached';
  return (
    <div className="flex flex-col gap-1">
      <p>The main checkout <span className="break-all font-mono">{plan.repo}</span> is {where}.</p>
      {plan.upstream && plan.upstreamHead && (
        <p className="text-muted-foreground">
          {plan.branch} is {plural(plan.ahead, 'commit')} ahead of {plan.upstream} and {plan.behind} behind, as of the last fetch.
        </p>
      )}
    </div>
  );
}
