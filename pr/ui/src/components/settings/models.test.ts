import { describe, expect, it } from 'vitest';
import { promptModelCatalog } from './models';
import type { RunContext } from '../todos/providers';

describe('promptModelCatalog', () => {
  it('forwards Captain model rows without translating runtime axes', () => {
    const models = [
      {
        id: 'claude-sonnet',
        provider: 'anthropic',
        label: 'Sonnet',
        reasoning: true,
        configured: true,
        modes: ['agent', 'cli', 'cmux'],
        runtime: { model: 'claude-sonnet' },
      },
    ];
    const context: RunContext = {
      efforts: [],
      tools: [],
      modes: [],
      runtimes: [],
      lifecycle: { steps: [] },
      models,
    };

    expect(promptModelCatalog(context)).toBe(models);
  });
});
