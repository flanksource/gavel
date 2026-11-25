import { describe, expect, it } from 'vitest';
import { defaultMergeMessage, readPRActionDraft, type MergeMessageSource } from './prActionDraft';

const commit = (messageHeadline: string, messageBody = '') => ({
  oid: `${messageHeadline}-sha`, messageHeadline, messageBody, committedDate: '2026-11-20T10:00:00Z', additions: 1, deletions: 0, changedFiles: 1,
});

const source: MergeMessageSource = {
  number: 27,
  title: 'Add widgets',
  body: 'Widgets for the dashboard',
  headRef: 'acme/widgets-branch',
  commits: [commit('feat: add widget', 'Renders a widget.'), commit('test: cover widget')],
};

describe('defaultMergeMessage', () => {
  it('squash uses the PR title with its number and lists every commit like GitHub', () => {
    expect(defaultMergeMessage('squash', source)).toEqual({
      title: 'Add widgets (#27)',
      body: '* feat: add widget\n\nRenders a widget.\n\n* test: cover widget',
    });
  });

  it('squash of a single commit keeps that commit body', () => {
    expect(defaultMergeMessage('squash', { ...source, commits: [commit('feat: add widget', 'Renders a widget.')] }))
      .toEqual({ title: 'Add widgets (#27)', body: 'Renders a widget.' });
  });

  it('rejoins a headline GitHub cut short with an ellipsis', () => {
    const cut = commit('fix(cli): use XDG cache d…', '…irectory\n\nRuns build:runtime.');
    expect(defaultMergeMessage('squash', { ...source, commits: [cut, commit('test: cover widget')] }).body)
      .toBe('* fix(cli): use XDG cache directory\n\nRuns build:runtime.\n\n* test: cover widget');
  });

  it('merge commit names the PR and head branch, with the PR title as body', () => {
    expect(defaultMergeMessage('merge', source)).toEqual({
      title: 'Merge pull request #27 from acme/widgets-branch',
      body: 'Add widgets',
    });
  });

  it('rebase writes no commit message', () => {
    expect(defaultMergeMessage('rebase', source)).toEqual({ title: '', body: '' });
  });
});

describe('readPRActionDraft', () => {
  const saved = {
    repo: 'acme/service', number: 27, url: 'https://github.com/acme/service/pull/27', base: 'main', nodeId: 'PR_27',
    headSha: 'abc', azure: false, comment: '', pending: { type: 'merge', method: 'squash', auto: false },
    commitTitle: 'Add widgets (#27)', commitBody: '* feat: add widget',
  };

  it('keeps the edited commit message', () => {
    expect(readPRActionDraft(saved)).toMatchObject({ commitTitle: saved.commitTitle, commitBody: saved.commitBody });
  });

  it('rejects a draft whose commit message is not text', () => {
    expect(() => readPRActionDraft({ ...saved, commitTitle: 7 })).toThrow('Invalid PR action draft');
  });
});
