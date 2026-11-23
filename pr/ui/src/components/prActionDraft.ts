import type { ModalDraft } from '@flanksource/clicky-ui/components';
import type { PRCommitInfo } from '../types';

export type MergeMethod = 'rebase' | 'squash' | 'merge';
export type PendingPRAction =
  | { type: 'merge'; method: MergeMethod; auto: boolean }
  | { type: 'approve' }
  | { type: 'update-branch' };

export const prActionFormKey = (type: PendingPRAction['type']) => `pr-action:${type}`;
export const prActionMutationKey = (repo: string, number: number) => ['prs', 'action', repo, number] as const;

export type PRActionAvailability = {
  disabled: boolean;
  changes: EventTarget;
  anchor: HTMLSpanElement | null;
};

// MergeMessageSource is what the default squash / merge commit message is
// written from. It is handed to the dialog alongside the draft, not saved with it.
export type MergeMessageSource = {
  number: number;
  title: string;
  body?: string;
  headRef: string;
  commits: PRCommitInfo[];
};

// defaultMergeMessage mirrors GitHub's defaults: a squash takes the PR title
// and the commit details, a merge commit names the PR and its head branch.
export function defaultMergeMessage(method: MergeMethod, source: MergeMessageSource) {
  switch (method) {
    case 'rebase':
      return { title: '', body: '' };
    case 'merge':
      return { title: `Merge pull request #${source.number} from ${source.headRef}`, body: source.title };
    case 'squash': {
      const messages = source.commits.map(commitMessage);
      const body = messages.length === 1
        ? messages[0]!.body
        : messages.map(m => [`* ${m.headline}`, m.body].filter(Boolean).join('\n\n')).join('\n\n');
      return { title: `${source.title} (#${source.number})`, body };
    }
  }
}

// commitMessage undoes GitHub's headline cut: a long first line is returned as
// "start…" in messageHeadline with "…rest" leading messageBody.
function commitMessage({ messageHeadline, messageBody = '' }: PRCommitInfo) {
  if (!messageHeadline.endsWith('…') || !messageBody.startsWith('…')) return { headline: messageHeadline, body: messageBody.trim() };
  const [rest = '', ...body] = messageBody.slice(1).split('\n');
  return { headline: messageHeadline.slice(0, -1) + rest, body: body.join('\n').trim() };
}

export function readPRActionDraft(saved: unknown) {
  if (!saved || typeof saved !== 'object') throw new Error('Invalid PR action draft');
  const { repo, number, url, base, nodeId, headSha, azure, pending, comment, commitTitle, commitBody } = saved as ModalDraft;
  if (typeof repo !== 'string' || typeof number !== 'number' || !Number.isSafeInteger(number) || number < 1 ||
      typeof url !== 'string' || typeof base !== 'string' || typeof nodeId !== 'string' || typeof headSha !== 'string' ||
      typeof azure !== 'boolean' || typeof comment !== 'string' || typeof commitTitle !== 'string' || typeof commitBody !== 'string' ||
      !pending || typeof pending !== 'object' || !('type' in pending)) {
    throw new Error('Invalid PR action draft');
  }
  const target = new URL(url);
  if (!['https:', 'http:'].includes(target.protocol) || target.username || target.password) throw new Error('Invalid PR URL');
  return { repo, number, url, base, nodeId, headSha, azure, pending: readPendingAction(pending), comment, commitTitle, commitBody };
}

function readPendingAction(saved: object & { type: unknown }): PendingPRAction {
  if (saved.type === 'approve' || saved.type === 'update-branch') return { type: saved.type };
  if (saved.type === 'merge' && 'method' in saved && 'auto' in saved &&
      (saved.method === 'rebase' || saved.method === 'squash' || saved.method === 'merge') && typeof saved.auto === 'boolean') {
    return { type: 'merge', method: saved.method, auto: saved.auto };
  }
  throw new Error('Invalid PR action');
}
