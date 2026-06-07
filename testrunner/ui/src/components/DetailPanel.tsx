import { useState, useRef, useCallback, type ReactNode } from 'react';
import type { Test, FixtureContext, GinkgoContext, GoTestContext, Violation, LinterResult, RunMeta, FailureDetail, TestEditAction, TestEditScope } from '../types';
import {
  statusIcon,
  statusColor,
  formatDuration,
  sum,
  frameworkIcon,
  lintNodeCount,
  formatRunTimestamp,
  formatRunDuration,
  hasTimeoutArgs,
  timeoutArgValue,
} from '../utils';
import { Clicky } from '@flanksource/clicky-ui/clicky';
import { JsonView } from './JsonView';
import { AnsiHtml } from './AnsiHtml';
import { ProgressBar } from './ProgressBar';
import { TestAttempts } from './TestAttempts';
import { DownloadMenu } from './DownloadMenu';
import { copyCurrentViewForAgent } from '../export';
import type { RouteState } from '../routes';
import {
  lintFolderActions,
  lintFileActions,
  collectFolderLintStats,
  folderPattern,
  type IgnoreRequest,
  type LintAction,
} from './lintActions';

export type { IgnoreRequest } from './lintActions';

interface Props {
  test: Test | null;
  lint?: LinterResult[];
  onRerun?: (t: Test) => void;
  rerunBusy?: boolean;
  onStop?: (t: Test) => void;
  stopBusy?: boolean;
  onIgnore?: (req: IgnoreRequest) => Promise<void> | void;
  ignoreBusy?: boolean;
  onTestEdit?: (t: Test, action: TestEditAction, scope: TestEditScope) => Promise<void> | void;
  testEditBusy?: boolean;
  testEditSupported?: boolean;
  runMeta?: RunMeta;
  nodeRouteState?: RouteState;
  failingOnlyRouteState?: RouteState;
}

function isLegacyClickyDocument(value: unknown): value is { version: unknown; node: unknown } {
  return (
    typeof value === 'object' &&
    value !== null &&
    'version' in value &&
    'node' in value
  );
}

function DetailValue({ value }: { value: unknown }) {
  if (isLegacyClickyDocument(value)) {
    return <Clicky data={value as any} />;
  }
  return <JsonView data={value} />;
}

function taskMeta(t: Test): { duration?: string; status?: string; type?: string } | null {
  if (t.framework !== 'task' || !t.context || typeof t.context !== 'object') return null;
  const ctx = t.context as Record<string, any>;
  return {
    duration: typeof ctx.duration === 'string' ? ctx.duration : undefined,
    status: typeof ctx.status === 'string' ? ctx.status : undefined,
    type: typeof ctx.type === 'string' ? ctx.type : undefined,
  };
}

export function DetailPanel({ test: t, lint, onRerun, rerunBusy, onStop, stopBusy, onIgnore, ignoreBusy, onTestEdit, testEditBusy, testEditSupported, runMeta, nodeRouteState, failingOnlyRouteState }: Props) {
  const [copyState, setCopyState] = useState<'idle' | 'copying' | 'copied' | 'error'>('idle');
  const [copyError, setCopyError] = useState('');
  const copyResetTimer = useRef<number | null>(null);

  const resetCopyFeedback = useCallback((nextState: 'copied' | 'error', error: string = '') => {
    setCopyState(nextState);
    setCopyError(error);
    if (copyResetTimer.current) window.clearTimeout(copyResetTimer.current);
    copyResetTimer.current = window.setTimeout(() => {
      setCopyState('idle');
      setCopyError('');
      copyResetTimer.current = null;
    }, nextState === 'copied' ? 2000 : 3000);
  }, []);

  const onCopyAIPrompt = useCallback(async () => {
    if (copyState === 'copying' || !failingOnlyRouteState) return;
    setCopyState('copying');
    setCopyError('');
    if (copyResetTimer.current) {
      window.clearTimeout(copyResetTimer.current);
      copyResetTimer.current = null;
    }
    try {
      await copyCurrentViewForAgent(failingOnlyRouteState);
      resetCopyFeedback('copied');
    } catch (e: any) {
      resetCopyFeedback('error', e?.message || 'Copy failed');
    }
  }, [copyState, failingOnlyRouteState, resetCopyFeedback]);

  if (!t) {
    return (
      <div className="flex items-center justify-center h-full text-gray-400 text-sm">
        <div className="text-center">
          <iconify-icon icon="codicon:list-tree" className="text-4xl mb-2 block" />
          Select a test to view details
        </div>
      </div>
    );
  }

  const hasChildren = (t.children?.length ?? 0) > 0;
  const s = hasChildren ? sum(t) : null;
  const fw = t.framework;
  const fwIcon = frameworkIcon(fw);
  const task = taskMeta(t);
  const isLint = t.kind === 'lint-root' || t.kind === 'lint-folder' || t.kind === 'linter'
    || t.kind === 'violation' || t.kind === 'lint-file' || t.kind === 'lint-rule' || t.kind === 'lint-rule-group';
  const canRerun = !!onRerun && t.kind !== 'violation' && t.framework !== 'task';
  const canStop = !!onStop && !!t.can_stop && !!t.task_id;
  const canExportNode = !!nodeRouteState && !!t.route_path && !isLint && t.kind !== 'violation' && t.framework !== 'task';
  const hasFailingContent = !!t.failed || !!t.timed_out || (s ? s.failed > 0 || s.timedout > 0 : false);
  const canCopyAIPrompt = canExportNode && !!failingOnlyRouteState && hasFailingContent;
  const canEditTest = !!testEditSupported && !!onTestEdit && !isLint && t.framework !== 'task' && editableFramework(t.framework) && !!t.file;
  const confirmTestEdit = (action: TestEditAction, scope: TestEditScope) => {
    if (!onTestEdit) return;
    const verb = action === 'skip' ? 'Skip' : 'Delete';
    const scopeLabel = scope === 'file' ? 'file' : 'test';
    const target = scope === 'file' ? (t.file || 'file') : (t.name || 'test');
    if (typeof window !== 'undefined' && !window.confirm(`${verb} ${scopeLabel} ${target}?`)) return;
    void onTestEdit(t, action, scope);
  };

  return (
    <div className="h-full overflow-y-auto p-5 space-y-4">
      {/* Header */}
      <div className="flex items-start gap-2">
        <iconify-icon icon={statusIcon(t)} className={`${statusColor(t)} text-2xl shrink-0 mt-0.5`} />
        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-2">
            <h2 className="text-lg font-bold text-gray-900 break-words">{t.name}</h2>
            <div className="flex items-center gap-1 shrink-0">
              {canExportNode && nodeRouteState && (
                <DownloadMenu routeState={nodeRouteState} align="right" title="Download this node as JSON or Markdown" />
              )}
              {canExportNode && (
                <button
                  className={`text-xs px-2 py-1 rounded border transition-colors flex items-center gap-1 ${
                    copyState === 'error'
                      ? 'border-red-300 text-red-700 bg-red-50 hover:bg-red-100'
                      : copyState === 'copied'
                        ? 'border-green-300 text-green-700 bg-green-50 hover:bg-green-100'
                        : 'border-gray-300 text-gray-600 hover:bg-gray-200'
                  } disabled:opacity-50 disabled:cursor-not-allowed`}
                  onClick={onCopyAIPrompt}
                  disabled={!canCopyAIPrompt || copyState === 'copying'}
                  title={!hasFailingContent ? 'No failures to copy' : (copyError || 'Copy failing output and repro steps as AI prompt')}
                >
                  <iconify-icon icon={copyState === 'copied' ? 'codicon:check' : copyState === 'copying' ? 'svg-spinners:ring-resize' : 'codicon:copy'} />
                  {copyState === 'copying' ? 'Copying...' : copyState === 'copied' ? 'Copied' : copyState === 'error' ? 'Copy failed' : 'Copy AI Prompt'}
                </button>
              )}
              {canRerun && (
                <button
                  className="text-xs px-2 py-1 rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                  onClick={() => onRerun!(t)}
                  disabled={rerunBusy}
                  title="Rerun this test or subtree"
                >
                  <iconify-icon icon="codicon:refresh" />
                  {rerunBusy ? 'Running...' : 'Rerun'}
                </button>
              )}
              {canStop && (
                <button
                  className="text-xs px-2 py-1 rounded bg-orange-600 text-white hover:bg-orange-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                  onClick={() => onStop!(t)}
                  disabled={stopBusy}
                  title="Stop this running task"
                >
                  <iconify-icon icon="codicon:debug-stop" />
                  {stopBusy ? 'Stopping...' : 'Stop'}
                </button>
              )}
            </div>
          </div>
          <div className="flex items-center gap-2 mt-1 flex-wrap">
            {fwIcon && (
              <span className="inline-flex items-center gap-1 text-xs bg-gray-100 rounded px-1.5 py-0.5 text-gray-600">
                <iconify-icon icon={fwIcon} className="text-sm" />
                {fw}
              </span>
            )}
            {t.duration ? (
              <span className="text-xs text-gray-500">
                <iconify-icon icon="codicon:clock" className="mr-0.5" />
                {formatDuration(t.duration)}
              </span>
            ) : task?.duration ? (
              <span className="text-xs text-gray-500">
                <iconify-icon icon="codicon:clock" className="mr-0.5" />
                {task.duration}
              </span>
            ) : null}
            {t.file && (
              <span className="text-xs text-gray-500 font-mono">
                {t.file}{t.line ? `:${t.line}` : ''}
              </span>
            )}
          </div>
        </div>
      </div>

      {t.kind === 'violation' && t.violation && <ViolationDetail v={t.violation} />}
      {canEditTest && (
        <Section title="Source Actions">
          <div className="flex flex-wrap gap-1.5">
            <TestEditButton
              icon="codicon:circle-slash"
              label="Skip Test"
              disabled={testEditBusy}
              onClick={() => confirmTestEdit('skip', 'test')}
            />
            <TestEditButton
              icon="codicon:trash"
              label="Delete Test"
              variant="danger"
              disabled={testEditBusy}
              onClick={() => confirmTestEdit('delete', 'test')}
            />
            <TestEditButton
              icon="codicon:file"
              label="Skip File"
              disabled={testEditBusy}
              onClick={() => confirmTestEdit('skip', 'file')}
            />
            <TestEditButton
              icon="codicon:trash"
              label="Delete File"
              variant="danger"
              disabled={testEditBusy}
              onClick={() => confirmTestEdit('delete', 'file')}
            />
          </div>
        </Section>
      )}
      {t.kind === 'linter' && <LinterDetail t={t} onIgnore={onIgnore} ignoreBusy={ignoreBusy} />}
      {t.kind === 'lint-folder' && (
        <FolderLintDetail t={t} lint={lint} onIgnore={onIgnore} ignoreBusy={ignoreBusy} />
      )}
      {t.kind === 'lint-file' && (
        <FileViolationsDetail t={t} onIgnore={onIgnore} ignoreBusy={ignoreBusy} />
      )}
      {(t.kind === 'lint-rule' || t.kind === 'lint-rule-group') && (
        <RuleViolationsDetail t={t} onIgnore={onIgnore} ignoreBusy={ignoreBusy} />
      )}

      {runMeta && (
        <Section title="Run">
          <div className="grid grid-cols-4 gap-3 text-sm">
            <MetaCard
              label={runMeta.kind === 'rerun' ? `Rerun #${runMeta.sequence}` : 'Initial run'}
              value={runMeta.started ? formatRunTimestamp(runMeta.started) : 'Pending'}
            />
            <MetaCard
              label="Duration"
              value={
                runMeta.ended
                  ? formatRunDuration(runMeta.started, runMeta.ended)
                  : runMeta.started
                    ? (
                      <span className="inline-flex items-center gap-1">
                        <iconify-icon icon="svg-spinners:ring-resize" className="text-blue-500" />
                        {formatRunDuration(runMeta.started, undefined)}
                      </span>
                    )
                    : '—'
              }
            />
            <MetaCard
              label="Exit"
              value={
                runMeta.exit_code !== undefined
                  ? (
                    <span className={runMeta.exit_code === 0 ? 'text-green-700' : 'text-red-700'}>
                      {runMeta.exit_code}
                    </span>
                  )
                  : runMeta.ended ? '—' : 'running'
              }
            />
            <MetaCard
              label="Timeout"
              value={
                runMeta.timed_out
                  ? <span className="text-amber-700 inline-flex items-center gap-1"><iconify-icon icon="mdi:clock-alert-outline" />triggered</span>
                  : runMeta.ended ? 'ok' : '—'
              }
            />
          </div>
          {(runMeta.pid || runMeta.command || (runMeta.frameworks && runMeta.frameworks.length > 0)) && (
            <div className="mt-3 grid grid-cols-1 gap-2 text-xs text-gray-600">
              {runMeta.pid !== undefined && runMeta.pid > 0 && (
                <div className="flex gap-2">
                  <span className="text-gray-400 w-20">PID</span>
                  <span className="font-mono">{runMeta.pid}</span>
                </div>
              )}
              {runMeta.frameworks && runMeta.frameworks.length > 0 && (
                <div className="flex gap-2">
                  <span className="text-gray-400 w-20">Frameworks</span>
                  <span>{runMeta.frameworks.join(', ')}</span>
                </div>
              )}
              {runMeta.command && (
                <div className="flex gap-2">
                  <span className="text-gray-400 w-20">Command</span>
                  <span className="font-mono truncate" title={runMeta.command}>{runMeta.command}</span>
                </div>
              )}
            </div>
          )}
          {hasTimeoutArgs(runMeta.args) && (
            <div className="mt-3 grid grid-cols-3 gap-3 text-sm">
              {timeoutArgValue(runMeta.args, 'timeout') && (
                <MetaCard label="Global timeout" value={timeoutArgValue(runMeta.args, 'timeout')!} />
              )}
              {timeoutArgValue(runMeta.args, 'test_timeout') && (
                <MetaCard label="Per-test" value={timeoutArgValue(runMeta.args, 'test_timeout')!} />
              )}
              {timeoutArgValue(runMeta.args, 'lint_timeout') && (
                <MetaCard label="Per-linter" value={timeoutArgValue(runMeta.args, 'lint_timeout')!} />
              )}
            </div>
          )}
        </Section>
      )}

      {/* Summary for containers */}
      {s && s.total > 0 && (
        <div className="space-y-3 border rounded-lg p-3 bg-gray-50">
          <div className="flex gap-4 text-sm">
            <Stat label="Total" value={s.total} color="text-gray-700" />
            <Stat label="Passed" value={s.passed} color="text-green-600" />
            <Stat label="Failed" value={s.failed} color="text-red-600" />
            {s.warned > 0 && <Stat label="Warned" value={s.warned} color="text-amber-600" />}
            {s.skipped > 0 && <Stat label="Skipped" value={s.skipped} color="text-yellow-600" />}
            {s.pending > 0 && <Stat label="Pending" value={s.pending} color="text-blue-600" />}
          </div>
          <ProgressBar
            segments={[
              { count: s.passed, color: 'bg-green-500', label: 'passed' },
              { count: s.warned, color: 'bg-amber-400', label: 'warned' },
              { count: s.skipped, color: 'bg-yellow-400', label: 'skipped' },
              { count: s.failed, color: 'bg-red-500', label: 'failed' },
              { count: s.pending, color: 'bg-blue-300', label: 'pending' },
            ]}
            total={s.total}
            height="h-2.5"
          />
        </div>
      )}

      {/* Suite path */}
      {t.suite && t.suite.length > 0 && (
        <Section title="Suite">
          <span className="text-sm text-gray-700">{t.suite.join(' > ')}</span>
        </Section>
      )}

      {/* Package */}
      {t.package_path && (
        <Section title="Package">
          <span className="text-sm text-gray-700 font-mono">{t.package_path}</span>
        </Section>
      )}

      {/* Error message — prefer the structured FailureDetail view when the
          parser recognised a known failure shape, otherwise fall back to the
          raw message. The raw message is still reachable as a <details>
          fallback so nothing is hidden by parsing heuristics. */}
      {t.failure_detail && t.failure_detail.kind && t.failure_detail.kind !== 'raw' ? (
        <Section title="Error">
          <FailureDetailView detail={t.failure_detail} rawMessage={t.message} />
        </Section>
      ) : t.message ? (
        <Section title="Error">
          <AnsiHtml
            text={t.message}
            className="text-sm text-red-700 whitespace-pre-wrap font-mono bg-red-50 rounded p-3 max-h-64 overflow-y-auto block"
          />
        </Section>
      ) : null}

      {/* Command */}
      {t.command && (
        <Section title="Command">
          <pre className="text-sm text-gray-700 whitespace-pre-wrap font-mono bg-blue-50 rounded p-3">
            <span className="text-gray-400">$ </span>{t.command}
          </pre>
        </Section>
      )}

      {task && (task.status || task.type) && (
        <Section title="Task">
          <div className="flex gap-4 text-sm text-gray-700">
            {task.type && <span>type: <span className="font-mono">{task.type}</span></span>}
            {task.status && <span>status: <span className="font-mono">{task.status}</span></span>}
          </div>
        </Section>
      )}

      {/* Framework-specific context */}
      {fw === 'fixture' && t.context && <FixtureDetail ctx={t.context as FixtureContext} />}
      {fw === 'ginkgo' && t.context && <GinkgoDetail ctx={t.context as GinkgoContext} failureLocation={t.failure_detail?.location} suiteAbove={t.suite?.join(' > ')} />}
      {fw === 'go test' && t.context && <GoTestDetail ctx={t.context as GoTestContext} />}

      {t.attempts && t.attempts.length > 0 && (
        <Section title={t.attempts.length > 1 ? `Attempts (${t.attempts.length})` : 'Attempt'}>
          <TestAttempts test={t} />
        </Section>
      )}

      {t.detail != null && (
        <Section title="Detail">
          <DetailValue value={t.detail} />
        </Section>
      )}

      {t.stdout && (
        <Section title="stdout">
          <AnsiHtml text={t.stdout} className="text-sm text-gray-700 whitespace-pre-wrap font-mono bg-gray-50 rounded p-3 max-h-80 overflow-y-auto" />
        </Section>
      )}

      {t.stderr && (
        <Section title="stderr">
          <AnsiHtml text={t.stderr} className="text-sm text-gray-700 whitespace-pre-wrap font-mono bg-gray-50 rounded p-3 max-h-80 overflow-y-auto" />
        </Section>
      )}

      {/* Children list for containers */}
      {hasChildren && (
        <Section title={isLint ? 'Tree' : 'Tests'}>
          <div className="space-y-0.5">
            {t.children!.map((child, i) => (
              <ChildRow key={i} test={child} />
            ))}
          </div>
        </Section>
      )}
    </div>
  );
}

const CEL_VAR_HIDDEN = new Set(['stdout', 'stderr', 'Stdout', 'Stderr']);

function filterCelVars(vars: Record<string, any>): Record<string, any> {
  const filtered: Record<string, any> = {};
  for (const [k, v] of Object.entries(vars)) {
    if (CEL_VAR_HIDDEN.has(k)) {
      filtered[k] = '(shown above)';
    } else {
      filtered[k] = v;
    }
  }
  return filtered;
}

function FailureDetailView({ detail: d, rawMessage }: { detail: FailureDetail; rawMessage?: string }) {
  const showRawFallback = !!rawMessage && rawMessage !== d.summary && rawMessage.includes('\n');

  return (
    <div className="space-y-2">
      {d.summary && (
        <div className="text-sm text-red-700 font-mono bg-red-50 rounded p-2">
          {d.summary}
        </div>
      )}

      {d.kind === 'gomega' && (d.actual || d.expected) && (
        <div className="grid grid-cols-2 gap-2">
          {d.actual !== undefined && (
            <div>
              <div className="text-xs font-semibold text-gray-500 mb-1">Actual</div>
              <pre className="text-sm font-mono bg-red-50 text-red-800 rounded p-2 whitespace-pre-wrap max-h-64 overflow-y-auto">{d.actual}</pre>
            </div>
          )}
          {d.expected !== undefined && (
            <div>
              <div className="text-xs font-semibold text-gray-500 mb-1">
                Expected{d.matcher ? ` (${d.matcher})` : ''}
              </div>
              <pre className="text-sm font-mono bg-green-50 text-green-800 rounded p-2 whitespace-pre-wrap max-h-64 overflow-y-auto">{d.expected}</pre>
            </div>
          )}
        </div>
      )}

      {d.kind === 'panic' && d.stack && (
        <details className="bg-gray-50 rounded p-2">
          <summary className="text-xs font-semibold text-gray-500 cursor-pointer">Stack trace</summary>
          <pre className="text-xs font-mono text-gray-800 whitespace-pre-wrap mt-2 max-h-80 overflow-y-auto">{d.stack}</pre>
        </details>
      )}

      {d.kind === 'go_test' && d.actual && (
        <pre className="text-sm font-mono bg-red-50 text-red-800 rounded p-2 whitespace-pre-wrap max-h-64 overflow-y-auto">{d.actual}</pre>
      )}

      {d.location && (
        <div className="text-xs text-gray-500 font-mono">
          <iconify-icon icon="codicon:location" className="mr-0.5" />{d.location}
        </div>
      )}

      {showRawFallback && (
        <details className="bg-gray-50 rounded p-2">
          <summary className="text-xs font-semibold text-gray-500 cursor-pointer">Full message</summary>
          <AnsiHtml
            text={rawMessage!}
            className="text-xs text-gray-700 font-mono whitespace-pre-wrap mt-2 max-h-64 overflow-y-auto block"
          />
        </details>
      )}
    </div>
  );
}

function FixtureDetail({ ctx }: { ctx: FixtureContext }) {
  const hasComparison = ctx.expected !== undefined && ctx.actual !== undefined;

  return (
    <>
      {ctx.command && (
        <Section title="Command">
          <pre className="text-sm text-gray-700 whitespace-pre-wrap font-mono bg-blue-50 rounded p-3">
            <span className="text-gray-400">$ </span>{ctx.command}
          </pre>
          <div className="flex gap-4 mt-1 text-xs text-gray-500">
            {ctx.cwd && <span><iconify-icon icon="codicon:folder" className="mr-0.5" />{ctx.cwd}</span>}
            {ctx.exit_code !== undefined && ctx.exit_code !== 0 && (
              <span className="text-red-500">exit code: {ctx.exit_code}</span>
            )}
          </div>
        </Section>
      )}

      {ctx.cel_expression && (
        <Section title="CEL Expression">
          <pre className="text-sm whitespace-pre-wrap font-mono bg-purple-50 text-purple-800 rounded p-3">
            {ctx.cel_expression}
          </pre>
        </Section>
      )}

      {ctx.cel_vars && Object.keys(ctx.cel_vars).length > 0 && (
        <Section title="CEL Variables">
          <div className="bg-gray-50 rounded p-2 max-h-80 overflow-y-auto">
            <JsonView data={filterCelVars(ctx.cel_vars)} />
          </div>
        </Section>
      )}

      {hasComparison && (
        <Section title="Expected vs Actual">
          <div className="grid grid-cols-2 gap-2">
            <div>
              <div className="text-xs font-semibold text-gray-500 mb-1">Expected</div>
              <pre className="text-sm font-mono bg-green-50 text-green-800 rounded p-2 whitespace-pre-wrap max-h-40 overflow-y-auto">
                {typeof ctx.expected === 'object' ? JSON.stringify(ctx.expected, null, 2) : String(ctx.expected)}
              </pre>
            </div>
            <div>
              <div className="text-xs font-semibold text-gray-500 mb-1">Actual</div>
              <pre className="text-sm font-mono bg-red-50 text-red-800 rounded p-2 whitespace-pre-wrap max-h-40 overflow-y-auto">
                {typeof ctx.actual === 'object' ? JSON.stringify(ctx.actual, null, 2) : String(ctx.actual)}
              </pre>
            </div>
          </div>
        </Section>
      )}
    </>
  );
}

function GinkgoDetail({
  ctx,
  failureLocation,
  suiteAbove,
}: {
  ctx: GinkgoContext;
  failureLocation?: string;
  suiteAbove?: string;
}) {
  // Skip the Failure Location block when FailureDetailView already shows the
  // same file:line above. Skip the suite_description line when the top
  // "Suite" section already shows the same string. suite_path (filesystem
  // location of the suite binary) is still useful — keep it when present.
  const showLocation = ctx.failure_location && ctx.failure_location !== failureLocation;
  const showDescription = ctx.suite_description && ctx.suite_description !== suiteAbove;
  const showSuite = showDescription || ctx.suite_path;
  return (
    <>
      {showSuite && (
        <Section title="Suite">
          {showDescription && <div className="text-sm text-gray-700 font-medium">{ctx.suite_description}</div>}
          {ctx.suite_path && <div className="text-xs text-gray-500 font-mono mt-0.5">{ctx.suite_path}</div>}
        </Section>
      )}

      {showLocation && (
        <Section title="Failure Location">
          <span className="text-sm text-red-600 font-mono">{ctx.failure_location}</span>
        </Section>
      )}
    </>
  );
}

function GoTestDetail({ ctx }: { ctx: GoTestContext }) {
  return (
    <>
      {ctx.parent_test && (
        <Section title="Parent Test">
          <span className="text-sm text-gray-700 font-mono">{ctx.parent_test}</span>
        </Section>
      )}
      {ctx.import_path && (
        <Section title="Import Path">
          <span className="text-sm text-gray-700 font-mono">{ctx.import_path}</span>
        </Section>
      )}
    </>
  );
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div>
      <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-1">{title}</h3>
      {children}
    </div>
  );
}

function Stat({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div className="text-center">
      <div className={`text-lg font-bold ${color}`}>{value}</div>
      <div className="text-xs text-gray-500">{label}</div>
    </div>
  );
}

function MetaCard({ label, value }: { label: string; value: any }) {
  return (
    <div className="rounded-lg border border-gray-200 bg-gray-50 px-3 py-2">
      <div className="text-xs uppercase tracking-wide text-gray-500">{label}</div>
      <div className="mt-1 text-sm font-medium text-gray-800">{value}</div>
    </div>
  );
}

function ViolationDetail({ v }: { v: Violation }) {
  const sevColor = v.severity === 'error' ? 'bg-red-100 text-red-800'
    : v.severity === 'warning' ? 'bg-yellow-100 text-yellow-800'
    : 'bg-blue-100 text-blue-800';
  return (
    <>
      <Section title="Severity">
        <span className={`inline-block text-xs font-semibold uppercase rounded px-2 py-0.5 ${sevColor}`}>
          {v.severity || 'info'}
        </span>
      </Section>
      {v.file && (
        <Section title="Location">
          <span className="text-sm font-mono text-gray-700">
            {v.file}{v.line ? `:${v.line}` : ''}{v.column ? `:${v.column}` : ''}
          </span>
        </Section>
      )}
      {v.rule?.method && (
        <Section title="Rule">
          <span className="text-sm font-mono text-gray-700">{v.rule.method}</span>
          {v.rule.description && <div className="text-xs text-gray-500 mt-0.5">{v.rule.description}</div>}
        </Section>
      )}
      {v.message && (
        <Section title="Message">
          <pre className="text-sm text-gray-700 whitespace-pre-wrap font-mono bg-gray-50 rounded p-3">{v.message}</pre>
        </Section>
      )}
      {v.code && (
        <Section title="Code">
          <pre className="text-sm text-gray-700 whitespace-pre-wrap font-mono bg-gray-50 rounded p-3 overflow-x-auto">{v.code}</pre>
        </Section>
      )}
    </>
  );
}

function groupViolationsByRule(violations: Violation[]): Array<{ rule: string; violations: Violation[] }> {
  const grouped = new Map<string, Violation[]>();
  for (const v of violations) {
    const key = v.rule?.method || '(no rule)';
    let bucket = grouped.get(key);
    if (!bucket) {
      bucket = [];
      grouped.set(key, bucket);
    }
    bucket.push(v);
  }
  return Array.from(grouped.keys()).sort().map(rule => ({
    rule,
    violations: grouped.get(rule)!.slice().sort((a, b) => (a.line || 0) - (b.line || 0)),
  }));
}

function LinterDetail({ t, onIgnore, ignoreBusy }: LintDetailProps) {
  const lr = t.linter;
  const noFileViolations = t.noFileViolations || [];
  const linter = t.linterName || lr?.linter || '';
  const workDir = t.work_dir || lr?.work_dir || '';
  return (
    <>
      {lr && (
        <>
          <Section title="Status">
            <div className="flex gap-3 text-sm">
              <span className={lr.skipped ? 'text-gray-500' : lr.timed_out ? 'text-amber-600' : lr.success ? 'text-green-600' : 'text-red-600'}>
                {lr.skipped ? 'skipped' : lr.timed_out ? 'timed out' : lr.success ? 'success' : 'failed'}
              </span>
              <span className="text-gray-500">{(lr.violations || []).length} violations</span>
              {lr.file_count !== undefined && <span className="text-gray-500">{lr.file_count} files</span>}
              {lr.rule_count !== undefined && <span className="text-gray-500">{lr.rule_count} rules</span>}
            </div>
          </Section>
          {lr.command && (
            <Section title="Command">
              <pre className="text-xs text-gray-800 whitespace-pre-wrap break-all font-mono bg-gray-50 rounded p-3">{formatLinterCommand(lr)}</pre>
              {workDir && (
                <div className="mt-1 text-xs text-gray-500 font-mono">cwd: {workDir}</div>
              )}
            </Section>
          )}
          {linter && (
            <Section title="Actions">
              <div className="flex flex-wrap gap-1.5">
                <IgnoreButton
                  label={`Disable ${linter} entirely`}
                  title="Add {source} to .gavel.yaml"
                  req={{ source: linter, work_dir: workDir }}
                  onIgnore={onIgnore}
                  disabled={ignoreBusy}
                />
              </div>
            </Section>
          )}
          {lr.error && (
            <Section title="Error">
              <pre className="text-sm text-red-700 whitespace-pre-wrap font-mono bg-red-50 rounded p-3">{lr.error}</pre>
            </Section>
          )}
          {lr.raw_output && (
            <Section title="Raw output">
              <pre className="text-xs text-gray-700 whitespace-pre-wrap font-mono bg-gray-50 rounded p-3 max-h-80 overflow-y-auto">{lr.raw_output}</pre>
            </Section>
          )}
        </>
      )}
      {noFileViolations.length > 0 && (
        <Section title="Violations Without File">
          <div className="space-y-3">
            {groupViolationsByRule(noFileViolations).map(({ rule, violations }) => (
              <div key={rule} className="border border-gray-200 rounded p-3 bg-gray-50">
                <div className="text-sm font-mono text-gray-700 mb-2">
                  {rule} <span className="text-xs text-gray-400">({violations.length})</span>
                </div>
                <div className="space-y-2">
                  {violations.map((v, i) => (
                    <ViolationRow
                      key={i}
                      v={v}
                      linter={linter}
                      file=""
                      workDir={workDir}
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
        </Section>
      )}
    </>
  );
}

// formatLinterCommand returns the shell-quoted argv that was executed. Args
// containing whitespace or shell metacharacters are single-quoted so the line
// can be copy-pasted into a terminal safely.
function formatLinterCommand(lr: LinterResult): string {
  if (!lr.command) return '';
  const parts = [shellQuote(lr.command), ...(lr.args || []).map(shellQuote)];
  return parts.join(' ');
}

function shellQuote(s: string): string {
  if (s === '') return "''";
  if (/^[A-Za-z0-9_./:=@%+,-]+$/.test(s)) return s;
  return `'${s.replace(/'/g, `'\\''`)}'`;
}

function FolderLintDetail({ t, lint, onIgnore, ignoreBusy }: LintDetailProps & { lint?: LinterResult[] }) {
  const folderDisplay = t.file || t.name;
  const pattern = folderPattern(t.target_path);
  const scopedRule = t.ruleName || '';
  const scopedLinter = t.linterName || '';
  const actions = lintFolderActions(t, lint);
  // The Linters section still wants stat counts to render — keep the
  // aggregation only for that informational block (not for action gating).
  const linterStats = scopedRule ? [] : collectFolderLintStats(t, lint);
  return (
    <>
      <Section title="Folder">
        <div className="flex items-center gap-2">
          <span className="text-sm font-mono text-gray-700">{folderDisplay}</span>
          {t.target_path !== undefined && (
            <span className="text-xs text-gray-400">{pattern}</span>
          )}
        </div>
      </Section>
      <ActionList actions={actions} onIgnore={onIgnore} ignoreBusy={ignoreBusy} />
      {scopedRule ? (
        <Section title="Scope">
          <div className="space-y-1 text-sm text-gray-700">
            <div><span className="text-gray-500">Linter:</span> <span className="font-mono">{scopedLinter}</span></div>
            <div><span className="text-gray-500">Rule:</span> <span className="font-mono">{scopedRule}</span></div>
          </div>
        </Section>
      ) : (
        <Section title="Linters">
          <div className="space-y-1">
            {linterStats.map(({ linter, count }) => (
              <div key={linter} className="flex items-center justify-between text-sm text-gray-700">
                <span className="font-mono">{linter}</span>
                <span className="text-xs text-gray-400">{count} violations</span>
              </div>
            ))}
            {linterStats.length === 0 && (
              <div className="text-sm text-gray-400">No violations found under this folder.</div>
            )}
          </div>
        </Section>
      )}
    </>
  );
}

function ActionList({ actions, onIgnore, ignoreBusy }: { actions: LintAction[]; onIgnore?: (req: IgnoreRequest) => Promise<void> | void; ignoreBusy?: boolean }) {
  if (!onIgnore || actions.length === 0) return null;
  return (
    <Section title="Actions">
      <div className="flex flex-wrap gap-1.5">
        {actions.map(a => (
          <IgnoreButton
            key={a.key}
            label={a.label}
            title={a.title}
            req={a.req}
            onIgnore={onIgnore}
            disabled={!!ignoreBusy || (!!a.disabledWithoutWorkDir && !a.req.work_dir)}
            variant={a.variant}
          />
        ))}
      </div>
    </Section>
  );
}

interface LintDetailProps {
  t: Test;
  onIgnore?: (req: IgnoreRequest) => Promise<void> | void;
  ignoreBusy?: boolean;
}

function IgnoreButton({
  label,
  title,
  req,
  onIgnore,
  disabled,
  variant,
}: {
  label: string;
  title: string;
  req: IgnoreRequest;
  onIgnore?: (req: IgnoreRequest) => Promise<void> | void;
  disabled?: boolean;
  variant?: 'primary' | 'subtle';
}) {
  if (!onIgnore) return null;
  const base = 'text-xs px-2 py-0.5 rounded border transition-colors disabled:opacity-40 disabled:cursor-not-allowed inline-flex items-center gap-1';
  const cls = variant === 'subtle'
    ? `${base} border-gray-200 text-gray-500 hover:bg-gray-50`
    : `${base} border-yellow-300 bg-yellow-50 text-yellow-800 hover:bg-yellow-100`;
  return (
    <button
      className={cls}
      title={title}
      disabled={disabled}
      onClick={(e) => { e.stopPropagation(); void onIgnore(req); }}
    >
      <iconify-icon icon="codicon:eye-closed" className="text-xs" />
      {label}
    </button>
  );
}

function editableFramework(framework?: string): boolean {
  return framework === 'go test' || framework === 'ginkgo' || framework === 'vitest';
}

function TestEditButton({
  icon,
  label,
  disabled,
  onClick,
  variant,
}: {
  icon: string;
  label: string;
  disabled?: boolean;
  onClick: () => void;
  variant?: 'danger';
}) {
  const base = 'text-xs px-2 py-0.5 rounded border transition-colors disabled:opacity-40 disabled:cursor-not-allowed inline-flex items-center gap-1';
  const cls = variant === 'danger'
    ? `${base} border-red-300 bg-red-50 text-red-700 hover:bg-red-100`
    : `${base} border-yellow-300 bg-yellow-50 text-yellow-800 hover:bg-yellow-100`;
  return (
    <button
      className={cls}
      title={label}
      disabled={disabled}
      onClick={(e) => { e.stopPropagation(); onClick(); }}
    >
      <iconify-icon icon={icon} className="text-xs" />
      {label}
    </button>
  );
}

function FileViolationsDetail({ t, onIgnore, ignoreBusy }: LintDetailProps) {
  const linter = t.linterName || '';
  const file = t.file || '';
  const targetPath = t.target_path || file;
  const workDir = t.work_dir || '';
  const vs = t.violations || [];
  const actions = lintFileActions(t);
  return (
    <>
      <Section title="File">
        <div className="flex items-center gap-2">
          <span className="text-sm font-mono text-gray-700">{file}</span>
          {vs.length > 0 && <span className="text-xs text-gray-400">({vs.length} violations)</span>}
        </div>
      </Section>
      <ActionList actions={actions} onIgnore={onIgnore} ignoreBusy={ignoreBusy} />
      {vs.length > 0 && (
        <Section title="Violations">
          <div className="space-y-2">
            {vs.map((v, i) => (
              <ViolationRow
                key={i}
                v={v}
                linter={linter}
                file={targetPath}
                workDir={workDir}
                onIgnore={onIgnore}
                ignoreBusy={ignoreBusy}
              />
            ))}
          </div>
        </Section>
      )}
    </>
  );
}

function RuleViolationsDetail({ t, onIgnore, ignoreBusy }: LintDetailProps) {
  const linter = t.linterName || '';
  const file = t.file || '';
  const targetPath = t.target_path || file;
  const rule = t.ruleName || '';
  const workDir = t.work_dir || '';
  const vs = t.violations || [];
  const noFileViolations = t.noFileViolations || [];
  const total = lintNodeCount(t);
  return (
    <>
      <Section title="Rule">
        <span className="text-sm font-mono text-gray-700">{rule}</span>
        <span className="text-xs text-gray-400 ml-2">({total} violations)</span>
      </Section>
      {file && (
        <Section title="File">
          <span className="text-sm font-mono text-gray-700">{file}</span>
        </Section>
      )}
      <Section title="Actions">
        <div className="flex flex-wrap gap-1.5">
          {file && (
            <IgnoreButton
              label={`Ignore ${rule} in this file`}
              title="Add {source, rule, file} to .gavel.yaml"
              req={{ source: linter, rule, file: targetPath, work_dir: workDir }}
              onIgnore={onIgnore}
              disabled={ignoreBusy}
            />
          )}
          <IgnoreButton
            label={`Ignore rule ${rule} everywhere`}
            title="Add {source, rule} to .gavel.yaml"
            req={{ source: linter, rule, work_dir: workDir }}
            onIgnore={onIgnore}
            disabled={ignoreBusy}
          />
          <IgnoreButton
            label={`Disable ${linter} entirely`}
            title="Add {source} to .gavel.yaml"
            req={{ source: linter, work_dir: workDir }}
            onIgnore={onIgnore}
            disabled={ignoreBusy}
            variant="subtle"
          />
        </div>
      </Section>
      {vs.length > 0 && (
        <Section title="Violations">
          <div className="space-y-2">
            {vs.map((v, i) => (
              <ViolationRow
                key={i}
                v={v}
                linter={linter}
                file={v.raw_file || targetPath}
                workDir={workDir}
                onIgnore={onIgnore}
                ignoreBusy={ignoreBusy}
                showFile={!file}
              />
            ))}
          </div>
        </Section>
      )}
      {noFileViolations.length > 0 && (
        <Section title="Violations Without File">
          <div className="space-y-2">
            {noFileViolations.map((v, i) => (
              <ViolationRow
                key={i}
                v={v}
                linter={linter}
                file=""
                workDir={workDir}
              />
            ))}
          </div>
        </Section>
      )}
    </>
  );
}

function ViolationRow({
  v, linter, file, workDir, onIgnore, ignoreBusy, showFile,
}: {
  v: Violation;
  linter: string;
  file: string;
  workDir: string;
  onIgnore?: (req: IgnoreRequest) => Promise<void> | void;
  ignoreBusy?: boolean;
  showFile?: boolean;
}) {
  const sev = v.severity || 'error';
  const sevColor = sev === 'error' ? 'text-red-600'
    : sev === 'warning' ? 'text-yellow-600' : 'text-blue-500';
  const sevIcon = sev === 'error' ? 'codicon:error'
    : sev === 'warning' ? 'codicon:warning' : 'codicon:info';
  const rule = v.rule?.method || '';
  return (
    <div className="border border-gray-200 rounded p-2 bg-white">
      <div className="flex items-start gap-2">
        <iconify-icon icon={sevIcon} className={`${sevColor} text-base shrink-0 mt-0.5`} />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2 flex-wrap text-xs text-gray-500 font-mono">
            {showFile && v.file && <span>{v.file}</span>}
            <span>
              :{v.line || 0}{v.column ? `:${v.column}` : ''}
            </span>
            {rule && <span className="text-gray-700">[{rule}]</span>}
          </div>
          {v.message && (
            <div className="text-sm text-gray-800 whitespace-pre-wrap mt-0.5">{v.message}</div>
          )}
          {v.code && (
            <pre className="text-xs font-mono bg-gray-50 rounded p-2 mt-1 overflow-x-auto">{v.code}</pre>
          )}
          {rule && file && (
            <div className="mt-1.5">
              <IgnoreButton
                label="Ignore this violation"
                title="Add {source, rule, file} to .gavel.yaml"
                req={{ source: linter, rule, file, work_dir: workDir }}
                onIgnore={onIgnore}
                disabled={ignoreBusy}
                variant="subtle"
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function ChildRow({ test: t }: { test: Test }) {
  return (
    <div className="flex items-center gap-1.5 py-0.5 text-sm">
      <iconify-icon icon={statusIcon(t)} className={`${statusColor(t)} text-base`} />
      <span className={`truncate ${t.failed ? 'text-red-700' : 'text-gray-700'}`}>{t.name}</span>
      {t.duration ? <span className="text-xs text-gray-400 ml-auto">{formatDuration(t.duration)}</span> : null}
    </div>
  );
}
