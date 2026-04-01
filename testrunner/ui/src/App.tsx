import { useState, useEffect, useRef, useMemo } from 'preact/hooks';
import type { Test, Snapshot } from './types';
import { Summary } from './components/Summary';
import { TestNode } from './components/TestNode';
import { DetailPanel } from './components/DetailPanel';
import { FilterBar, type Filters } from './components/FilterBar';
import { SplitPane } from './components/SplitPane';
import { sum, collectFrameworks, filterTests } from './utils';

function applySnapshot(
  snap: Snapshot,
  startTime: React.MutableRefObject<number | null>,
  endTime: React.MutableRefObject<number | null>,
  doneRef: React.MutableRefObject<boolean>,
  setTests: (t: Test[]) => void,
  setDone: (d: boolean) => void,
  setStatus: (s: string) => void,
) {
  if (!startTime.current) startTime.current = Date.now();
  setTests(snap.tests || []);
  if (snap.done) {
    endTime.current = Date.now();
    doneRef.current = true;
    setDone(true);
    setStatus('Test run complete');
  } else {
    setStatus('Running tests...');
  }
}

export function App() {
  const [tests, setTests] = useState<Test[]>([]);
  const [done, setDone] = useState(false);
  const [status, setStatus] = useState('Loading...');
  const [expandAll, setExpandAll] = useState<boolean | null>(null);
  const [selected, setSelected] = useState<Test | null>(null);
  const [filters, setFilters] = useState<Filters>({ status: new Set(), framework: new Set() });
  const startTime = useRef<number | null>(null);
  const endTime = useRef<number | null>(null);
  const [, tick] = useState(0);
  const doneRef = useRef(false);

  useEffect(() => {
    fetch('/api/tests')
      .then(r => r.json())
      .then((snap: Snapshot) => {
        applySnapshot(snap, startTime, endTime, doneRef, setTests, setDone, setStatus);
      })
      .catch(() => {});

    const es = new EventSource('/api/tests/stream');

    es.addEventListener('message', (e: MessageEvent) => {
      const snap: Snapshot = JSON.parse(e.data);
      applySnapshot(snap, startTime, endTime, doneRef, setTests, setDone, setStatus);
      if (snap.done) es.close();
    });

    es.addEventListener('done', () => {
      endTime.current = Date.now();
      doneRef.current = true;
      setDone(true);
      setStatus('Test run complete');
      es.close();
    });

    es.onerror = () => {
      if (!doneRef.current) setStatus('Connection lost — retrying...');
    };

    const timer = setInterval(() => {
      if (startTime.current && !doneRef.current) tick(n => n + 1);
    }, 1000);

    return () => { es.close(); clearInterval(timer); };
  }, []);

  const totals = useMemo(() => {
    const t = { total: 0, passed: 0, failed: 0, skipped: 0, pending: 0 };
    for (const test of tests) {
      const s = sum(test);
      t.total += s.total;
      t.passed += s.passed;
      t.failed += s.failed;
      t.skipped += s.skipped;
      t.pending += s.pending;
    }
    return t;
  }, [tests]);

  const frameworks = useMemo(() => collectFrameworks(tests), [tests]);
  const filtered = useMemo(() => filterTests(tests, filters.status, filters.framework), [tests, filters]);

  return (
    <div class="bg-gray-100 h-screen flex flex-col">
      <div class="border-b bg-white px-6 py-3">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <h1 class="text-xl font-bold text-gray-900">
              <iconify-icon icon="codicon:beaker" class="mr-1.5 text-blue-600" />
              Test Results
            </h1>
            {tests.length > 0 && (
              <div class="flex gap-1">
                <button
                  class="text-xs px-2 py-1 rounded border border-gray-300 text-gray-600 hover:bg-gray-200 transition-colors"
                  onClick={() => setExpandAll(true)}
                  title="Expand all"
                >
                  <iconify-icon icon="codicon:expand-all" class="mr-0.5" />
                  Expand
                </button>
                <button
                  class="text-xs px-2 py-1 rounded border border-gray-300 text-gray-600 hover:bg-gray-200 transition-colors"
                  onClick={() => setExpandAll(false)}
                  title="Collapse all"
                >
                  <iconify-icon icon="codicon:collapse-all" class="mr-0.5" />
                  Collapse
                </button>
              </div>
            )}
            <span class="text-sm text-gray-400">{status}</span>
          </div>
          <Summary tests={tests} startTime={startTime.current} endTime={endTime.current} done={done} />
        </div>
        {tests.length > 0 && (
          <div class="mt-2">
            <FilterBar filters={filters} onChange={setFilters} counts={totals} frameworks={frameworks} />
          </div>
        )}
      </div>

      <SplitPane
        defaultSplit={50}
        left={
          <>
            {filtered.map((t, i) => (
              <TestNode key={i} test={t} depth={0} expandAll={expandAll} selected={selected} onSelect={setSelected} />
            ))}
            {tests.length === 0 && !done && (
              <div class="p-8 text-center text-gray-400">
                <iconify-icon icon="svg-spinners:ring-resize" class="text-3xl text-blue-500" />
                <p class="mt-2">Waiting for test results...</p>
              </div>
            )}
            {filtered.length === 0 && tests.length > 0 && (
              <div class="p-8 text-center text-gray-400 text-sm">
                No tests match the current filters
              </div>
            )}
          </>
        }
        right={<DetailPanel test={selected} />}
      />
    </div>
  );
}
