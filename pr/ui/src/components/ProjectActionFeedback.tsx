import { AnsiHtml } from '@flanksource/clicky-ui/data';
import { UiCheck, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import { Spinner } from '../icons/Spinner';
import { Elapsed } from './Elapsed';
import type { ProjectAction } from './ProjectActionDialog';

export interface ProjectActionStatus {
  action?: ProjectAction;
  runId?: string;
  href?: string;
  running: boolean;
  startedAt?: string;
  endedAt?: string;
  exitCode?: number;
  output?: string;
  error?: string;
}

const runningLabels: Record<ProjectAction, string> = {
  commit: 'Creating commit',
  lint: 'Running lint',
  test: 'Running tests',
};

const completedLabels: Record<ProjectAction, string> = {
  commit: 'Commit completed',
  lint: 'Lint completed',
  test: 'Tests completed',
};

const failedLabels: Record<ProjectAction, string> = {
  commit: 'Commit failed',
  lint: 'Lint failed',
  test: 'Tests failed',
};

export function ProjectActionFeedback({ status }: { status: ProjectActionStatus }) {
  if (!status.action) return null;
  const failed = !status.running && Boolean(status.error);
  const label = status.running
    ? runningLabels[status.action]
    : failed ? failedLabels[status.action] : completedLabels[status.action];
  const output = status.output?.trimEnd();

  return (
    <section className="shrink-0 border-b border-border bg-muted/30" aria-live="polite">
      <div className="flex items-center gap-2 px-4 py-2 text-xs">
        {status.running ? <Spinner /> : failed ? <UiWarningTriangle className="text-red-600" /> : <UiCheck className="text-green-600" />}
        <span className={`font-medium ${failed ? 'text-red-600' : status.running ? 'text-primary' : 'text-green-600'}`}>{label}</span>
        <Elapsed startedAt={status.startedAt} endedAt={status.endedAt} running={status.running} />
        {status.href && (
          <a href={status.href} aria-label={`${actionLabel(status.action)} details`} className="ml-auto font-medium text-primary hover:underline">
            Details
          </a>
        )}
        {!status.running && status.exitCode !== undefined && <span className={`${status.href ? '' : 'ml-auto'} font-mono text-muted-foreground`}>Exit {status.exitCode}</span>}
      </div>
      {output ? (
        <AnsiHtml as="pre" text={output} className="max-h-40 overflow-auto border-t border-border bg-black px-4 py-2 text-[11px] leading-relaxed text-gray-100 whitespace-pre-wrap" />
      ) : (
        <div className="border-t border-border px-4 py-2 text-[11px] text-muted-foreground">
          {status.running ? 'Waiting for command output…' : status.error || 'Command completed without output.'}
        </div>
      )}
      {output && status.error && <div className="border-t border-red-500/20 px-4 py-1.5 text-[11px] text-red-600">{status.error}</div>}
    </section>
  );
}

function actionLabel(action: ProjectAction) {
  return action.charAt(0).toUpperCase() + action.slice(1);
}
