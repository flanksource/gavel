import { describe, expect, it } from 'vitest';
import { promptModelCatalog } from './SettingsDialog';
import type { RunContext } from './todos/providers';

describe('promptModelCatalog', () => {
  it('preserves backend membership when de-duplicating shared models', () => {
    const context: RunContext = {
      efforts: [],
      tools: [],
      backends: [
        {
          id: 'claude-agent',
          label: 'Claude Agent',
          provider: 'anthropic',
          agent: 'claude',
          defaultModel: 'claude-sonnet',
          driver: 'claude-headless',
          mechanisms: [],
          models: [
            {
              id: 'claude-sonnet',
              provider: 'anthropic',
              label: 'Sonnet',
              reasoning: true,
            },
          ],
        },
        {
          id: 'claude-cmux',
          label: 'Claude cmux',
          provider: 'anthropic',
          agent: 'claude',
          defaultModel: 'claude-sonnet',
          driver: 'claude-cmux',
          mechanisms: [],
          models: [
            {
              id: 'claude-sonnet',
              provider: 'anthropic',
              label: 'Sonnet',
              reasoning: true,
              backends: ['legacy-backend'],
            },
          ],
        },
      ],
    };

    expect(promptModelCatalog(context)).toEqual([
      {
        id: 'claude-sonnet',
        provider: 'anthropic',
        label: 'Sonnet',
        reasoning: true,
        backends: ['claude-agent', 'legacy-backend', 'claude-cmux'],
      },
    ]);
  });
});
