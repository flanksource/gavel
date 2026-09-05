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
}));
vi.mock('@flanksource/clicky-ui/components', async importOriginal => ({
  ...(await importOriginal<object>()),
  Modal: ({ children }: { children: ReactNode }) => <div>{children}</div>,
  Combobox: ({ options, onChange }: { options: Array<{ value: string; label: string }>; onChange: (value: string) => void }) => <div>{options.map(option => <Button key={option.value} onClick={() => onChange(option.value)}>{option.label}</Button>)}</div>,
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
  runtimeProfiles: [{ id: 'review-profile', name: 'Review profile', model: 'example-profile-model', presets: [] }],
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
  it('seeds verification from the verify prompt default', () => {
    expect(loadPromptRunOptions('verification', context)).toMatchObject({
      step: 'verify', spec: { mode: 'agent', model: 'example-verify-model' },
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

  it.each([
    { scope: 'approval' as const, step: 'run' },
    { scope: 'verification' as const, step: 'verify' },
  ])('submits $scope as an explicit step and spec', ({ scope, step }) => {
    const spec: AISpecRuntimeValue = { mode: 'agent', model: 'example-run-model', effort: 'medium' };
    const onRun = vi.fn();
    render(<PromptRunAdvancedDialog dir="/repo" scope={scope} open initial={{ step, spec }} onClose={() => {}} onRun={onRun} />);
    fireEvent.click(screen.getByRole('button', { name: 'Save runtime' }));
    expect(JSON.parse(JSON.stringify(onRun.mock.calls[0]?.[0]))).toEqual({ step, spec });
  });

  it.each([
    { scope: 'approval' as const, step: 'run', model: 'example-run-model' },
    { scope: 'verification' as const, step: 'verify', model: 'example-verify-model' },
  ])('selects a profile for $scope without sending seeded runtime overrides', ({ scope, step, model }) => {
    const onRun = vi.fn();
    render(<PromptRunAdvancedDialog dir="/repo" scope={scope} open initial={{ step, spec: { mode: 'agent', model, effort: 'medium' } }} onClose={() => {}} onRun={onRun} />);
    fireEvent.click(screen.getByRole('button', { name: 'Review profile' }));
    fireEvent.click(screen.getByRole('button', { name: 'Save runtime' }));
    expect(JSON.parse(JSON.stringify(onRun.mock.lastCall?.[0]))).toEqual({ step, runtimeProfile: 'review-profile', spec: {} });
  });
});
