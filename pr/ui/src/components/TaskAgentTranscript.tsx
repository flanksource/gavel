import { SessionViewer, type SessionEntry, type SessionUIMessage } from '@flanksource/clicky-ui/ai';
import type { TaskExtraTabs, TaskProcessDetails, TaskSnapshot } from '@flanksource/clicky-ui/data';
import { useTodoSession } from './todos/TodoSession';

// The annotation captain puts the agent's own session id under, and the labels
// gavel contributes from the todo the run belongs to.
const SESSION_LABEL = 'session';
const RUN_SESSION_LABEL = 'runSession';
const DIR_LABEL = 'dir';

function isProcessDetails(details: TaskSnapshot['details']): details is TaskProcessDetails {
  return !!details && typeof details === 'object' && 'metrics' in details;
}

/**
 * TaskAgentTranscript renders what an agent actually did. It exists because the
 * stdout of an agent process is its JSON-RPC transport — tool-call frames and
 * file contents — so the raw stream answers "is it alive" but never "what is it
 * working on". The transcript answers the second question.
 */
function TaskAgentTranscript({ sessionId, dir }: { sessionId: string; dir: string }) {
  const { entries, connected, error } = useTodoSession(dir, sessionId, true);

  if (error) {
    return <div role="alert" className="p-2 text-xs text-red-300">{error}</div>;
  }
  if (entries.length === 0) {
    return (
      <div className="p-2 text-xs text-gray-400">
        {connected ? 'Waiting for the first entry…' : 'Connecting to the session stream…'}
      </div>
    );
  }
  return (
    <div className="max-h-64 overflow-auto bg-background p-2">
      {/* The header carries the model and context meter, which is exactly the
          "what is this agent running on" the task row was missing. */}
      <SessionViewer session={entries as SessionEntry[] | SessionUIMessage[]} className="text-xs" />
    </div>
  );
}

/**
 * agentTranscriptTabs contributes the Transcript tab to supervised agent runs.
 * The identity it needs lives on the group snapshot — clicky only populates
 * labels there — and a run whose session has not been established yet simply
 * contributes nothing, leaving the row with its own streams.
 */
export const agentTranscriptTabs: TaskExtraTabs = (task: TaskSnapshot, group: TaskSnapshot) => {
  if (group.kind !== 'agent') return [];
  const labels = group.labels ?? {};
  // The provider's own session id only exists once its handshake lands, so it
  // arrives as a live annotation; the run session the host allocated up front
  // is on the labels and stands in until then.
  const annotations = isProcessDetails(task.details) ? (task.details.annotations ?? {}) : {};
  const sessionId = annotations[SESSION_LABEL] || labels[RUN_SESSION_LABEL];
  if (!sessionId) return [];
  const dir = labels[DIR_LABEL] ?? '';
  return [
    {
      id: 'transcript',
      label: 'Transcript',
      render: () => <TaskAgentTranscript sessionId={sessionId} dir={dir} />,
    },
  ];
};
