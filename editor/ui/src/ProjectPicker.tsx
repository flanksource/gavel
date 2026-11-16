import { useEffect, useState } from "react";
import { Button, Combobox, Modal } from "@flanksource/clicky-ui/components";
import { ApiError, request, type EditProject, type OpenedFiles } from "./api";

/** The project open in the editor: the registered project whose directory is the only open folder. */
export function currentProject(projects: EditProject[], roots: string[]): EditProject | undefined {
  return roots.length === 1 ? projects.find((project) => project.dir === roots[0]) : undefined;
}

/**
 * Switches every edited file to a registered project's directory
 * (POST /api/edit/project). Unsaved drafts are confirmed before they are closed.
 */
export function ProjectPicker({
  roots,
  unsaved,
  onSwitched,
  onAccessDenied,
  onError,
}: {
  roots: string[];
  /** The number of files whose drafts are not saved yet. */
  unsaved: () => number;
  onSwitched: (opened: OpenedFiles) => void;
  onAccessDenied: () => void;
  onError: (message: string) => void;
}) {
  const [projects, setProjects] = useState<EditProject[]>([]);
  const [confirm, setConfirm] = useState<{ project: EditProject; unsaved: number }>();
  const [switching, setSwitching] = useState(false);
  useEffect(() => {
    void request<{ projects: EditProject[] }>("/api/edit/projects")
      .then((result) => setProjects(result.projects))
      .catch((cause) => onError(`loading projects: ${String(cause)}`));
  }, []);

  const current = currentProject(projects, roots);
  const switchTo = async (project: EditProject) => {
    setConfirm(undefined);
    setSwitching(true);
    try {
      onSwitched(await request<OpenedFiles>("/api/edit/project", { dir: project.dir }));
    } catch (cause) {
      if (cause instanceof ApiError && cause.status === 403) onAccessDenied();
      onError(`switching to ${project.name}: ${cause instanceof Error ? cause.message : String(cause)}`);
    } finally {
      setSwitching(false);
    }
  };
  const choose = (dir: string) => {
    const project = projects.find((candidate) => candidate.dir === dir);
    if (!project || project === current) return;
    const count = unsaved();
    if (count > 0) setConfirm({ project, unsaved: count });
    else void switchTo(project);
  };

  return (
    <>
      <Combobox
        ariaLabel="Project"
        searchPlacement="menu"
        title={current ? `Project ${current.name} · ${current.dir}` : "Switch project"}
        triggerContent={<span className="truncate font-medium">{switching ? "Switching…" : (current?.name ?? "Project…")}</span>}
        inputPlaceholder="Search projects…"
        value={current?.dir ?? ""}
        disabled={switching}
        size="sm"
        menuWidth={360}
        onChange={choose}
        options={projects.map((project) => ({ value: project.dir, label: project.name, description: project.dir }))}
      />
      {confirm && (
        <Modal
          open
          onClose={() => setConfirm(undefined)}
          title={`Switch to ${confirm.project.name}?`}
          size="sm"
          footer={
            <div className="flex justify-end gap-2">
              <Button variant="ghost" onClick={() => setConfirm(undefined)}>
                Keep editing
              </Button>
              <Button variant="destructive" onClick={() => void switchTo(confirm.project)}>
                Discard and switch
              </Button>
            </div>
          }
        >
          <p className="text-sm">
            {confirm.unsaved === 1 ? "1 file has" : `${confirm.unsaved} files have`} changes that are not saved yet. Switching closes every open file, and those changes are lost.
          </p>
        </Modal>
      )}
    </>
  );
}
