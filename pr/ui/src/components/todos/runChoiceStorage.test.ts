import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { normalizeRunOptions, readRunChoiceState, writeRunChoiceState } from './runChoiceStorage';
import { rememberTodoRunOptionsForMode } from './run';

beforeEach(() => {
  const values = new Map<string, string>();
  vi.stubGlobal('localStorage', {
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => values.set(key, value),
  });
});
afterEach(() => vi.unstubAllGlobals());

describe('named lifecycle step history', () => {
  it.each(['plan', 'verify', 'review-security'])('retains %s without legacy request fields', step => {
    const spec = { mode: 'agent', model: 'example-model', effort: 'high' as const };
    expect(normalizeRunOptions(step, { spec, driver: 'agent', runMode: 'run', plan: true, prompt: 'run', resume: true })).toEqual({
      step, spec, resume: true,
    });
  });

  it('round-trips custom step options in their own history bucket', () => {
    const options = { step: 'review-security', spec: { mode: 'agent', model: 'example-model' } };
    writeRunChoiceState({ last: { 'review-security': options }, recentAdvanced: { 'review-security': [options] } });
    expect(readRunChoiceState()).toEqual({ last: { 'review-security': options }, recentAdvanced: { 'review-security': [options] } });
  });

  it('remembers the submitted step instead of the step that originally opened the dialog', () => {
    const options = { step: 'plan', spec: { mode: 'agent', model: 'example-model' } };
    expect(rememberTodoRunOptionsForMode(options, true)).toEqual(options);
    const state = readRunChoiceState();
    expect(state.last.plan).toEqual(options);
    expect(state.last.run).toBeUndefined();
  });
});
