import { useEffect, useRef, useState } from 'react';
import { openEventStream } from '../eventHub';
import { mutationJSON } from '../query';

export type PromptRunState =
  | { status: 'idle' }
  | { status: 'starting' }
  | { status: 'running'; runId: string; sessionId?: string }
  | { status: 'failed'; runId?: string; sessionId?: string; error: string };

interface PromptRunStarted { runId: string }
interface PromptRunFrame { sessionId?: string }
interface PromptRunSummary { sessionId?: string; success: boolean; error?: string }

function parseFrame<T>(event: Event): T {
  return JSON.parse((event as MessageEvent<string>).data) as T;
}

/**
 * Runs a runtime-discovered library prompt through Captain
 * (`POST /api/v1/prompt/run?args=<prompt>`) and follows it on
 * `/api/captain/prompt/runs/{runId}/stream`: the `run` frame names the session
 * to tail, `done`/`error` end it. `label` names the run in error messages;
 * onSuccess fires once a run succeeds and onSettled after any terminal frame.
 */
export function usePromptRun({ prompt, label, onSuccess, onSettled }: {
  prompt: string;
  label: string;
  onSuccess?: () => void;
  onSettled?: () => void;
}) {
  const [state, setState] = useState<PromptRunState>({ status: 'idle' });
  // Callers pass fresh callbacks per render; the stream must not reopen for them.
  const callbacks = useRef({ onSuccess, onSettled, label });
  callbacks.current = { onSuccess, onSettled, label };
  const runId = state.status === 'running' ? state.runId : '';

  useEffect(() => {
    if (!runId) return;
    const settle = () => callbacks.current.onSettled?.();
    const stream = openEventStream(`/api/captain/prompt/runs/${encodeURIComponent(runId)}/stream`);
    stream.addEventListener('run', event => {
      const { sessionId } = parseFrame<PromptRunFrame>(event);
      if (sessionId) setState(current => current.status === 'running' ? { ...current, sessionId } : current);
    });
    stream.addEventListener('done', event => {
      stream.close();
      const summary = parseFrame<PromptRunSummary>(event);
      if (!summary.success) {
        setState({ status: 'failed', runId, sessionId: summary.sessionId, error: summary.error || `The ${callbacks.current.label} did not succeed.` });
      } else {
        setState({ status: 'idle' });
        callbacks.current.onSuccess?.();
      }
      settle();
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
          setState({ status: 'failed', runId, error: `Captain no longer knows this ${callbacks.current.label}; check its result before retrying.` });
          settle();
        });
        return;
      }
      stream.close();
      const summary = JSON.parse(data) as PromptRunSummary;
      setState({ status: 'failed', runId, sessionId: summary.sessionId, error: summary.error || `The ${callbacks.current.label} failed.` });
      settle();
    });
    return () => stream.close();
  }, [runId]);

  const start = async (variables: Record<string, unknown>) => {
    setState({ status: 'starting' });
    try {
      const started = await mutationJSON<PromptRunStarted>({
        url: `/api/v1/prompt/run?args=${encodeURIComponent(prompt)}`,
        method: 'POST',
        body: { variables },
        context: `Could not start the ${label}`,
      });
      if (!started.runId) throw new Error(`Could not start the ${label}: no run id in the response`);
      setState({ status: 'running', runId: started.runId });
    } catch (error) {
      setState({ status: 'failed', error: error instanceof Error ? error.message : String(error) });
    }
  };

  // The run's terminal frame reports the stop; only a refused request lands here.
  const stop = () => mutationJSON<unknown>({
    url: `/api/captain/prompt/runs/${encodeURIComponent(runId)}/stop`,
    method: 'POST',
    context: `Could not stop the ${label}`,
  }).catch((error: unknown) => {
    setState({ status: 'failed', runId, error: error instanceof Error ? error.message : String(error) });
  });

  const active = state.status === 'starting' || state.status === 'running';
  return { state, active, start, stop };
}
