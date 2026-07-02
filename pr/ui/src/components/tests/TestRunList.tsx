import type { ComponentType } from 'react';
import { UiBeaker, UiDebugStepOver, UiError, UiFolder, UiPass, UiWarningTriangle, type IconProps } from '@flanksource/clicky-ui/icons';
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
          <UiBeaker className="mb-2 text-4xl" />
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
            <UiFolder />
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
  const KindIcon = run.kind === 'lint' ? UiWarningTriangle : UiBeaker;
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
          <KindIcon />
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
          <Pill icon={UiPass} tone="text-green-600" value={run.passed} />
          <Pill icon={UiError} tone={run.failed > 0 ? 'text-red-600' : 'text-muted-foreground'} value={run.failed} />
          {run.skipped > 0 && <Pill icon={UiDebugStepOver} tone="text-muted-foreground" value={run.skipped} />}
          {run.warned > 0 && <Pill icon={UiWarningTriangle} tone="text-amber-600" value={run.warned} />}
        </>
      )}
      {run.lintLinters > 0 && (
        <Pill
          icon={UiWarningTriangle}
          tone={run.lintViolations > 0 ? 'text-amber-600' : 'text-green-600'}
          value={run.lintViolations}
          label="lint"
        />
      )}
    </div>
  );
}

function Pill({ icon, tone, value, label }: { icon: ComponentType<IconProps>; tone: string; value: number; label?: string }) {
  const Icon = icon;
  return (
    <span className={`flex items-center gap-0.5 ${tone}`}>
      <Icon />
      {value}
      {label && <span className="text-muted-foreground">{label}</span>}
    </span>
  );
}
