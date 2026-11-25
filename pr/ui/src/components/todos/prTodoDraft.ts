import type { PRItem, TodoPriority, TodoStatus } from '../../types';
import { priorities, statuses } from './format';
import type { PRTodoCandidate } from './PRTodoContent';

export const newPRTodoFormKey = 'pr-todo:create';
export const prTodoRunRecoveryFormKey = 'pr-todo:run-recovery';

export function prTodoHref(ref: string) {
  return `/todos/${ref.split('/').map(encodeURIComponent).join('/')}`;
}

export type PRTodoDraft = {
  dir: string;
  title: string;
  priority: TodoPriority;
  status: TodoStatus;
  notes: string;
  selectedKeys: string[];
  candidates: PRTodoCandidate[];
  step: string;
};

export function savePRTodoDraft(pr: PRItem, draft: PRTodoDraft) {
  const { number, repo, url, title, author, source, target, state, isDraft, updatedAt } = pr;
  const selected = new Set(draft.selectedKeys);
  const candidates = draft.candidates.map(candidate => selected.has(candidate.key) ? candidate : {
    ...candidate,
    detail: { heading: candidate.detail.heading, location: candidate.detail.location, url: candidate.detail.url },
  });
  return { snapshot: JSON.stringify({
    pr: { number, repo, url, title, author, source, target, state, isDraft, updatedAt },
    draft: { ...draft, candidates },
  }) };
}

export function readPRTodoDraft(saved: unknown) {
  if (!saved || typeof saved !== 'object' || !('snapshot' in saved) || typeof saved.snapshot !== 'string') {
    throw new Error('Invalid saved PR todo');
  }
  const { pr, draft } = JSON.parse(saved.snapshot) as { pr: PRItem; draft: PRTodoDraft };
  if (!pr || !Number.isSafeInteger(pr.number) || pr.number < 1 ||
      ![pr.repo, pr.url, pr.title, pr.author, pr.source, pr.target, pr.state, pr.updatedAt].every(value => typeof value === 'string') ||
      typeof pr.isDraft !== 'boolean' || !draft ||
      ![draft.dir, draft.title, draft.notes, draft.step].every(value => typeof value === 'string') ||
      !priorities.includes(draft.priority) || !statuses.includes(draft.status) ||
      !Array.isArray(draft.selectedKeys) || !draft.selectedKeys.every(key => typeof key === 'string') ||
      !Array.isArray(draft.candidates) || !draft.candidates.every(validCandidate)) {
    throw new Error('Invalid saved PR todo');
  }
  const url = new URL(pr.url);
  if (!['http:', 'https:'].includes(url.protocol) || url.username || url.password) throw new Error('Invalid PR URL');
  if (draft.selectedKeys.some(key => !draft.candidates.some(candidate => candidate.key === key))) throw new Error('Invalid saved PR signal selection');
  return { pr, draft };
}

function validCandidate(candidate: PRTodoCandidate) {
  return !!candidate && ['tests', 'lint', 'checks', 'comments'].includes(candidate.group) &&
    [candidate.key, candidate.text, candidate.primary, candidate.detail?.heading].every(value => typeof value === 'string') &&
    [candidate.secondary, candidate.actionPattern, candidate.author, candidate.avatarUrl, candidate.detail.location,
      candidate.detail.meta, candidate.detail.message, candidate.detail.markdown, candidate.detail.log, candidate.detail.url]
      .every(value => value === undefined || typeof value === 'string') &&
    (candidate.commentId === undefined || Number.isSafeInteger(candidate.commentId));
}
