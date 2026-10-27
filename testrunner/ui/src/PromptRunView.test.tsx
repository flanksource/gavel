import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen, within } from '@testing-library/react';
import type { PromptRunStreamState } from '@flanksource/clicky-ui/ai';
import type { LinterResult } from '@flanksource/clicky-ui/data';
import PromptRunView from './PromptRunView';

const stream = vi.hoisted(() => ({ state: { status: 'idle', entries: [] } as PromptRunStreamState, runId: undefined as string | undefined }));

vi.mock('@flanksource/clicky-ui/ai', () => ({
  usePromptRunStream: (runId: string | undefined) => {
    stream.runId = runId;
    return stream.state;
  },
  SessionInspector: () => <div data-testid="session-inspector" />,
}));

const RUN_ID = 'run-7f3a';
const SESSION_ID = '0b9c2d1e-session';
const result = (over: Partial<LinterResult> = {}): LinterResult => ({ linter: 'messaging-review:report.mdx', violations: [], run_id: RUN_ID, ...over });

function header(): HTMLElement {
  return screen.getByRole('group', { name: 'Prompt run' });
}

beforeEach(() => {
  stream.state = { status: 'running', entries: [] };
});
afterEach(cleanup);

describe('PromptRunView header', () => {
  it('follows the result run id and shows its status, model, provider, mode and run id from the first run frame', () => {
    stream.state = { status: 'running', entries: [], run: { runId: RUN_ID, status: 'running', model: 'sonnet-5', provider: 'anthropic', mode: 'api' } };
    render(<PromptRunView result={result()} baseUrl="/base" />);

    expect(stream.runId).toBe(RUN_ID);
    expect(within(header()).getByText('running')).toBeTruthy();
    expect(within(header()).getByText('sonnet-5 · anthropic · api')).toBeTruthy();
    expect(within(header()).getByText(RUN_ID)).toBeTruthy();
    expect(within(header()).queryByText(/Session/)).toBeNull();
  });

  it('shows the session id, copyable, once the run frame carries it', () => {
    stream.state = { status: 'running', entries: [], sessionId: SESSION_ID, run: { sessionId: SESSION_ID } };
    render(<PromptRunView result={result()} baseUrl="" />);

    expect(within(header()).getByText(SESSION_ID)).toBeTruthy();
    expect(within(header()).getByRole('button', { name: 'Copy session id' })).toBeTruthy();
    expect(within(header()).getByRole('button', { name: 'Copy run id' })).toBeTruthy();
  });

  it('falls back to the result session id while the stream has none', () => {
    render(<PromptRunView result={result({ session_id: SESSION_ID })} baseUrl="" />);

    expect(within(header()).getByText(SESSION_ID)).toBeTruthy();
  });

  it('prefers the stream session id over the result one', () => {
    stream.state = { status: 'running', entries: [], sessionId: SESSION_ID };
    render(<PromptRunView result={result({ session_id: 'stale-session' })} baseUrl="" />);

    expect(within(header()).getByText(SESSION_ID)).toBeTruthy();
    expect(within(header()).queryByText('stale-session')).toBeNull();
  });

  it('shows the elapsed time only once the summary carries a duration', () => {
    render(<PromptRunView result={result()} baseUrl="" />);
    expect(within(header()).queryByText(/Elapsed/)).toBeNull();
    cleanup();

    stream.state = { status: 'succeeded', entries: [], summary: { success: true, duration: '3m2.5s' } };
    render(<PromptRunView result={result()} baseUrl="" />);
    expect(within(header()).getByText('Elapsed 3m2.5s')).toBeTruthy();
    expect(within(header()).getByText('succeeded')).toBeTruthy();
  });

  it('keeps the failure alert on a failed run', () => {
    stream.state = { status: 'failed', entries: [], error: 'provider quota exceeded' };
    render(<PromptRunView result={result()} baseUrl="" />);

    expect(screen.getByRole('alert').textContent).toBe('provider quota exceeded');
    expect(within(header()).getByText('failed')).toBeTruthy();
  });
});
