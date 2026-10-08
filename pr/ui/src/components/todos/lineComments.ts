import type { TodoEvent } from '../../types';

export type DiffSide = 'old' | 'new';

/** Where a line comment was left: the file line of a diff range of a run branch. */
export interface LineCommentAnchor {
  path: string;
  side: DiffSide;
  line: number;
  lineText: string;
  /** The range base; empty/absent for a single-commit range. */
  base?: string;
  commit: string;
  branch?: string;
  attemptId?: string;
}

/** A todo comment anchored to a diff line, with its folded resolved state. */
export interface LineComment {
  event: TodoEvent;
  anchor: LineCommentAnchor;
  resolved: boolean;
}

/** The todo PATCH fields that create and resolve line comments. */
export interface LineCommentPatch {
  lineComment?: { anchor: LineCommentAnchor; body: string };
  resolveComment?: { id: string; resolved: boolean };
}

export const COMMENT_KIND = 'comment';
export const COMMENT_RESOLVED_KIND = 'comment_resolved';

function requireString(value: unknown, field: string, id: string): string {
  if (typeof value !== 'string') throw new Error(`Line comment ${id}: anchor.${field} must be a string, got ${JSON.stringify(value)}`);
  return value;
}

function optionalString(value: unknown, field: string, id: string): string | undefined {
  if (value === undefined || value === null) return undefined;
  return requireString(value, field, id);
}

function parseAnchor(raw: unknown, id: string): LineCommentAnchor {
  if (typeof raw !== 'object' || raw === null || Array.isArray(raw)) throw new Error(`Line comment ${id}: anchor must be an object`);
  const anchor = raw as Record<string, unknown>;
  const side = anchor.side;
  if (side !== 'old' && side !== 'new') throw new Error(`Line comment ${id}: anchor.side must be "old" or "new", got ${JSON.stringify(side)}`);
  if (typeof anchor.line !== 'number' || !Number.isInteger(anchor.line)) throw new Error(`Line comment ${id}: anchor.line must be an integer, got ${JSON.stringify(anchor.line)}`);
  return {
    path: requireString(anchor.path, 'path', id),
    side,
    line: anchor.line,
    lineText: requireString(anchor.lineText, 'lineText', id),
    base: optionalString(anchor.base, 'base', id),
    commit: requireString(anchor.commit, 'commit', id),
    branch: optionalString(anchor.branch, 'branch', id),
    attemptId: optionalString(anchor.attemptId, 'attemptId', id),
  };
}

function parseResolution(event: TodoEvent): { commentId: string; resolved: boolean } {
  const id = event.id ?? '(no id)';
  const { commentId, resolved } = event.payload ?? {};
  if (typeof commentId !== 'string' || !commentId) throw new Error(`Resolution event ${id}: payload.commentId must be a non-empty string`);
  if (typeof resolved !== 'boolean') throw new Error(`Resolution event ${id}: payload.resolved must be a boolean, got ${JSON.stringify(resolved)}`);
  return { commentId, resolved };
}

/**
 * The anchored comments of a todo's events, in event order, each with the
 * resolved state its latest `comment_resolved` event left it in. Mirrors the
 * server's fold: a comment is unresolved until an event resolves it, and a later
 * event wins, so a comment can be resolved and reopened repeatedly.
 */
export function lineComments(events: TodoEvent[]): LineComment[] {
  const comments = new Map<string, LineComment>();
  for (const event of events) {
    if (event.kind === COMMENT_KIND && event.payload?.anchor !== undefined) {
      if (!event.id) throw new Error('Line comment without an id cannot be resolved');
      comments.set(event.id, { event, anchor: parseAnchor(event.payload.anchor, event.id), resolved: false });
    } else if (event.kind === COMMENT_RESOLVED_KIND) {
      const { commentId, resolved } = parseResolution(event);
      const comment = comments.get(commentId);
      if (comment) comment.resolved = resolved;
    }
  }
  return [...comments.values()];
}

export function unresolvedLineCommentCount(events: TodoEvent[]): number {
  return lineComments(events).filter(comment => !comment.resolved).length;
}

/** Whether an anchor belongs to the diff of `path` over the range `base..commit`. */
export function matchesRange(anchor: LineCommentAnchor, range: { path: string; base: string; commit: string }): boolean {
  return anchor.path === range.path && (anchor.base ?? '') === range.base && anchor.commit === range.commit;
}
