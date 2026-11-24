import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Button, Combobox, Field, Modal } from '@flanksource/clicky-ui/components';
import { ActionError } from '@flanksource/clicky-ui/data';
import { inputClass } from '../todos/format';
import { usePromptRun } from '../promptRun';
import { promptCatalogOptions, promptCatalogQuery, scopeQueryFor } from './promptCatalog';

const pathVariables = new Set(['path', 'file', 'files', 'paths', 'target', 'folder']);
const dirVariables = new Set(['dir', 'cwd', 'workdir', 'workDir', 'worktree', 'repo', 'root']);
const projectVariables = new Set(['project', 'projectName']);

// promptVariableDefaults prefills a prompt's declared variables from where it was
// launched: the selected file or folder, the project and its checkout.
export function promptVariableDefaults({ variables, project, workDir, path }: {
  variables: readonly string[];
  project: string;
  workDir: string;
  path: string;
}): Record<string, string> {
  return Object.fromEntries(variables.map(name => [name,
    pathVariables.has(name) ? path
      : dirVariables.has(name) ? workDir
        : projectVariables.has(name) ? project
          : '',
  ]));
}

// ProjectPromptRunDialog runs a catalog prompt in a project's checkout, for the
// whole project or one file or folder of it.
export function ProjectPromptRunDialog({ project, workDir, path = '', onClose }: {
  project: string;
  workDir: string;
  path?: string;
  onClose: () => void;
}) {
  const catalog = useQuery(promptCatalogQuery(scopeQueryFor(project)));
  const options = useMemo(() => promptCatalogOptions(catalog.data ?? []), [catalog.data]);
  const [prompt, setPrompt] = useState('');
  const [values, setValues] = useState<Record<string, string>>({});
  const [finished, setFinished] = useState('');
  const entry = catalog.data?.find(item => item.id === prompt);
  const declared = entry?.variables ?? [];
  const variables = { ...promptVariableDefaults({ variables: declared, project, workDir, path }), ...values };
  const run = usePromptRun({ prompt, label: `${prompt} run`, onSuccess: () => setFinished(prompt) });
  const subject = path || project;

  return <Modal open title={`Run prompt on ${subject}`} onClose={onClose}
    footer={<div className="flex justify-end gap-2">
      <Button variant="outline" onClick={onClose}>{run.active ? 'Close (keeps running)' : 'Close'}</Button>
      {run.state.status === 'running' && <Button variant="outline" onClick={() => void run.stop()}>Stop</Button>}
      <Button disabled={!entry || run.active} loading={run.active} onClick={() => {
        setFinished('');
        void run.start(variables, { spec: { setup: { cwd: workDir } } });
      }}>Run prompt</Button>
    </div>}>
    <div className="space-y-4 text-sm">
      <Field label="Prompt">
        <Combobox ariaLabel="Prompt" required options={options} loading={catalog.isPending} value={prompt} disabled={run.active}
          onChange={next => { setPrompt(next); setValues({}); }} />
      </Field>
      {catalog.error && <ActionError error={catalog.error} title="Could not load prompts" />}
      {entry?.description && <p className="text-muted-foreground">{entry.description}</p>}
      {declared.map(name => (
        <Field key={name} label={name}>
          <input aria-label={name} className={inputClass} value={variables[name] ?? ''} disabled={run.active}
            onChange={event => { const value = event.currentTarget.value; setValues(current => ({ ...current, [name]: value })); }} />
        </Field>
      ))}
      {entry && <p className="text-xs text-muted-foreground">Runs in <span className="font-mono">{workDir}</span></p>}
      {run.state.status === 'starting' && <p role="status">Starting {prompt}…</p>}
      {run.state.status === 'running' && <p role="status">Running {prompt}{run.state.sessionId ? ` in session ${run.state.sessionId}` : ''}…</p>}
      {run.state.status === 'failed' && <ActionError error={run.state.error} />}
      {finished && !run.active && run.state.status === 'idle' && <p role="status" className="text-green-600">{finished} finished.</p>}
    </div>
  </Modal>;
}
