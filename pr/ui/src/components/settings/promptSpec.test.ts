import { describe, expect, it } from 'vitest';
import { runtimeValueToPayload, specToRuntimeValue } from './promptSpec';

describe('promptSpec adapters', () => {
  it('folds the body into prompt.user and back out as the document body', () => {
    const value = specToRuntimeValue({ model: 'claude-x' }, 'Review {{diff}}.');
    expect(value.prompt?.user).toBe('Review {{diff}}.');
    expect(value.model).toBe('claude-x');

    const payload = runtimeValueToPayload(value);
    expect(payload.body).toBe('Review {{diff}}.');
    expect(payload.spec.model).toBe('claude-x');
    // The body must not be duplicated into the frontmatter prompt block.
    expect(payload.spec.prompt?.user).toBeUndefined();
  });

  it('keeps a frontmatter system prompt while sending the body separately', () => {
    const value = specToRuntimeValue({ prompt: { system: 'Be terse.' } }, 'do it');
    const payload = runtimeValueToPayload(value);
    expect(payload.body).toBe('do it');
    expect(payload.spec.prompt?.system).toBe('Be terse.');
    expect(payload.spec.prompt?.user).toBeUndefined();
  });

  it('drops an empty body so a body-only prompt carries no prompt frontmatter', () => {
    const payload = runtimeValueToPayload(specToRuntimeValue({ model: 'm' }, ''));
    expect(payload.body).toBe('');
    expect(payload.spec.prompt).toBeUndefined();
    expect(payload.spec.model).toBe('m');
  });
});
