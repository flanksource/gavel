import { describe, expect, it } from 'vitest';
import type { PRDetail, PRItem } from '../../types';
import {
  buildPRTodoBody,
  buildPRTodoCandidates,
  buildPRTodoVerification,
  isPRTodoCriterion,
  type PRTodoCandidate,
} from './PRTodoContent';

const pr: PRItem = {
  number: 17,
  title: 'Keep PR failure context',
  author: 'octocat',
  repo: 'acme/widget',
  source: 'failure-details',
  target: 'main',
  state: 'OPEN',
  isDraft: false,
  url: 'https://github.com/acme/widget/pull/17',
  updatedAt: '2026-07-26T12:00:00Z',
  checkStatus: {
    passed: 0,
    failed: 1,
    running: 0,
    pending: 0,
    failures: [{
      name: 'unit-tests',
      failedSteps: ['Run tests'],
      logTail: 'check log must stay out of the body',
    }],
  },
};

const detail: PRDetail = {
  gavelResults: [
    {
      stickyId: 'linux',
      artifactId: 101,
      artifactUrl: 'https://example.test/artifacts/101',
      testsPassed: 10,
      testsFailed: 2,
      testsSkipped: 0,
      testsTotal: 12,
      lintViolations: 1,
      lintLinters: 1,
      hasBench: false,
      failures: [
        {
          suite: ['storage'],
          name: 'saves Unicode café',
          file: 'pkg/store/save_test.go',
          line: 41,
          failed: true,
          message: 'expected record to persist',
          stderr: '\u001b[31mstderr: persistence failed\u001b[0m\nstdout: retry exhausted',
        },
        {
          suite: ['storage'],
          name: 'deselected linux failure',
          file: 'pkg/store/delete_test.go',
          line: 52,
          failed: true,
          message: 'must not be copied',
          stderr: 'deselected test log',
        },
      ],
      lint: [{
        linter: 'golangci-lint',
        success: false,
        duration: 0,
        violations: [{
          rule: { method: 'errcheck' },
          file: 'pkg/store/save.go',
          line: 23,
          message: 'return value is not checked',
        }],
      }],
    },
    {
      stickyId: 'macos',
      artifactId: 102,
      artifactUrl: 'https://example.test/artifacts/102',
      testsPassed: 8,
      testsFailed: 2,
      testsSkipped: 0,
      testsTotal: 10,
      lintViolations: 1,
      lintLinters: 1,
      hasBench: false,
      failures: [
        {
          suite: ['storage'],
          name: 'saves Unicode café',
          file: 'pkg/store/save_test.go',
          line: 41,
          failed: true,
          message: 'macOS shard failed independently',
          stderr: 'distinct shard output',
        },
      ],
      lint: [{
        linter: 'staticcheck',
        success: false,
        duration: 0,
        violations: [{
          rule: { method: 'SA4006' },
          file: 'pkg/store/delete.go',
          line: 19,
          message: 'deselected lint violation',
        }],
      }],
    },
  ],
  comments: [{
    id: 901,
    body: 'review comment body must stay out of the failure details',
    author: 'reviewer',
    url: 'https://github.com/acme/widget/pull/17#discussion_r901',
    createdAt: '2026-07-26T12:30:00Z',
    path: 'pkg/store/save.go',
    line: 23,
  }],
};

describe('PRTodoContent', () => {
  it('builds stable candidates and retains repeated failures from distinct shards', () => {
    const candidates = buildPRTodoCandidates(pr, detail);

    expect(candidates.map(candidate => candidate.key)).toEqual([
      'test:0:0:saves Unicode café',
      'test:0:1:deselected linux failure',
      'test:1:0:saves Unicode café',
      'lint:0:golangci-lint:0:pkg/store/save.go:23:errcheck',
      'lint:1:staticcheck:0:pkg/store/delete.go:19:SA4006',
      'check:0:unit-tests',
      'comment:901:0',
    ]);
    expect(candidates.filter(candidate => candidate.primary === 'saves Unicode café')).toHaveLength(2);
  });

  it('copies every available selected test and lint detail into the body', () => {
    const candidates = buildPRTodoCandidates(pr, detail);
    const selectedKeys = new Set([
      'test:0:0:saves Unicode café',
      'test:1:0:saves Unicode café',
      'lint:0:golangci-lint:0:pkg/store/save.go:23:errcheck',
    ]);
    const body = buildPRTodoBody(pr, 'Keep the public API stable.', candidates.filter(candidate => selectedKeys.has(candidate.key)));

    expect(body).toContain('Keep the public API stable.');
    expect(body).toContain('_From [acme/widget#17](https://github.com/acme/widget/pull/17)._');
    expect(body).toContain('## Failure details');
    expect(body).toContain('### Failing tests');
    expect(body).toContain('### Lint violations');
    expect(body).toContain('storage › saves Unicode café');
    expect(body).toContain('pkg/store/save_test.go:41');
    expect(body).toContain('expected record to persist');
    expect(body).toContain('stderr: persistence failed');
    expect(body).toContain('stdout: retry exhausted');
    expect(body).toContain('macOS shard failed independently');
    expect(body).toContain('distinct shard output');
    expect(body).toContain('golangci-lint (errcheck)');
    expect(body).toContain('pkg/store/save.go:23');
    expect(body).toContain('return value is not checked');
    expect(body).not.toContain('\u001b[31m');
    expect(body).not.toContain('deselected linux failure');
    expect(body).not.toContain('deselected test log');
    expect(body).not.toContain('deselected lint violation');
  });

  it('keeps checks and comments out of failure details while preserving verification selectors', () => {
    const candidates = buildPRTodoCandidates(pr, detail);
    const selected = candidates.filter(candidate => candidate.group === 'checks' || candidate.group === 'comments');

    expect(buildPRTodoBody(pr, '', selected)).toBe(
      '_From [acme/widget#17](https://github.com/acme/widget/pull/17)._',
    );
    expect(buildPRTodoVerification(pr, candidates, selected)).toEqual({
      prNumber: 17,
      repo: 'acme/widget',
      commentIds: [901],
      actions: ['*'],
    });
  });

  it('omits failure details without selected tests or lint violations', () => {
    expect(buildPRTodoBody(pr, 'Notes only.', [])).toBe(
      'Notes only.\n\n_From [acme/widget#17](https://github.com/acme/widget/pull/17)._',
    );
  });

  it('quotes multiline Markdown-looking content with a longer fence', () => {
    const candidate: PRTodoCandidate = {
      key: 'test:markdown',
      group: 'tests',
      text: 'Test `Markdown output` passes',
      primary: 'Markdown output',
      detail: {
        heading: 'Markdown output',
        message: 'message line\n## Acceptance Criteria',
        log: 'before\n````\n## Verification\n````\nafter',
      },
    };
    const body = buildPRTodoBody(pr, '', [candidate]);

    expect(body).toContain('> ```');
    expect(body).toContain('> ## Acceptance Criteria');
    expect(body).toContain('> `````');
    expect(body).toContain('> ## Verification');
    expect(body.match(/^## Acceptance Criteria$/gm)).toBeNull();
    expect(body.match(/^## Verification$/gm)).toBeNull();
  });

  it('classifies only tests and lint as acceptance criteria', () => {
    const candidates = buildPRTodoCandidates(pr, detail);

    expect(candidates.filter(isPRTodoCriterion).map(candidate => candidate.group)).toEqual([
      'tests',
      'tests',
      'tests',
      'lint',
      'lint',
    ]);
  });
});
