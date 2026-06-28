import { GavelIcon } from '../GavelIcon';
import { RelativeTime } from '../RelativeTime';
import type { ProjectRuns, TestRunView } from './types';

const KIND_LABEL: Record<TestRunView['kind'], string> = {
  test: 'Test',
  lint: 'Lint',
  'test+lint': 'Test + Lint',
};

export function TestRunList({
  projects,
  selectedPath,
  onSelect,
}: {
  projects: ProjectRuns[];
  selectedPath: string;
  onSelect: (path: string) => void;
}) {
  const withRuns = projects.filter(p => p.runs.length > 0);
  if (withRuns.length === 0) {
    return (
      <div className="flex h-full items-center justify-center p-6 text-center text-sm text-muted-foreground">
        <div>
          <GavelIcon name="codicon:beaker" className="mb-2 text-4xl" />
          <p>No test runs found.</p>
          <p className="mt-1 text-xs">
            Run <code className="rounded bg-muted px-1">gavel test</code> in a registered workspace.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full overflow-auto">
      {withRuns.map(project => (
        <div key={project.name}>
          <div className="sticky top-0 z-10 flex items-center gap-1.5 border-b border-border bg-background px-3 py-1.5 text-xs font-semibold text-muted-foreground">
            <GavelIcon name="codicon:folder" />
            {project.name}
            <span className="tabular-nums text-[10px] opacity-70">{project.runs.length}</span>
          </div>
          {project.runs.map(run => {
            const path = `${project.name}/${run.runId}`;
            return (
              <TestRunRow key={run.runId} run={run} selected={selectedPath === path} onClick={() => onSelect(path)} />
            );
          })}
        </div>
      ))}
    </div>
  );
}

function TestRunRow({ run, selected, onClick }: { run: TestRunView; selected: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex w-full flex-col gap-1 border-b border-border px-3 py-2 text-left transition-colors ${
        selected ? 'bg-primary/10' : 'hover:bg-muted'
      }`}
    >
      <div className="flex items-center justify-between gap-2">
        <span className="flex items-center gap-1.5 text-xs font-medium">
          <GavelIcon name={run.kind === 'lint' ? 'codicon:warning' : 'codicon:beaker'} />
          {KIND_LABEL[run.kind]}
        </span>
        {run.started && <RelativeTime iso={run.started} title={run.started} />}
      </div>
      <Counts run={run} />
    </button>
  );
}

function Counts({ run }: { run: TestRunView }) {
  return (
    <div className="flex flex-wrap items-center gap-x-2 gap-y-0.5 text-[11px] tabular-nums">
      {run.total > 0 && (
        <>
          <Pill icon="codicon:pass" tone="text-green-600" value={run.passed} />
          <Pill icon="codicon:error" tone={run.failed > 0 ? 'text-red-600' : 'text-muted-foreground'} value={run.failed} />
          {run.skipped > 0 && <Pill icon="codicon:debug-step-over" tone="text-muted-foreground" value={run.skipped} />}
          {run.warned > 0 && <Pill icon="codicon:warning" tone="text-amber-600" value={run.warned} />}
        </>
      )}
      {run.lintLinters > 0 && (
        <Pill
          icon="codicon:warning"
          tone={run.lintViolations > 0 ? 'text-amber-600' : 'text-green-600'}
          value={run.lintViolations}
          label="lint"
        />
      )}
    </div>
  );
}

function Pill({ icon, tone, value, label }: { icon: string; tone: string; value: number; label?: string }) {
  return (
    <span className={`flex items-center gap-0.5 ${tone}`}>
      <GavelIcon name={icon} />
      {value}
      {label && <span className="text-muted-foreground">{label}</span>}
    </span>
  );
}
