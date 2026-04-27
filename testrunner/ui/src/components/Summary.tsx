import type { Test, RunMeta } from '../types';
import { hasTimeoutArgs, sumNonTaskTests, timeoutArgValue } from '../utils';
import { ProgressBar } from './ProgressBar';

interface Props {
  tests: Test[];
  startTime: number | null;
  endTime: number | null;
  done: boolean;
  runMeta?: RunMeta;
}

export function Summary({ tests, startTime, endTime, done, runMeta }: Props) {
  const totals = { total: 0, passed: 0, failed: 0, skipped: 0, pending: 0, timedout: 0 };
  for (const t of tests) {
    const s = sumNonTaskTests(t);
    totals.total += s.total;
    totals.passed += s.passed;
    totals.failed += s.failed;
    totals.skipped += s.skipped;
    totals.pending += s.pending;
    totals.timedout += s.timedout;
  }
  const now = done && endTime ? endTime : Date.now();
  const elapsed = startTime ? ((now - startTime) / 1000).toFixed(1) + 's' : '';
  const passPct = totals.total > 0 ? Math.round((totals.passed / totals.total) * 100) : 0;
  const failPct = totals.total > 0 ? Math.round((totals.failed / totals.total) * 100) : 0;
  const pendingPct = totals.total > 0 ? Math.round((totals.pending / totals.total) * 100) : 0;

  const showTimeouts = runMeta && hasTimeoutArgs(runMeta.args);
  const globalTimeout = runMeta ? timeoutArgValue(runMeta.args, 'timeout') : null;
  const testTimeout = runMeta ? timeoutArgValue(runMeta.args, 'test_timeout') : null;
  const lintTimeout = runMeta ? timeoutArgValue(runMeta.args, 'lint_timeout') : null;

  return (
    <div class="flex flex-col items-end gap-2 min-w-[21rem]">
      {showTimeouts && (
        <div class="flex gap-2 text-[11px] text-gray-500 items-center justify-end flex-wrap">
          <iconify-icon icon="codicon:watch" class="text-gray-400" />
          {globalTimeout && (
            <span class="inline-flex items-center gap-1 rounded-full bg-gray-50 border border-gray-200 px-2 py-0.5" title="Global --timeout">
              <span class="opacity-70">global</span>
              <span class="font-medium text-gray-700">{globalTimeout}</span>
            </span>
          )}
          {testTimeout && (
            <span class="inline-flex items-center gap-1 rounded-full bg-gray-50 border border-gray-200 px-2 py-0.5" title="--test-timeout (per-package)">
              <span class="opacity-70">per-test</span>
              <span class="font-medium text-gray-700">{testTimeout}</span>
            </span>
          )}
          {lintTimeout && (
            <span class="inline-flex items-center gap-1 rounded-full bg-gray-50 border border-gray-200 px-2 py-0.5" title="--lint-timeout (per-linter)">
              <span class="opacity-70">per-lint</span>
              <span class="font-medium text-gray-700">{lintTimeout}</span>
            </span>
          )}
        </div>
      )}
      <div class="flex gap-3 text-sm text-gray-500 items-center justify-end flex-wrap">
        <span class="font-medium text-gray-700">{totals.total} tests</span>
        {runMeta && (
          <>
            <Sep />
            <span class={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${runMeta.kind === 'rerun' ? 'bg-blue-50 text-blue-700' : 'bg-gray-100 text-gray-600'}`}>
              <iconify-icon icon={runMeta.kind === 'rerun' ? 'codicon:history' : 'codicon:debug-restart'} />
              {runMeta.kind === 'rerun' ? `Rerun #${runMeta.sequence}` : 'Initial run'}
            </span>
          </>
        )}
        {totals.passed > 0 && <><Sep /><span class="text-green-600">{totals.passed} passed</span></>}
        {totals.failed > 0 && <><Sep /><span class="text-red-600">{totals.failed} failed</span></>}
        {totals.timedout > 0 && <><Sep /><span class="text-amber-600">{totals.timedout} timed out</span></>}
        {totals.skipped > 0 && <><Sep /><span class="text-yellow-600">{totals.skipped} skipped</span></>}
        {totals.pending > 0 && <><Sep /><span class="text-blue-500">{totals.pending} pending</span></>}
        {elapsed && (
          <>
            <Sep />
            <span class="text-gray-400">
              <iconify-icon icon="codicon:clock" class="mr-0.5" />
              {elapsed}
            </span>
          </>
        )}
        {done && <iconify-icon icon="codicon:check" class="text-green-600" />}
        {!done && <iconify-icon icon="svg-spinners:ring-resize" class="text-blue-500" />}
      </div>
      {totals.total > 0 && (
        <div class="w-80">
          <ProgressBar
            segments={[
              { count: totals.passed, color: 'bg-green-500', label: 'passed' },
              { count: totals.skipped, color: 'bg-yellow-400', label: 'skipped' },
              { count: totals.failed, color: 'bg-red-500', label: 'failed' },
              { count: totals.timedout, color: 'bg-amber-500', label: 'timed out' },
              { count: totals.pending, color: 'bg-blue-300', label: 'pending' },
            ]}
            total={totals.total}
            height="h-2.5"
          />
          <div class="grid grid-cols-3 gap-2 mt-2 text-[11px]">
            <Gauge label="Pass rate" value={`${passPct}%`} tone="text-green-700 bg-green-50 border-green-200" />
            <Gauge label="Failures" value={`${failPct}%`} tone="text-red-700 bg-red-50 border-red-200" />
            <Gauge label="Pending" value={`${pendingPct}%`} tone="text-blue-700 bg-blue-50 border-blue-200" />
          </div>
        </div>
      )}
    </div>
  );
}

function Sep() {
  return <span class="text-gray-300">|</span>;
}

function Gauge({ label, value, tone }: { label: string; value: string; tone: string }) {
  return (
    <div class={`rounded-lg border px-2 py-1.5 text-right ${tone}`}>
      <div class="text-[10px] uppercase tracking-wide opacity-70">{label}</div>
      <div class="text-sm font-semibold">{value}</div>
    </div>
  );
}
