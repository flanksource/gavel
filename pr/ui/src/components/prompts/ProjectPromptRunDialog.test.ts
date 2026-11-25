import { describe, expect, it } from 'vitest';
import { promptVariableDefaults } from './ProjectPromptRunDialog';

const launch = { project: 'acme', workDir: '/work/acme', path: 'src/app.ts' };

describe('promptVariableDefaults', () => {
  it('prefills path, directory and project variables and leaves the rest blank', () => {
    expect(promptVariableDefaults({ ...launch, variables: ['file', 'worktree', 'project', 'question'] })).toEqual({
      file: 'src/app.ts',
      worktree: '/work/acme',
      project: 'acme',
      question: '',
    });
  });

  it('leaves a path variable blank for a project-level run', () => {
    expect(promptVariableDefaults({ ...launch, path: '', variables: ['path'] })).toEqual({ path: '' });
  });
});
