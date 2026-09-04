import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Button,
  JsonSchemaForm,
  Modal,
} from '@flanksource/clicky-ui/components';
import {
  projectActionSchemaQuery,
  type ProjectActionName,
} from './oneShotQueries';

export type ProjectAction = ProjectActionName;

interface Props {
  projectName: string;
  action: ProjectAction | null;
  selectedFiles: string[];
  onClose: () => void;
  onRun: (action: ProjectAction, options: Record<string, unknown>) => Promise<void>;
}

interface RememberedActionOptions {
  schemaVersion: number;
  value: Record<string, unknown>;
}

type RememberedOptions = Record<string, Partial<Record<ProjectAction, RememberedActionOptions>>>;

const storageKey = 'gavel.project-action-options.v1';

export function ProjectActionDialog({ projectName, action, selectedFiles, onClose, onRun }: Props) {
  const [value, setValue] = useState<Record<string, unknown>>({});
  const [running, setRunning] = useState(false);
  const [runError, setRunError] = useState('');
  const definitionResult = useQuery({
    ...projectActionSchemaQuery(projectName, action),
    enabled: action !== null,
  });
  const definition = definitionResult.data ?? null;
  const loading = action !== null && definitionResult.isPending;
  const error = runError || (definitionResult.error instanceof Error ? definitionResult.error.message : '');

  useEffect(() => {
    setRunError('');
  }, [action, projectName]);

  useEffect(() => {
    if (!action || !definition) return;
    const remembered = readRememberedOptions(projectName, action);
    const initial = remembered?.schemaVersion === definition.schemaVersion
      ? { ...definition.defaults, ...remembered.value }
      : { ...definition.defaults };
    if (action !== 'test' && selectedFiles.length > 0) initial.files = [...selectedFiles].sort();
    setValue(initial);
    setRunError('');
  }, [action, definition, projectName, selectedFiles]);

  const run = async () => {
    if (!action || !definition) return;
    setRunning(true);
    setRunError('');
    try {
      await onRun(action, value);
      rememberOptions(projectName, action, definition.schemaVersion, value);
      onClose();
    } catch (cause) {
      setRunError(cause instanceof Error ? cause.message : `Failed to start ${action}`);
    } finally {
      setRunning(false);
    }
  };

  return (
    <Modal
      open={action !== null}
      onClose={onClose}
      title={action ? `Advanced ${action} options` : 'Advanced options'}
      size="lg"
      footer={
        <div className="flex items-center justify-end gap-2">
          {error && <span role="alert" className="mr-auto text-sm text-red-600 dark:text-red-400">{error}</span>}
          <Button type="button" variant="ghost" disabled={running} onClick={onClose}>Cancel</Button>
          <Button type="button" loading={running} disabled={loading || !definition} onClick={() => void run()}>
            {actionButtonLabel(action)}
          </Button>
        </div>
      }
    >
      {loading && <div className="py-8 text-center text-sm text-muted-foreground">Loading options…</div>}
      {!loading && definition && (
        <JsonSchemaForm
          schema={definition.schema}
          value={value}
          onChange={setValue}
          size="sm"
          inline
          idPrefix={`project-${projectName}-${action}`}
          showPreferencesMenu={false}
          persistPreferences={false}
        />
      )}
    </Modal>
  );
}

function actionButtonLabel(action: ProjectAction | null) {
  if (action === 'commit') return 'Commit selected';
  if (action === 'test') return 'Run tests';
  return 'Run lint';
}

function readRememberedOptions(project: string, action: ProjectAction): RememberedActionOptions | undefined {
  const raw = localStorage.getItem(storageKey);
  if (!raw) return undefined;
  const parsed = JSON.parse(raw) as RememberedOptions;
  return parsed[project]?.[action];
}

function rememberOptions(project: string, action: ProjectAction, schemaVersion: number, value: Record<string, unknown>) {
  const raw = localStorage.getItem(storageKey);
  const remembered = raw ? JSON.parse(raw) as RememberedOptions : {};
  remembered[project] = {
    ...remembered[project],
    [action]: { schemaVersion, value },
  };
  localStorage.setItem(storageKey, JSON.stringify(remembered));
}
