import { SessionInspector, usePromptRunStream, type PromptRunStreamState } from '@flanksource/clicky-ui/ai';
import { CopyButton } from '@flanksource/clicky-ui/components';
import { Badge, type LinterResult } from '@flanksource/clicky-ui/data';

const STATUS_TONE = { running: 'info', succeeded: 'success', failed: 'danger', idle: 'neutral' } as const;

function CopyableId({ label, value }: { label: string; value: string }) {
  return (
    <span className="inline-flex items-center gap-1">
      <span className="text-muted-foreground">{label}</span>
      <code className="select-all font-mono">{value}</code>
      <CopyButton value={value} label={`Copy ${label.toLowerCase()} id`} />
    </span>
  );
}

function RunHeader({ run, result }: { run: PromptRunStreamState; result: LinterResult }) {
  const model = run.run?.model ?? run.summary?.model;
  const provider = run.run?.provider ?? run.summary?.provider;
  const mode = run.run?.mode ?? run.summary?.mode;
  const engine = [model, provider, mode].filter(Boolean).join(' · ');
  const sessionId = run.sessionId ?? result.session_id;
  return (
    <div role="group" aria-label="Prompt run" className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs">
      <Badge tone={STATUS_TONE[run.status]} variant="soft" size="sm">{run.status}</Badge>
      {engine && <span className="text-muted-foreground">{engine}</span>}
      {result.run_id && <CopyableId label="Run" value={result.run_id} />}
      {sessionId && <CopyableId label="Session" value={sessionId} />}
      {run.summary?.duration && <span className="text-muted-foreground">Elapsed {run.summary.duration}</span>}
    </div>
  );
}

/**
 * Follows the Captain prompt run behind a prompt lint result and shows its
 * transcript as it streams, under a header with the run's status, model,
 * ids and, once it finishes, its duration. The transcript is the run's own
 * `entry` frames, so it shows whether or not Captain records a session. The
 * run's findings reach the lint tree through the run snapshot.
 */
export default function PromptRunView({ result, baseUrl }: { result: LinterResult; baseUrl: string }) {
  const run = usePromptRunStream(result.run_id, { baseUrl });
  const lifecycleStatus = run.status === 'running' ? 'running' : run.status === 'succeeded' ? 'succeeded' : 'failed';
  return (
    <div className="flex h-[32rem] min-h-0 flex-col gap-2">
      <RunHeader run={run} result={result} />
      {run.status === 'failed' && <p role="alert" className="text-sm text-red-600">{run.error}</p>}
      {run.entries.length === 0 ? (
        <p role="status" className="text-sm text-muted-foreground">
          {run.status === 'running' ? 'Waiting for the prompt run to start…' : 'The prompt run produced no transcript.'}
        </p>
      ) : (
        <SessionInspector
          session={{ ...(run.sessionId ? { id: run.sessionId } : {}), lifecycleStatus, messages: run.entries }}
          className="min-h-0 flex-1 text-xs"
          layout="compact"
          follow={false}
        />
      )}
    </div>
  );
}
