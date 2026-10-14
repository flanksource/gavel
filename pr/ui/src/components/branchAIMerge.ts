import { useEffect, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { openEventStream } from '../eventHub';
import { mutationJSON, queryKeys } from '../query';
import type { BranchMergeMode, BranchMergeResult } from '../types';
import { projectDiffQueryKey } from './projectMutations';

/** The runtime-discovered prompt (prompts/library/branch-merge.prompt) that rebases, resolves and lands. */
export const BRANCH_MERGE_PROMPT = 'branch-merge';

export interface BranchAIMergeInput {
  /** The primary checkout the branch lands into. */
  repo: string;
  /** The linked worktree the branch is checked out in; the agent works there. */
  worktree: string;
  branch: string;
  base: string;
  mode: BranchMergeMode;
  /** Squash commit message; empty keeps git's squash message. */
  message: string;
  /** Paths the failed land conflicted on, listed for the agent. */
  conflicts: string[];
  /** Commits ahead of base when the run started, reported once it lands. */
  ahead: number;
}

export type BranchAIMergeState =
  | { status: 'idle' }
  | { status: 'starting' }
  | { status: 'running'; runId: string; sessionId?: string }
  | { status: 'failed'; runId?: string; sessionId?: string; error: string };

interface PromptRunStarted { runId: string }
interface PromptRunFrame { sessionId?: string }
interface PromptRunSummary { sessionId?: string; success: boolean; error?: string }

export function branchAIMergeVariables(input: BranchAIMergeInput) {
  return {
    repo: input.repo,
    worktree: input.worktree,
    branch: input.branch,
    base: input.base,
    mode: input.mode,
    message: input.message.trim(),
    conflicts: input.conflicts,
  };
}

function parseFrame<T>(event: Event): T {
  return JSON.parse((event as MessageEvent<string>).data) as T;
}

/**
 * Runs the branch-merge prompt through Captain (`POST /api/v1/prompt/run`) and
 * follows it on `/api/captain/prompt/runs/{runId}/stream`: the `run` frame
 * names the session to tail, `done`/`error` end it. The prompt's own post-verify
 * steps land the branch and remove its worktree, so success reports a merge.
 */
export function useBranchAIMerge({ projectName, onMerged }: {
  projectName: string;
  onMerged?: (result: BranchMergeResult) => void;
}) {
  const client = useQueryClient();
  const [state, setState] = useState<BranchAIMergeState>({ status: 'idle' });
  const input = useRef<BranchAIMergeInput | null>(null);
  // Callers pass a fresh onMerged per render; the stream must not reopen for it.
  const onMergedRef = useRef(onMerged);
  onMergedRef.current = onMerged;
  const runId = state.status === 'running' ? state.runId : '';

  useEffect(() => {
    const landing = input.current;
    if (!runId || !landing) return;
    const refresh = () => Promise.all([
      client.invalidateQueries({ queryKey: queryKeys.projectGitSummary() }),
      client.invalidateQueries({ queryKey: queryKeys.projectGit(projectName) }),
      client.invalidateQueries({ queryKey: queryKeys.projectStatusScope(projectName) }),
      client.invalidateQueries({ queryKey: projectDiffQueryKey(projectName) }),
    ]);
    const stream = openEventStream(`/api/captain/prompt/runs/${encodeURIComponent(runId)}/stream`);
    stream.addEventListener('run', event => {
      const { sessionId } = parseFrame<PromptRunFrame>(event);
      if (sessionId) setState(current => current.status === 'running' ? { ...current, sessionId } : current);
    });
    stream.addEventListener('done', event => {
      stream.close();
      const summary = parseFrame<PromptRunSummary>(event);
      if (!summary.success) {
        setState({ status: 'failed', runId, sessionId: summary.sessionId, error: summary.error || 'The merge run did not succeed.' });
        void refresh();
        return;
      }
      setState({ status: 'idle' });
      // The prompt's after steps either all ran or failed the run, so a success
      // means the branch landed and its worktree and branch are gone.
      onMergedRef.current?.({
        targetBranch: landing.base, landedSha: '', mode: landing.mode, commits: landing.ahead,
        worktreeRemoved: true, branchDeleted: true,
      });
      void refresh();
    });
    stream.addEventListener('error', event => {
      // A dropped connection is a data-less error the stream reconnects from
      // (the server replays the run) — unless Captain has already forgotten the
      // run, which only its snapshot route can tell; only the server's terminal
      // `error` frame carries a summary.
      const data = (event as MessageEvent<string>).data;
      if (!data) {
        void fetch(`/api/captain/prompt/runs/${encodeURIComponent(runId)}`).then(response => {
          if (response.status !== 404) return;
          stream.close();
          setState({ status: 'failed', runId, error: 'Captain no longer knows this merge run; check the branch and the project before retrying.' });
          void refresh();
        });
        return;
      }
      stream.close();
      const summary = JSON.parse(data) as PromptRunSummary;
      setState({ status: 'failed', runId, sessionId: summary.sessionId, error: summary.error || 'The merge run failed.' });
      void refresh();
    });
    return () => stream.close();
  }, [runId, projectName, client]);

  const start = async (next: BranchAIMergeInput) => {
    input.current = next;
    setState({ status: 'starting' });
    try {
      const started = await mutationJSON<PromptRunStarted>({
        url: `/api/v1/prompt/run?args=${encodeURIComponent(BRANCH_MERGE_PROMPT)}`,
        method: 'POST',
        body: { variables: branchAIMergeVariables(next) },
        context: `Could not start the AI merge of ${next.branch}`,
      });
      if (!started.runId) throw new Error(`Could not start the AI merge of ${next.branch}: no run id in the response`);
      setState({ status: 'running', runId: started.runId });
    } catch (error) {
      setState({ status: 'failed', error: error instanceof Error ? error.message : String(error) });
    }
  };

  // The run's terminal frame reports the stop; only a refused request lands here.
  const stop = () => mutationJSON<unknown>({
    url: `/api/captain/prompt/runs/${encodeURIComponent(runId)}/stop`,
    method: 'POST',
    context: 'Could not stop the AI merge',
  }).catch((error: unknown) => {
    setState({ status: 'failed', runId, error: error instanceof Error ? error.message : String(error) });
  });

  return { state, start, stop };
}
