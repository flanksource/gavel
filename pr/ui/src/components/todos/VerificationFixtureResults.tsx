import { useMemo, useState } from 'react';
import {
  TestRunner,
  emptyTestFilters,
  filterTests,
  type Test,
  type TestFilters,
} from '@flanksource/clicky-ui/data';
import { attemptFixtureTests, type VerificationAttempt } from './verificationAttempts';

export function VerificationFixtureResults({ entry }: { entry: VerificationAttempt }) {
  const tests = useMemo(() => attemptFixtureTests(entry), [entry]);
  const [selected, setSelected] = useState<Test | null>(null);
  const [filters, setFilters] = useState<TestFilters>(() => emptyTestFilters());
  const [expandAll, setExpandAll] = useState<boolean | null>(null);
  const visible = useMemo(
    () => filterTests(tests, filters.status, filters.framework),
    [tests, filters],
  );

  if (tests.length === 0) {
    return <p className="px-3 py-4 text-xs text-muted-foreground">This attempt recorded no fixture execution tree.</p>;
  }

  return (
    <div className="h-[30rem] min-h-80">
      <TestRunner
        tests={visible}
        selected={selected}
        filters={filters}
        expandAll={expandAll}
        done={entry.outcome !== 'running'}
        status={{ running: entry.outcome === 'running' }}
        statusText={entry.outcome === 'running' ? 'Running verification…' : undefined}
        title={null}
        onSelect={setSelected}
        onFiltersChange={setFilters}
        onExpandAllChange={setExpandAll}
      />
    </div>
  );
}
