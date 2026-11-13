/*
 * Gavel's agent delivery contract (`POST /api/v1/comments/{id}/deliver`,
 * `/agent-sessions`, and the settings-based comment agents at
 * `/api/v1/comments/agent`). Browser safe: the playground server and the
 * Chrome extension's gavel client share these shapes.
 */

/**
 * The Captain prompt a gavel-managed session runs. Gavel's library has no
 * generic comment prompt; `editor-comments` is the closest (it works comment
 * threads in a project directory with `list_threads`, `reply_to_thread` and
 * `resolve_thread`), and its file-specific wording is not load-bearing.
 */
export const DEFAULT_RUN_PROMPT = "editor-comments";

/** A run target's `runOptions`, in Captain's run-request shape. */
export type RunOptions = {
  presets?: string[];
  spec?: Record<string, unknown>;
  variables?: Record<string, unknown>;
  /** Appended to the thread the new session is given. */
  instructions?: string;
};

export type DeliverTarget =
  | { kind: "session"; sessionId: string }
  | {
      kind: "run";
      prompt: string;
      runOptions?: RunOptions;
      groupKey?: string;
      /** Skips the group's recorded session, launches a new one and records it for the group. */
      fresh?: boolean;
    };

/**
 * A comment agent as gavel's settings layers hold it (`comments.agents` in
 * `.gavel.yaml` or a database layer, gavel/verify/config_comments.go): the run
 * new human threads on a page start. Every field is optional so a layer can
 * override one and inherit the rest. It never names a live session: which
 * session a page's threads go to is the browser's choice (localStorage).
 */
export type CommentAgent = {
  /** Whether new human threads are delivered automatically; unset is off. */
  enabled?: boolean;
  assign?: boolean;
  prompt?: string;
  presets?: string[];
  spec?: Record<string, unknown>;
  variables?: Record<string, unknown>;
  instructions?: string;
  /** The delivery group key: runs of one group reuse its recorded live session. */
  group?: string;
};

/** One settings layer: a position (`user`, `parent`, `project`, `path`) and its file ("" for the database layer). */
export type SettingsRef = { kind: "user" | "parent" | "project" | "path"; path?: string; file?: string };

/**
 * `GET`/`PUT /api/v1/comments/agent`: a target's settings layers and the
 * comment agent that applies to it, each field traced to the layer that set it.
 */
export type CommentAgentSettings = {
  settings: {
    positions: Array<{ kind: SettingsRef["kind"]; path?: string; dir: string; files: string[]; db: boolean }>;
    layers: Array<SettingsRef & { location: string; config: { comments?: { agents?: { default?: CommentAgent; pages?: Record<string, CommentAgent> } } } }>;
    provenance: Record<string, string>;
  };
  agent: { agent: CommentAgent; provenance: Record<string, { layer: string; page?: string }>; found: boolean };
};

/** `PUT /api/v1/comments/agent`: save `agent` at `layer` (its default, or with `page` its entry for the target); `null` removes it. */
export type SaveCommentAgentRequest = {
  target: { scopeUri: string; uri: string };
  layer: SettingsRef;
  page: boolean;
  agent: CommentAgent | null;
};

/**
 * A Claude or Codex session gavel can deliver a thread to: a live one, or one
 * gavel launched in this scope (`managed`) that has since ended — delivering to
 * it resumes that same session (`resumable`).
 */
export type GavelAgentSession = {
  id: string;
  source: string;
  title: string;
  model?: string;
  cwd?: string;
  /** Captain's live status (`active`, `sleeping`, …), or `ended` for a managed session that is no longer running. */
  status?: string;
  inCmux: boolean;
  /** Gavel launched this session for a thread in the scope. */
  managed?: boolean;
  /** Delivering resumes the ended session instead of failing. */
  resumable?: boolean;
  /** How much of the model's context window the session fills; absent when gavel has no reading. */
  context?: { usedTokens: number; windowTokens: number };
};

/**
 * `POST /api/v1/comments/deliver`: hands several open thread roots to one
 * target in a single message. Gavel answers the updated threads; a stale or
 * closed thread fails the whole batch.
 */
export type DeliverBatchRequest = {
  requestId: string;
  target: DeliverTarget;
  threads: Array<{ id: string; expectedVersion: number }>;
};
