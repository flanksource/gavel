import { Button } from "@flanksource/clicky-ui";
import { UiCheck, UiWarningTriangle } from "@flanksource/clicky-ui/icons";
import { ProjectSelect } from "./ProjectSelect";

const muted = "text-muted-foreground";

/** Chrome withholds the extension's access to the dashboard until the user allows it, so nothing can load. */
export function BlockedNotice({ host, onAllow, onUseAnother, className = "" }: { host: string; onAllow: () => void; onUseAnother?: () => void; className?: string }) {
  return (
    <div role="alert" className={`rounded-lg border border-amber-500/40 bg-amber-500/10 p-3 ${className}`}>
      <div className="flex items-start gap-2">
        <UiWarningTriangle className="mt-0.5 size-4 shrink-0 text-amber-700 dark:text-amber-400" aria-hidden />
        <div className="min-w-0">
          <p className="font-semibold">Chrome is blocking Gavel from reaching {host}</p>
          <p className={`mt-1 ${muted}`}>Projects and sessions can't load until you allow it.</p>
          <div className="mt-2.5 flex items-center gap-2">
            <Button size="sm" onClick={onAllow}>Allow access</Button>
            {onUseAnother && <Button size="sm" variant="ghost" onClick={onUseAnother}>Use another dashboard</Button>}
          </div>
        </div>
      </div>
    </div>
  );
}

function StepMark({ number, done }: { number: number; done: boolean }) {
  const shape = "inline-flex size-5 shrink-0 items-center justify-center rounded-full text-xs font-semibold";
  return done
    ? <span role="img" aria-label="Done" className={`${shape} bg-green-600 text-white`}><UiCheck className="size-3" aria-hidden /></span>
    : <span aria-hidden="true" className={`${shape} bg-primary text-primary-foreground`}>{number}</span>;
}

type FirstRunProps = {
  dashboard: { host: string; projects: string[] | undefined; error: string };
  onChangeDashboard: () => void;
  onProject: (project: string) => void;
};

/** What a site needs before Gavel can be switched on: a reachable dashboard, then a project. */
export function FirstRunSteps({ dashboard, onChangeDashboard, onProject }: FirstRunProps) {
  const connected = dashboard.projects !== undefined && !dashboard.error;
  return (
    <section className="p-4">
      <h2 className="font-semibold">Set up Gavel</h2>
      <p className={`mt-1 ${muted}`}>Choose a project, then switch Gavel on above.</p>
      <ol className="mt-3 divide-y divide-border rounded-lg border border-border">
        <li className="flex items-center gap-2 p-3">
          <StepMark number={1} done={connected} />
          <span className="font-semibold">Dashboard</span>
          <span className={`min-w-0 flex-1 truncate ${dashboard.error ? "text-destructive" : muted}`} title={dashboard.error || undefined}>
            {dashboard.error || `${dashboard.host} · ${connected ? "connected" : "connecting…"}`}
          </span>
          <Button size="sm" variant="ghost" onClick={onChangeDashboard}>Change</Button>
        </li>
        <li className="p-3">
          <div className="flex items-center gap-2">
            <StepMark number={2} done={false} />
            <span className="font-semibold">Project</span>
          </div>
          <ProjectSelect className="mt-2" projects={dashboard.projects} value={undefined} onChange={onProject} />
          <p className={`mt-1.5 ${muted}`}>Where this site's comments are filed.</p>
        </li>
      </ol>
    </section>
  );
}
