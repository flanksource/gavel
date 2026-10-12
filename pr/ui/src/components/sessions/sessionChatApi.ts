import { SPEC_PERMISSION_MODES, type SessionChatCapabilities, type SessionChatQueuedMessage, type SpecPermissionMode } from '@flanksource/clicky-ui/ai';
import { fetchJSON, HttpError, mutationJSON } from '../../query';
import { captainSessionUrl } from '../todos/TodoSessionDetail';

const CHAT_CONTEXT = 'Load session chat';
const INTERRUPT_CONTEXT = 'Interrupt session';
const PERMISSION_MODE_CONTEXT = 'Change session permission mode';
const CHAT_STATUSES = ['starting', 'running', 'idle', 'interrupting', 'stopping'] as const;

export type SessionChatStatus = (typeof CHAT_STATUSES)[number];

/** Captain's SessionChat (pkg/cli/session_chat_http.go): the live chat run attached to a session. */
export interface SessionChat {
  runId: string;
  capabilities: SessionChatCapabilities & { setPermissionMode: boolean };
  status: SessionChatStatus;
  queued: SessionChatQueuedMessage[];
  permissionMode?: SpecPermissionMode;
  permissionModes: SpecPermissionMode[];
}

export const sessionChatUrl = (sessionId: string) => `${captainSessionUrl(sessionId)}/chat`;
const sessionInterruptUrl = (sessionId: string) => `${captainSessionUrl(sessionId)}/interrupt`;
const sessionPermissionModeUrl = (sessionId: string) => `${captainSessionUrl(sessionId)}/permission-mode`;

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function invalid(context: string, payload: unknown, expected: string): never {
  throw new Error(`${context}: invalid response ${JSON.stringify(payload)}, expected ${expected}`);
}

function parsePermissionMode(value: unknown, payload: unknown): SpecPermissionMode {
  if (typeof value !== 'string' || !(SPEC_PERMISSION_MODES as readonly string[]).includes(value)) {
    return invalid(CHAT_CONTEXT, payload, `permission mode one of ${SPEC_PERMISSION_MODES.join(', ')}`);
  }
  return value as SpecPermissionMode;
}

function parseSessionChat(payload: unknown): SessionChat {
  const expected = '{runId, capabilities, state: {status, queued?}, permissionMode?, permissionModes?}';
  if (!isRecord(payload) || typeof payload.runId !== 'string' || !payload.runId
    || !isRecord(payload.capabilities) || !isRecord(payload.state)) {
    return invalid(CHAT_CONTEXT, payload, expected);
  }
  const caps = payload.capabilities;
  const flags = ['interrupt', 'steer', 'followUp', 'resume', 'setPermissionMode'] as const;
  if (!flags.every(flag => typeof caps[flag] === 'boolean')) return invalid(CHAT_CONTEXT, payload, `${expected} with boolean capability flags`);
  const { status, queued } = payload.state;
  if (!(CHAT_STATUSES as readonly unknown[]).includes(status)) return invalid(CHAT_CONTEXT, payload, `state.status one of ${CHAT_STATUSES.join(', ')}`);
  if (queued !== undefined && (!Array.isArray(queued)
    || !queued.every(item => isRecord(item) && typeof item.messageId === 'string' && typeof item.text === 'string'))) {
    return invalid(CHAT_CONTEXT, payload, 'state.queued as [{messageId, text}]');
  }
  const modes = payload.permissionModes;
  if (modes !== undefined && !Array.isArray(modes)) return invalid(CHAT_CONTEXT, payload, 'permissionModes as an array');
  return {
    runId: payload.runId,
    capabilities: {
      interrupt: caps.interrupt as boolean,
      steer: caps.steer as boolean,
      followUp: caps.followUp as boolean,
      resume: caps.resume as boolean,
      setPermissionMode: caps.setPermissionMode as boolean,
    },
    status: status as SessionChatStatus,
    queued: (queued ?? []) as SessionChatQueuedMessage[],
    ...(payload.permissionMode === undefined ? {} : { permissionMode: parsePermissionMode(payload.permissionMode, payload) }),
    permissionModes: (modes ?? []).map(mode => parsePermissionMode(mode, payload)),
  };
}

/**
 * The live chat Captain holds for a session, or null when none is attached in
 * this process (404): sessions not started through gavel have none, so that is
 * an ordinary state and the caller falls back to liveness-derived controls.
 * Every other failure throws.
 */
export async function fetchSessionChat({ sessionId, signal }: { sessionId: string; signal: AbortSignal }): Promise<SessionChat | null> {
  try {
    return parseSessionChat(await fetchJSON<unknown>({ url: sessionChatUrl(sessionId), signal, context: CHAT_CONTEXT }));
  } catch (error) {
    if (error instanceof HttpError && error.status === 404) return null;
    throw error;
  }
}

export async function postSessionInterrupt(sessionId: string): Promise<void> {
  const payload = await mutationJSON<unknown>({ url: sessionInterruptUrl(sessionId), method: 'POST', context: INTERRUPT_CONTEXT });
  if (!isRecord(payload) || typeof payload.status !== 'string') invalid(INTERRUPT_CONTEXT, payload, '{status}');
}

export async function postSessionPermissionMode({ sessionId, mode }: { sessionId: string; mode: SpecPermissionMode }): Promise<void> {
  const payload = await mutationJSON<unknown>({
    url: sessionPermissionModeUrl(sessionId),
    method: 'POST',
    body: { mode },
    context: PERMISSION_MODE_CONTEXT,
  });
  if (!isRecord(payload) || typeof payload.runId !== 'string' || typeof payload.permissionMode !== 'string') {
    invalid(PERMISSION_MODE_CONTEXT, payload, '{runId, permissionMode}');
  }
}
