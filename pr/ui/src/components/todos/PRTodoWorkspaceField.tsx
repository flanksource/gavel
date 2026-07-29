import { Button, Field, Select } from '@flanksource/clicky-ui/components';
import type { Project } from '../../types';
import { inputClass } from './format';

export function PRTodoWorkspaceField({
  choices,
  value,
  busy,
  onChange,
  onNewProject,
}: {
  choices: Project[];
  value: string;
  busy: boolean;
  onChange: (dir: string) => void;
  onNewProject: () => void;
}) {
  return (
    <Field label="Workspace">
      <div className="flex items-center gap-2">
        <Select
          value={value}
          onChange={event => onChange(event.currentTarget.value)}
          className={inputClass}
          aria-label="Workspace"
          disabled={choices.length === 0}
        >
          <option value="" disabled>
            {choices.length === 0 ? 'No workspace selected' : 'Select a workspace'}
          </option>
          {choices.map(workspace => <option key={workspace.dir} value={workspace.dir}>{workspace.name}</option>)}
        </Select>
        <Button
          type="button"
          variant="outline"
          className="shrink-0"
          disabled={busy}
          onClick={onNewProject}
        >
          New project
        </Button>
      </div>
      {!value && (
        <p className="mt-1 text-xs text-muted-foreground">
          Select a workspace or create a project before adding the todo.
        </p>
      )}
    </Field>
  );
}
