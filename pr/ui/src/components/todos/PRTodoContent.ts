import { stripAnsi } from '../../ansi';
import type { PRDetail, PRItem, Violation } from '../../types';
import { extractCommentTitle, isDeploymentComment } from '../../utils';

export type PRTodoSourceGroup = 'tests' | 'lint' | 'checks' | 'comments';

export interface PRTodoCandidateDetail {
  heading: string;
  headingMarkdown?: boolean;
  location?: string;
  meta?: string;
  message?: string;
  markdown?: string;
  log?: string;
  url?: string;
}

export interface PRTodoCandidate {
  key: string;
  group: PRTodoSourceGroup;
  text: string;
  primary: string;
  secondary?: string;
  actionPattern?: string;
  commentId?: number;
  author?: string;
  avatarUrl?: string;
  detail: PRTodoCandidateDetail;
}

export interface PRTodoVerificationPayload {
  prNumber: number;
  repo?: string;
  commentIds?: number[];
  actions?: string[];
}

function sourceLocation(file?: string, line?: number): string {
  if (!file) return '';
  return line ? `${file}:${line}` : file;
}

// violationRule mirrors gavel's own precedence: the linter's own code when it
// has one, otherwise the matched rule.
function violationRule(violation: Violation): string {
  return violation.code || violation.rule?.method || '';
}

export function buildPRTodoCandidates(pr: PRItem, detail: PRDetail | null): PRTodoCandidate[] {
  const candidates: PRTodoCandidate[] = [];
  const shards = detail?.gavelResults ?? [];

  shards.forEach((shard, shardIndex) => {
    (shard.failures ?? []).forEach((failure, failureIndex) => {
      const suite = failure.suite?.join(' › ');
      const name = suite ? `${suite} › ${failure.name}` : failure.name;
      const location = sourceLocation(failure.file, failure.line);
      candidates.push({
        key: `test:${shardIndex}:${failureIndex}:${failure.name}`,
        group: 'tests',
        text: `Test \`${name}\` passes`,
        primary: failure.name,
        secondary: location || suite,
        detail: {
          heading: name,
          location,
          message: failure.message,
          log: failure.stderr || failure.stdout,
        },
      });
    });
  });

  shards.forEach((shard, shardIndex) => {
    (shard.lint ?? []).forEach(result => {
      (result.violations ?? []).forEach((violation, violationIndex) => {
        const location = sourceLocation(violation.file, violation.line);
        const ruleName = violationRule(violation);
        const rule = ruleName ? ` (${ruleName})` : '';
        candidates.push({
          key: `lint:${shardIndex}:${result.linter}:${violationIndex}:${violation.file ?? ''}:${violation.line ?? ''}:${ruleName}`,
          group: 'lint',
          text: `Resolve ${result.linter}${rule} violation${location ? ` at ${location}` : ''}`,
          primary: `${result.linter}${rule}`,
          secondary: location || violation.message,
          detail: {
            heading: `${result.linter}${rule}`,
            location,
            message: violation.message,
          },
        });
      });
    });
  });

  (pr.checkStatus?.failures ?? []).forEach((check, checkIndex) => {
    const steps = (check.failedSteps ?? []).join(', ');
    candidates.push({
      key: `check:${checkIndex}:${check.name}`,
      group: 'checks',
      text: `CI check "${check.name}" passes`,
      primary: check.name,
      secondary: steps || undefined,
      actionPattern: check.name,
      detail: {
        heading: check.name,
        meta: steps ? `Failed steps: ${steps}` : undefined,
        log: check.logTail,
        url: check.detailsUrl,
      },
    });
  });

  (detail?.comments ?? [])
    .filter(comment => !isDeploymentComment(comment) && !comment.isResolved && !comment.isOutdated)
    .forEach((comment, commentIndex) => {
      const title = extractCommentTitle(comment.body);
      const location = sourceLocation(comment.path, comment.line);
      candidates.push({
        key: `comment:${comment.id}:${commentIndex}`,
        group: 'comments',
        text: `Address @${comment.author}'s comment${location ? ` on ${location}` : ''}: ${title}`,
        primary: title || `Comment by @${comment.author}`,
        secondary: location || undefined,
        commentId: comment.id,
        author: comment.author,
        avatarUrl: comment.avatarUrl,
        detail: {
          heading: title || `Comment by @${comment.author}`,
          headingMarkdown: true,
          meta: `@${comment.author}`,
          location,
          markdown: comment.body,
          url: comment.url,
        },
      });
    });

  return candidates;
}

export function isPRTodoCriterion(candidate: PRTodoCandidate): boolean {
  return candidate.group === 'tests' || candidate.group === 'lint';
}

function uniqueValues<T>(values: T[]): T[] {
  return Array.from(new Set(values));
}

export function buildPRTodoVerification(
  pr: PRItem,
  candidates: PRTodoCandidate[],
  selected: PRTodoCandidate[],
): PRTodoVerificationPayload | undefined {
  const commentIds = uniqueValues(
    selected
      .map(candidate => candidate.commentId)
      .filter((id): id is number => typeof id === 'number' && id > 0),
  );
  const allActionCandidates = candidates.filter(candidate => candidate.actionPattern);
  const selectedActionCandidates = selected.filter(candidate => candidate.actionPattern);
  let actions = uniqueValues(
    selectedActionCandidates
      .map(candidate => candidate.actionPattern)
      .filter((action): action is string => Boolean(action)),
  );

  if (selectedActionCandidates.length > 0 && selectedActionCandidates.length === allActionCandidates.length) {
    const selectedActionKeys = new Set(selectedActionCandidates.map(candidate => candidate.key));
    if (allActionCandidates.every(candidate => selectedActionKeys.has(candidate.key))) actions = ['*'];
  }

  if (commentIds.length === 0 && actions.length === 0) return undefined;
  return {
    prNumber: pr.number,
    repo: pr.repo,
    ...(commentIds.length > 0 ? { commentIds } : {}),
    ...(actions.length > 0 ? { actions } : {}),
  };
}

function longestBacktickRun(value: string): number {
  return Math.max(0, ...Array.from(value.matchAll(/`+/g), match => match[0].length));
}

function inlineCode(value: string): string {
  const clean = stripAnsi(value).replace(/\r?\n/g, ' ');
  const delimiter = '`'.repeat(Math.max(1, longestBacktickRun(clean) + 1));
  return `${delimiter} ${clean} ${delimiter}`;
}

function quotedPreformatted(value: string): string {
  const clean = stripAnsi(value).replace(/\r\n?/g, '\n');
  const fence = '`'.repeat(Math.max(3, longestBacktickRun(clean) + 1));
  return [`> ${fence}`, ...clean.split('\n').map(line => `> ${line}`), `> ${fence}`].join('\n');
}

function formatCandidate(candidate: PRTodoCandidate): string {
  const detail = candidate.detail;
  const parts = [`#### ${stripAnsi(detail.heading).replace(/\r?\n/g, ' ')}`];
  if (detail.location) parts.push(`**Location:** ${inlineCode(detail.location)}`);
  if (detail.meta) parts.push(`**Details:** ${stripAnsi(detail.meta).replace(/\r?\n/g, ' ')}`);
  if (detail.message) parts.push(`**Message**\n\n${quotedPreformatted(detail.message)}`);
  if (detail.log) parts.push(`**Log**\n\n${quotedPreformatted(detail.log)}`);
  return parts.join('\n\n');
}

export function buildPRTodoBody(
  pr: PRItem,
  notes: string,
  selected: PRTodoCandidate[],
): string {
  const parts = [
    notes.trim(),
    `_From [${pr.repo}#${pr.number}](${pr.url})._`,
  ].filter(Boolean);
  const failures = selected.filter(isPRTodoCriterion);
  if (failures.length === 0) return parts.join('\n\n');

  const grouped = new Map<'tests' | 'lint', PRTodoCandidate[]>();
  for (const candidate of failures) {
    const group = candidate.group as 'tests' | 'lint';
    const groupCandidates = grouped.get(group);
    if (groupCandidates) groupCandidates.push(candidate);
    else grouped.set(group, [candidate]);
  }

  const details = ['## Failure details'];
  for (const [group, candidates] of grouped) {
    details.push(
      `### ${group === 'tests' ? 'Failing tests' : 'Lint violations'}`,
      candidates.map(formatCandidate).join('\n\n'),
    );
  }
  parts.push(details.join('\n\n'));
  return parts.join('\n\n');
}
