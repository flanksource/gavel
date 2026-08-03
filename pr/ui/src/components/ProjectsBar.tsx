import { useEffect, useMemo, useState, type ComponentType } from 'react';
import { Button } from '@flanksource/clicky-ui/components';
import {
  UiAdd,
  UiBeaker,
  UiChevronDown,
  UiChevronRight,
  UiCog,
  UiDebugStepOver,
  UiError,
  UiFolder,
  UiPass,
  UiWarningTriangle,
  type IconProps,
} from '@flanksource/clicky-ui/icons';
import type { Project, ProcStatus } from '../types';
import { GitChangesBadge } from './GitChangesBadge';
import { ProcControl } from './ProcControl';
import { RelativeTime } from './RelativeTime';
import { TodoBadge } from './TodoBadge';
import type { ProjectRuns, TestRunView } from './tests/types';

interface Props {
  projects: Project[];
  runs: ProjectRuns[];
  procStatus: Record<string, ProcStatus>;
  selected: string;
  selectedRunId: string;
  runError?: string;
  runsLoading?: boolean;
  historyEnabled: boolean;
  onHistoryChange: (enabled: boolean) => void;
  onSelect: (project: Project) => void;
  onSelectRun: (project: string, runId: string) => void;
  onChanged: () => void;
  onAdd: () => void;
  onSettings: (project: Project) => void;
}

const KIND_LABEL: Record<TestRunView['kind'], string> = {
  test: 'Test',
  lint: 'Lint',
  'test+lint': 'Test + Lint',
};

const KIND_ARIA_LABEL: Record<TestRunView['kind'], string> = {
  test: 'test',
  lint: 'lint',
  'test+lint': 'test and lint',
};

export function ProjectsBar(props: Props) {
  const [expanded, setExpanded] = useState<Set<string>>(() => new Set(props.historyEnabled && props.selected ? [props.selected] : []));
  const runsByProject = useMemo(() => new Map(props.runs.map(project => [project.name, project.runs])), [props.runs]);

  useEffect(() => {
    if (!props.historyEnabled || !props.selected) return;
    setExpanded(current => current.has(props.selected) ? current : new Set(current).add(props.selected));
  }, [props.historyEnabled, props.selected]);

  return (
    <div className="h-full overflow-auto border-b border-border">
      <div className="sticky top-0 z-20 flex items-center justify-between gap-2 bg-muted px-3 py-1 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
        <span>Projects</span>
        <Button
          variant="ghost"
          type="button"
          onClick={() => props.onHistoryChange(!props.historyEnabled)}
          aria-label={`${props.historyEnabled ? 'Hide' : 'Show'} test and lint history`}
          className="h-auto px-1.5 py-0.5 text-[10px] font-medium normal-case tracking-normal"
        >
          {props.historyEnabled ? 'Hide history' : 'Show history'}
        </Button>
      </div>
      {props.historyEnabled && props.runError && <div role="alert" className="border-b border-red-200 px-3 py-2 text-xs text-red-600 dark:border-red-900 dark:text-red-400">{props.runError}</div>}
      {props.historyEnabled && props.runsLoading && !props.runError && <div role="status" className="border-b border-border px-3 py-2 text-xs text-muted-foreground">Loading test and lint runs…</div>}
      {props.projects.map(project => (
        <ProjectBranch
          key={project.name}
          {...props}
          project={project}
          projectRuns={runsByProject.get(project.name) ?? []}
          open={expanded.has(project.name)}
          onToggle={() => setExpanded(current => {
            const next = new Set(current);
            if (next.has(project.name)) next.delete(project.name);
            else next.add(project.name);
            return next;
          })}
          onOpen={() => {
            setExpanded(current => current.has(project.name) ? current : new Set(current).add(project.name));
            props.onSelect(project);
          }}
        />
      ))}
      <Button
        variant="ghost"
        type="button"
        onClick={props.onAdd}
        title="Add a local workspace directory"
        className="flex h-auto w-full items-center justify-start gap-2 py-1.5 pl-6 pr-3 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
      >
        <UiAdd className="shrink-0" />
        Add directory
      </Button>
    </div>
  );
}

function ProjectBranch({ project, projectRuns, open, selected, selectedRunId, procStatus, runsLoading, historyEnabled, onToggle, onOpen, onSelectRun, onChanged, onSettings }: Props & {
  project: Project;
  projectRuns: TestRunView[];
  open: boolean;
  onToggle: () => void;
  onOpen: () => void;
}) {
  const Chevron = open ? UiChevronDown : UiChevronRight;
  const key = project.repos[0] || project.name;
  const projectActive = selected === project.name && !selectedRunId;

  return (
    <div>
      <div className={`flex items-center gap-1 pr-3 transition-colors ${projectActive ? 'bg-primary/10' : 'hover:bg-muted'}`}>
        {historyEnabled ? (
          <Button variant="ghost" size="icon" type="button" onClick={onToggle} aria-label={`${open ? 'Collapse' : 'Expand'} ${project.name} runs`} className="h-7 w-7 shrink-0 rounded-none text-muted-foreground">
            <Chevron />
          </Button>
        ) : <span className="h-7 w-7 shrink-0" />}
        <Button variant="ghost" type="button" onClick={onOpen} aria-label={`Open ${project.name} project`} className="h-auto min-w-0 flex-1 justify-start gap-2 rounded-none py-1.5 pr-2 text-left">
          <UiFolder className="shrink-0 text-muted-foreground" />
          <span className="truncate text-sm font-medium text-foreground" title={project.dir}>{project.name}</span>
          {historyEnabled && <span className="text-[10px] tabular-nums text-muted-foreground" title={`${projectRuns.length} test and lint runs`}>{projectRuns.length}</span>}
        </Button>
        <TodoBadge counts={project.todoCounts} />
        <GitChangesBadge count={procStatus[project.name]?.gitChanges} />
        <Button variant="ghost" size="icon" type="button" onClick={() => onSettings(project)} title={`Edit ${project.name} .gavel.yaml`} aria-label={`Edit ${project.name} settings`} className="h-6 w-6 shrink-0 text-muted-foreground hover:text-foreground">
          <UiCog />
        </Button>
        <ProcControl repo={key} project={project} status={procStatus[project.name]} onChanged={onChanged} />
      </div>
      {historyEnabled && open && (
        <div role="group" aria-label={`${project.name} test and lint runs`}>
          {projectRuns.length > 0 ? projectRuns.map(run => (
            <TestRunRow key={run.runId} project={project.name} run={run} selected={selected === project.name && selectedRunId === run.runId} onSelect={onSelectRun} />
          )) : (
            <div className="border-b border-border py-2 pl-9 pr-3 text-xs text-muted-foreground">
              {runsLoading ? 'Loading runs…' : 'No test or lint runs'}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function TestRunRow({ project, run, selected, onSelect }: { project: string; run: TestRunView; selected: boolean; onSelect: (project: string, runId: string) => void }) {
  const KindIcon = run.kind === 'lint' ? UiWarningTriangle : UiBeaker;
  return (
    <Button
      variant="ghost"
      type="button"
      aria-label={`Open ${KIND_ARIA_LABEL[run.kind]} run ${run.runId} for ${project}`}
      onClick={() => onSelect(project, run.runId)}
      className={`h-auto w-full flex-col items-stretch gap-1 rounded-none border-b border-border py-2 pl-9 pr-3 text-left transition-colors ${selected ? 'bg-primary/10' : 'hover:bg-muted'}`}
    >
      <div className="flex w-full items-center justify-between gap-2">
        <span className="flex items-center gap-1.5 text-xs font-medium"><KindIcon />{KIND_LABEL[run.kind]}</span>
        {run.started && <RelativeTime iso={run.started} title={run.started} />}
      </div>
      <RunCounts run={run} />
    </Button>
  );
}

function RunCounts({ run }: { run: TestRunView }) {
  return (
    <div className="flex flex-wrap items-center gap-x-2 gap-y-0.5 text-[11px] tabular-nums">
      {run.total > 0 && <>
        <Pill icon={UiPass} tone="text-green-600" value={run.passed} />
        <Pill icon={UiError} tone={run.failed > 0 ? 'text-red-600' : 'text-muted-foreground'} value={run.failed} />
        {run.skipped > 0 && <Pill icon={UiDebugStepOver} tone="text-muted-foreground" value={run.skipped} />}
        {run.warned > 0 && <Pill icon={UiWarningTriangle} tone="text-amber-600" value={run.warned} />}
      </>}
      {run.lintLinters > 0 && <Pill icon={UiWarningTriangle} tone={run.lintViolations > 0 ? 'text-amber-600' : 'text-green-600'} value={run.lintViolations} label="lint" />}
    </div>
  );
}

function Pill({ icon: Icon, tone, value, label }: { icon: ComponentType<IconProps>; tone: string; value: number; label?: string }) {
  return <span className={`flex items-center gap-0.5 ${tone}`}><Icon />{value}{label && <span className="text-muted-foreground">{label}</span>}</span>;
}
