import { useState, type ReactNode } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { compactAISpecRuntime, SessionTail, SpecRuntimeBar, type AISpecRuntimeValue } from '@flanksource/clicky-ui/ai';
import { Button, Combobox, Modal, SplitButton, type DropdownMenuItem } from '@flanksource/clicky-ui/components';
import { ActionError, AnsiHtml } from '@flanksource/clicky-ui/data';
import { UiCog, UiGitMerge, UiSparkles, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import { HttpError, mutationJSON, queryKeys } from '../query';
import { useProjectRebaseContext } from '../projectGitQueries';
import { chatModelsQuery, chatRuntimesQuery } from './oneShotQueries';
import type { RebaseOperation, RebaseStart, RebaseStrategy } from '../types/branchRebase';
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
  const [runtimeOpen, setRuntimeOpen] = useState(false);
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [notice, setNotice] = useState('');
  const [target, setTarget] = useState('refs/remotes/origin/main');
  const url = `/api/projects/${encodeURIComponent(projectName)}/branch/rebase`;
  const context = useProjectRebaseContext({ projectName, worktree, generation });
  const refresh = () => Promise.all([
    client.invalidateQueries({ queryKey: ['projects', projectName, 'rebase'] }),
    client.invalidateQueries({ queryKey: queryKeys.projectGit(projectName) }),
    client.invalidateQueries({ queryKey: queryKeys.projectGitSummary() }),
    client.invalidateQueries({ queryKey: queryKeys.projectStatusScope(projectName) }),
    client.invalidateQueries({ queryKey: projectDiffQueryKey(projectName) }),
  ]);
  const mutation = useMutation({
    onMutate: () => setNotice(''),
    mutationFn: ({ abort = false, operation, start }: { abort?: boolean; operation?: RebaseOperation; start?: RebaseStart }) => mutationJSON<RebaseOperation>({
      url: abort ? `${url}/abort` : url, method: 'POST', context: abort ? `Abort ${strategyNoun(operation?.strategy)}` : start?.strategy === 'merge' ? 'Merge into branch' : 'Rebase branch',
      body: { worktree, branch: context.data?.branch, expectedHead: context.data?.head, ...start, ...(operation && { operation: operation.id }) },
    }),
    onSuccess: result => setNotice(`${strategyNoun(result.strategy, true)} ${result.phase === 'aborted' ? 'aborted' : 'complete'}; saved edits restored.`),
    onSettled: refresh,
  });
  const ai = usePromptRun({ prompt: BRANCH_REBASE_PROMPT, label: 'rebase run', onSuccess: () => { mutation.reset(); setNotice('Rebase complete; saved edits restored.'); }, onSettled: () => { void refresh(); } });
  const body = mutation.error instanceof HttpError ? mutation.error.body as { operation?: RebaseOperation } | null : null;
  const operation = context.data?.operation ?? body?.operation;
  const busy = mutation.isPending || ai.active;
  const reason = !worktree ? 'Select a checked-out branch to rebase.' : !branch && !operation ? 'Detached worktrees cannot start a rebase.' : '';
  const error = mutation.error?.message || context.error?.message || operation?.error;
  const active = operation && operation.phase !== 'done' && operation.phase !== 'aborted';
  const send = (runtime: AISpecRuntimeValue = {}) => {
    if (!operation || !context.data?.executable) return;
    setRuntimeOpen(false);
    void ai.start({ gavelExecutable: context.data.executable, worktree: operation.worktree, operation: operation.id, branch: operation.branch, target: operation.target, targetSha: operation.targetSha, phase: operation.phase, error: operation.error, conflicts: operation.conflicts, stashSha: operation.stashSha }, { spec: compactAISpecRuntime(runtime) });
  };
  const openDialog = () => { mutation.reset(); setOpen(true); void context.refetch(); };
  const startDisabled = busy || !context.data?.branch || context.isFetching || !!context.error;
  const begin = (start: RebaseStart) => {
    setAdvancedOpen(false);
    mutation.mutate({ start }, { onSettled: () => setOpen(false) });
  };
  const targetLabel = context.data?.targets.find(item => item.ref === target)?.label ?? target;
  const disabled = !!reason || busy || !!active;

  const action: DropdownMenuItem = { label: <span className="flex items-center gap-2"><UiGitMerge />Rebase…</span>, title: reason || undefined, disabled, onSelect: openDialog };

  const showStatus = (notice && !active) || error || active || ai.active || ai.state.status === 'failed';
  const status = <>
    {showStatus && <div className="flex min-w-0 shrink-0 flex-col gap-1.5 border-b border-border px-3 py-1.5 text-xs md:px-4">
      {notice && !active && <span role="status">{notice}</span>}
      {error && <RebaseFailure error={error} operation={operation} />}
      {active && <div className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1.5">
        <span className="min-w-0 break-words">{strategyNoun(operation.strategy, true)} of {operation.branch}: {operation.phase}{operation.stashSha && ` · saved edits: ${operation.stashSha.slice(0, 8)}`}</span>
        {!error && <ConflictPaths paths={operation.conflicts} />}
        <SplitButton variant="outline" size="sm" disabled={busy || !context.data?.executable} title="Rebase prompt options" label="Send to rebase prompt" onClick={() => send()}
          items={[{ label: <span className="flex items-center gap-2"><UiSparkles />Choose runtime…</span>, onSelect: () => setRuntimeOpen(true) }]} />
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
      <SplitButton label="Rebase" title="Rebase options" disabled={startDisabled} loading={mutation.isPending} onClick={() => begin({ target, strategy: 'rebase' })} items={[
        { label: <span className="flex items-center gap-2"><UiGitMerge />Merge {targetLabel} in</span>, onSelect: () => begin({ target, strategy: 'merge' }) },
        { label: <span className="flex items-center gap-2"><UiCog />Advanced…</span>, onSelect: () => setAdvancedOpen(true) },
      ]} />
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
    {advancedOpen && <RebaseAdvancedDialog
      branch={context.data?.branch ?? ''}
      targets={context.data?.targets ?? []}
      initialTarget={target}
      busy={startDisabled}
      onClose={() => setAdvancedOpen(false)}
      onStart={start => { setTarget(start.target); begin(start); }}
    />}
    {runtimeOpen && <RebaseRuntimeDialog busy={busy} onClose={() => setRuntimeOpen(false)} onSend={send} />}
  </>;

  return <>{children({ action, status })}</>;
}

// strategyNoun names an operation in status text; operations recorded before
// strategies existed are rebases.
function strategyNoun(strategy: RebaseStrategy | undefined, capitalized = false) {
  const noun = strategy === 'merge' ? 'merge' : 'rebase';
  return capitalized ? noun[0]!.toUpperCase() + noun.slice(1) : noun;
}

// RebaseAdvancedDialog chooses how the branch is brought up to date: rebase
// onto the target (optionally keeping merge commits) or merge the target in
// (optionally always recording a merge commit).
function RebaseAdvancedDialog({ branch, targets, initialTarget, busy, onClose, onStart }: {
  branch: string;
  targets: { ref: string; label: string }[];
  initialTarget: string;
  busy: boolean;
  onClose: () => void;
  onStart: (start: RebaseStart) => void;
}) {
  const [target, setTarget] = useState(initialTarget);
  const [strategy, setStrategy] = useState<RebaseStrategy>('rebase');
  const [rebaseMerges, setRebaseMerges] = useState(false);
  const [noFastForward, setNoFastForward] = useState(false);
  const start: RebaseStart = strategy === 'merge'
    ? { target, strategy, ...(noFastForward && { noFastForward }) }
    : { target, strategy, ...(rebaseMerges && { rebaseMerges }) };
  return (
    <Modal open onClose={onClose} title="Advanced update options" size="md" footer={<>
      <Button variant="ghost" onClick={onClose}>Cancel</Button>
      <Button disabled={busy || !target} onClick={() => onStart(start)}>{strategy === 'merge' ? 'Merge' : 'Rebase'}</Button>
    </>}>
      <div className="flex flex-col gap-3 text-sm">
        <label className="flex flex-col gap-1">Target
          <Combobox ariaLabel="Target" value={target} onChange={setTarget} options={targets.map(item => ({ value: item.ref, label: item.label }))} allowCustomValue={false} required />
        </label>
        <fieldset className="flex flex-col gap-1.5">
          <legend className="mb-1 font-medium">Strategy</legend>
          <label className="flex items-center gap-1.5"><input type="radio" name="rebase-strategy" checked={strategy === 'rebase'} onChange={() => setStrategy('rebase')} />Rebase {branch} onto the target</label>
          {strategy === 'rebase' && <label className="ml-5 flex items-center gap-1.5 text-muted-foreground"><input type="checkbox" checked={rebaseMerges} onChange={event => setRebaseMerges(event.currentTarget.checked)} />Keep merge commits (--rebase-merges)</label>}
          <label className="flex items-center gap-1.5"><input type="radio" name="rebase-strategy" checked={strategy === 'merge'} onChange={() => setStrategy('merge')} />Merge the target into {branch}</label>
          {strategy === 'merge' && <label className="ml-5 flex items-center gap-1.5 text-muted-foreground"><input type="checkbox" checked={noFastForward} onChange={event => setNoFastForward(event.currentTarget.checked)} />Always create a merge commit (--no-ff)</label>}
        </fieldset>
        <p className="text-xs text-muted-foreground">Uncommitted files are stashed first and restored afterwards, either way.</p>
      </div>
    </Modal>
  );
}

// Picks the Captain runtime (mode, model, effort…) layered over the branch-rebase
// prompt's frontmatter; unset fields inherit it.
function RebaseRuntimeDialog({ busy, onClose, onSend }: { busy: boolean; onClose: () => void; onSend: (runtime: AISpecRuntimeValue) => void }) {
  const [runtime, setRuntime] = useState<AISpecRuntimeValue>({});
  const models = useQuery(chatModelsQuery());
  const families = useQuery(chatRuntimesQuery());
  const catalogError = models.error ?? families.error;
  return (
    <Modal open onClose={onClose} title="Send to rebase prompt" size="md" footer={<>
      <Button variant="ghost" onClick={onClose}>Cancel</Button>
      <Button disabled={busy || !models.data || !families.data} onClick={() => onSend(runtime)}>Send to rebase prompt</Button>
    </>}>
      <div className="flex flex-col gap-2 text-sm">
        {catalogError ? <ActionError error={catalogError} title="Could not load agent runtimes" /> : models.data && families.data
          ? <SpecRuntimeBar ariaLabel="Rebase prompt runtime" models={models.data} families={families.data} value={runtime} onChange={setRuntime} readOnly={busy} />
          : <p role="status" className="text-muted-foreground">Loading agent runtimes…</p>}
        <p className="text-xs text-muted-foreground">Unset options inherit the branch-rebase prompt's runtime.</p>
      </div>
    </Modal>
  );
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
