import { Button, Modal } from '@flanksource/clicky-ui/components';
import { UiCheck, UiStop, UiWarningTriangle } from '@flanksource/clicky-ui/icons';
import { useTestRun } from '@flanksource/gavel/testrunner/hooks';
import { Spinner } from '../icons/Spinner';
import { TestRunResults } from './tests/TestRunResults';
import type { RunSnapshot } from './tests/types';

export type ProjectRunnerAction = 'lint' | 'test';

export function ProjectActionRunDialog({
  projectName,
  projectDir,
  action,
  runId,
  onClose,
  onTodoCreated,
}: {
  projectName: string;
  projectDir: string;
  action: ProjectRunnerAction;
  runId: string;
  onClose: () => void;
  onTodoCreated?: () => void;
}) {
  const run = useTestRun({ baseUrl: '/api/project-runs', enabled: true, runId });
  const label = action === 'test' ? 'Test' : 'Lint';

  return (
    <Modal
      open
      onClose={onClose}
      title={`${label} details · ${projectName}`}
      size="xl"
      footer={(
        <div className="flex w-full justify-between gap-2">
          <div>
            {run.status.stop_supported && run.status.running && (
              <Button type="button" variant="outline" onClick={() => void run.stop()}>
                <UiStop /> Stop
              </Button>
            )}
          </div>
          <Button type="button" variant="outline" onClick={onClose}>Close</Button>
        </div>
      )}
    >
      <div className="flex h-[75vh] min-h-0 flex-col">
        <div className="flex shrink-0 items-center gap-2 border-b border-border px-3 py-2 text-xs" aria-live="polite">
          {run.status.running ? <Spinner /> : run.error ? <UiWarningTriangle className="text-red-600" /> : <UiCheck className="text-green-600" />}
          <span className={run.error ? 'text-red-600' : 'text-muted-foreground'}>{run.error || run.statusText}</span>
        </div>
        <div className="min-h-0 flex-1">
          {run.snapshot ? (
            <TestRunResults
              snapshot={run.snapshot as unknown as RunSnapshot}
              done={run.done}
              runKey={runId}
              projectName={projectName}
              projectDir={projectDir}
              onTodoCreated={onTodoCreated}
              emptyMessage={run.statusText}
            />
          ) : (
            <Centered>{run.error || run.statusText}</Centered>
          )}
        </div>
      </div>
    </Modal>
  );
}

function Centered({ children }: { children: React.ReactNode }) {
  return <div className="flex h-full items-center justify-center p-6 text-sm text-muted-foreground">{children}</div>;
}
