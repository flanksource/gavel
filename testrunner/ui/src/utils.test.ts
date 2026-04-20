import { describe, it, expect } from 'vitest';
import { collapseLintSingleChildChains, collapseSingleChildChains, formatCount, groupLintByLinterRuleFile } from './utils';
import type { LinterResult, Test } from './types';

describe('formatCount', () => {
  const cases: Array<[number, string]> = [
    [0, '0'],
    [1, '1'],
    [42, '42'],
    [999, '999'],
    [1000, '1k'],
    [1500, '1.5k'],
    [1550, '1.6k'],
    [9999, '10k'],
    [10_000, '10k'],
    [12_345, '12k'],
    [99_999, '100k'],
    [100_000, '100k'],
    [999_999, '1000k'],
    [1_000_000, '1M'],
    [1_234_567, '1.2M'],
    [12_345_678, '12M'],
  ];

  for (const [n, want] of cases) {
    it(`formats ${n} as ${want}`, () => {
      expect(formatCount(n)).toBe(want);
    });
  }
});

describe('collapseSingleChildChains', () => {
  it('compresses a three-deep single-child chain into one node', () => {
    const tree: Test[] = [{
      name: 'pkg',
      children: [{
        name: 'sub',
        children: [{
          name: 'Describe',
          children: [{ name: 'It', passed: true }],
        }],
      }],
    }];

    const [merged] = collapseSingleChildChains(tree);
    expect(merged.name).toBe('pkg > sub > Describe > It');
    expect(merged.passed).toBe(true);
    expect(merged.children ?? []).toHaveLength(0);
  });

  it('stops merging at the branch point when a node has siblings', () => {
    const tree: Test[] = [{
      name: 'pkg',
      children: [{
        name: 'Describe',
        children: [
          { name: 'ItA', passed: true },
          { name: 'ItB', failed: true },
        ],
      }],
    }];

    const [merged] = collapseSingleChildChains(tree);
    expect(merged.name).toBe('pkg > Describe');
    expect(merged.children).toHaveLength(2);
    expect(merged.children![0].name).toBe('ItA');
    expect(merged.children![1].name).toBe('ItB');
  });

  it('does not swallow a leaf that already carries a status into its parent', () => {
    // Here the "parent" is a passed leaf with one failed child — the
    // parent itself is a test result, not a container, so we must not
    // merge it with the child.
    const tree: Test[] = [{
      name: 'ParentTest',
      passed: true,
      children: [{ name: 'HelperTest', failed: true }],
    }];

    const [out] = collapseSingleChildChains(tree);
    expect(out.name).toBe('ParentTest');
    expect(out.children).toHaveLength(1);
    expect(out.children![0].name).toBe('HelperTest');
  });

  it('preserves the deepest node kind, file, and target_path', () => {
    const tree: Test[] = [{
      name: 'linter',
      kind: 'linter',
      children: [{
        name: 'src',
        kind: 'lint-folder',
        children: [{
          name: 'foo.ts',
          kind: 'lint-file',
          file: 'src/foo.ts',
          target_path: '/repo/src/foo.ts',
          violations: [{ severity: 'error', message: 'boom' }],
        }],
      }],
    }];

    const [merged] = collapseSingleChildChains(tree);
    expect(merged.name).toBe('linter > src > foo.ts');
    expect(merged.kind).toBe('lint-file');
    expect(merged.file).toBe('src/foo.ts');
    expect(merged.target_path).toBe('/repo/src/foo.ts');
    expect(merged.violations).toHaveLength(1);
  });

  it('keeps multi-root trees independent', () => {
    const tree: Test[] = [
      { name: 'alpha', children: [{ name: 'one', passed: true }] },
      { name: 'beta', children: [{ name: 'two', failed: true }] },
    ];

    const out = collapseSingleChildChains(tree);
    expect(out).toHaveLength(2);
    expect(out[0].name).toBe('alpha > one');
    expect(out[1].name).toBe('beta > two');
  });

  it('recurses into deeper subtrees after a branch', () => {
    // pkg has two children. The first is itself a single-leaf chain that
    // should collapse; the second has a sibling and must not.
    const tree: Test[] = [{
      name: 'pkg',
      children: [
        {
          name: 'groupA',
          children: [{ name: 'Describe', children: [{ name: 'It', passed: true }] }],
        },
        {
          name: 'groupB',
          children: [
            { name: 'ItX', passed: true },
            { name: 'ItY', failed: true },
          ],
        },
      ],
    }];

    const [pkg] = collapseSingleChildChains(tree);
    expect(pkg.name).toBe('pkg');
    expect(pkg.children).toHaveLength(2);
    expect(pkg.children![0].name).toBe('groupA > Describe > It');
    expect(pkg.children![1].name).toBe('groupB');
    expect(pkg.children![1].children).toHaveLength(2);
  });

  it('preserves linter and rule grouping while still collapsing folder chains', () => {
    const tree: Test[] = [{
      name: 'golangci-lint',
      kind: 'linter',
      children: [{
        name: 'errcheck',
        kind: 'lint-rule-group',
        linterName: 'golangci-lint',
        ruleName: 'errcheck',
        children: [{
          name: 'src',
          kind: 'lint-folder',
          children: [{
            name: 'foo.go',
            kind: 'lint-file',
            file: 'src/foo.go',
            target_path: 'src/foo.go',
            linterName: 'golangci-lint',
            ruleName: 'errcheck',
            violations: [{ severity: 'error', message: 'boom' }],
          }],
        }],
      }],
    }];

    const [linter] = collapseLintSingleChildChains(tree);
    expect(linter.name).toBe('golangci-lint');
    expect(linter.kind).toBe('linter');
    expect(linter.children).toHaveLength(1);
    expect(linter.children![0].name).toBe('errcheck');
    expect(linter.children![0].kind).toBe('lint-rule-group');
    expect(linter.children![0].children).toHaveLength(1);
    expect(linter.children![0].children![0].name).toBe('src > foo.go');
    expect(linter.children![0].children![0].kind).toBe('lint-file');
    expect(linter.children![0].children![0].ruleName).toBe('errcheck');
  });
});

describe('groupLintByLinterRuleFile', () => {
  const filters = { severity: new Map(), linter: new Map() };

  it('groups results by linter, then rule, then folder/file', () => {
    const lint: LinterResult[] = [{
      linter: 'golangci-lint',
      success: false,
      duration: 0,
      violations: [
        {
          file: 'pkg/foo.go',
          line: 4,
          severity: 'error',
          rule: { method: 'errcheck' },
          message: 'check error',
        },
        {
          file: 'pkg/foo.go',
          line: 8,
          severity: 'error',
          rule: { method: 'errcheck' },
          message: 'check error again',
        },
        {
          file: 'pkg/sub/bar.go',
          line: 3,
          severity: 'warning',
          rule: { method: 'unused' },
          message: 'unused value',
        },
      ],
    }];

    const [linter] = groupLintByLinterRuleFile(lint, filters);
    expect(linter.kind).toBe('linter');
    expect(linter.children).toHaveLength(2);

    const errcheck = linter.children!.find(child => child.ruleName === 'errcheck');
    expect(errcheck?.kind).toBe('lint-rule-group');
    expect(errcheck?.children).toHaveLength(1);
    expect(errcheck?.children![0].name).toBe('pkg');
    expect(errcheck?.children![0].kind).toBe('lint-folder');
    expect(errcheck?.children![0].children).toHaveLength(1);
    expect(errcheck?.children![0].children![0].name).toBe('foo.go');
    expect(errcheck?.children![0].children![0].kind).toBe('lint-file');
    expect(errcheck?.children![0].children![0].linterName).toBe('golangci-lint');
    expect(errcheck?.children![0].children![0].ruleName).toBe('errcheck');
    expect(errcheck?.children![0].children![0].violations).toHaveLength(2);

    const unused = linter.children!.find(child => child.ruleName === 'unused');
    expect(unused?.children).toHaveLength(1);
    expect(unused?.children![0].name).toBe('pkg');
  });

  it('keeps fileless violations on the rule group', () => {
    const lint: LinterResult[] = [{
      linter: 'betterleaks',
      success: false,
      duration: 0,
      violations: [
        {
          line: 1,
          severity: 'error',
          rule: { method: 'global-rule' },
          message: 'secret found',
        },
      ],
    }];

    const [linter] = groupLintByLinterRuleFile(lint, filters);
    expect(linter.children).toHaveLength(1);
    expect(linter.children![0].kind).toBe('lint-rule-group');
    expect(linter.children![0].ruleName).toBe('global-rule');
    expect(linter.children![0].children ?? []).toHaveLength(0);
    expect(linter.children![0].noFileViolations).toHaveLength(1);
  });
});
