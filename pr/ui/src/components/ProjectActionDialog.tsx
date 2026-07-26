import { useEffect, useState } from 'react';
import {
  Button,
  JsonSchemaForm,
  Modal,
  type JsonSchemaObject,
} from '@flanksource/clicky-ui/components';

export type ProjectAction = 'commit' | 'lint' | 'test';

interface ProjectActionSchemaResponse {
  schemaVersion: number;
  action: ProjectAction;
  schema: JsonSchemaObject;
  defaults: Record<string, unknown>;
}

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
  const [definition, setDefinition] = useState<ProjectActionSchemaResponse | null>(null);
  const [value, setValue] = useState<Record<string, unknown>>({});
  const [loading, setLoading] = useState(false);
  const [running, setRunning] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!action) return;
    const controller = new AbortController();
    setDefinition(null);
    setLoading(true);
    setError('');
    fetch(`/api/projects/${encodeURIComponent(projectName)}/actions/schema?action=${action}`, { signal: controller.signal })
      .then(async response => {
        const payload = await response.json();
        if (!response.ok) throw new Error(payload.error || `Failed to load ${action} options`);
        return payload as ProjectActionSchemaResponse;
      })
      .then(next => {
        const remembered = readRememberedOptions(projectName, action);
        const initial = remembered?.schemaVersion === next.schemaVersion
          ? { ...next.defaults, ...remembered.value }
          : { ...next.defaults };
        if (action !== 'test' && selectedFiles.length > 0) initial.files = [...selectedFiles].sort();
        setDefinition(next);
        setValue(initial);
      })
      .catch(cause => {
        if (cause instanceof DOMException && cause.name === 'AbortError') return;
        setError(cause instanceof Error ? cause.message : `Failed to load ${action} options`);
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });
    return () => controller.abort();
  }, [action, projectName, selectedFiles]);

  const run = async () => {
    if (!action || !definition) return;
    setRunning(true);
    setError('');
    try {
      await onRun(action, value);
      rememberOptions(projectName, action, definition.schemaVersion, value);
      onClose();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : `Failed to start ${action}`);
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
