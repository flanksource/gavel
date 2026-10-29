import { useEffect, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Button,
  JsonSchemaForm,
  Modal,
  type PreExtension,
} from '@flanksource/clicky-ui/components';
import {
  projectActionSchemaQuery,
  type ProjectActionName,
} from './oneShotQueries';
import { RuntimeOptions } from './project-action/RuntimeOptions';
import { runtimeKeys, withoutRemovedAIOptions } from './project-action/runtime';

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

const storageKey = 'gavel.project-action-options.v2';

const inheritedBoolean: PreExtension = field => (
  field.kind === 'boolean' && field.value === undefined && typeof field.schema.default === 'boolean'
    ? { ...field, value: field.schema.default }
    : field
);
const formExtensions = [inheritedBoolean];
const commitControls: PreExtension = field => {
  if (runtimeKeys.some(key => key === field.key)) return null;
  if (field.schema['x-enum-display'] !== 'segmented') return field;
  return {
    ...field,
    options: field.options?.map(option => ({ ...option, label: field.schema['x-enum-labels']?.[option.value] ?? option.label })),
    value: field.value ?? '',
    onChange: next => field.onChange(next === '' ? undefined : next),
  };
};
const commitExtensions = [...formExtensions, commitControls];

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
    let initial = remembered?.schemaVersion === definition.schemaVersion
      ? { ...remembered.value }
      : {};
    if (action === 'commit') initial = withoutRemovedAIOptions(initial);
    if (action !== 'test') delete initial.files;
    setValue(initial);
    setRunError('');
  }, [action, definition, projectName]);

  useEffect(() => {
    if (!action || action === 'test' || !definition) return;
    setValue(current => {
      const next = { ...current };
      delete next.files;
      if (selectedFiles.length > 0) next.files = [...selectedFiles].sort();
      return next;
    });
  }, [action, definition, projectName, selectedFiles]);

  const run = async () => {
    if (!action || !definition) return;
    setRunning(true);
    setRunError('');
    try {
      const options = action === 'commit' ? withoutRemovedAIOptions(value) : value;
      await onRun(action, options);
      const remembered = { ...options };
      if (action !== 'test') delete remembered.files;
      rememberOptions(projectName, action, definition.schemaVersion, remembered);
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
        <fieldset disabled={running} className="min-w-0 space-y-4">
          {action === 'commit' && <RuntimeOptions value={value} onChange={setValue} />}
          <JsonSchemaForm
            schema={definition.schema}
            value={value}
            onChange={setValue}
            applyDefaults={false}
            pre={action === 'commit' ? commitExtensions : formExtensions}
            size="sm"
            layout={{ mode: 'stacked' }}
            idPrefix={`project-${projectName}-${action}`}
            showPreferencesMenu={false}
            persistPreferences={false}
          />
        </fieldset>
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
