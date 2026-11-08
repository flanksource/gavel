import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import type { PromptFormProps, PromptSnapshot } from '@flanksource/clicky-ui/data';
import { StepPendingForm } from './StepPendingForm';

// PromptForm's own behaviour (fields, Submit/Cancel, errors) is covered by
// clicky-ui's PromptForm spec; vitest here resolves the published clicky-ui, so
// this spec checks only how StepPendingForm finds the step's prompt and wires it.
const OWNER = 'run-2';
const STEP = 'approval';

const promptsOptions: unknown[] = [];
let pending: PromptSnapshot[] = [];
const formProps: PromptFormProps[] = [];
vi.mock('@flanksource/clicky-ui/data', () => ({
  usePrompts: (options: unknown) => {
    promptsOptions.push(options);
    return { prompts: pending, pending, status: 'connected' };
  },
  PromptForm: (props: PromptFormProps) => {
    formProps.push(props);
    return <div data-testid="prompt-form">{props.prompt.title}</div>;
  },
}));

const prompt = (id: string, title: string): PromptSnapshot => ({
  id, title, state: 'pending', owner: OWNER, labels: { node: STEP }, schema: { type: 'object', properties: {} },
});

afterEach(() => {
  cleanup();
  promptsOptions.length = 0;
  formProps.length = 0;
  pending = [];
});

describe('StepPendingForm', () => {
  it("subscribes to the run's pending prompts labelled with the step's node name", () => {
    render(<StepPendingForm node={{ name: STEP }} owner={OWNER} scope={{ runId: 'r9' }} />);
    expect(promptsOptions.at(-1)).toEqual({
      basePath: '/r9/api', owner: OWNER, labels: { node: STEP }, state: 'pending', enabled: true,
    });
  });

  it('renders an answerable form for each pending prompt of the step, on the run\'s prompt API', () => {
    pending = [prompt('p1', 'Purge staged activities?'), prompt('p2', 'Confirm reload')];
    render(<StepPendingForm node={{ name: STEP }} owner={OWNER} />);

    expect(screen.getAllByTestId('prompt-form').map((form) => form.textContent)).toEqual(['Purge staged activities?', 'Confirm reload']);
    expect(formProps.map((props) => ({ id: props.prompt.id, basePath: props.basePath, idPrefix: props.idPrefix })))
      .toEqual([
        { id: 'p1', basePath: '/api', idPrefix: 'step-prompt-p1' },
        { id: 'p2', basePath: '/api', idPrefix: 'step-prompt-p2' },
      ]);
    expect(formProps.every((props) => typeof props.lookupFetcher === 'function')).toBe(true);
  });

  it('says so when the step has no pending form', () => {
    render(<StepPendingForm node={{ name: STEP }} owner={OWNER} />);
    expect(screen.queryByText('No pending form for this step')).not.toBeNull();
    expect(screen.queryByTestId('prompt-form')).toBeNull();
  });

  it('does not subscribe for a run with no prompt owner, and says why no form shows', () => {
    render(<StepPendingForm node={{ name: STEP }} owner={undefined} />);
    expect(promptsOptions.at(-1)).toMatchObject({ enabled: false });
    expect(screen.queryByText('No pending form for this step: the run publishes no prompts')).not.toBeNull();
  });
});
