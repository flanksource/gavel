import { Button } from '@flanksource/clicky-ui/components';
import type { ReactNode } from 'react';
import type { AISpecRuntimeValue } from '@flanksource/clicky-ui/ai';
import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { loadPromptRunOptions, PromptRunAdvancedDialog, rememberPromptRunOptions } from './PromptRunButton';
import type { RunContext } from './providers';

vi.mock('./run', async importOriginal => ({
  ...(await importOriginal<object>()),
  useTodoRunContext: () => ({ context, loading: false, error: '' }),
}));
vi.mock('@flanksource/clicky-ui/ai', async importOriginal => ({
  ...(await importOriginal<object>()),
  SpecRuntimeEditor: ({ onSave, beforeSections }: { onSave: () => void; beforeSections: ReactNode }) => <div>{beforeSections}<Button onClick={onSave}>Save runtime</Button></div>,
  OrderedPresetSelect: ({ presets, onChange }: { presets: Array<{ id: string; name: string }>; onChange: (value: string[]) => void }) => (
    <div>{presets.map(preset => <Button key={preset.id} onClick={() => onChange([preset.id])}>{preset.name}</Button>)}</div>
  ),
}));
vi.mock('@flanksource/clicky-ui/components', async importOriginal => ({
  ...(await importOriginal<object>()),
  Modal: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

const context: RunContext = {
  modes: [{
    id: 'agent', label: 'Agent', provider: 'openai', agent: 'codex',
    defaultModel: 'example-run-model', driver: 'agent', mechanisms: [],
    models: ['example-run-model', 'example-verify-model'].map(id => ({ id, label: id, provider: 'openai', reasoning: true })),
  }],
  defaultMode: 'agent', defaultProvider: 'openai', efforts: ['medium'], tools: [], models: [],
  runtimes: [{ family: 'codex', provider: 'openai', catalogPrefix: 'openai', modes: [{ mode: 'agent', schema: { type: 'object' } }] }],
  promptDefaults: { verify: { mode: 'agent', model: 'example-verify-model' } },
  runtimePresets: [{ id: 'review-preset', name: 'Review preset', scope: 'context', spec: { model: 'example-preset-model' } }],
  lifecycle: { steps: [{ name: 'verify', label: 'Verify', prompt: 'verify', readOnly: false }] },
};

beforeEach(() => {
  const values = new Map<string, string>();
  vi.stubGlobal('localStorage', {
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => values.set(key, value),
  });
});
afterEach(() => vi.unstubAllGlobals());

describe('prompt lifecycle options', () => {
  it('leaves verification sparse so the shared resolver applies its prompt default', () => {
    expect(loadPromptRunOptions('verification', context)).toMatchObject({
      step: 'verify', spec: {},
    });
  });

  it('remembers a verification step without implementation-only fields', () => {
    const options = rememberPromptRunOptions('verification', {
      step: 'run', spec: {
        mode: 'agent', model: 'example-run-model', setup: { cwd: '/repo' },
        prompt: { user: 'Implementation prompt' }, workflow: { commits: [{ on: 'run', gates: 'full' }] },
      },
    }, context);
    expect(JSON.parse(JSON.stringify(options))).toEqual({ step: 'verify', spec: { mode: 'agent', model: 'example-run-model', effort: 'medium' } });
  });

  it('remembers an approval run without the prompt body it was dispatched with', () => {
    const spec: AISpecRuntimeValue = { mode: 'agent', model: 'example-run-model', effort: 'medium' };
    const remembered = rememberPromptRunOptions('approval', {
      step: 'run', spec: { ...spec, prompt: { user: 'Prompt rendered for another todo', attachments: [{ id: 'attachment-1' }] } },
    }, context);
    expect(JSON.parse(JSON.stringify(remembered))).toEqual({ step: 'run', spec });
    expect(JSON.parse(JSON.stringify(loadPromptRunOptions('approval', context)))).toEqual({ step: 'run', spec });
  });

  it('strips a prompt body already persisted in approval history', () => {
    const spec: AISpecRuntimeValue = { mode: 'agent', model: 'example-run-model', effort: 'medium' };
    localStorage.setItem('gavel.pr-ui.promptRunChoices.v2', JSON.stringify({
      approval: { last: { step: 'run', spec: { ...spec, prompt: { user: 'Prompt rendered for another todo' } } } },
    }));
    expect(JSON.parse(JSON.stringify(loadPromptRunOptions('approval', context)))).toEqual({ step: 'run', spec });
  });

  it.each([
    { scope: 'approval' as const, step: 'run' },
    { scope: 'verification' as const, step: 'verify' },
  ])('opens $scope advanced options empty even when a last-used entry exists', ({ scope, step }) => {
    const spec: AISpecRuntimeValue = { mode: 'agent', model: 'example-run-model', effort: 'medium' };
    rememberPromptRunOptions(scope, { step, spec }, context);
    const onRun = vi.fn();
    render(<PromptRunAdvancedDialog dir="/repo" scope={scope} open onClose={() => {}} onRun={onRun} />);
    fireEvent.click(screen.getByRole('button', { name: 'Save runtime' }));
    expect(JSON.parse(JSON.stringify(onRun.mock.calls[0]?.[0]))).toEqual({ step, spec: {} });
  });

  it.each([
    { scope: 'approval' as const, step: 'run' },
    { scope: 'verification' as const, step: 'verify' },
  ])('restores the last-used $scope spec on demand', ({ scope, step }) => {
    const spec: AISpecRuntimeValue = { mode: 'agent', model: 'example-run-model', effort: 'medium' };
    rememberPromptRunOptions(scope, { step, spec }, context);
    const onRun = vi.fn();
    render(<PromptRunAdvancedDialog dir="/repo" scope={scope} open onClose={() => {}} onRun={onRun} />);
    fireEvent.click(screen.getByRole('button', { name: /^Last used/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Save runtime' }));
    expect(JSON.parse(JSON.stringify(onRun.mock.calls[0]?.[0]))).toEqual({ step, spec });
  });

  // The catalog only offers `medium`, so reconciling on restore or on save would
  // rewrite the stored `high`; the restored choice must be dispatched unchanged.
  it.each([
    { scope: 'approval' as const, step: 'run' },
    { scope: 'verification' as const, step: 'verify' },
  ])('dispatches a restored $scope runtime without reconciling its effort', ({ scope, step }) => {
    const spec: AISpecRuntimeValue = { mode: 'agent', model: 'example-run-model', effort: 'high' };
    localStorage.setItem('gavel.pr-ui.promptRunChoices.v2', JSON.stringify({ [scope]: { last: { step, spec } } }));
    const onRun = vi.fn();
    render(<PromptRunAdvancedDialog dir="/repo" scope={scope} open onClose={() => {}} onRun={onRun} />);
    fireEvent.click(screen.getByRole('button', { name: /^Last used/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Save runtime' }));
    expect(JSON.parse(JSON.stringify(onRun.mock.calls[0]?.[0]))).toEqual({ step, spec });
  });

  it.each([
    { scope: 'approval' as const, step: 'run', model: 'example-run-model' },
    { scope: 'verification' as const, step: 'verify', model: 'example-verify-model' },
  ])('preserves the restored runtime when selecting a preset for $scope', ({ scope, step, model }) => {
    const spec: AISpecRuntimeValue = { mode: 'agent', model, effort: 'medium' };
    rememberPromptRunOptions(scope, { step, spec }, context);
    const onRun = vi.fn();
    render(<PromptRunAdvancedDialog dir="/repo" scope={scope} open onClose={() => {}} onRun={onRun} />);
    fireEvent.click(screen.getByRole('button', { name: /^Last used/ }));
    fireEvent.click(screen.getByRole('button', { name: 'Review preset' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save runtime' }));
    expect(JSON.parse(JSON.stringify(onRun.mock.lastCall?.[0]))).toEqual({ step, presets: ['review-preset'], spec: { mode: 'agent', model, effort: 'medium' } });
  });
});
