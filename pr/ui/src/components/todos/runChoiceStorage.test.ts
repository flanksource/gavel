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

// A prompt body is rendered from one todo. Remembering it as a reusable run
// choice replays that todo's prompt against every todo run afterwards.
describe('run choice history without prompt content', () => {
  const runtime = { mode: 'agent', model: 'example-model', effort: 'xhigh' as const };
  const otherTodoPrompt = '## Other todo title\n\nOther todo body';
  const attachment = { id: 'attachment-1', mediaType: 'image/png' };

  it('drops the prompt body and attachments but keeps the rest of the prompt section', () => {
    const options = { step: 'plan', spec: { ...runtime, prompt: { user: otherTodoPrompt, system: 'Shared system prompt', attachments: [attachment] } } };
    const expected = { step: 'plan', spec: { ...runtime, prompt: { system: 'Shared system prompt' } } };
    expect(rememberTodoRunOptionsForMode(options, true)).toEqual(expected);
    expect(readRunChoiceState()).toEqual({ last: { plan: expected }, recentAdvanced: { plan: [expected] } });
  });

  it('removes a prompt section left empty instead of storing an explicit empty prompt', () => {
    const remembered = rememberTodoRunOptionsForMode({ step: 'plan', spec: { ...runtime, prompt: { user: otherTodoPrompt } } });
    expect(remembered).toEqual({ step: 'plan', spec: runtime });
    expect(readRunChoiceState().last.plan?.spec).not.toHaveProperty('prompt');
  });

  it('strips prompt content already persisted by an earlier session', () => {
    const persisted = { step: 'plan', spec: { ...runtime, prompt: { user: otherTodoPrompt } } };
    localStorage.setItem('gavel.pr-ui.todoRunChoices.v3', JSON.stringify({ last: { plan: persisted }, recentAdvanced: { plan: [persisted] } }));
    expect(readRunChoiceState()).toEqual({ last: { plan: { step: 'plan', spec: runtime } }, recentAdvanced: { plan: [{ step: 'plan', spec: runtime }] } });
  });
});
