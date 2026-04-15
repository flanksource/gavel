import type { Test } from '../types';
import { sumNonTaskTests } from '../utils';
import { ProgressBar } from './ProgressBar';

interface Props {
  tests: Test[];
  startTime: number | null;
  endTime: number | null;
  done: boolean;
}

export function Summary({ tests, startTime, endTime, done }: Props) {
  const totals = { total: 0, passed: 0, failed: 0, skipped: 0, pending: 0 };
  for (const t of tests) {
    const s = sumNonTaskTests(t);
    totals.total += s.total;
    totals.passed += s.passed;
    totals.failed += s.failed;
    totals.skipped += s.skipped;
    totals.pending += s.pending;
  }
  const now = done && endTime ? endTime : Date.now();
  const elapsed = startTime ? ((now - startTime) / 1000).toFixed(1) + 's' : '';

  return (
    <div class="flex flex-col items-end gap-1">
      <div class="flex gap-3 text-sm text-gray-500 items-center">
        <span class="font-medium text-gray-700">{totals.total} tests</span>
        {totals.passed > 0 && <><Sep /><span class="text-green-600">{totals.passed} passed</span></>}
        {totals.failed > 0 && <><Sep /><span class="text-red-600">{totals.failed} failed</span></>}
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
        <div class="w-64">
          <ProgressBar
            segments={[
              { count: totals.passed, color: 'bg-green-500', label: 'passed' },
              { count: totals.skipped, color: 'bg-yellow-400', label: 'skipped' },
              { count: totals.failed, color: 'bg-red-500', label: 'failed' },
              { count: totals.pending, color: 'bg-blue-300', label: 'pending' },
            ]}
            total={totals.total}
          />
        </div>
      )}
    </div>
  );
}

function Sep() {
  return <span class="text-gray-300">|</span>;
}
