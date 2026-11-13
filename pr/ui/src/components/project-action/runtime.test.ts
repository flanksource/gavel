import { describe, expect, it } from 'vitest';
import { commitRuntimeValue, withCommitRuntime, withoutRemovedAIOptions } from './runtime';

describe('commit runtime projection', () => {
  it('does not create inherited runtime values', () => {
    expect(commitRuntimeValue({ files: ['one.go'], 'add-metadata': false })).toEqual({});
  });

  it('projects authored compact mode and effort without changing saved options', () => {
    const options = { model: 'api:example-model:high', effort: 'low', message: 'fix: selected' };
    expect(commitRuntimeValue(options)).toEqual({ model: 'example-model', mode: 'api', effort: 'high' });
    expect(options).toEqual({ model: 'api:example-model:high', effort: 'low', message: 'fix: selected' });
  });

  it('replaces compact pins when the runtime is explicitly edited', () => {
    expect(withCommitRuntime({ model: 'api:example-model:high', message: '', 'add-metadata': false }, {
      model: 'example-model', mode: 'agent', effort: 'low', id: 'catalog-only', budget: { cost: 1 },
    })).toEqual({ model: 'example-model', mode: 'agent', effort: 'low', message: '', 'add-metadata': false });
  });

  it('removes cleared runtime keys without materializing defaults', () => {
    expect(withCommitRuntime({ model: 'example-model', mode: 'api', effort: 'high', files: ['one.go'] }, {}))
      .toEqual({ files: ['one.go'] });
  });

  it('removes only AI settings omitted from the form', () => {
    expect(withoutRemovedAIOptions({
      model: 'example-model', 'group-model': 'group', fallback: ['fallback'], temperature: '0', 'no-cache': false,
      'add-metadata': false, 'ai-fix-max-iterations': 3, message: '',
    })).toEqual({ model: 'example-model', 'add-metadata': false, 'ai-fix-max-iterations': 3, message: '' });
  });

  it('rejects invalid runtime field types rather than coercing them', () => {
    expect(() => commitRuntimeValue({ effort: false })).toThrow('effort');
  });
});
