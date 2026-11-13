import type { Violation } from '@flanksource/clicky-ui/data';
import type { CommentVocabularyItem } from '@flanksource/clicky-ui/comments';
import type { CommentElementContext } from './comments-model';
export const COMMENT_RATINGS = ["positive", "negative"] as const;
export type CommentRating = (typeof COMMENT_RATINGS)[number];
export const REVIEW_DECISIONS = [
  "approved",
  "rejected",
  "revision_requested",
] as const;
export type ReviewDecision = (typeof REVIEW_DECISIONS)[number];

/**
 * Playground API projection of feedback. Deliberately structural rather than an
 * import of the library's `Comment` type: this module is loaded by Vite's config
 * bundler, where the `@flanksource/clicky-ui` source aliases do not apply.
 */
export type StoredAuthor = {
  id?: string;
  name: string;
  model?: string;
  effort?: string;
  kind?: "user" | "agent";
  /** The agent that posted on `name`'s behalf, when another agent wrote the text. */
  via?: string;
  /** The session that produced the text, so the run can be inspected later. */
  sessionId?: string;
  /** AI response cost in USD for producing the text. */
  cost?: number;
};

export const AGENT_RUN_STATES = ["queued", "sent", "done", "error"] as const;
export type AgentRunState = (typeof AGENT_RUN_STATES)[number];

/**
 * Hand-off of a root to the playground's shared agent session. Several roots
 * share one `sessionId`; the agent itself replies and resolves each one, which
 * moves its state to `done`.
 */
export type StoredAgentRun = {
  /** Absent when the hand-off failed before a session existed. */
  sessionId?: string;
  requestedAt: string;
  state: AgentRunState;
  /** Why the hand-off failed; present only when `state` is `error`. */
  error?: string;
  /** The live captain session it went to, when not the playground's own; `channel` once delivered. */
  target?: { label: string; channel?: string };
  /** When the agent first replied in the thread after the hand-off. */
  repliedAt?: string;
};

/** One finding of a structured comment, rendered as a chip and toggled by status. */
export type StoredChecklistItem = { label: string; status: string };

/** A labelled group of references (files, selectors) rendered as chips. */
export type StoredRefGroup = { label: string; items: string[]; mono?: boolean };

/** The captain prompt run a root was produced by. */
export type StoredPromptRun = {
  promptId: string;
  promptName: string;
  runId: string;
  sessionId?: string;
  variables?: Record<string, unknown>;
};

/** One attribute a visual style edit wanted to set; `from` is null when the element did not carry it. */
export type StoredStyleChange = {
  attribute: string;
  from: string | number | null;
  to: string | number | null;
};

/**
 * A visual edit Design mode could not write back (its value is computed in
 * code, spread, ...), asked of an agent instead. `marker` is the element's
 * `data-editable-source`; the refusal is the visual-edits server's 422.
 */
export type StoredStyleEdit = {
  marker: string;
  tag: string;
  changes: StoredStyleChange[];
  refusal: {
    reason: string;
    file: string;
    line: number;
    column: number;
    message: string;
  };
};

export type StoredComment = {
  id: string;
  version?: number;
  rootId?: string;
  body: string;
  createdAt: string;
  updatedAt?: string;
  author: StoredAuthor | null;
  status?: string;
  parentId?: string | null;
  anchor?: string | null;
  rating?: CommentRating;
  reviewDecision?: ReviewDecision;
  closedAt?: string;
  closedBy?: StoredAuthor;
  /** Who holds the thread while it is in progress; kept on record once resolved. */
  assignee?: StoredAuthor;
  /** When `acknowledgedBy` picked the thread up. Absent while merely assigned. */
  acknowledgedAt?: string;
  acknowledgedBy?: StoredAuthor;
  agentRun?: StoredAgentRun;
  element?: CommentElementContext;
  checklist?: StoredChecklistItem[];
  vocabularyRefs?: CommentVocabularyItem[];
  refs?: StoredRefGroup[];
  promptRun?: StoredPromptRun;
  /** The lint findings a prompt run reported, kept verbatim. */
  lint?: { violations: Violation[] };
  /** The refused visual style edit this comment asks an agent to make. */
  styleEdit?: StoredStyleEdit;
};
