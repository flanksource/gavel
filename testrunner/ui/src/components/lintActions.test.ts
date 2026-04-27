import { describe, expect, it } from 'vitest';
import type { Test, LinterResult } from '../types';
import { lintFolderActions, lintFileActions, folderPattern } from './lintActions';

function folderNode(opts: Partial<Test> & { target_path?: string }): Test {
  return {
    name: opts.name || 'pkg',
    framework: 'lint',
    kind: 'lint-folder',
    file: opts.file ?? 'pkg',
    target_path: opts.target_path ?? 'pkg',
    work_dir: opts.work_dir,
    linterName: opts.linterName,
    ruleName: opts.ruleName,
    children: opts.children,
  };
}

function fileNode(opts: Partial<Test>): Test {
  return {
    name: opts.name || 'foo.go',
    framework: 'lint',
    kind: 'lint-file',
    file: opts.file ?? 'pkg/foo.go',
    target_path: opts.target_path ?? 'pkg/foo.go',
    work_dir: opts.work_dir,
    linterName: opts.linterName,
    children: opts.children,
    violations: opts.violations,
  };
}

describe('folderPattern', () => {
  it('returns ** for empty path', () => {
    expect(folderPattern(undefined)).toBe('**');
    expect(folderPattern('')).toBe('**');
  });
  it('appends /** to a path and strips trailing slash', () => {
    expect(folderPattern('pkg')).toBe('pkg/**');
    expect(folderPattern('pkg/')).toBe('pkg/**');
    expect(folderPattern('pkg/sub')).toBe('pkg/sub/**');
  });
});

describe('lintFolderActions', () => {
  it('rule-scoped folder: 3 buttons (folder + repo + disable-linter)', () => {
    const node = folderNode({
      target_path: 'pkg/sub',
      linterName: 'golangci-lint',
      ruleName: 'errcheck',
      work_dir: '/repo',
    });
    const actions = lintFolderActions(node, []);
    expect(actions.map(a => a.label)).toEqual([
      'Ignore errcheck in this folder',
      'Ignore rule errcheck everywhere',
      'Disable golangci-lint entirely',
    ]);
    expect(actions[0].req).toEqual({
      source: 'golangci-lint',
      rule: 'errcheck',
      file: 'pkg/sub/**',
      work_dir: '/repo',
    });
    expect(actions[1].req).toEqual({
      source: 'golangci-lint',
      rule: 'errcheck',
      work_dir: '/repo',
    });
    expect(actions[2].req).toEqual({
      source: 'golangci-lint',
      work_dir: '/repo',
    });
    expect(actions[2].variant).toBe('subtle');
  });

  it('single-linter folder with no direct violations: shows linter-scoped 3-button matrix', () => {
    const node = folderNode({
      target_path: 'pkg/empty',
      linterName: 'eslint',
      work_dir: '/repo',
      // children empty — no descendant violations
    });
    const actions = lintFolderActions(node, []);
    expect(actions.map(a => a.label)).toEqual([
      'Ignore everything in this folder',
      'Ignore eslint in this folder',
      'Disable eslint entirely',
    ]);
    // The disable-entirely button should not require work_dir to render.
    expect(actions[2].disabledWithoutWorkDir).toBeFalsy();
  });

  it('multi-linter folder: per-linter pair plus everything-in-folder', () => {
    const lint: LinterResult[] = [
      {
        linter: 'golangci-lint',
        work_dir: '/repo',
        success: true,
        duration: 0,
        violations: [{ file: '/repo/pkg/sub/a.go', line: 1, message: 'x', rule: { method: 'errcheck' } }],
      },
      {
        linter: 'betterleaks',
        work_dir: '/repo',
        success: true,
        duration: 0,
        violations: [{ file: '/repo/pkg/sub/b.go', line: 2, message: 'y', rule: { method: 'gen-key' } }],
      },
    ];
    const node = folderNode({ target_path: 'pkg/sub' });
    const actions = lintFolderActions(node, lint);
    // First action is always "everything in folder", then per-linter pairs.
    expect(actions[0].label).toBe('Ignore everything in this folder');
    const labels = actions.map(a => a.label);
    expect(labels).toContain('Ignore golangci-lint in this folder');
    expect(labels).toContain('Disable golangci-lint entirely');
    expect(labels).toContain('Ignore betterleaks in this folder');
    expect(labels).toContain('Disable betterleaks entirely');
    // 1 base + 2 linters * 2 buttons = 5
    expect(actions).toHaveLength(5);
    // disable-entirely should not be disabled without workDir
    const disableGo = actions.find(a => a.label === 'Disable golangci-lint entirely');
    expect(disableGo?.disabledWithoutWorkDir).toBeFalsy();
  });

  it('multi-linter folder with no descendant violations: only the everything-in-folder button', () => {
    const node = folderNode({ target_path: 'pkg/orphan' });
    const actions = lintFolderActions(node, []);
    expect(actions).toHaveLength(1);
    expect(actions[0].label).toBe('Ignore everything in this folder');
  });
});

describe('lintFileActions', () => {
  it('single-linter file: ignore-in-file + disable-entirely', () => {
    const node = fileNode({
      file: 'pkg/foo.go',
      target_path: 'pkg/foo.go',
      linterName: 'golangci-lint',
      work_dir: '/repo',
    });
    const actions = lintFileActions(node);
    expect(actions.map(a => a.label)).toEqual([
      'Ignore all golangci-lint in this file',
      'Disable golangci-lint entirely',
    ]);
    expect(actions[0].req).toEqual({
      source: 'golangci-lint',
      file: 'pkg/foo.go',
      work_dir: '/repo',
    });
  });

  it('multi-linter file parent: everything-in-file + per-linter pair', () => {
    // Multi-linter file built by buildFileTree has linter-kind children.
    const node = fileNode({
      file: 'pkg/foo.go',
      target_path: 'pkg/foo.go',
      work_dir: '/repo',
      children: [
        {
          name: 'golangci-lint',
          framework: 'lint',
          kind: 'linter',
          linterName: 'golangci-lint',
        },
        {
          name: 'eslint',
          framework: 'lint',
          kind: 'linter',
          linterName: 'eslint',
        },
      ],
    });
    const actions = lintFileActions(node);
    expect(actions[0].label).toBe('Ignore everything in this file');
    expect(actions[0].req).toEqual({ file: 'pkg/foo.go', work_dir: '/repo' });
    const labels = actions.map(a => a.label);
    expect(labels).toContain('Ignore all golangci-lint in this file');
    expect(labels).toContain('Disable golangci-lint entirely');
    expect(labels).toContain('Ignore all eslint in this file');
    expect(labels).toContain('Disable eslint entirely');
    expect(actions).toHaveLength(5);
  });
});
