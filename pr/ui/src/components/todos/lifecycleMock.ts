import type { TodoLifecycle, TodoLifecycleStep } from '../../types';

// Fixture lifecycle payloads matching the GET /api/todos/item `lifecycle`
// contract, for suites that construct a TodoItem without going through the
// server. Replaces workflowPhasesMock.ts's WORKFLOW_PHASES stand-in: that mock
// existed only because phaseMachine (deleted) read label/glyph/tone from
// clicky-ui's WORKFLOW_PHASES and every module reaching it — including
// format.tsx, and therefore every list row — needed the mock present. The
// header's phase widget now renders from `todo.lifecycle` directly, so tests
// exercising it need a lifecycle fixture instead of a WORKFLOW_PHASES stand-in.

function step(overrides: Partial<TodoLifecycleStep> & Pick<TodoLifecycleStep, 'name' | 'label'>): TodoLifecycleStep {
  return { applicable: true, suggested: false, done: false, lastRun: null, ...overrides };
}

// A fresh todo: nothing has run, planning comes first.
export const LIFECYCLE_MOCK_DRAFT: TodoLifecycle = {
  steps: [
    step({ name: 'plan', label: 'Plan', suggested: true }),
    step({ name: 'run', label: 'Run' }),
    step({ name: 'verify', label: 'Verify' }),
  ],
  next: 'plan',
  reason: 'Nothing has run yet, so planning comes first.',
};

// Work landed but the definition of done has not been checked.
export const LIFECYCLE_MOCK_IMPLEMENTED: TodoLifecycle = {
  steps: [
    step({ name: 'plan', label: 'Plan', done: true }),
    step({ name: 'run', label: 'Run', done: true }),
    step({ name: 'verify', label: 'Verify', suggested: true }),
  ],
  next: 'verify',
  reason: 'Work landed but the definition of done has not been checked.',
};

// A plan is waiting on a human decision: nothing applies until it is resolved.
export const LIFECYCLE_MOCK_REVIEW: TodoLifecycle = {
  steps: [
    step({ name: 'plan', label: 'Plan', done: true, applicable: false }),
    step({ name: 'run', label: 'Run', applicable: false }),
    step({ name: 'verify', label: 'Verify', applicable: false }),
  ],
  next: null,
  reason: 'A plan is waiting on a human — approve or reject it below.',
};
