import type { Test } from '@flanksource/clicky-ui/data';
import { describe, expect, it } from 'vitest';
import { buildRunFailureCandidates, defaultRunTodoTitle, type RunFailureCandidate } from './RunFailureCandidates';
import type { RunSnapshot } from './types';

describe('buildRunFailureCandidates', () => {
  it('accepts lint-only snapshots whose Go tests slice is null', () => {
    const snapshot = {
      tests: null,
      lint: [{ linter: 'eslint', violations: [{ file: 'src/app.ts', message: 'Invalid import.' }] }],
    } satisfies Pick<RunSnapshot, 'tests' | 'lint'>;

    expect(buildRunFailureCandidates(snapshot)).toHaveLength(1);
  });

  it('selects failed leaves, aggregate failures without failed children, lint violations, and linter errors', () => {
    const tests: Test[] = [
      {
        name: 'package tests',
        package_path: 'github.com/acme/gavel/pkg',
        failed: true,
        children: [
          {
            name: 'rejects invalid input',
            suite: ['service'],
            package_path: 'github.com/acme/gavel/pkg',
            file: 'pkg/service_test.go',
            line: 27,
            failed: true,
            message: 'expected status 400',
          },
          { name: 'accepts valid input', passed: true },
        ],
      },
      {
        name: 'package setup',
        package_path: 'github.com/acme/gavel/pkg',
        failed: true,
        children: [{ name: 'discovery', passed: true }],
      },
      { name: 'slow test', warned: true },
    ];
    const snapshot = {
      tests,
      lint: [
        {
          linter: 'eslint',
          violations: [{
            file: 'src/app.ts',
            line: 12,
            column: 7,
            message: 'Unexpected any.',
            rule: { package: 'typescript-eslint', method: 'no-explicit-any' },
          }],
        },
        { linter: 'golangci-lint', work_dir: './backend', violations: null, error: 'configuration is invalid' },
        { linter: 'disabled', skipped: true, violations: null, error: 'must not be selected' },
      ],
    } satisfies Pick<RunSnapshot, 'tests' | 'lint'>;

    expect(buildRunFailureCandidates(snapshot)).toEqual([
      {
        key: 'test:0.0',
        group: 'tests',
        title: 'github.com/acme/gavel/pkg › service › rejects invalid input',
        location: 'pkg/service_test.go:27',
        criterion: 'Test `github.com/acme/gavel/pkg › service › rejects invalid input` passes at `pkg/service_test.go:27`.',
        detail: 'expected status 400',
      },
      {
        key: 'test:1',
        group: 'tests',
        title: 'github.com/acme/gavel/pkg › package setup',
        criterion: 'Test `github.com/acme/gavel/pkg › package setup` passes.',
      },
      {
        key: 'lint:0:0',
        group: 'lint',
        title: 'typescript-eslint:no-explicit-any',
        location: 'src/app.ts:12:7',
        criterion: 'Resolve `typescript-eslint:no-explicit-any` at `src/app.ts:12:7`: Unexpected any.',
        detail: 'Unexpected any.',
      },
      {
        key: 'linter:1',
        group: 'lint',
        title: 'golangci-lint failed',
        location: './backend',
        criterion: 'Linter `golangci-lint` completes successfully in `./backend`.',
        detail: 'configuration is invalid',
      },
    ]);
  });
});

describe('defaultRunTodoTitle', () => {
  const candidate = (group: RunFailureCandidate['group']): RunFailureCandidate => ({
    key: group,
    group,
    title: group,
    criterion: group,
  });

  it.each([
    [[candidate('tests')], 'Fix test failures in gavel'],
    [[candidate('lint')], 'Fix lint failures in gavel'],
    [[candidate('tests'), candidate('lint')], 'Fix test and lint failures in gavel'],
  ] as const)('derives a grouped title from the selected failure kinds', (candidates, expected) => {
    expect(defaultRunTodoTitle('gavel', [...candidates])).toBe(expected);
  });
});
