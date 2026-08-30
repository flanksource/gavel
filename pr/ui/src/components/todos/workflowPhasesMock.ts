// A stand-in for clicky-ui's WORKFLOW_PHASES, for suites that mock
// '@flanksource/clicky-ui/ai' with an allowlist.
//
// phaseMachine reads the pipeline's label, glyph and tone from the library so a
// phase looks the same everywhere, which means every module that reaches
// phaseMachine — including format.tsx, and therefore every list row — needs
// this export present in the mock. Shared here so the four suites that mock the
// module do not each invent a different shape.
export const WORKFLOW_PHASES_MOCK = {
  plan: { label: 'Plan', icon: () => null, tone: 'sky' },
  run: { label: 'Run', icon: () => null, tone: 'emerald' },
  verify: { label: 'Verify', icon: () => null, tone: 'teal' },
};
