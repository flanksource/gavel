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
import { ErrorBoundary } from './ErrorBoundary';
import { RuntimeOptions } from './project-action/RuntimeOptions';
import { runtimeKeys, withoutRemovedAIOptions } from './project-action/runtime';

export type ProjectAction = ProjectActionName;

interface Props {
  projectName: string;
  action: ProjectAction | null;
  stagedFiles: string[];
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

export function ProjectActionDialog(props: Props) {
  const { action, projectName, onClose } = props;
  return (
    <ErrorBoundary
      key={`${projectName}/${action ?? ''}`}
      fallback={error => (
        <Modal
          open={action !== null}
          onClose={onClose}
          title={`Advanced ${action} options`}
          footer={<Button type="button" variant="ghost" onClick={onClose}>Close</Button>}
        >
          <div role="alert">
            <p className="text-sm font-medium">Could not render advanced {action} options.</p>
            <p className="text-sm text-muted-foreground">{error.message}</p>
          </div>
        </Modal>
      )}
    >
      <ProjectActionFormDialog {...props} />
    </ErrorBoundary>
  );
}

function ProjectActionFormDialog({ projectName, action, stagedFiles, onClose, onRun }: Props) {
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
    if (takesStagedFiles(action)) delete initial.files;
    setValue(initial);
    setRunError('');
  }, [action, definition, projectName]);

  useEffect(() => {
    if (!action || !takesStagedFiles(action) || !definition) return;
    setValue(current => {
      const next = { ...current };
      delete next.files;
      if (stagedFiles.length > 0) next.files = [...stagedFiles].sort();
      return next;
    });
  }, [action, definition, projectName, stagedFiles]);

  const run = async () => {
    if (!action || !definition) return;
    setRunning(true);
    setRunError('');
    try {
      const sparse = Object.fromEntries(Object.entries(value).filter(([, entry]) => entry !== undefined));
      const options = action === 'commit' ? withoutRemovedAIOptions(sparse) : sparse;
      await onRun(action, options);
      const remembered = { ...options };
      if (takesStagedFiles(action)) delete remembered.files;
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
      title={action ? `Advanced ${actionTitle(action)} options` : 'Advanced options'}
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
            pre={action === 'commit' || action === 'open-pr' ? commitExtensions : formExtensions}
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

// The staged files ride in the options of the actions whose schema lists them;
// open-pr commits the staged files beside its push options, test takes paths.
function takesStagedFiles(action: ProjectAction) {
  return action === 'commit' || action === 'lint';
}

function actionTitle(action: ProjectAction) {
  return action === 'open-pr' ? 'Open PR' : action;
}

function actionButtonLabel(action: ProjectAction | null) {
  if (action === 'commit') return 'Commit staged';
  if (action === 'open-pr') return 'Open PR';
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
