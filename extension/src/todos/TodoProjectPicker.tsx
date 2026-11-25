import { Combobox, Field } from "@flanksource/clicky-ui";
import type { Project } from "./model";

export function TodoProjectPicker({ projects, dir, disabled, onChange }: { projects: Project[]; dir: string; disabled: boolean; onChange: (dir: string) => void }) {
  return <Field label="Project" htmlFor="todo-project" required helper={dir || "Choose a saved project, or enter a workspace directory."}>
    <Combobox id="todo-project" ariaLabel="Project" ariaRequired required value={dir} disabled={disabled} onChange={onChange}
      options={projects.filter(project => project.dir).map(project => ({ value: project.dir, label: project.name, description: project.dir }))}
      onNew={value => ({ value, label: value, description: "Use this workspace directory" })} />
  </Field>;
}
