import { useEffect, useState } from "react";
import { Button, Field, InputField, Modal } from "@flanksource/clicky-ui/components";
import type { MissingProject, ProjectInput } from "./file-comments";

/**
 * Offered when the open file is outside every gavel project: comments live under
 * a project, so this creates one (prefilled with the file's repository) first.
 */
export function CreateProjectDialog({
  missing,
  onCreate,
  onClose,
}: {
  missing: MissingProject;
  onCreate: (project: ProjectInput) => Promise<void>;
  onClose: () => void;
}) {
  const [project, setProject] = useState<ProjectInput>(missing.suggestion);
  const [error, setError] = useState<string>();
  const [saving, setSaving] = useState(false);
  useEffect(() => setProject(missing.suggestion), [missing.suggestion.name, missing.suggestion.dir]);

  const create = async () => {
    setSaving(true);
    setError(undefined);
    try {
      await onCreate({ name: project.name.trim(), dir: project.dir.trim() });
    } catch (cause) {
      setError(String(cause));
    } finally {
      setSaving(false);
    }
  };
  const valid = project.name.trim() !== "" && project.dir.trim() !== "";
  return (
    <Modal
      open
      onClose={onClose}
      title="Create a project to comment"
      size="md"
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="ghost" onClick={onClose}>
            Not now
          </Button>
          <Button onClick={() => void create()} disabled={!valid || saving}>
            {saving ? "Creating…" : "Create project"}
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-3">
        <p className="text-sm text-muted-foreground">{missing.error}</p>
        <Field label="Name" htmlFor="create-project-name" required>
          <InputField id="create-project-name" value={project.name} onChange={(name) => setProject({ ...project, name })} />
        </Field>
        <Field label="Directory" htmlFor="create-project-dir" required helper="Every file under it is commented on within this project.">
          <InputField id="create-project-dir" value={project.dir} onChange={(dir) => setProject({ ...project, dir })} />
        </Field>
        {error && (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        )}
      </div>
    </Modal>
  );
}
