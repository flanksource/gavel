import { SessionInspector, usePromptRunStream } from '@flanksource/clicky-ui/ai';
import type { LinterResult } from '@flanksource/clicky-ui/data';

/**
 * Follows the Captain prompt run behind a prompt lint result and shows its
 * transcript as it streams. The transcript is the run's own `entry` frames, so
 * it shows whether or not Captain records a session. The run's findings reach
 * the lint tree through the run snapshot.
 */
export default function PromptRunView({ result, baseUrl }: { result: LinterResult; baseUrl: string }) {
  const run = usePromptRunStream(result.run_id, { baseUrl });
  const lifecycleStatus = run.status === 'running' ? 'running' : run.status === 'succeeded' ? 'succeeded' : 'failed';
  return (
    <div className="flex h-[32rem] min-h-0 flex-col gap-2">
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
