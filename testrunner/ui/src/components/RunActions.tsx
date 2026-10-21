import { Badge } from '@flanksource/clicky-ui/data';
import { Button } from '@flanksource/clicky-ui/components';
import { UiStop, UiWatch } from '@flanksource/clicky-ui/icons';
import type { RouteState } from '../routes';
import type { RunMeta } from '../types';
import { timeoutArgValue } from '../utils';
import { CopyAIPromptButton, DownloadMenu } from './ExportControls';

const TIMEOUT_ARGS = [
  { key: 'timeout', label: 'global', title: 'Global --timeout' },
  { key: 'test_timeout', label: 'per-test', title: '--test-timeout (per-package)' },
  { key: 'lint_timeout', label: 'per-lint', title: '--lint-timeout (per-linter)' },
] as const;

/** The run's timeout knobs (--timeout, --test-timeout, --lint-timeout) that were set, as pills. */
export function TimeoutPills({ runMeta }: { runMeta: RunMeta | undefined }) {
  const set = TIMEOUT_ARGS.flatMap(arg => {
    const value = timeoutArgValue(runMeta?.args, arg.key);
    return value ? [{ ...arg, value }] : [];
  });
  if (set.length === 0) return null;
  return (
    <span className="inline-flex items-center gap-1" data-timeout-pills="">
      <UiWatch className="text-muted-foreground" />
      {set.map(arg => (
        <span key={arg.key} title={arg.title}>
          <Badge tone="neutral" variant="outline" size="xs">{arg.label} {arg.value}</Badge>
        </span>
      ))}
    </span>
  );
}

function runInfo(runMeta: RunMeta): string {
  return [
    runMeta.pid ? `PID ${runMeta.pid}` : '',
    runMeta.frameworks?.length ? `Frameworks ${runMeta.frameworks.join(', ')}` : '',
    runMeta.command ? `Command ${runMeta.command}` : '',
  ].filter(Boolean).join('\n');
}

export interface RunActionsProps {
  runMeta: RunMeta | undefined;
  /** The run can be stopped and is running. */
  canStop: boolean;
  stopping: boolean;
  onStop: () => void;
  /** The route state that exports the whole current view; absent when there is nothing to export. */
  exportState: RouteState | undefined;
}

/**
 * The run-level actions of the page header: the timeout pills, the global
 * Stop, and the whole view's Download and AI-prompt copy.
 */
export function RunActions({ runMeta, canStop, stopping, onStop, exportState }: RunActionsProps) {
  const info = runMeta ? runInfo(runMeta) : '';
  return (
    <>
      {info && <span className="hidden text-xs text-muted-foreground lg:inline" title={info}>run info</span>}
      <TimeoutPills runMeta={runMeta} />
      {canStop && (
        <Button type="button" size="sm" variant="destructive" disabled={stopping} onClick={onStop} title="Stop the active run">
          <UiStop className="mr-1" />
          {stopping ? 'Stopping...' : 'Stop'}
        </Button>
      )}
      {exportState && <DownloadMenu routeState={exportState} title="Download the whole result as JSON or Markdown" />}
      {exportState && <CopyAIPromptButton routeState={exportState} />}
    </>
  );
}
