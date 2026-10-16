import { describe, expect, it } from 'vitest';
import { buildLintTree, findLintNode, type LinterResult } from '@flanksource/clicky-ui/data';
import { buildLintRerunBody } from './lintRerun';

const lint: LinterResult[] = [
  {
    linter: 'eslint', work_dir: '/repo/web', success: false, duration: 1,
    violations: [
      { file: '/repo/web/src/a.ts', line: 3, message: 'no-var', rule: { method: 'no-var' } },
      { file: '/repo/web/src/b.ts', line: 9, message: 'eqeqeq', rule: { method: 'eqeqeq' } },
    ],
  },
  {
    linter: 'golangci-lint', work_dir: '/repo/api', success: false, duration: 1,
    violations: [{ file: '/repo/api/main.go', line: 1, message: 'unused', rule: { method: 'unused' } }],
  },
];

function nodeAt(path: string[]) {
  const tree = buildLintTree({ results: lint, grouping: 'linter-rule-file', filters: { severity: new Map(), linter: new Map() } });
  let routePath = '';
  let level = tree;
  for (const name of path) {
    const next = level.find(candidate => candidate.name === name || candidate.name.endsWith(name));
    if (!next) throw new Error(`no lint node ${name} under ${routePath || 'root'}; have ${level.map(n => n.name).join(', ')}`);
    routePath = next.route_path;
    level = next.children ?? [];
  }
  const node = findLintNode(tree, routePath);
  if (!node) throw new Error(`lint node ${routePath} not found`);
  return node;
}

describe('buildLintRerunBody', () => {
  it('reruns every linter when nothing is selected, but only when they share one work dir', () => {
    expect(buildLintRerunBody(null, lint)).toEqual({ lint: true, work_dir: '', lint_linters: ['eslint', 'golangci-lint'], lint_files: [] });
    expect(buildLintRerunBody(null, [lint[0]!])).toEqual({ lint: true, work_dir: '/repo/web', lint_linters: ['eslint'], lint_files: [] });
  });

  it('scopes a linter node to its linter and work dir', () => {
    expect(buildLintRerunBody(nodeAt(['eslint']), lint)).toEqual({ lint: true, work_dir: '/repo/web', lint_linters: ['eslint'], lint_files: [] });
  });

  it('scopes a rule node to the files that still have that rule\'s findings', () => {
    expect(buildLintRerunBody(nodeAt(['eslint', 'no-var']), lint)).toEqual({
      lint: true, work_dir: '/repo/web', lint_linters: ['eslint'], lint_files: ['src/a.ts'],
    });
  });
});
