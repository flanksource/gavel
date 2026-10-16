// "@flanksource/gavel/testrunner" — the testrunner surface a host builds its own
// view from: the data hooks, the trace query client, the result/progress types
// and the gavel adapters that plug gavel's domain rendering into clicky-ui's
// TestRunner. Presentation itself (tree, detail, lint, process tree and detail)
// is clicky-ui's.
export * from './hooks';
export * from './types';

// Trace query client: the run's trace store is served through commons-db query.
export {
  TraceStreamNotFound,
  checkpointSession,
  fetchLatestSeq,
  fetchTraceKinds,
  fetchTraceRecords,
  openInQueryUrl,
  parseNdjson,
  processMount,
  queryUrl,
  rebaseTraces,
  runTraces,
  traceExportUrl,
} from './traces';
export type { CheckpointRequest, TraceExportFormat, TraceKindInfo, TraceRecord } from './traces';

// Live attempt metrics, read from the run's process_samples.
export { fetchProcessMetricsFromSamples } from './processMetrics';

// Adapters for clicky-ui's TestRunner.
export {
  detailAdapter,
  gavelAdapters,
  gavelRegistry,
  nodeActionsAdapter,
  processesAdapter,
  resourcesAdapter,
  tracesAdapter,
} from './adapters';
export type { GavelAdapterOptions, NodeActionsConfig } from './adapters';

// Components gavel adds on top of clicky-ui's.
export { RunProcesses } from './components/RunProcesses';
export type { RunProcessesProps } from './components/RunProcesses';
export { TracesPanel } from './adapters/TracesPanel';
export { StepProcesses } from './adapters/StepProcesses';
export { ResourcesPanel } from './adapters/ResourcesPanel';
export { BenchView } from './components/BenchView';
export { FixtureBenchView } from './components/FixtureBenchView';
export { RerunDialog } from './components/RerunDialog';
export { RunActions, TimeoutPills } from './components/RunActions';
export { CopyAIPromptButton, DownloadMenu } from './components/ExportControls';

// Run helpers.
export { apiBase, apiUrl, basePath } from './config';
export { buildLintRerunBody } from './lintRerun';
export { buildExportRoute, buildRoute, defaultStatusFilter, parseRoute } from './routes';
export type { ExportFormat, RouteState, TabKey } from './routes';
export { isLintOnlyPhase, processLinksByPid, stripTestTaskGroup, timeoutArgValue } from './utils';
export type { ProcessStepLink } from './utils';

export { App } from './App';
export type { AppProps } from './App';
