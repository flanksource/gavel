import { useEffect, useRef, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { SessionChatCapabilities, SpecPermissionMode } from '@flanksource/clicky-ui/ai';
import type { AgentSession } from '../../types';
import { mutationJSON, queryKeys } from '../../query';
import { captainSessionUrl } from '../todos/TodoSessionDetail';
import { fetchSessionChat, postSessionInterrupt, postSessionPermissionMode } from './sessionChatApi';

// While a run this view started is in flight the sessions list and the live
// chat are polled at this cadence, so the live indicator (and with it `follow`)
// and the chat's queue/status catch up within a few seconds instead of the 15s
// idle poll.
const RUN_POLL_MS = 3_000;
// A sent message must show its process live within this window, or the run is
// treated as never started and `follow` stops instead of staying on forever.
export const RUN_START_TIMEOUT_MS = 60_000;
const SEND_CONTEXT = 'Send session message';

/** Captain's ChatMessageResponse (pkg/cli/prompt_chat_types.go). */
export interface SessionMessageResponse {
  runId: string;
  messageId: string;
  status: string;
  capabilities: SessionChatCapabilities;
}

export function sessionMessageUrl(sessionId: string): string {
  return `${captainSessionUrl(sessionId)}/message`;
}

function parseSessionMessageResponse(payload: unknown): SessionMessageResponse {
  const response = payload as Partial<SessionMessageResponse> | null;
  const caps = response?.capabilities as Record<string, unknown> | undefined;
  const flagsValid = !!caps && (['interrupt', 'steer', 'followUp', 'resume'] as const).every(flag => typeof caps[flag] === 'boolean');
  if (typeof response?.runId !== 'string' || !response.runId
    || typeof response.messageId !== 'string'
    || typeof response.status !== 'string'
    || !flagsValid) {
    throw new Error(`${SEND_CONTEXT}: invalid response ${JSON.stringify(payload)}, expected {runId, messageId, status, capabilities}`);
  }
  return response as SessionMessageResponse;
}

/** Sends to the live session, or resumes a finished one; Captain decides which. */
export async function postSessionMessage({ sessionId, text, messageId }: {
  sessionId: string;
  text: string;
  messageId: string;
}): Promise<SessionMessageResponse> {
  const payload = await mutationJSON<unknown>({
    url: sessionMessageUrl(sessionId),
    method: 'POST',
    body: { text, messageId },
    context: SEND_CONTEXT,
  });
  return parseSessionMessageResponse(payload);
}

/**
 * Drives a SessionChatComposer for one dashboard session. Capabilities come, in
 * order, from the live chat Captain has attached to the session (which also
 * supplies the queue, interrupt and permission-mode controls), else the last
 * send's response, else the row's liveness (a live process takes follow-ups, a
 * finished one is resumed) — Captain arbitrates, and a rejection lands in
 * `error`. No attached chat (404) is the normal state of a session not started
 * through gavel; any other chat failure lands in `error`. After a send `follow`
 * stays on for the run it started: through the poll below it ends once the
 * process has been seen live and then gone, or fails into `error` when it never
 * shows up live within `startTimeoutMs`.
 */
export function useSessionMessageComposer(session: AgentSession, { startTimeoutMs = RUN_START_TIMEOUT_MS }: {
  /** How long a sent message may wait for its process to appear live. */
  startTimeoutMs?: number;
} = {}) {
  const queryClient = useQueryClient();
  const [capabilities, setCapabilities] = useState<SessionChatCapabilities>();
  const [following, setFollowing] = useState(false);
  const [actionError, setActionError] = useState<string>();
  const sawLive = useRef(false);
  const { processActive } = session;
  const live = processActive || following;
  const chatKey = queryKeys.agentSessionChat(session.id);
  const refreshChat = () => void queryClient.invalidateQueries({ queryKey: chatKey });
  const action = { onMutate: () => setActionError(undefined), onError: (error: Error) => setActionError(error.message) };

  const chatQuery = useQuery({
    queryKey: chatKey,
    queryFn: ({ signal }) => fetchSessionChat({ sessionId: session.id, signal }),
    retry: false,
    refetchInterval: live ? RUN_POLL_MS : false,
  });
  const chat = chatQuery.data ?? undefined;

  const send = useMutation({
    mutationFn: (text: string) => postSessionMessage({ sessionId: session.id, text, messageId: crypto.randomUUID() }),
    ...action,
    onSuccess: response => {
      sawLive.current = processActive;
      setCapabilities(response.capabilities);
      setFollowing(true);
      void queryClient.invalidateQueries({ queryKey: queryKeys.agentSessions() });
      refreshChat();
    },
  });
  const interrupt = useMutation({ mutationFn: () => postSessionInterrupt(session.id), ...action, onSuccess: refreshChat });
  const permissionMode = useMutation({
    mutationFn: (mode: SpecPermissionMode) => postSessionPermissionMode({ sessionId: session.id, mode }),
    ...action,
    onSuccess: refreshChat,
  });

  useEffect(() => {
    if (!following) return;
    if (processActive) {
      sawLive.current = true;
      return;
    }
    if (sawLive.current) {
      sawLive.current = false;
      setFollowing(false);
      setCapabilities(undefined);
    }
  }, [following, processActive]);

  useEffect(() => {
    if (!following || processActive || sawLive.current) return;
    const timer = setTimeout(() => {
      setFollowing(false);
      setCapabilities(undefined);
      setActionError(`Resumed run did not start within ${startTimeoutMs / 1000}s`);
    }, startTimeoutMs);
    return () => clearTimeout(timer);
  }, [following, processActive, startTimeoutMs]);

  useEffect(() => {
    if (!following) return;
    const timer = setInterval(() => { void queryClient.invalidateQueries({ queryKey: queryKeys.agentSessions() }); }, RUN_POLL_MS);
    return () => clearInterval(timer);
  }, [following, queryClient]);

  const canSetPermissionMode = !!chat?.capabilities.setPermissionMode && !!chat.permissionMode && chat.permissionModes.length > 0;
  return {
    status: send.isPending ? ('starting' as const) : chat ? chat.status : processActive ? ('running' as const) : ('idle' as const),
    capabilities: chat?.capabilities ?? capabilities ?? {
      interrupt: false,
      steer: processActive,
      followUp: processActive,
      resume: !processActive,
    },
    queued: chat?.queued ?? [],
    error: actionError ?? chatQuery.error?.message,
    follow: live,
    send: (text: string) => send.mutate(text),
    onInterrupt: chat?.capabilities.interrupt ? () => interrupt.mutate() : undefined,
    permissionMode: canSetPermissionMode ? chat.permissionMode : undefined,
    permissionModes: canSetPermissionMode ? chat.permissionModes : undefined,
    permissionFamily: session.source,
    onPermissionModeChange: canSetPermissionMode ? (mode: SpecPermissionMode) => permissionMode.mutate(mode) : undefined,
  };
}
