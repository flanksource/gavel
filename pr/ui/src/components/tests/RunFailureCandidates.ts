import type { Test } from '@flanksource/clicky-ui/data';
import type { RunSnapshot } from './types';

export type RunFailureGroup = 'tests' | 'lint';

export interface RunFailureCandidate {
  key: string;
  group: RunFailureGroup;
  title: string;
  location?: string;
  criterion: string;
  detail?: string;
}

export function buildRunFailureCandidates(snapshot: Pick<RunSnapshot, 'tests' | 'lint'>): RunFailureCandidate[] {
  const candidates = (snapshot.tests ?? []).flatMap((test, index) => failedTestCandidates(test, [index]));
  for (const [resultIndex, result] of (snapshot.lint ?? []).entries()) {
    if (result.skipped) continue;
    for (const [violationIndex, violation] of (result.violations ?? []).entries()) {
      const location = sourceLocation(violation.file, violation.line, violation.column);
      const title = violation.code
        || (violation.rule?.package && violation.rule.method ? `${violation.rule.package}:${violation.rule.method}` : '')
        || violation.rule?.method
        || violation.rule?.pattern
        || violation.source
        || result.linter;
      candidates.push({
        key: `lint:${resultIndex}:${violationIndex}`,
        group: 'lint',
        title,
        ...(location ? { location } : {}),
        criterion: `Resolve \`${title}\`${location ? ` at \`${location}\`` : ''}${violation.message ? `: ${violation.message}` : '.'}`,
        ...(violation.message ? { detail: violation.message } : {}),
      });
    }
    if (result.error) {
      const location = result.work_dir;
      candidates.push({
        key: `linter:${resultIndex}`,
        group: 'lint',
        title: `${result.linter} failed`,
        ...(location ? { location } : {}),
        criterion: `Linter \`${result.linter}\` completes successfully${location ? ` in \`${location}\`` : ''}.`,
        detail: result.error,
      });
    }
  }
  return candidates;
}

export function defaultRunTodoTitle(projectName: string, candidates: RunFailureCandidate[]): string {
  const groups = new Set(candidates.map(candidate => candidate.group));
  if (groups.has('tests') && groups.has('lint')) return `Fix test and lint failures in ${projectName}`;
  if (groups.has('tests')) return `Fix test failures in ${projectName}`;
  return `Fix lint failures in ${projectName}`;
}

function failedTestCandidates(test: Test, path: number[]): RunFailureCandidate[] {
  const descendants = (test.children ?? []).flatMap((child, index) => failedTestCandidates(child, [...path, index]));
  if (descendants.length > 0 || !test.failed) return descendants;
  const title = [test.package_path || test.package, ...(test.suite ?? []), test.name].filter(Boolean).join(' › ');
  const location = sourceLocation(test.file, test.line);
  return [{
    key: `test:${path.join('.')}`,
    group: 'tests',
    title,
    ...(location ? { location } : {}),
    criterion: `Test \`${title}\` passes${location ? ` at \`${location}\`` : ''}.`,
    ...(test.message ? { detail: test.message } : {}),
  }];
}

function sourceLocation(file?: string, line?: number, column?: number): string {
  if (!file) return '';
  return [file, line, column].filter(value => value !== undefined && value !== 0).join(':');
}
